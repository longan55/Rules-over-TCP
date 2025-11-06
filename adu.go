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

func NewProtocolBuilder() *ProtocolBuilder {
	return &ProtocolBuilder{
		du: &ProtocolImpl{
			elements: make([]ProtocolElement, 0, 3),
		},
		fh: make(map[FunctionCode]*FunctionHandler, 32),
	}
}

//协议元素组成规则
//1. 第一个元素必须是起始符
//2. 第一个元素到数据域之前的元素组成 消息头部元素.
//3. 数据域称为消息体元素
//4. 最后一个元素必须是校验码元素

type ProtocolBuilder struct {
	du *ProtocolImpl
	fh map[FunctionCode]*FunctionHandler
}

func (duBuilder *ProtocolBuilder) AddElement(element ProtocolElement) *ProtocolBuilder {
	duBuilder.du.elements = append(duBuilder.du.elements, element)
	return duBuilder
}

func (duBuilder *ProtocolBuilder) AddCryptConfig(cryptConfig *CryptConfig) *ProtocolBuilder {
	duBuilder.du.cryptLib = cryptConfig.GetCryptMap()
	return duBuilder
}

func (duBuilder *ProtocolBuilder) AddHandler(fc FunctionCode, f *FunctionHandler) *ProtocolBuilder {
	duBuilder.du.AddHandler(fc, f)
	return duBuilder
}

func (duBuilder *ProtocolBuilder) AddHandlerConfig(config *HandlerConfig) *ProtocolBuilder {
	duBuilder.du.handlerMap = config.handlerMap
	return duBuilder
}

func (duBuilder *ProtocolBuilder) NewHandler(fc FunctionCode) *FunctionHandler {
	fh := &FunctionHandler{
		fc: fc,
	}
	duBuilder.fh[fc] = fh
	return fh
}

func (duBuilder *ProtocolBuilder) Build() (Protocol, error) {
	// 添加协议元素验证
	if len(duBuilder.du.elements) == 0 {
		return nil, errors.New("协议元素不能为空")
	}

	// 验证第一个元素必须是起始符
	if duBuilder.du.elements[0].Type() != Preamble {
		return nil, errors.New("第一个协议元素必须是起始符")
	}

	// 验证最后一个元素必须是校验码
	lastIdx := len(duBuilder.du.elements) - 1
	if duBuilder.du.elements[lastIdx].Type() != Checksum {
		return nil, errors.New("最后一个协议元素必须是校验码")
	}

	fmt.Println("协议元素组成部分：")
	fmt.Println("------------------------------------------------------")
	fmt.Printf("元素名称\t元素类型\t元素长度\t默认值\n")
	//todo 起始码+长度码 的长度
	for index, element := range duBuilder.du.elements {
		element.SetIndex(index)
		fmt.Printf("%s\t%v\t\t%d\t\t%#x\n", element.GetName(), element.Type(), element.Length(), element.RealValue())
	}
	fmt.Println("------------------------------------------------------")
	return duBuilder.du, nil
}

type Protocol interface {
	AddHandler(fc FunctionCode, f *FunctionHandler)
	Handle(ctx context.Context, conn net.Conn)
	SetDataLength(length int)
	//Parse(adu [][]byte) error
	// Serialize(f *FucntionHandler) []byte
}

var _ Protocol = (*ProtocolImpl)(nil)

// dph 应用数据单元 结构体
type ProtocolImpl struct {
	counts uint64
	//当前数据单元的 数据域长度
	dataLength   int
	functionCode FunctionCode
	//加密标志
	encryptionFlag int
	cryptLib       map[int]CryptFunc

	conn net.Conn
	//存储协议元素信息
	elements []ProtocolElement
	// parserMap  map[FunctionCode]Parser
	handlerMap map[FunctionCode]*FunctionHandler
}

func (dph *ProtocolImpl) AddHandler(fc FunctionCode, f *FunctionHandler) {
	if dph.handlerMap == nil {
		dph.handlerMap = make(map[FunctionCode]*FunctionHandler)
	}
	dph.handlerMap[fc] = f
}

