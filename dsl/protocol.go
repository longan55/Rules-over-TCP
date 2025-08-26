package dsl

type Protocol struct {
	Start *Start
}

type Start struct {
	Index int    `json:"index"`
	Value []byte `json:"value"`
}
type DataLength struct {
	Index      int    `json:"index"`
	SelfLength int    //表示"长度"的数据长度不超过2字节
	Encoding   string //BCD、BIN、ASCII
	Endian     int    //0小端,1大端
}

// SerialNumber 默认2字节
type SerialNumber struct {
	Index    int
	Encoding string
	Endian   int //0小端,1大端
}

// Encryption 加密标志
type Encryption struct {
	Index     int
	Algorithm string
	Range     [2]int //范围,起始索引 - 终止索引
}

// DataType 数据类型
type DataType struct {
	Index      int
	SelfLength int //一般1字节
	Encoding   string
	Endian     int //0小端,1大端
}

// DataDomain 数据域
type DataDomain struct {
	Index int
}

// CheckSum 校验码
type CheckSum struct {
	Index      int
	SelfLength int //一般1字节
	Algorithm  string
	Encoding   string
	Endian     int //0小端,1大端
}

/*
	StartCode          any      //起始码 V
	DataLength         chan int //数据长度 V
	EncryptionFlag     byte     //加密标志
	SerialNumber       any      //序列号
	ConfirmationNumber any      //确认号
	Data               []byte   //数据域 V
	CheckCode          any      //校验码 V
*/

// GlobalAttribute 全局属性
// type GlobalAttribute struct {
// 	DefaultEndian bool //默认字节序 false小端, true大端
// }

// ElementDescription 元素描述
// type ElementDescription struct {
// 	Index     int              //元素索引
// 	Name      string           //元素名称
// 	Content   string           //元素内容/中文名
// 	Length    int              //元素自身长度
// 	Value     []byte           //元素默认值
// 	EncodAttr EncodingAttrbute //元素编码属性
// }

// EncodingAttrbute 编码属性
// type EncodingAttrbute struct {
// 	Endian   bool //0小端,1大端
// 	Encoding string
// 	Zoom     int //缩放因子
// }

// 定义操作
// const (
// 	OP_Equal = iota //比较值、返回结果
// 	OP_Read         //读取值、并返回
// 	OP_Function
// )

// type OptionAttribute struct {
// 	Operator int
// }

// DataUnitDescription 数据单元描述
// type DataUnitDescription struct {
// 	GlobalAttr          GlobalAttribute
// 	ElementDescriptions []ElementDescription
// }
