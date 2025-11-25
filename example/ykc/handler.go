package main

import (
	"fmt"

	rot "github.com/longan55/Rules-over-TCP"
)

func Handle01(parsedData map[string]rot.ParsedData) error {
	fmt.Println("parsedData01:", parsedData)
	return nil
}

func Parse01(fh *rot.FunctionHandler) {
	fh.AddField("a", rot.WithBin(), rot.WithLength(4), rot.WithInteger(true, 1, 0))
	fh.AddField("b", rot.WithBin(), rot.WithLength(4), rot.WithInteger(true, 1, 0))
	fh.AddField("c", rot.WithBin(), rot.WithLength(2), rot.WithInteger(true, 2, 0))
	fh.AddField("d", rot.WithBin(), rot.WithLength(2), rot.WithFloat(true, 0.01, 0))
	fh.AddField("e", rot.WithBin(), rot.WithLength(1), rot.WithInteger(true, 1, 0), rot.WithEnum("Other", map[int]any{0: "A", 1: "B", 2: "C"}))
}

// Handle02 处理函数02
func Handle02(parsedData map[string]rot.ParsedData) error {
	fmt.Println("parsedData02:", parsedData)
	return nil
}

// Parse02 解析函数02
func Parse02(fh *rot.FunctionHandler) {
	fh.AddField("code", rot.WithBcd(), rot.WithLength(4), rot.WithString())
	fh.AddField("price", rot.WithBcd(), rot.WithLength(4), rot.WithFloat(true, 0.0001, 0))
	fh.AddField("intPrice", rot.WithBcd(), rot.WithLength(4), rot.WithInteger(true, 1, 0))
}

// Handle03 处理函数03
func Handle03(parsedData map[string]rot.ParsedData) error {
	fmt.Println("parsedData03:", parsedData)
	return nil
}

// Parse03 解析函数03
func Parse03(fh *rot.FunctionHandler) {
	fh.AddField("ascii", rot.WithAscii(), rot.WithLength(4), rot.WithString())
	fh.SetHandler(func(parsedData map[string]rot.ParsedData) error {
		fmt.Println("parsedData:", parsedData)
		return nil
	})
}
