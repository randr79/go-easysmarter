package typed

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
)

type Model string

func (Model) Type() Type { return TypeModel }

func (m Model) Encode() ([]byte, error)    { return stringToBytes(m, 0, 65535) }
func (m *Model) Decode(value []byte) error { return bytesTostring(value, m) }

type Description string

func (Description) Type() Type { return TypeDescription }

func (m Description) Encode() ([]byte, error)    { return stringToBytes(m, 0, 32) }
func (m *Description) Decode(value []byte) error { return bytesTostring(value, m) }

type MacAddress net.HardwareAddr

func (MacAddress) Type() Type { return TypeMacAddress }

func (m MacAddress) Encode() ([]byte, error)    { return encodeBytes(m, 6) }
func (m *MacAddress) Decode(value []byte) error { return decodeBytes(value, 6, m) }

type FirmwareVersion string

func (FirmwareVersion) Type() Type { return TypeFirmwareVersion }

func (m FirmwareVersion) Encode() ([]byte, error)    { return stringToBytes(m, 0, 65535) }
func (m *FirmwareVersion) Decode(value []byte) error { return bytesTostring(value, m) }

type HardwareRevision string

func (HardwareRevision) Type() Type { return TypeHardwareRevision }

func (m HardwareRevision) Encode() ([]byte, error)    { return stringToBytes(m, 0, 65535) }
func (m *HardwareRevision) Decode(value []byte) error { return bytesTostring(value, m) }

type Autosaving bool

func (Autosaving) Type() Type { return TypeAutosaving }

func (m Autosaving) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *Autosaving) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type FactoryDefault bool

func (FactoryDefault) Type() Type { return TypeFactoryDefault }

func (m FactoryDefault) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *FactoryDefault) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type FlashTypeId uint32

func (FlashTypeId) Type() Type { return TypeFlashTypeId }
func (m FlashTypeId) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *FlashTypeId) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type PortOrderReversed bool

func (PortOrderReversed) Type() Type { return TypePortOrderReversed }

func (m PortOrderReversed) Encode() ([]byte, error) { return binary.Append(nil, binary.BigEndian, m) }
func (m *PortOrderReversed) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type DhcpEnabled bool

func (DhcpEnabled) Type() Type { return TypeDhcpEnabled }

func (m DhcpEnabled) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *DhcpEnabled) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type IpAddress net.IP

func (v IpAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(v).String())
}
func (v *IpAddress) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	} else {
		*v = IpAddress(net.ParseIP(str))
	}
	return nil
}

func (IpAddress) Type() Type { return TypeIpAddress }

func (m IpAddress) Encode() ([]byte, error)    { return encodeBytes(m, 4) }
func (m *IpAddress) Decode(value []byte) error { return decodeBytes(value, 4, m) }

type SubnetMask net.IP

func (v SubnetMask) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(v).String())
}
func (v *SubnetMask) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	} else {
		*v = SubnetMask(net.ParseIP(str))
	}
	return nil
}

func (SubnetMask) Type() Type { return TypeSubnetMask }

func (m SubnetMask) Encode() ([]byte, error)    { return encodeBytes(m, 4) }
func (m *SubnetMask) Decode(value []byte) error { return decodeBytes(value, 4, m) }

type GatewayAddress net.IP

func (v GatewayAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(v).String())
}
func (v *GatewayAddress) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	} else {
		*v = GatewayAddress(net.ParseIP(str))
	}
	return nil
}

func (GatewayAddress) Type() Type { return TypeGatewayAddress }

func (m GatewayAddress) Encode() ([]byte, error)    { return encodeBytes(m, 4) }
func (m *GatewayAddress) Decode(value []byte) error { return decodeBytes(value, 4, m) }

type ManagementVlan uint16

func (m ManagementVlan) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *ManagementVlan) Decode(value []byte) error {
	if vl := len(value); vl != 2 {
		return fmt.Errorf("value must be 2 bytes, got %d", vl)
	}
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e
}

type LedEnabled bool

func (LedEnabled) Type() Type { return TypeLedEnabled }

func (m LedEnabled) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *LedEnabled) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}
