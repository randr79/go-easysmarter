package rc4

import (
	"encoding/hex"
	"testing"
)

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}

var rc4rs8key *Rc4Key = must(NewRc4Key([]byte("Wzxre7rc")))

func TestRC4(t *testing.T) {
	in := []byte{0x5, 0x3, 0x67, 0x87, 0xe7, 0x69, 0x12, 0x2e, 0xf9, 0xe2, 0xe0, 0xef, 0x4c, 0xf2, 0x0, 0xfa, 0x5, 0x3, 0x1, 0x5, 0x7, 0x3e, 0x6, 0x4, 0x0, 0x4, 0x33, 0x6f, 0x4, 0x3, 0x6, 0x6, 0x7, 0x4, 0x3, 0x4, 0x60, 0x60, 0x6d, 0x6b, 0x6d, 0x7, 0x1, 0x4, 0x3, 0xc, 0x6b, 0x79, 0x76, 0x36, 0x76, 0x32, 0x61, 0x64, 0x4, 0xf9, 0xff, 0x0, 0x4}
	want := "01" +
		"03" +
		"6083e768112d" +
		"f8e4e3eb4bf0" +
		"02fb" +
		"00000000" +
		"003b" +
		"0000" +
		"0000" +
		"3569" +
		"00000000" +

		"0200" +
		"0006" +
		"61646d696e00" +

		"0202" +
		"0009" +
		"6b7973317430676100" +

		"ffff" +
		"0000"

	out := rc4rs8key.XORKey(in)

	if got := hex.EncodeToString(out); got != want {
		t.Errorf("RC4Eight() = %s, want %s", got, want)
	}
}

func TestRC8again(t *testing.T) {
	in := "05016787e769122ef9e2e0ef4cf200fb05030105072d0604000495f9040306060c050302fefb0002"
	want := "01" +
		"01" +
		"6083e768112d" +
		"f8e4e3eb4bf0" +
		"02fa" +
		"00000000" +
		"0028" +
		"0000" +
		"0000" +
		"93ff" +
		"00000000" +
		"0901" +
		"0000" +
		"ffff0000"
	inb, err := hex.DecodeString(in)
	if err != nil {
		t.Fatal(err)
	}
	out := rc4rs8key.XORKey(inb)

	if got := hex.EncodeToString(out); got != want {
		t.Errorf("RC4Eight() = %s, want %s", got, want)
	}
}
