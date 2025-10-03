package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// 编码方式常量
type EncodingType string

const (
	EncodingBIN   EncodingType = "BIN"
	EncodingBCD   EncodingType = "BCD"
	EncodingASCII EncodingType = "ASCII"
	EncodingHEX   EncodingType = "HEX"
)

// 数据类型常量
type DataType string

const (
	DataTypeInt    DataType = "INT"
	DataTypeFloat  DataType = "FLOAT"
	DataTypeString DataType = "STRING"
	DataTypeBit    DataType = "BIT"
)

// CodecItem 单个编解码项的接口
type CodecItem interface {
	// 设置数据长度
	SetLength(length int)
	// 设置倍数（用于浮点数转换）
	SetMultiple(multiple float64)
	// 从字节切片解码值
	Decode(data []byte) (any, error)
	// 将Go数据类型编码为字节切片
	Encode(value any) ([]byte, error)
	// 获取数据长度
	GetLength() int
}

// Codec 编解码器接口
type Codec interface {
	// 添加一个编解码项
	AddItem(name string, item CodecItem)
	// 从字节切片解码多个值
	Decode(data []byte) (map[string]any, error)
	// 将多个Go数据类型编码为字节切片
	Encode(values map[string]any) ([]byte, error)
}

// BaseCodecItem 基础编解码项实现
type BaseCodecItem struct {
	Length   int
	Multiple float64
	Encoding EncodingType
	DataType DataType
	Endian   binary.ByteOrder
}

// SetLength 设置数据长度
func (b *BaseCodecItem) SetLength(length int) {
	b.Length = length
}

// SetMultiple 设置倍数
func (b *BaseCodecItem) SetMultiple(multiple float64) {
	b.Multiple = multiple
}

// GetLength 获取数据长度
func (b *BaseCodecItem) GetLength() int {
	return b.Length
}

// IntCodecItem 整数编解码项
type IntCodecItem struct {
	BaseCodecItem
}

// NewIntCodecItem 创建整数编解码项
func NewIntCodecItem(byteLength int, encoding EncodingType, endian binary.ByteOrder) *IntCodecItem {
	return &IntCodecItem{
		BaseCodecItem: BaseCodecItem{
			Length:   byteLength, // 默认2字节
			Multiple: 1.0,        // 默认倍数1
			Encoding: encoding,
			DataType: DataTypeInt,
			Endian:   endian,
		},
	}
}

