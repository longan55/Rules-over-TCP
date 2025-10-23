package rot

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

func NewBuilder() *Builder {
	return &Builder{
		du: &DataHandler{
			Fields: make([]Fielder, 0, 3),
		},
		fh: make(map[FunctionCode]*FucntionHandler, 32),
	}
}

//协议元素组成规则
//1. 第一个元素必须是起始符
//2. 第一个元素到数据域之前的元素组成 消息头部元素.
//3. 数据域称为消息体元素
//4. 最后一个元素必须是校验码元素

type Builder struct {
	du *DataHandler
	fh map[FunctionCode]*FucntionHandler
}

func (duBuilder *Builder) AddFielder(field Fielder) *Builder {
	duBuilder.du.Fields = append(duBuilder.du.Fields, field)
	return duBuilder
}

func (duBuilder *Builder) AddCryptConfig(cryptConfig *CryptConfig) *Builder {
	duBuilder.du.cryptLib = cryptConfig.Config()
	return duBuilder
}

func (duBuilder *Builder) AddHandler(fc FunctionCode, f *FucntionHandler) *Builder {
	duBuilder.du.AddHandler(fc, f)
	return duBuilder
}

func (duBuilder *Builder) AddHandlerConfig(config *HandlerConfig) *Builder {
	duBuilder.du.handlerMap = config.handlerMap
	return duBuilder
}

func (duBuilder *Builder) NewHandler(fc FunctionCode) *FucntionHandler {
	fh := &FucntionHandler{
		fc: fc,
	}
	duBuilder.fh[fc] = fh
	return fh
}

func (duBuilder *Builder) Build() (MainHandler, error) {
	//todo 起始码+长度码 的长度
	for index, field := range duBuilder.du.Fields {
		field.SetIndex(index)
		fmt.Printf("元素名称:%s, 元素类型:%v, 自身长度:%v\n", field.GetName(), field.Type(), field.Length())
	}
	//TODO: 校验元素是否符合协议规范
	return duBuilder.du, nil
}

type MainHandler interface {
	AddHandler(fc FunctionCode, f *FucntionHandler)
	Handle(ctx context.Context, conn net.Conn)
	SetDataLength(length uint64)
	Parse(adu [][]byte) error
	// Serialize(f *FucntionHandler) []byte
}

var _ MainHandler = (*DataHandler)(nil)

// dph 应用数据单元 结构体
type DataHandler struct {
	//当前数据单元的 数据域长度
	dataLength   uint64
	functionCode FunctionCode
	//加密标志
	encryptionFlag uint64
	cryptLib       map[uint64]CryptFunc

	conn net.Conn
	//存储协议元素信息
	Fields []Fielder
	// parserMap  map[FunctionCode]Parser
	handlerMap map[FunctionCode]*FucntionHandler
}

func (dph *DataHandler) AddHandler(fc FunctionCode, f *FucntionHandler) {
	if dph.handlerMap == nil {
		dph.handlerMap = make(map[FunctionCode]*FucntionHandler)
	}
	dph.handlerMap[fc] = f
}

// 字段顺序已有---》新增处理顺序
func (dph *DataHandler) Handle(ctx context.Context, conn net.Conn) {
	dph.conn = conn
	for {
		select {
		case <-ctx.Done():
			//停止读取
			return
		default:
			alldata := make([][]byte, 0, len(dph.Fields))
			//第一遍遍历fields, 读取一个完整的数据单元
			for _, field := range dph.Fields {
				//定义好合适长度的buf,接收该元素数据
				var buf []byte
				if field.Type() != DATA {
					buf = make([]byte, field.Length())
				} else {
					buf = make([]byte, dph.dataLength)
				}
				//读取
				_, err := io.ReadFull(dph.conn, buf)
				if err != nil {
					fmt.Println("读取数据失败:", err)
					return
				}

				//is start with correct code?
				startFlag := true

				typ, a, err := field.Deal(alldata)
				switch typ {
				case START:
					if err != nil {
						fmt.Println("起始符校验失败: ", err)
						startFlag = false
					}
				case LENGTH:
					dph.dataLength = a.(uint64)
				case ENCRYPTION:
					dph.encryptionFlag = a.(uint64)
				}
				if !startFlag {
					break
				}
				//将数据拼接
				alldata = append(alldata, buf)
			}
			//第二次遍历fields, 解析数据单元
			for _, field := range dph.Fields {
				if field.Type() == START {
					continue
				}
				typ, a, err := field.Deal(alldata)
				if err != nil {
					fmt.Println("数据解析失败:", err)
					break
				}
				switch typ {
				case LENGTH:
					dph.dataLength = a.(uint64)
				case ENCRYPTION:
					dph.encryptionFlag = a.(uint64)
				case FUNCTION:
					dph.functionCode = a.(FunctionCode)
				case DATA:
					data := a.([]byte)
					var err error
					data, err = dph.cryptLib[dph.encryptionFlag](data)
					if err != nil {
						return
					}
					hd, ok := dph.handlerMap[dph.functionCode]
					if !ok {
						fmt.Println("未注册功能码:", dph.functionCode)
						return
					}
					err = hd.Handle(data)
					if err != nil {
						return
					}
				}
			}
			fmt.Println("读取数据:", alldata)
		}
	}
}

