package protocol

type CryptFunc func(data []byte) ([]byte, error)

func CryptNone(data []byte) ([]byte, error) {
	return data, nil
}

type CryptConfig struct {
	cryptMap map[uint64]CryptFunc
}

func NewCryptConfig() *CryptConfig {
	return &CryptConfig{
		cryptMap: map[uint64]CryptFunc{
			0: CryptNone,
		},
	}
}

func (cc *CryptConfig) AddCrypt(flag uint64, crypt CryptFunc) {
	cc.cryptMap[flag] = crypt
}

func (cc *CryptConfig) Config() map[uint64]CryptFunc {
	return cc.cryptMap
}