// Decode 解码整数
func (i *IntCodecItem) Decode(data []byte) (any, error) {
	if len(data) < i.Length {
		return nil, errors.New("数据长度不足")
	}

	switch i.Encoding {
	case EncodingBIN:
		val, err := BIN2Uint64(data[:i.Length], i.Endian)
		if err != nil {
			return nil, err
		}
		return int64(val), nil
	case EncodingBCD:
		val := BCD2Int(data[:i.Length])
		return int64(val), nil
	case EncodingASCII:
		valStr := string(data[:i.Length])
		val, err := strconv.ParseInt(strings.TrimSpace(valStr), 10, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	case EncodingHEX:
		valStr := strings.ToUpper(string(data[:i.Length]))
		val, err := strconv.ParseInt(valStr, 16, 64)
		if err != nil {
			return nil, err
		}
		return val, nil
	default:
		return nil, errors.New("不支持的编码方式")
	}
}

// Encode 编码整数
func (i *IntCodecItem) Encode(value any) ([]byte, error) {
	var num int64

	switch v := value.(type) {
	case int:
		num = int64(v)
	case int64:
		num = v
	case float64:
		num = int64(v)
	default:
		return nil, errors.New("无效的整数值类型")
	}

	switch i.Encoding {
	case EncodingBIN:
		return Int2Bin(num, byte(i.Length), i.Endian), nil
	case EncodingBCD:
		return UInt2Bcd2Bytes(num, i.Endian == binary.LittleEndian), nil
	case EncodingASCII:
		format := "%0" + strconv.Itoa(i.Length) + "d"
		str := fmt.Sprintf(format, num)
		if len(str) > i.Length {
			return nil, errors.New("数字长度超过指定长度")
		}
		return []byte(str), nil
	case EncodingHEX:
		format := "%0" + strconv.Itoa(i.Length) + "X"
		str := fmt.Sprintf(format, num)
		if len(str) > i.Length {
			return nil, errors.New("数字长度超过指定长度")
		}
		return []byte(str), nil
	default:
		return nil, errors.New("不支持的编码方式")
	}
}

// FloatCodecItem 浮点数编解码项
type FloatCodecItem struct {
	BaseCodecItem
	DecimalPlaces int // 小数位数
}

// NewFloatCodecItem 创建浮点数编解码项
func NewFloatCodecItem(encoding EncodingType, endian binary.ByteOrder) *FloatCodecItem {
	return &FloatCodecItem{
		BaseCodecItem: BaseCodecItem{
			Length:   4,   // 默认4字节
			Multiple: 1.0, // 默认倍数1
			Encoding: encoding,
			DataType: DataTypeFloat,
			Endian:   endian,
		},
		DecimalPlaces: 2, // 默认2位小数
	}
}

// SetDecimalPlaces 设置小数位数
func (f *FloatCodecItem) SetDecimalPlaces(places int) {
	f.DecimalPlaces = places
}

// Decode 解码浮点数
func (f *FloatCodecItem) Decode(data []byte) (interface{}, error) {
	if len(data) < f.Length {
		return nil, errors.New("数据长度不足")
	}

	// 先解码为整数，再根据倍数和小数位数转换为浮点数
	switch f.Encoding {
	case EncodingBIN:
		val, err := BIN2Uint64(data[:f.Length], f.Endian)
		if err != nil {
			return nil, err
		}
		return float64(val) * f.Multiple / math.Pow10(f.DecimalPlaces), nil
	case EncodingBCD:
		val := BCD2Int(data[:f.Length])
		return float64(val) * f.Multiple / math.Pow10(f.DecimalPlaces), nil
	case EncodingASCII:
		valStr := string(data[:f.Length])
		val, err := strconv.ParseFloat(strings.TrimSpace(valStr), 64)
		if err != nil {
			return nil, err
		}
		return val * f.Multiple, nil
	case EncodingHEX:
		valStr := strings.ToUpper(string(data[:f.Length]))
		val, err := strconv.ParseInt(valStr, 16, 64)
		if err != nil {
			return nil, err
		}
		return float64(val) * f.Multiple / math.Pow10(f.DecimalPlaces), nil
	default:
		return nil, errors.New("不支持的编码方式")
	}
}

// Encode 编码浮点数
func (f *FloatCodecItem) Encode(value interface{}) ([]byte, error) {
	var num float64

	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	default:
		return nil, errors.New("无效的浮点数值类型")
	}

	// 根据倍数和小数位数转换为整数
	intValue := int64(num / f.Multiple * math.Pow10(f.DecimalPlaces))

	switch f.Encoding {
	case EncodingBIN:
		return Int2Bin(intValue, byte(f.Length), f.Endian), nil
	case EncodingBCD:
		return UInt2Bcd2Bytes(intValue, f.Endian == binary.LittleEndian), nil
	case EncodingASCII:
		format := "%0" + strconv.Itoa(f.Length) + "." + strconv.Itoa(f.DecimalPlaces) + "f"
		str := fmt.Sprintf(format, num)
		if len(str) > f.Length {
			return nil, errors.New("浮点数长度超过指定长度")
		}
		return []byte(str), nil
	case EncodingHEX:
		return nil, errors.New("浮点数不支持HEX编码")
	default:
		return nil, errors.New("不支持的编码方式")
	}
}

// StringCodecItem 字符串编解码项
type StringCodecItem struct {
	BaseCodecItem
}

// NewStringCodecItem 创建字符串编解码项
func NewStringCodecItem(encoding EncodingType) *StringCodecItem {
	return &StringCodecItem{
		BaseCodecItem: BaseCodecItem{
			Length:   16, // 默认16字节
			Multiple: 1.0,
			Encoding: encoding,
			DataType: DataTypeString,
		},
	}
}

// Decode 解码字符串
func (s *StringCodecItem) Decode(data []byte) (interface{}, error) {
	if len(data) < s.Length {
		return nil, errors.New("数据长度不足")
	}

	switch s.Encoding {
	case EncodingASCII:
		return strings.TrimSpace(string(data[:s.Length])), nil
	case EncodingHEX:
		return strings.ToUpper(string(data[:s.Length])), nil
	default:
		return nil, errors.New("字符串只支持ASCII和HEX编码")
	}
}

