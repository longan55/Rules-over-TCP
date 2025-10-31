package rot

import "encoding/binary"

func NewEncoder(fieldName string, order binary.ByteOrder) *EncoderImpl {
	encoder := &EncoderImpl{order: order}
	return encoder
}

type EncoderImpl struct {
	byteLength int
	fh         *FunctionHandler
	order      binary.ByteOrder
}

func (encoder *EncoderImpl) BIN() *EncodeBIN {
	temp := &EncodeBIN{
		encoder: encoder,
		order:   encoder.order,
	}
	return temp
}

type EncodeBIN struct {
	encoder *EncoderImpl
	order   binary.ByteOrder
}

func (bin *EncodeBIN) SetByteLength(byteLength int) *EncodeBIN {
	bin.encoder.byteLength = byteLength
	return bin
}
func (bin *EncodeBIN) Encode(data any) []byte {
	return nil
}

func (e *EncodeBIN) Integer() *EncodeInteger {
	return &EncodeInteger{
		encoder: e.encoder,
		order:   e.order,
	}
}
func (bin *EncodeBIN) Float() *EncodeBIN {
	return bin
}

type EncodeInteger struct {
	encoder *EncoderImpl
	order   binary.ByteOrder
}
