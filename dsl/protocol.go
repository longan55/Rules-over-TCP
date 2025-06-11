package dsl

type Protocol struct {
	Start *Start
}

type Start struct {
	Index int
	Value []byte
}
type DataLength struct {
	Index    int
	Length   int
	Encoding string //BCD、BIN、ASCII
}

/*
	StartCode          any      //起始码
	DataLength         chan int //数据长度
	EncryptionFlag     byte     //加密标志
	SerialNumber       any      //序列号
	ConfirmationNumber any      //确认号
	Data               []byte   //数据域
	CheckCode          any      //校验码
*/
