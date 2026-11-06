package typed

import (
	"encoding/binary"
)

type MultiTenantUplink struct {
	Enabled bool
	Uplink  Port
}

func (MultiTenantUplink) Type() Type { return TypeMultiTenantUplink }

func (m *MultiTenantUplink) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e

}
func (m MultiTenantUplink) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
