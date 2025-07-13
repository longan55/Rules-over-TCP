package main

import "github.com/longan55/proto/protocol"

// 起始符 数据长度    数据域  校验
// 0x68  1      N      crc16
//
// 数据单元格式 -》 按功能码解析
//
// 两种方式
// 1.统一的Field结构体，统一的方法
// 2.不同的结构体，同一个接口

func main() {
	//原始数据
	var sourceData = []byte{0x68, 0x01, 0x97}
	//处理器构建器
	builder := protocol.NewDUHBuilder()
	//构建处理器
	builder.AddFielder(protocol.NewStarter([]byte{0x68})).
		AddFielder(protocol.NewDataLen(1))
	dataHander := builder.Build()
	//解析
	dataHander.Parse(sourceData)
}

// func main() {
// 	var data = []byte{0x68, 0x01, 0x97}
// 	builder := protocol.NewProtoBuilder()
// 	builder.SetStart(1, 0x68, binary.BigEndian).
// 		SetDataLength(1, binary.LittleEndian).
// 		SetData(1, binary.LittleEndian)
// 	//SetVerify(2, binary.LittleEndian, nil)
// 	proto := builder.Build()
// 	m, err := proto.UnWrap(data)
// 	if err != nil {
// 		fmt.Printf("unwrap error:%v\n", err)
// 		return
// 	}
// 	fmt.Printf("result:%v\n", m)
// }
