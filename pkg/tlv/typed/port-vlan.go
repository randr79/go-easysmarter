package typed

import (
	"encoding/binary"
	"fmt"
)

type PortVlanCount uint8

func (PortVlanCount) Type() Type { return TypePortVlanCount }

func (m PortVlanCount) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *PortVlanCount) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type PortVlanEnabled bool

func (PortVlanEnabled) Type() Type { return TypePortVlanEnabled }

func (m PortVlanEnabled) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *PortVlanEnabled) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type PortVlanGroup byte

type PortVlanMembers struct {
	GroupID PortVlanGroup
	Ports   Ports
}

func (PortVlanMembers) Type() Type { return TypePortVlanMembers }

func (m *PortVlanMembers) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}
func (m PortVlanMembers) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