// 字段顺序已有---》新增处理顺序
func (dph *ProtocolImpl) Handle(ctx context.Context, conn net.Conn) {
	dph.conn = conn
	for {
		select {
		case <-ctx.Done():
			//停止读取
			return
		default:
			fmt.Printf("[第%v个数据单元解析开始]\n", dph.counts)
			alldata := make([][]byte, 0, len(dph.elements))
			//第一遍遍历elements, 读取一个完整的数据单元
			for _, element := range dph.elements {
				//定义好合适长度的buf,接收该元素数据
				var buf []byte
				if element.Type() == Payload {
					buf = make([]byte, dph.dataLength-2)
				} else {
					buf = make([]byte, element.Length())
				}
				_, err := io.ReadFull(dph.conn, buf)
				if err != nil {
					fmt.Println("读取数据失败:", err)
					return
				}

				alldata = append(alldata, buf)
				if element.Type() == Preamble {
					_, _, err := element.Deal(alldata)
					if err != nil {
						fmt.Println("起始符校验失败: ", err)
						break
					}
				} else if element.Type() == Length {
					_, a, err := element.Deal(alldata)
					if err != nil {
						fmt.Println("数据长度校验失败: ", err)
						break
					}
					dph.dataLength = a.(int)
				}
			}
			//第二次遍历elements, 解析数据单元
			for _, element := range dph.elements {
				switch element.Type() {
				case Preamble, Length:
					continue
				case EncryptionFlag:
					_, a, err := element.Deal(alldata)
					if err != nil {
						fmt.Println("加密标志校验失败: ", err)
						break
					}
					dph.encryptionFlag = a.(int)
				case Function:
					_, a, err := element.Deal(alldata)
					if err != nil {
						fmt.Println("功能码校验失败: ", err)
						break
					}
					dph.functionCode = a.(FunctionCode)
				case Payload:
					_, a, err := element.Deal(alldata)
					if err != nil {
						fmt.Println("数据解析失败:", err)
						break
					}
					data := a.([]byte)
					data, err = dph.cryptLib[dph.encryptionFlag](data)
					if err != nil {
						fmt.Println("解密失败:", err)
						return
					}
					hd, ok := dph.handlerMap[dph.functionCode]
					if !ok {
						fmt.Println("未注册功能码:", dph.functionCode)
						return
					}
					err = hd.Handle(data)
					if err != nil {
						fmt.Println("处理数据失败:", err)
						return
					}
				}
			}
			fmt.Printf("[第%v个数据单元解析完成]\n", dph.counts)
			fmt.Println()
			dph.counts++
		}
	}
}

func (dph *ProtocolImpl) SetDataLength(length int) {
	dph.dataLength = length
}

func (dph *ProtocolImpl) Info() {
	for _, v := range dph.elements {
		of := reflect.TypeOf(v)
		fmt.Println("类型:", of, " 长度:", v.Length())
	}
}

// ProtocolElement 元素接口
type ProtocolElement interface {
	GetIndex() int
	SetIndex(index int)
	GetName() string
	Type() ProtocolElementType
	RealValue() []byte
	Length() int
	// GetScale() uint8
	GetOrder() binary.ByteOrder
	// GetRange() (start, end uint8)
	Deal([][]byte) (ProtocolElementType, any, error)
	ChecksumType() uint8
}

type ProtocolElementType byte

const (
	// 帧首符
	Preamble ProtocolElementType = iota
	// 当前元素之后的数据长度
	Length
	// 加密标志
	EncryptionFlag
	// 功能码
	Function
	// 消息负载
	Payload
	// 校验和
	Checksum
)

var _ ProtocolElement = (*ProtocolElementImpl)(nil)

// ProtocolElementImpl 基础元素结构体
type ProtocolElementImpl struct {
	//元数据: 存储该元素的元数据(用于描述说明)
	index        int                                                                            //说明该元素的索引
	Typ          ProtocolElementType                                                            //元素类型
	name         string                                                                         //元素名字
	selfLength   int                                                                            //元素本身长度
	defaultValue []byte                                                                         //默认值
	order        binary.ByteOrder                                                               //大小端
	start        uint8                                                                          //开始索引: 该元素影响的元素区域的第一个元素索引
	end          uint8                                                                          //结束索引: 该元素影响的元素区域的最后一个元素索引
	DealFunc     func(element ProtocolElement, data [][]byte) (ProtocolElementType, any, error) //处理函数
	checksumType uint8
}

func (f *ProtocolElementImpl) GetIndex() int {
	return f.index
}

func (f *ProtocolElementImpl) SetIndex(index int) {
	f.index = index
}

func (f *ProtocolElementImpl) GetName() string {
	return f.name
}

func (f *ProtocolElementImpl) Type() ProtocolElementType {
	return f.Typ
}

func (f *ProtocolElementImpl) RealValue() []byte {
	return f.defaultValue
}
func (f *ProtocolElementImpl) SetLen(l int) {
	f.selfLength = l
}
func (f *ProtocolElementImpl) Length() int {
	return f.selfLength
}

func (f *ProtocolElementImpl) GetOrder() binary.ByteOrder {
	return f.order
}

func (f *ProtocolElementImpl) GetRange() (start, end uint8) {
	return f.start, f.end
}

func (f *ProtocolElementImpl) Deal(data [][]byte) (ProtocolElementType, any, error) {
	return f.DealFunc(f, data)
}

func (f *ProtocolElementImpl) ChecksumType() uint8 {
	return f.checksumType
}

