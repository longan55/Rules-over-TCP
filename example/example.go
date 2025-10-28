package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	rot "github.com/longan55/Rules-over-TCP"
)

func main() {
	//原始数据
	// var sourceData = [][]byte{{0x68},
	// 	{0x0F},
	// 	{0x00},
	// 	{0x01},
	// 	{0x7F, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x01},
	// 	{0xC0, 0xA7}}
	//处理器构建器
	builder := rot.NewProtocolBuilder()

	//添加加密配置
	cryptConfig := rot.NewCryptConfig()
	cryptConfig.AddCrypt(0, rot.CryptNone)
	builder.AddCryptConfig(cryptConfig)

	//配置处理器，来数据时自动处理
	setHandlerConfig(builder)

	//TODO:
	//配置编码器，用于主动发送数据或被处理器调用

	//构建处理器
	builder.AddElement(rot.NewStarter([]byte{0x68})).
		AddElement(rot.NewDataLen(1)).
		AddElement(rot.NewCyptoFlag()).
		AddElement(rot.NewFuncCode()).
		AddElement(rot.NewPayload()).
		AddElement(rot.NewCheckSum(0, 2))
	dataHander, err := builder.Build()
	if err != nil {
		fmt.Println("构建处理器失败:", err)
		return
	}

	// 创建FakeConn并设置测试数据
	fakeConn := NewFakeConn([]byte{0x68, 0x0F, 0x00, 0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x01})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用fake conn运行Handle方法
	go dataHander.Handle(ctx, fakeConn)

	// 给Handle方法一些时间处理数据
	time.Sleep(100 * time.Millisecond)

	//解析
	// if err := dataHander.Parse(sourceData); err != nil {
	// 	fmt.Println("解析失败:", err)
	// 	return
	// } else {
	// 	fmt.Println("解析成功!")
	// }
}

func setHandlerConfig(builder *rot.ProtocolBuilder) {
	handlerConfig := rot.NewHandlerConfig()
	fh := new(rot.FucntionHandler)
	fh.NewDecoder("a", binary.BigEndian).BIN().SetByteLength(4).Integer()
	fh.NewDecoder("b", binary.BigEndian).BIN().SetByteLength(4).Integer()
	fh.NewDecoder("c", binary.BigEndian).BIN().SetByteLength(2).Integer()
	fh.NewDecoder("d", binary.BigEndian).BIN().SetByteLength(2).Float1().Multiple(0.01)
	fh.NewDecoder("e", binary.BigEndian).BIN().SetByteLength(1).Integer().SetEnum(map[int]any{
		0: "A",
		1: "B",
		2: "C",
		3: "D",
	})
	fh.SetHandle(func(parsedData map[string]rot.ParsedData) error {
		fmt.Println("parsedData:", parsedData)
		return nil
	})
	handlerConfig.AddHandler(rot.FunctionCode(0x01), fh)
	builder.AddHandlerConfig(handlerConfig)
}

// FakeConn 实现net.Conn接口，用于测试
type FakeConn struct {
	reader *bytes.Reader
	writer bytes.Buffer
	closed bool
}

// NewFakeConn 创建一个新的FakeConn实例
func NewFakeConn(data []byte) *FakeConn {
	return &FakeConn{
		reader: bytes.NewReader(data),
		writer: bytes.Buffer{},
		closed: false,
	}
}

// SetData 设置要读取的数据
func (c *FakeConn) SetData(data []byte) {
	c.reader = bytes.NewReader(data)
}

// Read 实现net.Conn接口的Read方法
func (c *FakeConn) Read(b []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}
	return c.reader.Read(b)
}

// Write 实现net.Conn接口的Write方法
func (c *FakeConn) Write(b []byte) (n int, err error) {
	if c.closed {
		return 0, errors.New("connection closed")
	}
	return c.writer.Write(b)
}

// Close 实现net.Conn接口的Close方法
func (c *FakeConn) Close() error {
	c.closed = true
	return nil
}

// LocalAddr 实现net.Conn接口的LocalAddr方法
func (c *FakeConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
}

// RemoteAddr 实现net.Conn接口的RemoteAddr方法
func (c *FakeConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8081}
}

// SetDeadline 实现net.Conn接口的SetDeadline方法
func (c *FakeConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline 实现net.Conn接口的SetReadDeadline方法
func (c *FakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline 实现net.Conn接口的SetWriteDeadline方法
func (c *FakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// GetWrittenData 获取写入的数据
func (c *FakeConn) GetWrittenData() []byte {
	return c.writer.Bytes()
}
