package transport

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type MACAddress [6]byte

func (m MACAddress) String() string {
	const hexDigit = "0123456789abcdef"
	buf := make([]byte, 0, 17)
	for i, b := range m {
		if i > 0 {
			buf = append(buf, ':')
		}
		buf = append(buf, hexDigit[b>>4], hexDigit[b&0xF])
	}
	return string(buf)
}

func (m MACAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())

}
func (m *MACAddress) UnmarshalJSON(data []byte) error {
	var val string
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	} else if x, err := hex.DecodeString(strings.ReplaceAll(val, ":", "")); err != nil {
		return err
	} else if len(x) != 6 {
		return fmt.Errorf("invalid mac address %s", val)
	} else {
		*m = MACAddress(x)
	}
	return nil
}
