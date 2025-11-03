package rot

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestCodec2(t *testing.T) {
	fc1 := NewFieldCodecConfig("1", WithMode(ModeDecode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithBinInteger(true, 1, 0))
	fc2 := NewFieldCodecConfig("2", WithMode(ModeDecode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithBinInteger(true, 1, -10))
	fc3 := NewFieldCodecConfig("3", WithMode(ModeDecode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithBinInteger(true, 2, -10))
	fc4 := NewFieldCodecConfig("4", WithMode(ModeDecode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithBinInteger(false, 2, -10))
	fc41 := NewFieldCodecConfig("41", WithMode(ModeDecode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithBinInteger(true, 1, 0), WithEnum("Other", map[int]any{0: "A", 1: "B", 2: "C"}))

	fc5 := NewFieldCodecConfig("5", WithMode(ModeDecode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithBinFloat(true, 0.01, 0))
	fc6 := NewFieldCodecConfig("6", WithMode(ModeDecode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithBinFloat(true, 0.01, 0.1))
	fc7 := NewFieldCodecConfig("7", WithMode(ModeDecode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithBinFloat(false, 0.01, 1))

	fcs := []*FieldCodecConfig{fc1, fc2, fc3, fc4, fc41, fc5, fc6, fc7}
	ds := [][]byte{
		{0x12, 0x34},
		{0x12, 0x34},
		{0x12, 0x34},
		{0x12, 0x34},
		{0x01},
		{0x12, 0x34},
		{0x12, 0x34},
		{0x12, 0x34},
	}
	for i, fc := range fcs {
		a, err := fc.Decode(ds[i])
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%v: %v\n", fc.name, a)
	}
}

func TestCodec2_Encode(t *testing.T) {
	fc1 := NewFieldCodecConfig("1", WithMode(ModeEncode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithLength(2), WithBinInteger(true, 1, 0))
	fc2 := NewFieldCodecConfig("2", WithMode(ModeEncode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithLength(2), WithBinInteger(true, 1, -10))
	fc3 := NewFieldCodecConfig("3", WithMode(ModeEncode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithLength(2), WithBinInteger(true, 2, -10))
	fc4 := NewFieldCodecConfig("4", WithMode(ModeEncode), WithCodec(&CodecBIN{order: binary.BigEndian}), WithLength(2), WithBinInteger(false, 2, -10))

	fcs := []*FieldCodecConfig{fc1, fc2, fc3, fc4}
	ds := []int{
		4660,
		4650,
		9310,
		9300,
	}

	for i, fc := range fcs {
		a, err := fc.Encode(ds[i])
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%v: %#x\n", fc.name, a)
	}
}
