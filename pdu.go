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
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
)

var (
	once         sync.Once
	defaultOrder binary.ByteOrder = binary.BigEndian
)

func SetDefaultOrder(order binary.ByteOrder) {
	once.Do(func() {
		defaultOrder = order
	})
}

// DefaultOrder 返回默认字节序
func DefaultOrder() binary.ByteOrder {
	return defaultOrder
}

func NewProtocolBuilder() *ProtocolBuilder {
	return &ProtocolBuilder{
		du: &ProtocolDataUnit{
			elements: make([]ProtocolElement, 0, 3),
		},
	}
}

//协议元素组成规则
//1. 第一个元素必须是起始符
//2. 第一个元素到数据域之前的元素组成 消息头部元素.
//3. 数据域称为消息体元素
//4. 最后一个元素必须是校验码元素

type ProtocolBuilder struct {
	du *ProtocolDataUnit
}

// SetDefaultOrder 设置默认字节序
func (duBuilder *ProtocolBuilder) SetDefaultOrder(order binary.ByteOrder) *ProtocolBuilder {
	defaultOrder = order
	return duBuilder
}

// AddElement 添加协议元素
func (duBuilder *ProtocolBuilder) AddElement(element ProtocolElement) *ProtocolBuilder {
	duBuilder.du.elements = append(duBuilder.du.elements, element)
	return duBuilder
}

// AddCryptConfig 添加加密配置,应该只被调用一次
func (duBuilder *ProtocolBuilder) AddCryptConfig(cryptConfig *CryptConfig) *ProtocolBuilder {
	duBuilder.du.cryptLib = cryptConfig.cryptMap
	return duBuilder
}

// AddCrypt 添加加密算法
func (duBuilder *ProtocolBuilder) AddCrypt(cryptFlag int, cipher Cipher) *ProtocolBuilder {
	duBuilder.du.AddCrypt(cryptFlag, cipher)
	return duBuilder
}

func (duBuilder *ProtocolBuilder) HandleFunc(fc FunctionCode, fields ...func(*FunctionHandler)) *ProtocolBuilder {
	// 创建一个新的FunctionHandler实例
	functionHandler := NewFunctionHandler()
	// 应用所有字段定义
	for _, fieldDef := range fields {
		fieldDef(functionHandler)
	}
	// 添加到handlerMap中
	duBuilder.du.AddHandler(fc, functionHandler)
	return duBuilder
}

// HandleFuncWithParse 带字段定义的处理函数注册，类似于http处理函数注册方式
// 可以直接传入匿名函数作为处理逻辑，并同时定义字段结构
// fields是一个字段定义列表，每个字段定义是一个函数，用于配置FunctionHandler
func (duBuilder *ProtocolBuilder) HandleFuncWithParse(fc FunctionCode, handler Handler, fields ...func(*FunctionHandler)) *ProtocolBuilder {
	// 创建一个新的FunctionHandler实例
	functionHandler := NewFunctionHandler()
	// 应用所有字段定义
	for _, fieldDef := range fields {
		fieldDef(functionHandler)
	}
	// 设置处理函数
	functionHandler.SetHandler(handler)
	// 添加到handlerMap中
	duBuilder.du.AddHandler(fc, functionHandler)
	return duBuilder
}

// Build 构建协议数据单元
func (duBuilder *ProtocolBuilder) Build() (Protocol, error) {
	// 添加协议元素验证
	if len(duBuilder.du.elements) == 0 {
		return nil, errors.New("协议元素不能为空")
	}

	// 验证第一个元素必须是起始符
	if duBuilder.du.elements[0].Type() != Preamble {
		return nil, errors.New("第一个协议元素必须是起始符")
	}

	// 验证最后一个元素必须是校验码
	lastIdx := len(duBuilder.du.elements) - 1
	if duBuilder.du.elements[lastIdx].Type() != Checksum {
		return nil, errors.New("最后一个协议元素必须是校验码")
	}

	fmt.Println("协议元素组成部分：")
	fmt.Println("------------------------------------------------------")
	fmt.Printf("元素名称\t元素类型\t元素长度\t默认值\n")
	//todo 起始码+长度码 的长度
	for index, element := range duBuilder.du.elements {
		element.SetIndex(index)
		fmt.Printf("%s\t%v\t\t%d\t\t%#x\n", element.GetName(), element.Type(), element.SelfLength(), element.DefaultValue())
	}
	fmt.Println("------------------------------------------------------")
	return duBuilder.du, nil
}

