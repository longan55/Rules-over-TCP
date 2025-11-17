package rot

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
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
	//获取元素的实际值(不包含默认值)
	RealValue() any
	//设置元素的实际值(不包含默认值)
	SetRealValue(value any)
	//获取元素的默认值
	DefaultValue() []byte
	//获取元素自身占用的字节长度
	Length() int
	//获取元素的字节序
	GetOrder() binary.ByteOrder
	//预处理: 对fullData进行预处理，返回处理后的fullData
	Preprocess(element ProtocolElement, fullData [][]byte) error
	//fullData按元素切割整个数据单元，分成多个切片，pdu提供对ProtocolDataUnit的访问
	Deal(fullData [][]byte, pdu ProtocolDataUnitAccessor) error
	//获取校验和类型
	ChecksumType() uint8
}

// NewStarter 起始符示例的DealFunc更新
// 其他元素创建函数的DealFunc也需要类似更新，此处略过具体实现

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
}

// DealFunction 处理函数
type DealFunction func(element ProtocolElement, data [][]byte, pdu ProtocolDataUnitAccessor) error

var _ ProtocolElement = (*ProtocolElementImpl)(nil)

// ProtocolElementImpl 基础元素结构体
type ProtocolElementImpl struct {
	//元数据: 存储该元素的元数据(用于描述说明)
	index        int                 //说明该元素的索引
	Typ          ProtocolElementType //元素类型
	name         string              //元素名字
	selfLength   int                 //元素本身长度
	defaultValue []byte              //默认值
	realValue    any                 //元素的实际值(不包含默认值)
	order        binary.ByteOrder    //大小端
	start        uint8               //开始索引: 该元素影响的元素区域的第一个元素索引
	end          uint8               //结束索引: 该元素影响的元素区域的最后一个元素索引
	//TODO: DealFunc可简化为func(element ProtocolElement, data [][]byte)error。。返回Type貌似是必须的。
	DealFunc     DealFunction //处理函数
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

func (f *ProtocolElementImpl) RealValue() any {
	return f.realValue
}

func (f *ProtocolElementImpl) SetRealValue(value any) {
	f.realValue = value
}

func (f *ProtocolElementImpl) DefaultValue() []byte {
	return f.defaultValue
}

//	func (f *ProtocolElementImpl) SetLen(l int) {
//		f.selfLength = l
//	}
func (f *ProtocolElementImpl) Length() int {
	return f.selfLength
}

func (f *ProtocolElementImpl) GetOrder() binary.ByteOrder {
	return f.order
}

func (f *ProtocolElementImpl) GetRange() (start, end uint8) {
	return f.start, f.end
}

func (f *ProtocolElementImpl) Preprocess(element ProtocolElement, fullData [][]byte) error {
	return nil
}

func (f *ProtocolElementImpl) Deal(data [][]byte, pdu ProtocolDataUnitAccessor) error {
	return f.DealFunc(f, data, pdu)
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
	element.DealFunc = func(element ProtocolElement, fullData [][]byte, pdu ProtocolDataUnitAccessor) error {
		if len(fullData) == 0 {
			return errors.New("数据为空")
		}
		if !bytes.Equal(fullData[0][:element.Length()], element.DefaultValue()) {
			return fmt.Errorf("起始符错误Need:%0X,But:%0X", element.DefaultValue(), fullData[0][:element.Length()])
		}
		fmt.Printf("起始符:\t\t\t[%#0X]\n", fullData[0][:element.Length()])
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
	element.DealFunc = func(element ProtocolElement, fullData [][]byte, pdu ProtocolDataUnitAccessor) error {
		if fullData == nil {
			return errors.New("数据为空")
		}
		data := fullData[element.GetIndex()]
		length := Bin2Int(data, element.GetOrder())
		fmt.Printf("帧长度:\t\t\t[%d]\n", length)
		element.SetRealValue(length)
		return nil
	}
	return element
}

// 序 列 号, 接收方返回相同的序列号，用于确认接收方是否成功接收数据
func NewSerialNumber() ProtocolElement {
	element := &ProtocolElementImpl{
		Typ:        SerialNumber,
		name:       "序 列 号 ",
		selfLength: 2,
	}
	element.DealFunc = func(element ProtocolElement, fullData [][]byte, pdu ProtocolDataUnitAccessor) error {
		if fullData == nil {
			return errors.New("数据为空")
		}
		serialdata := fullData[element.GetIndex()]
		sn := Bin2Int(serialdata, element.GetOrder())
		//实际类型为int16的any可以强转为int吗？
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
		fmt.Printf("%s:\t\t[%#0X]\n", element.GetName(), fullData[element.GetIndex()])
		return nil
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
	element.DealFunc = func(element ProtocolElement, fullData [][]byte, pdu ProtocolDataUnitAccessor) error {
		if fullData == nil {
			return errors.New("数据为空")
		}
		flagdata := fullData[element.GetIndex()]
		flag := Bin2Int(flagdata, element.GetOrder())
		fmt.Printf("加密标识:\t\t[%#0X]\n", fullData[element.GetIndex()])
		element.SetRealValue(flag)

		// 示例：通过pdu访问其他元素
		// 例如，如果需要在设置加密标志时检查其他元素
		payloadElement := pdu.GetElementByType(Payload)
		if payloadElement != nil {
			fmt.Printf("检测到存在Payload元素，加密标志将用于解密操作\n")
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
	element.DealFunc = func(element ProtocolElement, fullData [][]byte, pdu ProtocolDataUnitAccessor) error {
		if fullData == nil {
			return errors.New("数据为空")
		}
		fmt.Printf("功能码:\t\t\t[%#0X]\n", fullData[element.GetIndex()])
		functionCode := FunctionCode(fullData[element.GetIndex()][0])
		element.SetRealValue(functionCode)
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
	element.DealFunc = func(element ProtocolElement, fullData [][]byte, pdu ProtocolDataUnitAccessor) error {
		if fullData == nil {
			return errors.New("数据为空")
		}

		// 示例：通过pdu获取Length元素的值作为payload长度
		lengthElement := pdu.GetElementByType(Length)
		if lengthElement != nil {
			lengthValue, ok := lengthElement.RealValue().(int)
			if ok && lengthValue > 0 {
				fmt.Printf("使用Length元素的值[%d]作为负载长度参考\n", lengthValue)
			}
		}

		fmt.Printf("帧负载:\t\t\t[% #0X]\n", fullData[element.GetIndex()])
		element.SetRealValue(fullData[element.GetIndex()])
		return nil
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
	element.DealFunc = func(element ProtocolElement, fullData [][]byte, pdu ProtocolDataUnitAccessor) error {
		if fullData == nil {
			return errors.New("数据为空")
		}
		fmt.Printf("校 验 码:\t\t[% #0X]\n", fullData[element.GetIndex()])
		checksum0 := fullData[element.GetIndex()]
		//将各切片连接为一个切片
		full := bytes.Join(fullData[2:element.GetIndex()], nil)
		checksum := CheckSum(element.ChecksumType(), full)
		if !bytes.Equal(checksum, checksum0) {
			return fmt.Errorf("校验码错误Need:%0X,But:%0X", checksum, checksum0)
		}
		fmt.Printf("校验码类型:%d,计算校验码:% #0X,校验通过\n", element.ChecksumType(), checksum)
		return nil
	}
	return element
}
