package protocol

type Encode interface {
	Integer(b []byte) (int, error)
	String(b []byte) (string, error)
}

type BIN struct{}

func (BIN) Integer(b []byte) (int, error) {
	return 0, nil
}

func (BIN) String(b []byte) (string, error) {
	return string(b), nil
}

type BCD struct{}

func (BCD) Integer(b []byte) (int, error) {
	return 0, nil
}

func (BCD) String(b []byte) (string, error) {
	return "", nil
}

type ASCII struct{}

func (ASCII) Integer(b []byte) (int, error) {
	return 0, nil
}

func (ASCII) String(b []byte) (string, error) {
	return "", nil
}

type HEX struct{}

func (HEX) Integer(b []byte) (int, error) {
	return 0, nil
}

func (HEX) String(b []byte) (string, error) {
	return "", nil
}
