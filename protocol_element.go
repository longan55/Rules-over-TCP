package rot

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

// ProtocolElement 元素接口
type ProtocolElement interface {
	//获取元素的索引
	GetIndex() int
	//设置元素的索引
	SetIndex(index int)
	//获取元素的名称
	GetName() string
	//获取元素的类型
	Type() ProtocolElementType
	//获取元素的字节原始数据
	Source() []byte
	//设置元素的字节原始数据
	SetSource(src []byte)
	//获取元素解析后的实际值
	RealValue() any
	//设置元素解析后的实际值
	SetRealValue(value any)
	//获取元素的默认值
	DefaultValue() []byte
	//获取元素自身占用的字节长度
	SelfLength() int
	//获取元素的字节序
	GetOrder() binary.ByteOrder
	//预处理: 从conn中读取数据并解析，将字节数据和实际值填充到element中，pdu用于访问整个元素集合或协议的元数据
	Preprocess(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error
	//pdu提供对ProtocolDataUnit的访问
	Deal(pdu ProtocolDataUnitAccessor) error
	//获取校验和类型
	ChecksumType() uint8
}

type ProtocolElementType byte

const (
	// 帧首符
	Preamble ProtocolElementType = iota
	// 当前元素之后的数据长度
	Length
	// 序 列 号
	SerialNumber
	// 加密标志
	EncryptionFlag
	// 功能码
	Function
	// 消息负载
	Payload
	// 校验和
	Checksum
)

// ProtocolDataUnitAccessor 提供对ProtocolDataUnit的访问接口
type ProtocolDataUnitAccessor interface {
	// GetElementByIndex 通过索引获取ProtocolElement
	GetElementByIndex(index int) ProtocolElement
	// GetElementByType 通过类型获取ProtocolElement
	GetElementByType(typ ProtocolElementType) ProtocolElement
	// GetAllElements 获取所有ProtocolElement
	GetAllElements() []ProtocolElement
	Decrypt(cryptFlag int, src []byte) ([]byte, error)
	DoHandle(code FunctionCode, payload []byte) error
}

type PreprocessFunction func(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error

// DealFunction 处理函数
type DealFunction func(element ProtocolElement, pdu ProtocolDataUnitAccessor) error

var _ ProtocolElement = (*ProtocolElementImpl)(nil)

// ProtocolElementImpl 基础元素结构体
type ProtocolElementImpl struct {
	//元数据: 存储该元素的元数据(用于描述说明)
	index          int                 //说明该元素的索引
	Typ            ProtocolElementType //元素类型
	name           string              //元素名字
	selfLength     int                 //元素本身长度
	defaultValue   []byte              //默认值
	src            []byte              //元素的实际值(原始字节数据)
	realValue      any                 //元素的实际值(解析后的值)
	order          binary.ByteOrder    //大小端
	start          uint8               //开始索引: 该元素影响的元素区域的第一个元素索引
	end            uint8               //结束索引: 该元素影响的元素区域的最后一个元素索引
	PreprocessFunc PreprocessFunction  //预处理函数
	DealFunc       DealFunction        //处理函数
	checksumType   uint8
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

func (f *ProtocolElementImpl) Source() []byte {
	return f.src
}

func (f *ProtocolElementImpl) SetSource(src []byte) {
	f.src = src
}

func (f *ProtocolElementImpl) RealValue() any {
	return f.realValue
}

func (f *ProtocolElementImpl) SetRealValue(value any) {
	f.realValue = value
}

func (f *ProtocolElementImpl) DefaultValue() []byte {
	return f.defaultValue
}

func (f *ProtocolElementImpl) SelfLength() int {
	return f.selfLength
}

func (f *ProtocolElementImpl) GetOrder() binary.ByteOrder {
	return f.order
}

func (f *ProtocolElementImpl) GetRange() (start, end uint8) {
	return f.start, f.end
}

func (f *ProtocolElementImpl) Preprocess(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
	return f.PreprocessFunc(conn, element, pdu)
}

func (f *ProtocolElementImpl) Deal(pdu ProtocolDataUnitAccessor) error {
	if f.DealFunc != nil {
		return f.DealFunc(f, pdu)
	}
	return nil
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
	// 起始符的预处理函数, 从conn中读取数据, 并将其添加到fullData中
	element.PreprocessFunc = func(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		buf := make([]byte, element.SelfLength())
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			fmt.Println("读取数据失败:", err)
			return err
		}
		fmt.Printf("起始符:\t\t\t[% #0X]\n", buf)
		element.SetSource(buf)
		return nil
	}
	// 起始符的处理函数, 验证起始符是否正确
	element.DealFunc = func(element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		if !bytes.Equal(element.Source(), element.DefaultValue()) {
			return fmt.Errorf("起始符错误Need:%0X,But:%0X", element.DefaultValue(), element.Source())
		}
		return nil
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
	element.PreprocessFunc = func(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		buf := make([]byte, element.SelfLength())
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			fmt.Println("读取数据失败:", err)
			return err
		}
		element.SetSource(buf)
		length := Bin2Int(buf, element.GetOrder())
		element.SetRealValue(length)
		fmt.Printf("帧长度:\t\t\t[%d]\n", length)
		return nil
	}
	return element
}

// TODO： 序 列 号, 接收方返回相同的序列号，用于确认接收方是否成功接收数据
func NewSerialNumber() ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:        SerialNumber,
		name:       "序 列 号 ",
		selfLength: 2,
	}
	element.DealFunc = func(element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		if pdu == nil {
			return errors.New("数据为空")
		}
		serialdata := pdu.GetElementByIndex(element.GetIndex()).Source()
		sn := Bin2Int(serialdata, element.GetOrder())
		oldsn, ok := element.RealValue().(int16)
		if !ok {
			return fmt.Errorf("序列号错误Need:%d,But:%d", element.RealValue(), sn)
		}
		if sn < int(oldsn) {
			return fmt.Errorf("序列号重复Latest:%d,But:%d", element.RealValue(), sn)
		}
		if sn == int(oldsn)+1 {
			element.SetRealValue(int16(sn))
		}
		return nil
	}
	return element
}

