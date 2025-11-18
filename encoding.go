/*
* Copyright 2025-2026 longan55 or authors.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*      https://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */
package rot

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

//编码方式：BIN、BCD、ASCII、HEX
//数据类型：INT、FLOAT、STRING、BITMAP
//对应方式：
//BIN：直接转换
//BCD：->INT ->STRING
//ASCII：每个字节表示一个ASCII码
//HEX：每个字节表示一个十六进制数

//ENCODE->DATATYPE

//如何解析一个字节序列
// 1. 字节顺序、字节长度、编码方式
// 2. 倍率、偏移量、BITMAP
// 3. 结果数据类型

//字节序：binary.LittleEndian, binary.BigEndian
//字节长度：1-8
//倍率：int类型
//偏移量：int类型
//BITMAP：[int]any, [string]any
//结果数据类型：INT、FLOAT、STRING、BITMAP(slice)
//
// BCD: 一般用于传输数字字符串，作为数值传输时不会丢失精度。最终解释的类型：整数、小数（需乘以倍率）、字符串
// BIN：二进制码、最基础的编码方式，直接将字节序列转换为无符号整数。最终解释的类型：整数、小数（需乘以倍率）
//		需要加工解释：BIN 码 1 版本号乘 10，v1.0 表示 0x0A
//		解释为：枚举切片（只表示其中一种意义）
//      解释为：BITMAP（可表示多种意义）
// ASCII: 每个字节表示一个ASCII码，直接将字节序列转换为字符串。最终解释的类型：字符串、也可以解释为数字字符的字面值数值（不常见）

func BIN2Uint64(bin []byte, order binary.ByteOrder) (uint64, error) {
	len := len(bin)
	switch len {
	case 1:
		return uint64(bin[0]), nil
	case 2:
		return uint64(order.Uint16(bin)), nil
	case 3, 4:
		bin4 := make([]byte, 8)
		copy(bin4[4-len:], bin) //前面字节填充0
		return uint64(order.Uint32(bin4)), nil
	case 5, 6, 7, 8:
		bin8 := make([]byte, 8)
		copy(bin8[8-len:], bin)
		return order.Uint64(bin8), nil
	default:
		return 0, errors.New("不符合字节长度范围1-8")
	}
}

// 无符号整形转BIN码
var ErrLength = errors.New("需要更大的长度存储该数值")

// len: 最长8字节
func Uint2BIN(n uint64, len uint8, order binary.ByteOrder) ([]byte, error) {
	switch len {
	case 1:
		if n > math.MaxUint8 {
			return nil, ErrLength
		}
		return []byte{byte(n)}, nil
	case 2:
		if n > math.MaxUint16 {
			return nil, ErrLength
		}
		b := make([]byte, 2)
		order.PutUint16(b, uint16(n))
		return b, nil
	case 3, 4:
		if n > math.MaxUint32 {
			return nil, ErrLength
		}
		b := make([]byte, 4)
		order.PutUint32(b, uint32(n))
		return b, nil
	case 5, 6, 7, 8:
		b := make([]byte, 8)
		order.PutUint64(b, n)
		return b, nil
	default:
		return nil, errors.New("非法字节长度")
	}
}

func Uint16ToBin(i uint16, order binary.ByteOrder) []byte {
	buf := make([]byte, 2)
	order.PutUint16(buf, i)
	return buf
}

// func Bin2UInt16(buf []byte, order binary.ByteOrder) uint16 {
// 	return order.Uint16(buf)
// }

// UInt2Bcd2Bytes uint16转两字节bcd码 isLittleEndian = true:小端, =false:大端
func UInt2Bcd2Bytes(n int64, isLittleEndian bool) []byte {
	if n < 0 {
		return nil
	}
	var b []byte
	if n < 256 {
		b = []byte{0}
	}
	for i := 0; ; i++ {
		h := (n / 10) % 10
		l := n % 10
		b = append(b, byte(h<<4|l))
		n = n / 100
		if n == 0 {
			break
		}
	}
	if !isLittleEndian {
		return b
	}
	l := len(b)
	var r = make([]byte, l)
	for i, v := range b {
		r[l-1-i] = v
	}
	return r
}

func Bcd2Int(b []byte) int {
	n, _ := strconv.Atoi(fmt.Sprintf("%x", Bin2Int(b)))
	return n
}

func BCD2Int(b []byte) (n int, err error) {
	l := len(b)
	for i := 1; i <= l; i++ {
		temp := b[l-i]
		n += int(temp&0xF) * int(math.Pow10(i*2-2))
		n += int(temp&0xF0) * int(math.Pow10(i*2-1))
	}
	return
}