func (dph *DataHandler) SetDataLength(length uint64) {
	dph.dataLength = length
}

func (dph *DataHandler) Parse(alldata [][]byte) error {
	fmt.Println("解析前数据:", alldata)
	for _, field := range dph.Fields {
		// if field.Type() == START {
		// 	fmt.Println("起始符:", alldata[field.GetIndex()])
		// 	continue
		// }
		if field.Type() == START {
			continue
		}
		typ, a, err := field.Deal(alldata)
		if err != nil {
			fmt.Println("数据解析失败:", err)
			break
		}
		switch typ {
		case LENGTH:
			dph.dataLength = a.(uint64)
		case ENCRYPTION:
			dph.encryptionFlag = a.(uint64)
		case FUNCTION:
			dph.functionCode = a.(FunctionCode)
		case DATA:
			data := a.([]byte)
			var err error
			data, err = dph.cryptLib[dph.encryptionFlag](data)
			if err != nil {
				return err
			}
			hd, ok := dph.handlerMap[dph.functionCode]
			if !ok {
				return fmt.Errorf("未注册功能码:%v", dph.functionCode)
			}
			err = hd.Handle(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (dph *DataHandler) Info() {
	for _, v := range dph.Fields {
		of := reflect.TypeOf(v)
		fmt.Println("类型:", of, " 长度:", v.Length())
	}
}

// Debug 解析数据
func (dph *DataHandler) Debug(r io.Reader, source []byte) {
	// 起始符 只需要判断是否相等
	// 数据域长度 要传给数据域元素作为长度
	// 加密标志 是否对指定元素的值进行加密或解密
	// 校验码 是否对指定元素进行校验计算
	// offset := 0
	// //遍历所有元素
	// for _, field := range dph.Fields {
	// 	//根据元素Field获取对应数据切片
	// 	data := source[offset : offset+field.Length()]
	// 	//更新偏移量
	// 	offset += field.Length()
	// 	//debug打印元素
	// 	if field.GetScale() == 0 {
	// 		fmt.Printf("[%s] = %0d", field.GetName(), data)
	// 	} else {
	// 		fmt.Printf("[%s] = %0x", field.GetName(), data)
	// 	}

	// 	//处理方法
	// 	_, err := field.Deal(data)
	// 	if err != nil { //log.Println("数据解析出错! [error]:", err)
	// 		fmt.Printf("数据解析出错! [error]: %v\n", err)
	// 	}
	// }
}

// Fielder 元素接口
type Fielder interface {
	GetIndex() int
	SetIndex(index int)
	GetName() string
	Type() FieldType
	RealValue() []byte
	Length() int
	GetScale() uint8
	GetOrder() binary.ByteOrder
	GetRange() (start, end uint8)
	Deal([][]byte) (FieldType, any, error)
}

type FieldType byte

const (
	// 起始符
	START FieldType = iota
	// 数据域长度
	LENGTH
	// 加密标志
	ENCRYPTION
	// 功能码
	FUNCTION
	// 数据域
	DATA
	// 校验码
	CHECK
)

var _ Fielder = (*Field)(nil)

// Field 基础元素结构体
type Field struct {
	//元数据: 存储该元素的元数据(用于描述说明)
	index    int                                                        //说明该元素的索引
	Typ      FieldType                                                  //元素类型
	name     string                                                     //元素名字
	scale    uint8                                                      // 1十六进制，0十进制
	len      int                                                        //元素本身长度
	defaultV []byte                                                     //默认值
	order    binary.ByteOrder                                           //大小端
	start    uint8                                                      //开始索引: 该元素影响的元素区域的第一个元素索引
	end      uint8                                                      //结束索引: 该元素影响的元素区域的最后一个元素索引
	DealFunc func(field Fielder, data [][]byte) (FieldType, any, error) //处理函数
	//临时数据: 存储当前adu的数据
	realData   []byte
	parsedData any
}

func (f *Field) GetIndex() int {
	return f.index
}

func (f *Field) SetIndex(index int) {
	f.index = index
}

func (f *Field) GetName() string {
	return f.name
}

func (f *Field) Type() FieldType {
	return f.Typ
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

func (f *Field) GetScale() uint8 {
	return f.scale
}

func (f *Field) GetOrder() binary.ByteOrder {
	return f.order
}

func (f *Field) GetRange() (start, end uint8) {
	return f.start, f.end
}

func (f *Field) Deal(data [][]byte) (FieldType, any, error) {
	return f.DealFunc(f, data)
}

// 起始符
func NewStarter(start []byte) Fielder {
	field := &Field{
		Typ:      START,
		name:     "起始符",
		defaultV: start,
		len:      len(start),
	}
	field.DealFunc = func(field Fielder, data [][]byte) (FieldType, any, error) {
		if data == nil {
			return field.Type(), nil, errors.New("数据为空")
		}
		if len(data) < field.Length() {
			return field.Type(), nil, errors.New("数据长度小于起始符长度")
		}
		if !bytes.Equal(data[0][:field.Length()], field.RealValue()) {
			return field.Type(), nil, fmt.Errorf("起始符错误Need:%0X,But:%0X", field.RealValue(), data[0][:field.Length()])
		}
		fmt.Printf("起始符:\t\t\t[%#0X]\n", data[0][:field.Length()])
		return field.Type(), nil, nil
	}
	return field
}

func NewDataLen(length int) Fielder {
	field := &Field{
		Typ:      LENGTH,
		name:     "数据域长度",
		defaultV: nil,
		len:      length,
	}
	field.DealFunc = func(field Fielder, data [][]byte) (FieldType, any, error) {
		if data == nil {
			return field.Type(), nil, errors.New("数据为空")
		}
		if len(data) < field.Length() {
			return field.Type(), nil, errors.New("数据长度小于数据域长度字段长度")
		}
		lenData := data[field.GetIndex()]
		u64, err := BIN2Uint64(lenData, field.GetOrder())
		if err != nil {
			return field.Type(), nil, err
		}
		fmt.Printf("数据域长度:\t\t[%d]\n", u64)
		return field.Type(), u64, nil
	}
	return field
}

// TODO: 加密标志，设置加密算法库。
func NewCyptoFlag() Fielder {
	field := &Field{
		Typ:      ENCRYPTION,
		name:     "加密标志",
		defaultV: []byte{0x01},
		len:      1,
	}
	field.DealFunc = func(field Fielder, data [][]byte) (FieldType, any, error) {
		if data == nil {
			return field.Type(), nil, errors.New("数据为空")
		}
		flagdata := data[field.GetIndex()]
		u64, err := BIN2Uint64(flagdata, field.GetOrder())
		if err != nil {
			return field.Type(), false, err
		}
		fmt.Printf("加密标志:\t\t[%#0X]\n", data[field.GetIndex()])
		return field.Type(), u64, nil
	}
	return field
}

func NewFuncCode() Fielder {
	field := &Field{
		Typ:      FUNCTION,
		name:     "功能码",
		defaultV: nil,
		len:      1,
	}
	field.DealFunc = func(field Fielder, data [][]byte) (FieldType, any, error) {
		if data == nil {
			return field.Type(), nil, errors.New("数据为空")
		}
		if len(data) < field.Length() {
			return field.Type(), nil, errors.New("数据长度小于功能码字段长度")
		}
		fmt.Printf("功能码:\t\t\t[%#0X]\n", data[field.GetIndex()])
		fc := FunctionCode(data[field.GetIndex()][0])
		return field.Type(), fc, nil
	}
	return field
}
func NewDataZone() Fielder {
	field := &Field{
		Typ:      DATA,
		name:     "数据域",
		defaultV: nil,
		len:      1,
	}
	field.DealFunc = func(field Fielder, data [][]byte) (FieldType, any, error) {
		if data == nil {
			return field.Type(), nil, errors.New("数据为空")
		}
		if len(data) < field.Length() {
			return field.Type(), nil, errors.New("数据长度小于数据域字段长度")
		}
		fmt.Printf("数据域:\t\t\t[% #0X]\n", data[field.GetIndex()])
		return field.Type(), data[field.GetIndex()], nil
	}
	return field
}