// TODO: 发送数据时，需要判断是否需要加密。
func NewCyptoFlag() ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:          EncryptionFlag,
		name:         "加密标识 ",
		defaultValue: []byte{0x01},
		selfLength:   1,
	}
	element.PreprocessFunc = func(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		buf := make([]byte, element.SelfLength())
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			fmt.Println("读取数据失败:", err)
			return err
		}
		element.SetSource(buf)
		flag := Bin2Int(buf, element.GetOrder())
		element.SetRealValue(flag)
		fmt.Printf("加密标识:\t\t[%#0X]\n", buf)
		return nil
	}
	element.DealFunc = func(element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		if pdu == nil {
			return errors.New("数据为空")
		}
		if flag, ok := element.RealValue().(int); !ok {
			return fmt.Errorf("加密标识类型错误")
		} else {
			payloadElement := pdu.GetElementByType(Payload)
			payload := payloadElement.Source()
			payload0, err := pdu.Decrypt(flag, payload)
			if err != nil {
				return err
			}
			payloadElement.SetRealValue(payload0)
		}
		return nil
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
	element.PreprocessFunc = func(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		buf := make([]byte, element.SelfLength())
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			fmt.Println("读取数据失败:", err)
			return err
		}
		element.SetSource(buf)
		functionCode := Bin2Int(buf, element.GetOrder())
		element.SetRealValue(FunctionCode(functionCode))
		fmt.Printf("功能码:\t\t\t[%#0X]\n", buf)
		return nil
	}
	return element
}

func NewPayload() ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:          Payload,
		name:         "帧 负 载",
		defaultValue: nil,
		selfLength:   -1,
	}
	//读取负载
	element.PreprocessFunc = func(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		lengthElement := pdu.GetElementByType(Length)
		if lengthElement == nil {
			return errors.New("未找到Length元素")
		}
		length, ok := lengthElement.RealValue().(int)
		if !ok {
			return errors.New("Length元素值不是整数")
		}
		buf := make([]byte, length-2)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			fmt.Println("读取数据失败:", err)
			return err
		}
		element.SetSource(buf)
		fmt.Printf("帧负载:\t\t\t[% #0X]\n", buf)
		return nil
	}
	element.DealFunc = func(element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		functionCodeElement := pdu.GetElementByType(Function)
		if functionCodeElement == nil {
			return errors.New("未找到Function元素")
		}
		functionCode, ok := functionCodeElement.RealValue().(FunctionCode)
		if !ok {
			return errors.New("Function元素值不是整数")
		}
		return pdu.DoHandle(functionCode, element.RealValue().([]byte))
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
	element.PreprocessFunc = func(conn net.Conn, element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		buf := make([]byte, element.SelfLength())
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			fmt.Println("读取数据失败:", err)
			return err
		}
		element.SetSource(buf)
		fmt.Printf("校 验 码:\t\t[% #0X]\n", buf)
		return nil
	}
	element.DealFunc = func(element ProtocolElement, pdu ProtocolDataUnitAccessor) error {
		checksum0 := element.Source()
		if len(checksum0) == 0 {
			return errors.New("校验码为空")
		}
		//将各切片连接为一个切片
		var full []byte
		for i := 2; i < element.GetIndex(); i++ {
			e := pdu.GetElementByIndex(i)
			if e == nil {
				return fmt.Errorf("未找到索引为%d的元素", i)
			}
			src := e.Source()
			if src == nil {
				return fmt.Errorf("索引为%d的元素源数据为空", i)
			}
			full = append(full, src...)
		}
		checksum := CheckSum(element.ChecksumType(), full)
		if !bytes.Equal(checksum, checksum0) {
			return fmt.Errorf("校验码错误Need:%0X,But:%0X", checksum, checksum0)
		}
		fmt.Printf("校验码类型:%d,计算校验码:% #0X,校验通过\n", element.ChecksumType(), checksum)
		return nil
	}
	return element
}
