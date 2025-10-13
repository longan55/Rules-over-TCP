package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// BIN码转无符号整形
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
var lengthErr = errors.New("需要更大的长度存储该数值")

// len: 最长8字节
func Uint2BIN(n uint64, len uint8, order binary.ByteOrder) ([]byte, error) {
	switch len {
	case 1:
		if n > math.MaxUint8 {
			return nil, lengthErr
		}
		return []byte{byte(n)}, nil
	case 2:
		if n > math.MaxUint16 {
			return nil, lengthErr
		}
		b := make([]byte, 2)
		order.PutUint16(b, uint16(n))
		return b, nil
	case 3, 4:
		if n > math.MaxUint32 {
			return nil, lengthErr
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
	n, _ := strconv.Atoi(fmt.Sprintf("%x", Bin2UInt(b)))
	return n
}

func BCD2Int(b []byte) (n int) {
	l := len(b)
	for i := 1; i <= l; i++ {
		temp := b[l-i]
		n += int(temp&0xF) * int(math.Pow10(i*2-2))
		n += int(temp&0xF0) * int(math.Pow10(i*2-1))
	}
	return
}

// Bin2Float64 b:Bin码[]byte, bit:最多保留小数位, 小端
func Bin2Float64(b []byte, bit int) float64 {
	f := float64(Bin2UInt(b, binary.LittleEndian)) / math.Pow10(bit)
	num, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'f', bit, 64), 64)
	return num
}

// Float64ToBin  f:浮点数, byteLength:字节数, bit:小数点位数
func Float64ToBin(f float64, byteLength byte, bit int) []byte {
	i := int64(f * math.Pow10(bit))
	return Int2Bin(i, byteLength, binary.LittleEndian)
}

// Hex2Str 不用这个函数 使用hex.EncodeToString()
func Hex2Str(b []byte) string {
	var s = ""
	for _, v := range b {
		s += fmt.Sprintf("%x", v)
	}
	return strings.ToUpper(s)
}

func Str2Hex(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		fmt.Printf("err:%v\n", err)
		return nil
	}
	return b
}

func Bcd2Str(b []byte) string {
	return hex.EncodeToString(b)
}

func Str2Bcd(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		fmt.Printf("err:%v\n", err)
		return nil
	}
	return b
	//var rNumber = str
	//for i := 0; i < 8-len(str); i++ {
	//	rNumber = "f" + rNumber
	//}
	//bcd := Hex2Byte(rNumber)
	//return bcd
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

func Bin2UInt(b []byte, orders ...binary.ByteOrder) int {
	var order binary.ByteOrder = binary.LittleEndian
	if orders != nil {
		order = orders[0]
	}
	if len(b) == 3 {
		b = append([]byte{0}, b...)
	}
	bytesBuffer := bytes.NewBuffer(b)
	switch len(b) {
	case 1:
		var tmp uint8
		binary.Read(bytesBuffer, order, &tmp)
		return int(tmp)
	case 2:
		var tmp uint16
		binary.Read(bytesBuffer, order, &tmp)
		return int(tmp)
	case 4:
		var tmp uint32
		binary.Read(bytesBuffer, order, &tmp)
		return int(tmp)
	default:
		return 0
	}
}

func Int2Bin(n int64, bytesLength byte, order binary.ByteOrder) []byte {
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
		tmp := n
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()[0:5]
	case 6:
		tmp := n
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()[0:6]
	case 7:
		tmp := n
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()[0:7]
	case 8:
		tmp := n
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, order, &tmp)
		return bytesBuffer.Bytes()
	}
	return nil
}
