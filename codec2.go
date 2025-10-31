package rot

import (
	"encoding/binary"
	"encoding/hex"
)

type Codec interface {
	Encode(data any, byteLength int) []byte
	Decode(data []byte) (any, error)
}

type CodecBIN struct {
	order binary.ByteOrder
}

var _ Codec = (*CodecBIN)(nil)

func (c *CodecBIN) Encode(data any, byteLength int) []byte {
	return Int2Bin(data.(int), byte(byteLength), c.order)
}
func (c *CodecBIN) Decode(data []byte) (any, error) {
	src := Bin2Int(data, c.order)
	return src, nil
}

type CodecBCD struct {
	order binary.ByteOrder
}

var _ Codec = (*CodecBCD)(nil)

func (c *CodecBCD) Encode(data any, byteLength int) []byte {
	bcdBytes, _ := hex.DecodeString(data.(string))
	return bcdBytes
}

func (c *CodecBCD) Decode(data []byte) (any, error) {
	// 根据字节序处理数据
	if c.order == binary.LittleEndian && len(data) > 1 {
		// 小端序需要反转字节顺序
		reversed := make([]byte, len(data))
		for i := range len(reversed) {
			reversed[i] = data[len(data)-1-i]
		}
		return hex.EncodeToString(reversed), nil
	} else {
		// 大端序（默认）直接使用原数据
		return hex.EncodeToString(data), nil
	}
}
