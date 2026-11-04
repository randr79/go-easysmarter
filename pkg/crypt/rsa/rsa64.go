package rsa

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
)

var (
	ErrInvalidKey   = errors.New("invalid RSA key")
	ErrInvalidValue = errors.New("RSA integer must be smaller than the modulus")
)

var (
	bigint0 big.Int = *(new(big.Int).SetUint64(0))
	bigint1 big.Int = *(new(big.Int).SetUint64(1))
)

type Rsa64Key struct {
	Modulus         *big.Int
	PublicExponent  *big.Int
	PrivateExponent *big.Int
}

type InvalidDataSizeError [3]int

func (k InvalidDataSizeError) Error() string {
	return fmt.Sprintf("invalid input, length must be  %d <= `%d` <= %d", k[0], k[1], k[2])
}

func (key Rsa64Key) Encrypt(value []byte) (*big.Int, error) {
	if vl := len(value); vl != 8 {
		return nil, InvalidDataSizeError{vl, 8, 8}
	}
	bi := new(big.Int).SetUint64(binary.LittleEndian.Uint64(value[:]))
	if bi.Cmp(key.Modulus) > 0 {
		return nil, ErrInvalidValue
	}
	return bi.Exp(bi, key.PublicExponent, key.Modulus), nil
}

func (key Rsa64Key) Decrypt(value *big.Int) ([]byte, error) {
	if key.Modulus.Cmp(&bigint1) < 0 {
		return nil, ErrInvalidKey
	} else if key.PrivateExponent.Cmp(&bigint0) == 0 {
		return nil, ErrInvalidKey
	}
	res := value.Exp(value, key.PrivateExponent, key.Modulus)
	return binary.LittleEndian.AppendUint64(nil, res.Uint64()), nil
}
