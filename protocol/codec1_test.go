package protocol

import (
	"fmt"
	"testing"
)

func TestCode1(t *testing.T) {
	b := []byte{0x01, 0x02, 0x03, 0x04}
	bin, err := BIN{}.String(b)
	if err != nil {
		t.Errorf("BIN.String() error = %v", err)
	}
	fmt.Printf("BIN:%v\n", bin)
	bcd, err := BCD{}.String(b)
	if err != nil {
		t.Errorf("BCD.String() error = %v", err)
	}
	fmt.Printf("BCD:%v\n", bcd)
	ascii, err := ASCII{}.String(b)
	if err != nil {
		t.Errorf("ASCII.String() error = %v", err)
	}
	fmt.Printf("ASCII:%v\n", ascii)
	hex, err := HEX{}.String(b)
	if err != nil {
		t.Errorf("HEX.String() error = %v", err)
	}
	fmt.Printf("HEX:%v\n", hex)
}
