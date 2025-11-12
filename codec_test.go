package rot

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/longan55/Rules-over-TCP/fake"
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

func TestCodec2_Parse2(t *testing.T) {
	//处理器构建器
	builder := NewProtocolBuilder()
	//添加加密配置
	cryptConfig := NewCryptConfig()
	cryptConfig.AddCrypt(0, &CryptNothing{})
	builder.AddCryptConfig(cryptConfig)

	//配置处理器，来数据时自动处理
	setHandlerConfig(builder)

	//构建处理器
	builder.AddElement(NewStarter([]byte{0x68})).
		AddElement(NewDataLen(1)).
		AddElement(NewCyptoFlag()).
		AddElement(NewFuncCode()).
		AddElement(NewPayload()).
		AddElement(NewCheckSum(0, 2))
	dataHander, err := builder.Build()
	if err != nil {
		fmt.Println("构建处理器失败:", err)
		return
	}

	// 创建FakeConn并设置测试数据
	fakeConn := fake.NewFakeConn()
	fakeConn.SetData([]byte{0x68, 0x0F, 0x00, 0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x01, 0x1a, 0x40})
	// fakeConn.SetData([]byte{0x68, 0x0E, 0x00, 0x02, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0xbd, 0x09})
	// fakeConn.SetData([]byte{0x68, 0x06, 0x00, 0x03, 0x30, 0x31, 0x32, 0x33, 0x4f, 0xa1})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用fake conn运行Handle方法
	go dataHander.Handle(ctx, fakeConn)

	// 给Handle方法一些时间处理数据
	time.Sleep(100 * time.Millisecond)
}

func setHandlerConfig(builder *ProtocolBuilder) {
	handlerConfig := NewHandlerConfig()
	//1. BIN编码,默认解释为整数，强烈建议只解释为整数或浮点数，不要解释为字符串。
	fh := new(FunctionHandler)
	fh.AddField("a", WithBin(), WithLength(4), WithBinInteger(true, 1, 0))
	fh.AddField("b", WithBin(), WithLength(4), WithBinInteger(true, 1, 0))
	fh.AddField("c", WithBin(), WithLength(2), WithBinInteger(true, 2, 0))
	fh.AddField("d", WithBin(), WithLength(2), WithBinFloat(true, 0.01, 0))
	fh.AddField("e", WithBin(), WithLength(1), WithBinInteger(true, 1, 0), WithEnum("Other", map[int]any{0: "A", 1: "B", 2: "C"}))

	fh.SetHandle(func(parsedData map[string]ParsedData) error {
		fmt.Println("parsedData:", parsedData)
		return nil
	})
	// //2. BCD编码，默认解释为字符串，还可以解释为整数或浮点数，浮点数较为常见（在需要高精度传输时）
	// fh1 := new(FunctionHandler)
	// fh1.NewDecoder("code", binary.BigEndian).BCD().SetByteLength(4).String()
	// fh1.NewDecoder("price", binary.BigEndian).BCD().SetByteLength(4).Float().DecimalPlace(4)
	// fh1.NewDecoder("intPrice", binary.BigEndian).BCD().SetByteLength(4).Integer()

	// fh1.SetHandle(func(parsedData map[string]ParsedData) error {
	// 	fmt.Println("parsedData:", parsedData)
	// 	return nil
	// })
	// //3. ASCII编码，仅解释为字符串
	// fh2 := new(FunctionHandler)
	// fh2.NewDecoder("ascii", binary.BigEndian).ASCII().SetByteLength(4).String()
	// fh2.SetHandle(func(parsedData map[string]ParsedData) error {
	// 	fmt.Println("parsedData:", parsedData)
	// 	return nil
	// })
	//end
	handlerConfig.AddHandler(FunctionCode(0x01), fh)
	// handlerConfig.AddHandler(FunctionCode(0x02), fh1)
	// handlerConfig.AddHandler(FunctionCode(0x03), fh2)
	builder.AddHandlerConfig(handlerConfig)
}
