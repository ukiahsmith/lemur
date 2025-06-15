package funcs

import (
	"errors"
)

// Mod returns in1 % in2
func Mod(in1, in2 int) (int64, error) {
	n1 := int64(in1)
	n2 := int64(in2)

	if n2 == 0 {
		return 0, errors.New("number can't be divided by zero at modulo operation")
	}

	return n1 % n2, nil
}

// ModBool returns true if in1 % in2 == 0
func ModBool(in1, in2 int) (bool, error) {
	result, err := Mod(in1, in2)
	if err != nil {
		return false, err
	}

	return result == int64(0), nil
}
