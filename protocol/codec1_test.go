package protocol

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestCode1(t *testing.T) {
	encoder := NewEncoder(binary.BigEndian)
	i := encoder.BIN().SetByteLength(4).Integer().Multiple(1).Offset(0)
	src := i.SourceValue([]byte{0, 0, 1, 0})
	t.Logf("source value:%v\n", src)

	encoder1 := NewEncoder(binary.BigEndian)
	f := encoder1.BIN().SetByteLength(4).Float1().Multiple(math.Pow10(-4))
	t.Logf("source value:%v\n", f.SourceValue([]byte{0, 0, 0, 1}))
}

func TestCode2(t *testing.T) {
	encoder := NewEncoder(binary.BigEndian)
	i := encoder.BIN().SetByteLength(1).Integer().SetEnum(map[int]any{
		0: "A",
		1: "B",
		2: "C",
		3: "D",
	})
	explained := i.ExplainedValue([]byte{2})
	t.Logf("explained value:%v\n", explained)
}
