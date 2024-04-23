package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
)

// Protocol 协议接口, 功能包括：1.封包 2.解包 3.检查协议，可以跟解包封包一起
type Protocol interface {
	Wrap(Protocol) []byte
	UnWrap([]byte) (map[string]any, error)
	Check() error
}

// DefProto 协议实体，存储协议信息。 协议字段顺序、编码方式、默认值等
type DefProto struct {
	Protocol
	StartCode          any      //起始码
	DataLength         chan int //数据长度
	EncryptionFlag     byte     //加密标志
	SerialNumber       any      //序列号
	ConfirmationNumber any      //确认号
	Data               []byte   //数据域
	CheckCode          any      //校验码
	//
	fields []field
}

func (p *DefProto) UnWrap(in []byte) (map[string]any, error) {
	offset := 0
	m := make(map[string]any)
	for _, field := range p.fields {
		field.RealValue = in[offset : offset+field.length()]
		offset += field.length()
		_, err := field.Check(field.RealValue)
		if err != nil {
			fmt.Printf("发生错误：%v\n", err)
			return nil, err
		}
		m[field.name] = field.RealValue
	}
	return m, nil
}

type fieldInterface interface {
	length() int
}

type field struct {
	name string //消息帧 元素名字
	//FType        fieldType      //消息帧 字段类型
	scale        uint8            // 1十六进制，0十进制
	Len          int              //消息帧 元素本身长度
	DefaultValue int64            //默认值
	RealValue    []byte           //真实值
	Order        binary.ByteOrder //大小端
	length       func() int       //获取长度
	Check        func([]byte) (any, error)
	IsAsciiChar  bool //true ASCII字符，false 数值
}

type ProtoBuilder struct {
	proto *DefProto
}

func NewProtoBuilder() *ProtoBuilder {
	return &ProtoBuilder{
		proto: &DefProto{
			fields: []field{},
		},
	}
}

// SetStart byteLength：占用字节长度，defaultValue：默认值,order：大小端
func (pb *ProtoBuilder) SetStart(selfLength int, defaultValue int64, order binary.ByteOrder) *ProtoBuilder {
	var f = field{
		name:         "Start Code",
		scale:        0,
		Len:          selfLength,
		DefaultValue: defaultValue,
		Order:        order,
		Check: func(start []byte) (any, error) {
			i, err := BIN2Uint64(start, order)
			if err != nil {
				return nil, errors.New("Start Code translate to uint64 wrong :" + err.Error())
			} else if i == uint64(defaultValue) {
				return nil, nil
			}
			return nil, errors.New("StartCode is not " + strconv.Itoa(int(defaultValue)))
		},
		length: func() int {
			return selfLength
		},
	}
	//加入fields队列
	pb.proto.fields = append(pb.proto.fields, f)
	pb.proto.StartCode = defaultValue
	return pb
}

func (pb *ProtoBuilder) SetDataLength(selfLength int, order binary.ByteOrder) *ProtoBuilder {
	var c = make(chan int, 1)
	pb.proto.DataLength = c
	var f = field{
		name:  "Data Length",
		scale: 0,
		Len:   selfLength,
		Order: order,
		Check: func(dataLength []byte) (any, error) {
			i, err := BIN2Uint64(dataLength, order)
			if err != nil {
				return nil, errors.New("Data Length translate to uint64 wrong :" + err.Error())
			}
			c <- int(i)
			return i, nil
		},
		length: func() int {
			return selfLength
		},
	}
	pb.proto.fields = append(pb.proto.fields, f)
	return pb
}

func (pb *ProtoBuilder) SetData(selfLength int, order binary.ByteOrder) *ProtoBuilder {
	var f = field{
		name:  "Data Length",
		scale: 0,
		Len:   selfLength,
		Order: order,
		Check: func(dataLength []byte) (any, error) {
			i, err := BIN2Uint64(dataLength, order)
			if err != nil {
				return nil, errors.New("Data Length translate to uint64 wrong :" + err.Error())
			}
			return i, nil
		},
		length: func() int {
			return <-pb.proto.DataLength
		},
	}
	pb.proto.fields = append(pb.proto.fields, f)
	return pb
}

func (pb *ProtoBuilder) SetVerify(selfLength int, order binary.ByteOrder, verify func([]byte) error) *ProtoBuilder {
	var f = field{
		name:  "Verify Code",
		scale: 0,
		Len:   selfLength,
		Order: order,
		Check: func(verifyCode []byte) (any, error) {
			err := verify(verifyCode)
			if err != nil {
				return nil, errors.New("Data Length translate to uint64 wrong :" + err.Error())
			}
			return nil, nil
		},
	}
	pb.proto.fields = append(pb.proto.fields, f)
	return pb
}

func (pb *ProtoBuilder) Build() Protocol {
	fmt.Printf("protocol fields len：%d\n", len(pb.proto.fields))
	return pb.proto
}