// Encode 编码字符串
func (s *StringCodecItem) Encode(value interface{}) ([]byte, error) {
	str, ok := value.(string)
	if !ok {
		return nil, errors.New("无效的字符串类型")
	}

	switch s.Encoding {
	case EncodingASCII:
		// 填充或截断到指定长度
		if len(str) > s.Length {
			str = str[:s.Length]
		} else if len(str) < s.Length {
			str = str + strings.Repeat(" ", s.Length-len(str))
		}
		return []byte(str), nil
	case EncodingHEX:
		// 验证HEX字符串
		if len(str)%2 != 0 {
			str = "0" + str
		}
		if len(str) > s.Length {
			return nil, errors.New("HEX字符串长度超过指定长度")
		}
		return []byte(str), nil
	default:
		return nil, errors.New("字符串只支持ASCII和HEX编码")
	}
}

// BitCodecItem 比特位编解码项
type BitCodecItem struct {
	BaseCodecItem
}

// NewBitCodecItem 创建比特位编解码项
func NewBitCodecItem(endian binary.ByteOrder) *BitCodecItem {
	return &BitCodecItem{
		BaseCodecItem: BaseCodecItem{
			Length:   1, // 默认1字节
			Multiple: 1.0,
			Encoding: EncodingBIN, // 比特位只能是BIN编码
			DataType: DataTypeBit,
			Endian:   endian,
		},
	}
}

// Decode 解码比特位
func (b *BitCodecItem) Decode(data []byte) (interface{}, error) {
	if len(data) < b.Length {
		return nil, errors.New("数据长度不足")
	}

	bits := make([]bool, b.Length*8)
	for i := 0; i < b.Length; i++ {
		for j := 0; j < 8; j++ {
			if b.Endian == binary.LittleEndian {
				bits[i*8+j] = (data[i] & (1 << uint(j))) != 0
			} else {
				bits[i*8+j] = (data[i] & (1 << uint(7-j))) != 0
			}
		}
	}

	return bits, nil
}

// Encode 编码比特位
func (b *BitCodecItem) Encode(value interface{}) ([]byte, error) {
	bits, ok := value.([]bool)
	if !ok {
		return nil, errors.New("无效的比特位类型")
	}

	// 确保比特位数不超过长度*8
	maxBits := b.Length * 8
	if len(bits) > maxBits {
		bits = bits[:maxBits]
	}

	data := make([]byte, b.Length)
	for i := 0; i < len(bits); i++ {
		byteIndex := i / 8
		bitIndex := i % 8
		if b.Endian == binary.LittleEndian {
			if bits[i] {
				data[byteIndex] |= (1 << uint(bitIndex))
			}
		} else {
			if bits[i] {
				data[byteIndex] |= (1 << uint(7-bitIndex))
			}
		}
	}

	return data, nil
}

// MultiCodec 多值编解码器实现
type MultiCodec struct {
	items []struct {
		name string
		item CodecItem
	}
	totalLength int
}

// NewMultiCodec 创建多值编解码器
func NewMultiCodec() *MultiCodec {
	return &MultiCodec{
		items: make([]struct {
			name string
			item CodecItem
		}, 0),
		totalLength: 0,
	}
}

// AddItem 添加编解码项
func (m *MultiCodec) AddItem(name string, item CodecItem) {
	m.items = append(m.items, struct {
		name string
		item CodecItem
	}{
		name: name,
		item: item,
	})
	m.totalLength += item.GetLength()
}

// Decode 解码多个值
func (m *MultiCodec) Decode(data []byte) (map[string]interface{}, error) {
	if len(data) < m.totalLength {
		return nil, errors.New("数据长度不足")
	}

	result := make(map[string]interface{})
	offset := 0

	for _, item := range m.items {
		value, err := item.item.Decode(data[offset:])
		if err != nil {
			return nil, errors.New("解码" + item.name + "时出错: " + err.Error())
		}
		result[item.name] = value
		offset += item.item.GetLength()
	}

	return result, nil
}

// Encode 编码多个值
func (m *MultiCodec) Encode(values map[string]interface{}) ([]byte, error) {
	result := make([]byte, 0, m.totalLength)

	for _, item := range m.items {
		value, exists := values[item.name]
		if !exists {
			return nil, errors.New("缺少" + item.name + "的值")
		}

		data, err := item.item.Encode(value)
		if err != nil {
			return nil, errors.New("编码" + item.name + "时出错: " + err.Error())
		}

		result = append(result, data...)
	}

	return result, nil
}
