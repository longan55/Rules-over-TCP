package protocol

import "encoding/binary"

//解析器	-> 整数数据类型
//			-> 浮点数数据类型
//			-> 字符串数据类型
//			-> 比特位数组数据类型
//解析器包含字节序，包含多个数据类型
//数据类型包含 数据长度、位数

//编码方式：BIN、BCD、ASCII
//数据类型：整数、浮点数、字符串、比特位数组
//操作：设置长度、设置倍数、获取值、编码值

type ParserConfig struct {
	parserMap map[FunctionCode]Parser
}

func NewParserConfig() *ParserConfig {
	return &ParserConfig{
		parserMap: make(map[FunctionCode]Parser),
	}
}

func (pc *ParserConfig) AddParser(fc FunctionCode, parser Parser) {
	pc.parserMap[fc] = parser
}

type Parser func(data []byte) (map[string]any, error)

func NewParser() Parser {
	return func(data []byte) (map[string]any, error) {
		return nil, nil
	}
}

type dataType byte

// const (
// 	DataTypeInteger dataType = 0x01
// 	DataTypeFloat   dataType = 0x02
// 	DataTypeString  dataType = 0x03
// 	DataTypeBit     dataType = 0x04
// )

type Integer struct {
	length   int
	multiple int64
}

func (integer *Integer) SetLength(length int) {
	integer.length = length
}

func (integer *Integer) Multiply(multiple int64) {
	integer.multiple = multiple
}

func (integer *Integer) Value(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

type Float struct {
	length   int
	multiple int64
}

func (float *Float) SetLength(length int) {
	float.length = length
}

func (float *Float) Multiply(multiple int64) {
	float.multiple = multiple
}

func (float *Float) Value(data []byte) float64 {
	return 0
}

type Easier struct {
	DefaultEndian binary.ByteOrder
	Parser        EasyParser
	Encoder       EasyEncoder
}

type EasyParser struct {
	dataType    dataType
	finalEndian binary.ByteOrder
}

func (ep *EasyParser) SetEndian(endian binary.ByteOrder) {
	ep.finalEndian = endian
}

// func (ep *EasyParser) Integer() {
// 	ep.dataType = DataTypeInteger
// }

// func (ep *EasyParser) Float() {
// 	ep.dataType = DataTypeFloat
// }

// func (ep *EasyParser) String() {
// 	ep.dataType = DataTypeString
// }

// func (ep *EasyParser) Bit() {
// 	ep.dataType = DataTypeBit
// }

type EasyEncoder struct {
}

type EncodeType interface {
	Integer() (int64, error)
	Float() (float64, error)
	String() (string, error)
	BitMap(map[int]any)
	Bit() ([]any, error)
}

type BIN struct {
	bitMap map[int]any
}

func (bin *BIN) Integer() (int64, error) {
	return 0, nil
}

func (bin *BIN) Float() (float64, error) {
	return 0, nil
}

func (bin *BIN) String() (string, error) {
	return "", nil
}

func (bin *BIN) BitMap(bitMap map[int]any) {
	bin.bitMap = bitMap
}

func (bin *BIN) Bit() ([]any, error) {
	return nil, nil
}
