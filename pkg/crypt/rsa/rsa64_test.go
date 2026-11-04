package rsa

import (
	"math/big"
	"slices"
	"testing"
)

func DefaultRsa64Key() Rsa64Key {
	return Rsa64Key{
		Modulus:         (new(big.Int).SetUint64(16400000136700000189)),
		PublicExponent:  (new(big.Int).SetUint64(65537)),
		PrivateExponent: (new(big.Int).SetUint64(13455672473792676021)),
	}
}

func TestAsciiString(t *testing.T) {
	data := []byte{0x00, 0x04, 0xf0, 0xf6, 0x3c, 0xc2, 0x90, 0xc8, 0x1a, 0x9b}
	v, err := DecodeTLV528Value(data)
	if err != nil {
		t.Error(err)
	}
	if bi, err := DefaultRsa64Key().Decrypt(v); err != nil {
		t.Error(err)
	} else if string(bi[:]) != "Wzxre7rc" {
		t.Errorf("bytes should have decoded to `Wzxre7rc` got `%s`", string(bi[:]))
	}
}
func TestBytesString(t *testing.T) {
	bi, err := DefaultRsa64Key().Encrypt([]byte("Wzxre7rc"))
	if err != nil {
		t.Error(err)
	} else {
		enc := EncodeTLV528Value(bi)
		if slices.Compare(enc, []byte{0x00, 0x04, 0xf0, 0xf6, 0x3c, 0xc2, 0x90, 0xc8, 0x1a, 0x9b}) != 0 {
			t.Errorf("bytes should have encoded to [00 04 f0 f6 3c c2 90 c8 1a 9b] got `% 02x`", enc)
		}
	}

}
