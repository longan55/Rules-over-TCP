package protocol

import "sync"

//功能结构体实现这个接口
type Function interface {
	Parse(data []byte) (map[string]any, error)
	Serialize(map[string]any) ([]byte, error)
}

type FuctionTest struct {
}

func (f *FuctionTest) Parse(data []byte) (map[string]any, error) {
	return map[string]any{"test": data[:]}, nil
}

func (f *FuctionTest) Serialize(data map[string]any) ([]byte, error) {
	return []byte{data["test"].(byte)}, nil
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

//全局功能码注册
var functionMap = map[FunctionCode]func() Function{}
var fmux sync.RWMutex

func Regidter(fc FunctionCode, f func() Function) {
	fmux.Lock()
	defer fmux.Unlock()

	functionMap[fc] = f
}