// Protocol 协议接口
type Protocol interface {
	AddHandler(fc FunctionCode, f *FunctionHandler)
	Serve(ctx context.Context, conn net.Conn)
}

var _ Protocol = (*ProtocolDataUnit)(nil)

// ProtocolDataUnit 协议数据单元, 保存所有协议的上下文信息
type ProtocolDataUnit struct {
	counts     uint64
	cryptLib   map[int]Cipher
	conn       net.Conn
	elements   []ProtocolElement
	handlerMap map[FunctionCode]*FunctionHandler
}

// GetElementByIndex 通过索引获取ProtocolElement
func (pdu *ProtocolDataUnit) GetElementByIndex(index int) ProtocolElement {
	if index >= 0 && index < len(pdu.elements) {
		return pdu.elements[index]
	}
	return nil
}

// GetElementByType 通过类型获取ProtocolElement
func (pdu *ProtocolDataUnit) GetElementByType(typ ProtocolElementType) ProtocolElement {
	for _, element := range pdu.elements {
		if element.Type() == typ {
			return element
		}
	}
	return nil
}

// GetAllElements 获取所有ProtocolElement
func (pdu *ProtocolDataUnit) GetAllElements() []ProtocolElement {
	return pdu.elements
}

// AddCrypt 添加加密算法
func (pdu *ProtocolDataUnit) AddCrypt(cryptFlag int, crypt Cipher) {
	if pdu.cryptLib == nil {
		pdu.cryptLib = make(map[int]Cipher)
	}
	pdu.cryptLib[cryptFlag] = crypt
}

// Decrypt 解密数据
func (pdu *ProtocolDataUnit) Decrypt(cryptFlag int, src []byte) ([]byte, error) {
	if pdu.cryptLib == nil {
		return nil, errors.New("未配置加密算法")
	}
	if _, ok := pdu.cryptLib[cryptFlag]; !ok {
		panic(fmt.Sprintf("未配置加密算法 %d", cryptFlag))
	}
	return pdu.cryptLib[cryptFlag].Decrypt(src)
}

// AddHandler 添加处理函数
func (pdu *ProtocolDataUnit) AddHandler(fc FunctionCode, f *FunctionHandler) {
	if pdu.handlerMap == nil {
		pdu.handlerMap = make(map[FunctionCode]*FunctionHandler)
	}
	pdu.handlerMap[fc] = f
}

// DoHandle 执行处理函数
func (pdu *ProtocolDataUnit) DoHandle(code FunctionCode, payload []byte) error {
	fmt.Printf("函数码:\t\t%v,%+v\n", code, pdu.handlerMap)
	if handler, ok := pdu.handlerMap[code]; !ok {
		return errors.New("未配置处理函数")
	} else {
		return handler.Handle(payload)
	}
}

// Serve 处理连接
func (pdu *ProtocolDataUnit) Serve(ctx context.Context, conn net.Conn) {
	pdu.conn = conn
	for {
		select {
		case <-ctx.Done():
			//停止读取
			return
		default:
			fmt.Printf("[第%v个数据单元解析开始]\n", pdu.counts)
			//第一遍遍历elements, 读取一个完整的数据单元
			for _, element := range pdu.elements {
				err := element.Preprocess(conn, element, pdu)
				if err != nil {
					fmt.Println("数据预处理失败:", err)
					return
				}
			}
			for _, element := range pdu.elements {
				err := element.Deal(pdu)
				if err != nil {
					fmt.Println("数据解析失败:", err)
					return
				}
			}
			fmt.Printf("[第%v个数据单元解析完成]\n", pdu.counts)
			fmt.Println()
			pdu.counts++
		}
	}
}