// 起始符
func NewStarter(start []byte) ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:          Preamble,
		name:         "帧 首 符",
		defaultValue: start,
		selfLength:   len(start),
	}
	element.DealFunc = func(element ProtocolElement, fullData [][]byte) (ProtocolElementType, any, error) {
		if fullData == nil {
			return element.Type(), nil, errors.New("数据为空")
		}
		// if len(fullData) < element.Length() {
		// 	return element.Type(), nil, errors.New("数据长度小于起始符长度")
		// }
		if !bytes.Equal(fullData[0][:element.Length()], element.RealValue()) {
			return element.Type(), nil, fmt.Errorf("起始符错误Need:%0X,But:%0X", element.RealValue(), fullData[0][:element.Length()])
		}
		fmt.Printf("起始符:\t\t\t[%#0X]\n", fullData[0][:element.Length()])
		return element.Type(), nil, nil
	}
	return element
}

func NewDataLen(selfLength int) ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:          Length,
		name:         "帧 长 度",
		defaultValue: nil,
		selfLength:   selfLength,
	}
	element.DealFunc = func(element ProtocolElement, fullData [][]byte) (ProtocolElementType, any, error) {
		if fullData == nil {
			return element.Type(), nil, errors.New("数据为空")
		}
		// if len(fullData) < element.Length() {
		// 	return element.Type(), nil, errors.New("数据长度小于帧长度字段长度")
		// }
		data := fullData[element.GetIndex()]
		length := Bin2Int(data, element.GetOrder())
		fmt.Printf("帧长度:\t\t\t[%d]\n", length)
		return element.Type(), length, nil
	}
	return element
}

func NewCyptoFlag() ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:          EncryptionFlag,
		name:         "加密标识 ",
		defaultValue: []byte{0x01},
		selfLength:   1,
	}
	element.DealFunc = func(element ProtocolElement, fullData [][]byte) (ProtocolElementType, any, error) {
		if fullData == nil {
			return element.Type(), nil, errors.New("数据为空")
		}
		flagdata := fullData[element.GetIndex()]
		flag := Bin2Int(flagdata, element.GetOrder())
		fmt.Printf("加密标识:\t\t[%#0X]\n", fullData[element.GetIndex()])
		return element.Type(), flag, nil
	}
	return element
}

func NewFuncCode() ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:          Function,
		name:         "功 能 码",
		defaultValue: nil,
		selfLength:   1,
	}
	element.DealFunc = func(element ProtocolElement, fullData [][]byte) (ProtocolElementType, any, error) {
		if fullData == nil {
			return element.Type(), nil, errors.New("数据为空")
		}
		// if len(fullData) < element.Length() {
		// 	return element.Type(), nil, errors.New("数据长度小于功能码字段长度")
		// }
		fmt.Printf("功能码:\t\t\t[%#0X]\n", fullData[element.GetIndex()])
		functionCode := FunctionCode(fullData[element.GetIndex()][0])
		return element.Type(), functionCode, nil
	}
	return element
}
func NewPayload() ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:          Payload,
		name:         "帧 负 载",
		defaultValue: nil,
		selfLength:   1,
	}
	element.DealFunc = func(element ProtocolElement, fullData [][]byte) (ProtocolElementType, any, error) {
		if fullData == nil {
			return element.Type(), nil, errors.New("数据为空")
		}
		// if len(fullData) < element.Length() {
		// 	return element.Type(), nil, errors.New("数据长度小于帧负载字段长度")
		// }
		fmt.Printf("帧负载:\t\t\t[% #0X]\n", fullData[element.GetIndex()])
		return element.Type(), fullData[element.GetIndex()], nil
	}
	return element
}

func NewCheckSum(checksumType uint8, selfLength int) ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:          Checksum,
		name:         "校 验 码",
		defaultValue: nil,
		selfLength:   selfLength,
		checksumType: checksumType,
	}
	element.DealFunc = func(element ProtocolElement, fullData [][]byte) (ProtocolElementType, any, error) {
		if fullData == nil {
			return element.Type(), nil, errors.New("数据为空")
		}
		// if len(fullData) < element.Length() {
		// 	return element.Type(), nil, errors.New("数据长度小于校验码字段长度")
		// }
		fmt.Printf("校 验 码:\t\t[% #0X]\n", fullData[element.GetIndex()])
		checksum0 := fullData[element.GetIndex()]
		//将各切片连接为一个切片
		full := bytes.Join(fullData[2:element.GetIndex()], nil)
		checksum := CheckSum(element.ChecksumType(), full)
		if !bytes.Equal(checksum, checksum0) {
			return element.Type(), nil, fmt.Errorf("校验码错误Need:%0X,But:%0X", checksum, checksum0)
		}
		fmt.Printf("校验码类型:%d,计算校验码:% #0X,校验通过\n", element.ChecksumType(), checksum)
		return element.Type(), checksum, nil
	}
	return element
}
