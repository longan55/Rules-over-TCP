package protocol

type Parser func(data []byte) (map[string]any, error)

//编码方式：BIN、BCD、ASCII
//数据类型：整数、浮点数、字符串
