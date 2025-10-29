package fake

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"
)

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
