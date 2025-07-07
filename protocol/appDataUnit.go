package protocol

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
)

//通过框架配置协议，框架自动解析和封装，无需自己开发。
//1. 定义一个结构体(Data Unit Handler)包含协议组成元素信息
//2. 每次解析和组装都要使用这个结构体(DPH)
//3. DPH需要一个解析数据获取Data域的方法，返回Data域的字节切片
//4. 定义功能码接口(Function),解析数据域和组装数据域

func NewDUHBuilder() DUHBuilder {
	return DUHBuilder{
		dph: &DPH{
			Fields: make([]Fielder, 0, 3),
		},
	}
}

type DUHBuilder struct {
	dph *DPH
}

func (dphBuilder *DUHBuilder) AddFielder(field Fielder) *DUHBuilder {
	dphBuilder.dph.Fields = append(dphBuilder.dph.Fields, field)
	return dphBuilder
}

func (dphBuilder *DUHBuilder) Build() DataUnitHandler {
	//todo 起始码+长度码 的长度
	dphBuilder.dph.headerLen = 1
	return dphBuilder.dph
}

type DataUnitHandler interface {
	Handle(ctx context.Context, conn net.Conn)
	Parse(adu []byte) (Function, error)
	Serialize(f Function) []byte
}

func NewDPH() DPH {
	return DPH{Fields: make([]Fielder, 0, 3)}
}

var _ DataUnitHandler = (*DPH)(nil)

// dph 应用数据单元 结构体
type DPH struct {
	headerLen int
	conn      net.Conn
	Fields    []Fielder
}

func (dph *DPH) Handle(ctx context.Context, conn net.Conn) {
	dph.conn = conn
	for {
		select {
		case <-ctx.Done():
			//停止读取
			return
		default:
			for _, field := range dph.Fields {
				buf := make([]byte, field.Length())
				n, err := io.ReadFull(dph.conn, buf)
				if err != nil {
					fmt.Println("读取数据失败:", err)
					return
				}
				fmt.Println("读取数据:", string(buf[:n]))
			}
		}
	}
}

func (dph *DPH) Parse(adu []byte) (Function, error) {
	for _, v := range dph.Fields {
		v.Deal(nil)
	}
	return nil, nil
}

func (dph *DPH) Serialize(f Function) []byte {
	return nil
}

func (dph *DPH) Info() {
	for _, v := range dph.Fields {
		of := reflect.TypeOf(v)
		fmt.Println("类型:", of, " 长度:", v.Length())
	}
}

// Debug 解析数据
func (dph *DPH) Debug(r io.Reader, source []byte) {
	// 起始符 只需要判断是否相等
	// 数据域长度 要传给数据域元素作为长度
	// 加密标志 是否对指定元素的值进行加密或解密
	// 校验码 是否对指定元素进行校验计算
	offset := 0
	//遍历所有元素
	for _, field := range dph.Fields {
		//根据元素Field获取对应数据切片
		data := source[offset : offset+field.Length()]
		//更新偏移量
		offset += field.Length()
		//debug打印元素
		if field.GetScale() == 0 {
			fmt.Printf("[%s] = %0d", field.GetName(), data)
		} else {
			fmt.Printf("[%s] = %0x", field.GetName(), data)
		}

		//处理方法
		err := field.Deal(data)
		if err != nil { //log.Println("数据解析出错! [error]:", err)
			fmt.Printf("数据解析出错! [error]: %v\n", err)
		}
	}
}

// Fielder 元素接口
type Fielder interface {
	//获取元素名称
	GetName() string
	SetName(name string)
	SetDefault(value []byte)
	GetDefault() []byte
	//获取实际值
	RealValue() []byte
	// SetLen 设置元素长度
	SetLen(int)
	// Length 获取元素长度
	Length() int
	SetScale(uint8)
	//获取进制
	GetScale() uint8

	// SetRange 设置范围
	SetRange(start, end uint8)
	// GetRange 获取范围
	GetRange() (start, end uint8)
	// Deal 解析元素
	Deal([]byte) error
}

type fieldType byte

var _ Fielder = (*Field)(nil)

// Field 基础元素结构体
type Field struct {
	name string //消息帧 元素名字
	//FType        fieldType      //消息帧 字段类型
	scale    uint8            // 1十六进制，0十进制
	len      int              //消息帧 元素本身长度
	defaultV []byte           //默认值
	order    binary.ByteOrder //大小端
	next     *Fielder
	start    uint8
	end      uint8
	DealFunc func(field Fielder, data []byte) error
}

func (f *Field) GetName() string {
	return f.name
}

func (f *Field) SetName(name string) {
	f.name = name
}

func (f *Field) SetDefault(val []byte) {
	f.defaultV = val
}

func (f *Field) GetDefault() []byte {
	return f.defaultV
}

func (f *Field) RealValue() []byte {
	return f.defaultV
}
func (f *Field) SetLen(l int) {
	f.len = l
}
func (f *Field) Length() int {
	return f.len
}

func (f *Field) SetScale(u uint8) {
	f.scale = u
}

func (f *Field) GetScale() uint8 {
	return f.scale
}

func (f *Field) GetRange() (start, end uint8) {
	return f.start, f.end
}
func (f *Field) SetRange(start, end uint8) {
	f.start = start
	f.end = end
}

func (f *Field) Deal(data []byte) error {
	return f.DealFunc(f, data)
}

// 起始符
func NewStarter(start []byte) Fielder {
	field := &Field{
		name:     "起始符",
		defaultV: start,
		len:      len(start),
	}
	field.DealFunc = func(field Fielder, data []byte) error {
		if data == nil {
			return errors.New("数据为空")
		}
		if len(data) < field.Length() {
			return errors.New("数据长度小于起始符长度")
		}
		if bytes.Equal(data[:field.Length()], field.GetDefault()) {
			return fmt.Errorf("起始符错误Need:%s,But:%s", string(field.GetDefault()), string(data[:field.Length()]))
		}
		return nil
	}
	return field
}

type Starter struct {
	Field
}

func (start Starter) Deal(data []byte) error {
	if len(data) != int(start.len) {
		return errors.New("起始长度不对")
	}
	for i, v := range start.defaultV {
		if v != data[i] {
			fmt.Printf("起始：%# 02x，预期：%# 02x\n", data, start.defaultV)
			return errors.New("起始值错误")
		}
	}
	fmt.Printf("[起始值]:% 02X\n", data)
	return nil
}

type DataLen struct {
	Field
}

func (d DataLen) Deal(data []byte) error {
	if len(data) != int(d.len) {
		return errors.New("数据域长度字段本省身长度不对")
	}
	var l = make([]byte, d.len)
	data = data[:d.len]
	buffer := bytes.NewBuffer(data)
	err := binary.Read(buffer, d.order, l)
	if err != nil {
		fmt.Println("binary read error", err)
		return err
	}
	bin2Uint64, err := BIN2Uint64(data, d.order)
	if err != nil {
		fmt.Println("b2i error: ", err)
		return err
	}
	fmt.Printf("[数据长度]:%d字节\n", bin2Uint64)
	return nil
}

type Datar struct {
	Field
}

func (datar Datar) Deal(data []byte) error {
	//解析再打印
	fmt.Printf("[数据]:% 0X\n", data)
	return nil
}
