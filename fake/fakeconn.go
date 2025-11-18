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
	buffer *bytes.Buffer // 使用buffer代替reader，支持数据追加
	writer bytes.Buffer
	closed bool
}

// NewFakeConn 创建一个新的FakeConn实例
func NewFakeConn() *FakeConn {
	buffer := bytes.NewBuffer([]byte{})
	return &FakeConn{
		buffer: buffer,
		writer: bytes.Buffer{},
		closed: false,
	}
}

// SetData 追加数据到读取缓冲区，不覆盖现有数据
func (c *FakeConn) SetData(data []byte) {
	c.buffer.Write(data)
}

// ClearData 清空读取缓冲区
func (c *FakeConn) ClearData() {
	c.buffer.Reset()
}

// Read 实现net.Conn接口的Read方法
func (c *FakeConn) Read(b []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}
	return c.buffer.Read(b)
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
