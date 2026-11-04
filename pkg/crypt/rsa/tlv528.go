package rsa

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

func EncodeTLV528Value(value *big.Int) []byte {
	return binary.LittleEndian.AppendUint64([]byte{0, 4}, value.Uint64())
}

func DecodeTLV528Value(encoded []byte) (*big.Int, error) {
	if len(encoded) != 10 || binary.BigEndian.Uint16(encoded[0:2]) != 4 {
		return nil, fmt.Errorf("invalid input, input should be 10 bytes starting with `00 04` got `% 02x`", encoded)
	} else {
		return new(big.Int).SetUint64(
			binary.LittleEndian.Uint64(encoded[2:10]),
		), nil
	}
}

type Tlv528 []byte

func (tlv Tlv528) Type() uint16 {
	return 528
}

func (tlv Tlv528) Value() []byte {
	return []byte(tlv)
}
func (tlv Tlv528) Length() uint16 {
	return uint16(len(tlv))
}