// Bin2Float64 b:Bin码[]byte, bit:最多保留小数位, 小端
func Bin2Float64(order binary.ByteOrder, b []byte, bit int) (float64, error) {
	f := float64(Bin2Int(b, order)) / math.Pow10(bit)
	num, err := strconv.ParseFloat(strconv.FormatFloat(f, 'f', bit, 64), 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// Float64ToBin  f:浮点数, byteLength:字节数, bit:小数点位数
func Float64ToBin(f float64, byteLength byte, bit int) []byte {
	i := int(f * math.Pow10(bit))
	return Int2Bin(i, byteLength, binary.LittleEndian)
}

func Hex2Byte(str string) []byte {
	l := len(str)
	bHex := make([]byte, len(str)/2)
	ii := 0
	for i := 0; i < len(str); i = i + 2 {
		if l != 1 {
			ss := string(str[i]) + string(str[i+1])
			bt, _ := strconv.ParseInt(ss, 16, 32)
			bHex[ii] = byte(bt)
			ii = ii + 1
			l = l - 2
		}
	}
	return bHex
}

// func Hex2Byte(str string) ([]byte, error) {
// 	// 验证输入长度
// 	strLen := len(str)
// 	if strLen%2 != 0 {
// 		return nil, errors.New("hex string has odd length")
// 	}

// 	// 预分配结果数组
// 	result := make([]byte, strLen/2)

// 	// 解析每个字节
// 	for i := 0; i < strLen; i += 2 {
// 		// 使用strconv.ParseUint直接解析两个字符，避免字符串拼接
// 		val, err := strconv.ParseUint(str[i:i+2], 16, 8)
// 		if err != nil {
// 			return nil, fmt.Errorf("invalid hex character at position %d: %v", i, err)
// 		}
// 		result[i/2] = byte(val)
// 	}

// 	return result, nil
// }

func Bin2Int(b []byte, orders ...binary.ByteOrder) int {
	var order binary.ByteOrder = binary.BigEndian // 默认使用大端序，与Int2Bin保持一致
	if orders != nil {
		order = orders[0]
	}
	if len(b) == 3 {
		b = append([]byte{0}, b...) // 3字节特殊处理，扩展为4字节
	}
	bytesBuffer := bytes.NewBuffer(b)
	switch len(b) {
	case 1:
		var tmp int8 // 使用有符号整数类型以正确处理负数
		err := binary.Read(bytesBuffer, order, &tmp)
		if err != nil {
			return 0
		}
		return int(tmp)
	case 2:
		var tmp int16
		err := binary.Read(bytesBuffer, order, &tmp)
		if err != nil {
			return 0
		}
		return int(tmp)
	case 4:
		var tmp int32
		err := binary.Read(bytesBuffer, order, &tmp)
		if err != nil {
			return 0
		}
		return int(tmp)
	case 8:
		var tmp int64
		err := binary.Read(bytesBuffer, order, &tmp)
		if err != nil {
			return 0
		}
		return int(tmp) // 在32位系统上可能会被截断，但这是int类型的限制
	default:
		// 对于不支持的字节长度，尝试作为有符号整数处理
		// 首先检查第一个字节的最高位是否为1（表示负数）
		signed := false
		if len(b) > 0 && b[0]&0x80 != 0 {
			signed = true
		}

		// 如果是有符号负数，使用补码规则处理
		if signed {
			// 计算对应正数的补码值
			complement := make([]byte, len(b))
			for i := range complement {
				complement[i] = ^b[i]
			}
			// 加1得到原码
			for i := len(complement) - 1; i >= 0; i-- {
				complement[i]++
				if complement[i] != 0 {
					break // 没有进位，结束
				}
			}
			// 计算补码对应的整数值
			val := 0
			for i := 0; i < len(complement); i++ {
				if order == binary.BigEndian {
					val = val<<8 | int(complement[i])
				} else {
					val = val | int(complement[i])<<(8*i)
				}
			}
			return -val
		}

		// 无符号数处理
		val := 0
		for i := 0; i < len(b); i++ {
			if order == binary.BigEndian {
				val = val<<8 | int(b[i])
			} else {
				val = val | int(b[i])<<(8*i)
			}
		}
		return val
	}
}

func Int2Bin(n int, bytesLength byte, order binary.ByteOrder) []byte {
	switch bytesLength {
	case 1:
		tmp := int8(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()
	case 2:
		tmp := int16(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()
	case 3:
		tmp := int32(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()[0:3]
	case 4:
		tmp := int32(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()
	case 5:
		tmp := int64(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()[0:5]
	case 6:
		tmp := int64(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()[0:6]
	case 7:
		tmp := int64(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()[0:7]
	case 8:
		tmp := int64(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()
	}
	return nil
}

// ParseCP56time2a 解析CP56time2a 为字符串时间  7字节
func ParseCP56time2a(b []byte) string {
	second := ((int64(b[1]) << 8) + int64(b[0])) / 1000
	min := b[2] & 0x3F
	hour := b[3] & 0x1F
	day := b[4] & 0x1F
	month := b[5] & 0xF
	year := b[6] & 0x7F
	return fmt.Sprintf("20%d-%02d-%02d %02d:%02d:%02d",
		year, month, day, hour, min, second)
}

// EncodeCP56time2a 字符串时间编码为CP56time2a格式
func EncodeCP56time2a(str string) []byte {
	var b = []byte{0, 0, 0, 0, 0, 0, 0}
	parse, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return b
	}
	second := Int2Bin(int(parse.Second())*1000, 2, binary.LittleEndian)
	min := byte(parse.Minute())
	hour := byte(parse.Hour())
	day := byte(parse.Day())
	month := byte(parse.Month())
	year := byte(parse.Year() - 2000)
	b[0] = second[0]
	b[1] = second[1]
	b[2] = min
	b[3] = hour
	b[4] = day
	b[5] = month
	b[6] = year
	return b
}
