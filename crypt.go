package rot

type CryptFunc func(data []byte) ([]byte, error)

func CryptNone(data []byte) ([]byte, error) {
	return data, nil
}

type CryptConfig struct {
	cryptMap map[int]CryptFunc
}

func NewCryptConfig() *CryptConfig {
	return &CryptConfig{
		cryptMap: map[int]CryptFunc{
			0: CryptNone,
		},
	}
}

func (cc *CryptConfig) AddCrypt(flag int, crypt CryptFunc) {
	cc.cryptMap[flag] = crypt
}

func (cc *CryptConfig) GetCryptMap() map[int]CryptFunc {
	return cc.cryptMap
}
