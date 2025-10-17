package protocol

import (
	"encoding/binary"
	"testing"
)

func TestCode1(t *testing.T) {
	encoder := NewEncoder(binary.LittleEndian)
	i := encoder.BIN().SetByteLength(4).Integer().Multiple(2).Offset(100)
	i.SourceValue()
}
