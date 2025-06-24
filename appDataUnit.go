package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/longan55/proto/protocol"
)

//通过框架配置协议，框架自动解析和封装，无需自己开发。

// AppDataUnit 应用数据单元接口
type AppDataUnit interface {
	Marshal(AppDataUnit) []byte
	UnMarshal([]byte) (AppDataUnit, error)
}

func NewAdu() Adu {
	return Adu{Fields: make([]Fielder, 0, 3)}
}

// Adu 应用数据单元 结构体
type Adu struct {
	Fields []Fielder
}

func (adu *Adu) Info() {
	for _, v := range adu.Fields {
		of := reflect.TypeOf(v)
		fmt.Println("类型:", of, " 长度:", v.Length())
	}
}

// AddField 用于添加元素
func (adu *Adu) AddField(field Fielder) {
	adu.Fields = append(adu.Fields, field)
}

// Debug 解析数据
func (adu *Adu) Debug(r io.Reader, source []byte) {
	// 起始符 只需要判断是否相等
	// 数据域长度 要传给数据域元素作为长度
	// 加密标志 是否对指定元素的值进行加密或解密
	// 校验码 是否对指定元素进行校验计算
	offset := uint16(0)
	//遍历所有元素
	for _, field := range adu.Fields {
		//根据元素Field获取对应数据切片
		data := source[offset : offset+field.Length()]
		//更新偏移量
		offset += field.Length()
		//debug打印元素
		if field.Scale() == 0 {
			fmt.Printf("[%s] = %0d", field.Name(), data)
		} else {
			fmt.Printf("[%s] = %0x", field.Name(), data)
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
	Name() string
	// Value 获取默认值
	Value() []byte
	// Scale 获取进制
	Scale() uint8
	// SetLen 设置元素长度
	SetLen(uint16)
	// Length 获取元素长度
	Length() uint16
	// Deal 解析元素
	Deal([]byte) error
}

type fieldType byte

// Field 基础元素结构体
type Field struct {
	Fielder
	name string //消息帧 元素名字
	//FType        fieldType      //消息帧 字段类型
	scale        uint8            // 1十六进制，0十进制
	Len          uint16           //消息帧 元素本身长度
	DefaultValue []byte           //默认值
	Order        binary.ByteOrder //大小端
	Next         *Fielder
}

func (f *Field) Name() string {
	return f.name
}

func (f *Field) SetName(name string) {
	f.name = name
}

func (f *Field) SetLen(l uint16) {
	f.Len = l
}
func (f *Field) Length() uint16 {
	return f.Len
}

func (f *Field) Scale() uint8 {
	return f.scale
}

func (f *Field) SetScale(u uint8) {
	f.scale = u
}

func (f *Field) SetDefaultValue(val []byte) {
	f.DefaultValue = val
}
func (f *Field) SetDefaultValueString(str string) {
	f.DefaultValue = []byte(str)
}
func (f *Field) Value() []byte {
	return f.DefaultValue
}

type Starter struct {
	Field
}

func (start Starter) Deal(data []byte) error {
	if len(data) != int(start.Len) {
		return errors.New("起始长度不对")
	}
	for i, v := range start.DefaultValue {
		if v != data[i] {
			fmt.Printf("起始：%# 02x，预期：%# 02x\n", data, start.DefaultValue)
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
	if len(data) != int(d.Len) {
		return errors.New("数据域长度字段本省身长度不对")
	}
	var l = make([]byte, d.Len)
	data = data[:d.Len]
	buffer := bytes.NewBuffer(data)
	err := binary.Read(buffer, d.Order, l)
	if err != nil {
		fmt.Println("binary read error", err)
		return err
	}
	bin2Uint64, err := protocol.BIN2Uint64(data, d.Order)
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
