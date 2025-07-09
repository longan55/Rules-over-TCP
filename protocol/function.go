package protocol

//功能结构体实现这个接口
type Function interface {
	Parse(data []byte) error
	Serialize() ([]byte, error)
}

// 功能码,长度应该更加广泛,但暂时使用1字节
type FunctionCode byte

func (fc FunctionCode) NewFunction() Function {
	//todo implate
	return nil
}

const (
	FunctionCodeLogin FunctionCode = iota
	FunctionCodeHeartbeat
)
