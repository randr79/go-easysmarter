package typed

import (
	"bytes"
	"fmt"
)

func stringToBytes[S ~string](str S, minLen int, maxLen int) ([]byte, error) {
	if sl := len(str); sl < minLen || sl > maxLen {
		return nil, fmt.Errorf("Invalid value, %T needs to be between %d and %d characters, got %d", str, minLen, maxLen, sl)
	} else {
		return append([]byte(str), 0), nil
	}
}

func bytesTostring[S ~string](value []byte, into *S) error {
	*into = S(bytes.TrimRight(value, "\x00"))
	return nil
}
func noContent[C any](v []byte, o C) error {
	if bl := len(v); bl != 0 {
		return fmt.Errorf("type %T is a zero length object, got %d bytes to decode", o, bl)
	}
	return nil
}

func encodeBytes[I ~[]byte](value I, size int) ([]byte, error) {
	if vs := len(value); vs != size {
		return nil, fmt.Errorf("Invalid value, %T needs to be %d bytes, got %d", value, vs, size)
	}
	return append([]byte{}, value...), nil
}
func decodeBytes[O ~[]byte](value []byte, size int, into *O) error {
	if vs := len(value); vs != size {
		return fmt.Errorf("Invalid value, %T needs to be %d bytes, got %d", value, vs, size)
	}
	*into = O(value)
	return nil
}
