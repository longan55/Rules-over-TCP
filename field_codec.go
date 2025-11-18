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
	"encoding/binary"
	"errors"
	"fmt"
)

// CodecMode 编解码模式
type CodecMode int

const (
	ModeDecode CodecMode = iota // 解码模式（反序列化）
	ModeEncode                  // 编码模式（序列化）
)

// FieldCodecConfig 字段编解码配置
type FieldCodecConfig struct {
	name          string
	length        int
	mode          CodecMode
	codec         Codec
	dataTyper     DataTyper
	explainConfig *ExplainConfig
}

// ExplainConfig 数据解释配置
type ExplainConfig struct {
	moflag   bool
	multiple float64     // 倍数
	offset   float64     // 偏移量
	other    string      // 其他解释
	enum     map[int]any // 枚举映射
	bitmap   map[int]any // 位图映射
}

// NewFieldCodecConfig 创建新的字段编解码配置
func NewFieldCodecConfig(name string, options ...CodecOption) *FieldCodecConfig {
	config := &FieldCodecConfig{
		name:          name,
		mode:          ModeDecode,
		explainConfig: &ExplainConfig{},
	}

	// 应用所有选项
	for _, option := range options {
		option.Apply(config)
	}

	// 如果没有指定编解码器，默认使用BIN编解码器（大端序）
	if config.codec == nil {
		config.codec = NewCodecBIN(binary.BigEndian)
	}
	return config
}

func (config *FieldCodecConfig) Decode(data []byte) (*ParsedData, error) {
	if config.mode != ModeDecode {
		return nil, errors.New("config is not in decode mode")
	}
	// 使用编解码器解码
	rawValue, err := config.codec.Decode(data)
	if err != nil {
		return nil, err
	}
	explainedValue := config.dataTyper.Explain(rawValue)

	parsed := &ParsedData{
		Bytes:     data,
		Origin:    explainedValue,
		Explained: explainedValue,
	}

	if config.explainConfig != nil && config.explainConfig.enum != nil {
		i, ok := explainedValue.(int)
		if !ok {
			return nil, fmt.Errorf("explained value is not int: %v", explainedValue)
		}
		if enumValue, ok := config.explainConfig.enum[i]; ok {
			parsed.Explained = enumValue
			return parsed, nil
		} else {
			parsed.Explained = config.explainConfig.other
			return parsed, nil
		}
	}

	return parsed, nil
}

// Encode 编码方法
func (config *FieldCodecConfig) Encode(data any) ([]byte, error) {
	if config.mode != ModeEncode {
		return nil, errors.New("config is not in encode mode")
	}
	if config.dataTyper != nil {
		data = config.dataTyper.UnExplain(data)
	}
	// 使用编解码器编码
	return config.codec.Encode(data, config.length)
}
