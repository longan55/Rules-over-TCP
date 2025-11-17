package rot

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

var (
	once         sync.Once
	defaultOrder binary.ByteOrder = binary.BigEndian
)

func SetDefaultOrder(order binary.ByteOrder) {
	once.Do(func() {
		defaultOrder = order
	})
}

func DefaultOrder() binary.ByteOrder {
	return defaultOrder
}

func NewProtocolBuilder() *ProtocolBuilder {
	return &ProtocolBuilder{
		du: &ProtocolDataUnit{
			elements: make([]ProtocolElement, 0, 3),
		},
	}
}

//协议元素组成规则
//1. 第一个元素必须是起始符
//2. 第一个元素到数据域之前的元素组成 消息头部元素.
//3. 数据域称为消息体元素
//4. 最后一个元素必须是校验码元素

type ProtocolBuilder struct {
	du *ProtocolDataUnit
}

func (duBuilder *ProtocolBuilder) SetDefaultOrder(order binary.ByteOrder) *ProtocolBuilder {
	defaultOrder = order
	return duBuilder
}

func (duBuilder *ProtocolBuilder) AddElement(element ProtocolElement) *ProtocolBuilder {
	duBuilder.du.elements = append(duBuilder.du.elements, element)
	return duBuilder
}

// AddCryptConfig 添加加密配置,应该只被调用一次
func (duBuilder *ProtocolBuilder) AddCryptConfig(cryptConfig *CryptConfig) *ProtocolBuilder {
	duBuilder.du.cryptLib = cryptConfig.cryptMap
	return duBuilder
}

func (duBuilder *ProtocolBuilder) AddCrypt(cryptFlag int, cipher Cipher) *ProtocolBuilder {
	duBuilder.du.AddCrypt(cryptFlag, cipher)
	return duBuilder
}

func (duBuilder *ProtocolBuilder) AddHandlerConfig(config *HandlerConfig) *ProtocolBuilder {
	duBuilder.du.handlerMap = config.handlerMap
	return duBuilder
}

func (duBuilder *ProtocolBuilder) AddHandler(fc FunctionCode, f *FunctionHandler) *ProtocolBuilder {
	duBuilder.du.AddHandler(fc, f)
	return duBuilder
}
func (duBuilder *ProtocolBuilder) NewHandler(fc FunctionCode) *FunctionHandler {
	fh := &FunctionHandler{
		fc: fc,
	}
	duBuilder.du.AddHandler(fc, fh)
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
		fmt.Printf("%s\t%v\t\t%d\t\t%#x\n", element.GetName(), element.Type(), element.Length(), element.DefaultValue())
	}
	fmt.Println("------------------------------------------------------")
	return duBuilder.du, nil
}

type Protocol interface {
	AddHandler(fc FunctionCode, f *FunctionHandler)
	Handle(ctx context.Context, conn net.Conn)
	SetDataLength(length int)
}

var _ Protocol = (*ProtocolDataUnit)(nil)

// ProtocolDataUnit 协议数据单元, 保存所有协议的上下文信息
type ProtocolDataUnit struct {
	counts uint64
	//当前数据单元的 数据域长度
	dataLength   int
	functionCode FunctionCode
	//加密标志
	encryptionFlag int
	cryptLib       map[int]Cipher

	conn net.Conn
	//存储协议元素信息
	elements   []ProtocolElement
	handlerMap map[FunctionCode]*FunctionHandler
}

// GetElementByIndex 通过索引获取ProtocolElement
func (pdu *ProtocolDataUnit) GetElementByIndex(index int) ProtocolElement {
	if index >= 0 && index < len(pdu.elements) {
		return pdu.elements[index]
	}
	return nil
}

// GetElementByType 通过类型获取ProtocolElement
func (pdu *ProtocolDataUnit) GetElementByType(typ ProtocolElementType) ProtocolElement {
	for _, element := range pdu.elements {
		if element.Type() == typ {
			return element
		}
	}
	return nil
}

// GetAllElements 获取所有ProtocolElement
func (pdu *ProtocolDataUnit) GetAllElements() []ProtocolElement {
	return pdu.elements
}

func (pdu *ProtocolDataUnit) AddCrypt(cryptFlag int, crypt Cipher) {
	if pdu.cryptLib == nil {
		pdu.cryptLib = make(map[int]Cipher)
	}
	pdu.cryptLib[cryptFlag] = crypt
}

