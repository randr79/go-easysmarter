package typed

import (
	"encoding/binary"
	"fmt"
)

type Port uint8
type LagGroup uint8
type PortSpeed uint8

const (
	SpeedLinkDown PortSpeed = 0
	SpeedAuto     PortSpeed = 1
	Speed10MH     PortSpeed = 2 // 10Mbps Half Duplex
	Speed10MF     PortSpeed = 3 // 10Mbps Full Duplex
	Speed100MH    PortSpeed = 4 // 100Mbps Half Duplex
	Speed100MF    PortSpeed = 5 // 100Mbps Full Duplex
	Speed1000MF   PortSpeed = 6 // 1000Mbps Full Duplex
)

type Ports uint32

func (Ports) max() uint8 { return 32 }
func (p *Ports) Set(number uint8, on bool) error {
	if number > 32 {
		return fmt.Errorf("port must be in the range 1-32, %d is not valid", number)
	}
	mask := Ports(1) << (number - 1)
	if on {
		*p |= mask
	} else {
		*p &= ^mask
	}
	return nil
}
func (p *Ports) Get(number uint8) (bool, error) {
	if number > 32 {
		return false, fmt.Errorf("port must be in the range 1-32, %d is not valid", number)
	}

	mask := Ports(1) << (number - 1)
	return *p&mask == mask, nil
}

func (p *Ports) GetAll() []Port {
	all := make([]Port, 0)
	pm := p.max()
	for i := uint8(0); i < pm; i++ {
		mask := Ports(1) << i
		if *p&mask == mask {
			all = append(all, Port(i+1))
		}
	}
	return all
}

type PortCount struct {
	value uint8
}

func (PortCount) Type() Type { return TypePortCount }

func NewPortCount(value uint8) (*PortCount, error) {
	if value > 32 {
		return nil, fmt.Errorf("portcount must be in the range 1-32, %d is not valid", value)
	}
	return &PortCount{
		value: value,
	}, nil
}

func (m PortCount) Get() uint8 { return m.value }
func (m PortCount) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m.value)
}
func (m *PortCount) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, &m.value); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type PortMirroring struct {
	Enabled         bool
	MirrorPort      Port
	IngressPortMask Ports
	EgressPortMask  Ports
}

func (PortMirroring) Type() Type { return TypePortMirroring }

func (m PortMirroring) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *PortMirroring) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e
}

type PortStatistics struct {
	PortNo    Port      // Byte 0: Poortnummer (Hex 01 t/m 10)
	Enabled   bool      // Byte 1: Admin status (0=Disabled, 1=Enabled)
	CurSpeed  PortSpeed // Byte 2: Actuele linksnelheid (0=Down, 6=1000MF, etc.)
	TxGoodPkt uint32    // Bytes 3-6: Goede verzonden pakketten
	TxBadPkt  uint32    // Bytes 7-10: Foutieve verzonden pakketten
	RxGoodPkt uint32    // Bytes 11-14: Goede ontvangen pakketten
	RxBadPkt  uint32    // Bytes 15-18: Foutieve ontvangen pakketten
}

func (PortStatistics) Type() Type { return TypePortStatistics }

func (m PortStatistics) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *PortStatistics) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e
}

type PortSettings struct {
	PortNo             Port      // Byte 0: Poortnummer (Hex 01 t/m 10)
	Enabled            bool      // Byte 1: Admin status (0=Disabled, 1=Enabled)
	LagID              LagGroup  // Byte 2: Trunk/LAG Groep ID (0=Geen, 1=LAG1, 2=LAG2, etc.)
	SetSpeed           PortSpeed // Byte 3: Geconfigureerde snelheid
	CurSpeed           PortSpeed // Byte 4: Actuele linksnelheid
	FlowControlEnabled bool      // Byte 5: Geconfigureerde Flow Control
	FlowControlActive  bool      // Byte 6: Actuele Flow Control Status onderhandeling
}

func (PortSettings) Type() Type { return TypePortSettings }

func (m PortSettings) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *PortSettings) Decode(value []byte) error {
	if vl := len(value); vl != 7 {
		return fmt.Errorf("value must be 7 bytes, got %d", vl)
	}
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e
}

type LoopPreventionEnabled bool

func (LoopPreventionEnabled) Type() Type { return TypeLoopPreventionEnabled }

func (m LoopPreventionEnabled) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *LoopPreventionEnabled) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e
}

type LagConfig struct {
	Group    LagGroup
	PortMask Ports
}

func (LagConfig) Type() Type { return TypeLagConfig }

func (m *LagConfig) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e
}

func (m LagConfig) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
