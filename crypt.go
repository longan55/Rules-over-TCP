package rot

type Cipher interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

var _ Cipher = (*CryptNothing)(nil)

type CryptNothing struct{}

func (c *CryptNothing) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (c *CryptNothing) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}

type CryptConfig struct {
	cryptMap map[int]Cipher
}

func NewCryptConfig() *CryptConfig {
	return &CryptConfig{
		cryptMap: map[int]Cipher{
			0: &CryptNothing{},
		},
	}
}

func (cc *CryptConfig) AddCrypt(flag int, crypt Cipher) {
	cc.cryptMap[flag] = crypt
}
