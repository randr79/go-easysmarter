package typed

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Ieee8021QVlanEnabled bool

func (Ieee8021QVlanEnabled) Type() Type { return TypeIeee8021QVlanEnabled }

func (m Ieee8021QVlanEnabled) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *Ieee8021QVlanEnabled) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type Ieee8021QMaxVlans uint8

func (Ieee8021QMaxVlans) Type() Type { return TypeIeee8021QMaxVlans }

func (m Ieee8021QMaxVlans) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *Ieee8021QMaxVlans) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type VlanId struct {
	value uint16
}

func (v VlanId) Get() uint16 {
	return v.value
}

func NewVlanId(vlan uint16) (*VlanId, error) {
	if vlan == 0 || vlan > 4094 {
		return nil, fmt.Errorf("Invalid vlan value, vlans must be in the range 1-4096, not %d", vlan)
	}
	return &VlanId{
		value: vlan,
	}, nil
}

type Ieee8021QVlanMembers struct {
	VlanID   VlanId //2
	Untagged Ports  //4
	Tagged   Ports  //4
	Name     string
}

func (Ieee8021QVlanMembers) Type() Type { return TypeIeee8021QVlanMembers }

func (m Ieee8021QVlanMembers) Encode() ([]byte, error) {
	if vn, err := stringToBytes(m.Name, 0, 10); err != nil {
		return nil, err
	} else {
		buffer := binary.BigEndian.AppendUint16(nil, uint16(m.VlanID.value))
		buffer = binary.BigEndian.AppendUint32(buffer, uint32(m.Untagged))
		buffer = binary.BigEndian.AppendUint32(buffer, uint32(m.Tagged))
		return append(buffer, vn...), nil
	}
}
func (m *Ieee8021QVlanMembers) Decode(value []byte) error {
	if vl := len(value); vl < 10 {
		return fmt.Errorf("need at least 10 bytes, got %d", vl)
	}
	m.VlanID.value = binary.BigEndian.Uint16(value[0:1])
	m.Untagged = Ports(binary.BigEndian.Uint32(value[1:5]))
	m.Tagged = Ports(binary.BigEndian.Uint16(value[5:10]))
	m.Name = string(bytes.TrimRight(value[10:], "\x00"))
	return nil
}

type Ieee8021QPortNativeVlan struct {
	PortNo     Port
	NativeVlan VlanId
}

func (Ieee8021QPortNativeVlan) Type() Type { return TypeIeee8021QPortNativeVlan }

func (m Ieee8021QPortNativeVlan) Encode() ([]byte, error) {
	return binary.BigEndian.AppendUint16(
		[]byte{byte(m.PortNo)},
		uint16(m.NativeVlan.value),
	), nil
}
func (m *Ieee8021QPortNativeVlan) Decode(value []byte) error {
	if vl := len(value); vl != 3 {
		return fmt.Errorf("value needs to be 3 bytes, got %d", vl)
	}
	m.PortNo = Port(value[0])
	m.NativeVlan.value = binary.BigEndian.Uint16(value[1:3])
	return nil
}