func (pdu *ProtocolDataUnit) Decrypt(cryptFlag int, src []byte) ([]byte, error) {
	if pdu.cryptLib == nil {
		return nil, errors.New("未配置加密算法")
	}
	if _, ok := pdu.cryptLib[cryptFlag]; !ok {
		panic(fmt.Sprintf("未配置加密算法 %d", cryptFlag))
	}
	return pdu.cryptLib[cryptFlag].Decrypt(src)
}

func (pdu *ProtocolDataUnit) AddHandler(fc FunctionCode, f *FunctionHandler) {
	if pdu.handlerMap == nil {
		pdu.handlerMap = make(map[FunctionCode]*FunctionHandler)
	}
	pdu.handlerMap[fc] = f
}

func (pdu *ProtocolDataUnit) SetDataLength(length int) {
	pdu.dataLength = length
}

// 字段顺序已有---》新增处理顺序
func (pdu *ProtocolDataUnit) Handle(ctx context.Context, conn net.Conn) {
	pdu.conn = conn
	for {
		select {
		case <-ctx.Done():
			//停止读取
			return
		default:
			fmt.Printf("[第%v个数据单元解析开始]\n", pdu.counts)
			alldata := make([][]byte, 0, len(pdu.elements))
			//第一遍遍历elements, 读取一个完整的数据单元
			for _, element := range pdu.elements {
				//定义好合适长度的buf,接收该元素数据
				var buf []byte
				//如果是数据域, 就需要从ProtocolDataUnit中获取长度
				if element.Type() == Payload {
					buf = make([]byte, pdu.dataLength-2)
				} else {
					//其他元素, 就直接通过Length()读取
					buf = make([]byte, element.Length())
				}
				_, err := io.ReadFull(pdu.conn, buf)
				if err != nil {
					fmt.Println("读取数据失败:", err)
					return
				}

				alldata = append(alldata, buf)
				if element.Type() == Preamble {
					err := element.Deal(alldata, pdu)
					if err != nil {
						fmt.Println("起始符校验失败: ", err)
						break
					}
				} else if element.Type() == Length {
					err := element.Deal(alldata, pdu)
					if err != nil {
						fmt.Println("数据长度校验失败: ", err)
						break
					}
					pdu.dataLength = element.RealValue().(int)
				}
			}
			//第二次遍历elements, 解析数据单元
			for _, element := range pdu.elements {
				switch element.Type() {
				case Preamble, Length:
					continue
				case EncryptionFlag:
					err := element.Deal(alldata, pdu)
					if err != nil {
						fmt.Println("加密标志校验失败: ", err)
						break
					}
					pdu.encryptionFlag = element.RealValue().(int)
				case Function:
					err := element.Deal(alldata, pdu)
					if err != nil {
						fmt.Println("功能码校验失败: ", err)
						break
					}
					pdu.functionCode = element.RealValue().(FunctionCode)
				case Payload:
					err := element.Deal(alldata, pdu)
					if err != nil {
						fmt.Println("数据解析失败:", err)
						break
					}
					data := element.RealValue().([]byte)
					data, err = pdu.Decrypt(pdu.encryptionFlag, data)
					if err != nil {
						fmt.Println("解密失败:", err)
						return
					}
					hd, ok := pdu.handlerMap[pdu.functionCode]
					if !ok {
						fmt.Println("未注册功能码:", pdu.functionCode)
						return
					}
					err = hd.Handle(data)
					if err != nil {
						fmt.Println("处理数据失败:", err)
						return
					}
				}
			}
			fmt.Printf("[第%v个数据单元解析完成]\n", pdu.counts)
			fmt.Println()
			pdu.counts++
		}
	}
}

func (pdu *ProtocolDataUnit) Handle1(ctx context.Context, conn net.Conn) {
	pdu.conn = conn
	for {
		select {
		case <-ctx.Done():
			//停止读取
			return
		default:
			fmt.Printf("[第%v个数据单元解析开始]\n", pdu.counts)
			alldata := make([][]byte, 0, len(pdu.elements))
			//第一遍遍历elements, 读取一个完整的数据单元
			for _, element := range pdu.elements {
				//定义好合适长度的buf,接收该元素数据
				var buf = make([]byte, element.Length())
				_, err := io.ReadFull(pdu.conn, buf)
				if err != nil {
					fmt.Println("读取数据失败:", err)
					return
				}
				alldata = append(alldata, buf)
			}
			fmt.Printf("[第%v个数据单元解析完成]\n", pdu.counts)
			fmt.Println()
			pdu.counts++
		}
	}
}
