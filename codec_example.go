package main

import (
	"encoding/binary"
	"fmt"

	"github.com/longan55/proto/protocol"
)

func Test() {
	// 示例1: 基本整数编解码
	fmt.Println("=== 示例1: 基本整数编解码 ===")
	integerCodec := protocol.NewIntCodecItem(2, protocol.EncodingBIN, binary.BigEndian)

	// 编码
	encodedInt, err := integerCodec.Encode(1234)
	if err != nil {
		fmt.Println("整数编码错误:", err)
	} else {
		fmt.Printf("整数1234编码为: %#v\n", encodedInt)
	}

	// 解码

	decodedInt, err := integerCodec.Decode(encodedInt)
	if err != nil {
		fmt.Println("整数解码错误:", err)
	} else {
		fmt.Printf("解码结果: %v\n", decodedInt)
	}

	// 示例2: 浮点数编解码（带倍数和小数位）
	fmt.Println("\n=== 示例2: 浮点数编解码 ===")
	floatCodec := protocol.NewFloatCodecItem(protocol.EncodingBIN, binary.BigEndian)
	floatCodec.SetLength(4)
	floatCodec.SetMultiple(1.0)
	floatCodec.SetDecimalPlaces(2)

	// 编码
	encodedFloat, err := floatCodec.Encode(123.45)
	if err != nil {
		fmt.Println("浮点数编码错误:", err)
	} else {
		fmt.Printf("浮点数123.45编码为: %#v\n", encodedFloat)
	}

	// 解码
	decodedFloat, err := floatCodec.Decode(encodedFloat)
	if err != nil {
		fmt.Println("浮点数解码错误:", err)
	} else {
		fmt.Printf("解码结果: %v\n", decodedFloat)
	}

	// 示例3: 字符串编解码
	fmt.Println("\n=== 示例3: 字符串编解码 ===")
	stringCodec := protocol.NewStringCodecItem(protocol.EncodingASCII)
	stringCodec.SetLength(10)

	// 编码
	encodedStr, err := stringCodec.Encode("Hello")
	if err != nil {
		fmt.Println("字符串编码错误:", err)
	} else {
		fmt.Printf("字符串'Hello'编码为: %#v\n", encodedStr)
		fmt.Printf("编码后的字符串表示: '%s'\n", string(encodedStr))
	}

	// 解码
	decodedStr, err := stringCodec.Decode(encodedStr)
	if err != nil {
		fmt.Println("字符串解码错误:", err)
	} else {
		fmt.Printf("解码结果: '%v'\n", decodedStr)
	}

	// 示例4: 比特位编解码
	fmt.Println("\n=== 示例4: 比特位编解码 ===")
	bitCodec := protocol.NewBitCodecItem(binary.LittleEndian)
	bitCodec.SetLength(1)

	// 创建比特位数组
	bits := []bool{true, false, true, false, false, true, true, true}

	// 编码
	encodedBits, err := bitCodec.Encode(bits)
	if err != nil {
		fmt.Println("比特位编码错误:", err)
	} else {
		fmt.Printf("比特位%v编码为: %#v\n", bits, encodedBits)
	}

	// 解码
	decodedBits, err := bitCodec.Decode(encodedBits)
	if err != nil {
		fmt.Println("比特位解码错误:", err)
	} else {
		fmt.Printf("解码结果: %v\n", decodedBits)
	}

	// 示例5: 多值编解码
	fmt.Println("\n=== 示例5: 多值编解码 ===")
	multiCodec := protocol.NewMultiCodec()

	// 添加各种类型的编解码项
	deviceIDCodec := protocol.NewIntCodecItem(2, protocol.EncodingBCD, binary.BigEndian)

	temperatureCodec := protocol.NewFloatCodecItem(protocol.EncodingBIN, binary.BigEndian)
	temperatureCodec.SetLength(2)
	temperatureCodec.SetDecimalPlaces(1)
	temperatureCodec.SetMultiple(1.0)

	statusCodec := protocol.NewBitCodecItem(binary.LittleEndian)
	statusCodec.SetLength(1)

	multiCodec.AddItem("deviceID", deviceIDCodec)
	multiCodec.AddItem("temperature", temperatureCodec)
	multiCodec.AddItem("status", statusCodec)

	// 创建要编码的值
	values := map[string]interface{}{
		"deviceID":    int64(1234),
		"temperature": float64(25.5),
		"status":      []bool{true, true, false, false, true, false, false, false},
	}

	// 编码多个值
	encodedData, err := multiCodec.Encode(values)
	if err != nil {
		fmt.Println("多值编码错误:", err)
	} else {
		fmt.Printf("多值编码结果: %#v\n", encodedData)
	}

	// 解码多个值
	decodedValues, err := multiCodec.Decode(encodedData)
	if err != nil {
		fmt.Println("多值解码错误:", err)
	} else {
		fmt.Println("多值解码结果:")
		for k, v := range decodedValues {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}

	// 示例6: 使用不同的编码方式
	fmt.Println("\n=== 示例6: 使用不同的编码方式 ===")
	// ASCII编码的整数
	asciiIntCodec := protocol.NewIntCodecItem(4, protocol.EncodingASCII, binary.BigEndian)

	encodedASCII, err := asciiIntCodec.Encode(5678)
	if err != nil {
		fmt.Println("ASCII整数编码错误:", err)
	} else {
		fmt.Printf("整数5678的ASCII编码: %#v\n", encodedASCII)
		fmt.Printf("ASCII编码的字符串表示: '%s'\n", string(encodedASCII))

		// 解码
		decodedASCII, err := asciiIntCodec.Decode(encodedASCII)
		if err != nil {
			fmt.Println("ASCII整数解码错误:", err)
		} else {
			fmt.Printf("ASCII解码结果: %v\n", decodedASCII)
		}
	}
}
