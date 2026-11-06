package typed

import (
	"encoding/binary"
	"fmt"
)

const maxUint32 = uint64(1<<32 - 1)

type BandwidthRate uint32

func (m BandwidthRate) Kbps() uint64 { return uint64(m) * 64 }
func (m BandwidthRate) Aprox(kbps uint64) {
	multiples := kbps >> 6
	if multiples > maxUint32 {
		m = BandwidthRate(maxUint32)
	} else {
		m = BandwidthRate(multiples)
	}

}

type QosMode byte

const (
	QosModePort  QosMode = 0
	QosMode8021P QosMode = 1
	QosModeDscp  QosMode = 2
)

type QosType QosMode

func (QosType) Type() Type { return TypeQosType }

func (m QosType) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *QosType) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type PortPriority byte

const (
	PortPriorityLowest  PortPriority = 0
	PortPriorityNormal  PortPriority = 1
	PortPriorityMedium  PortPriority = 2
	PortPriorityHighest PortPriority = 3
)

type QosPortPriority struct {
	PortNo   Port
	Priority PortPriority
}

func (QosPortPriority) Type() Type { return TypeQosPortPriority }

func (m *QosPortPriority) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e

}
func (m QosPortPriority) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}

type StormControl struct {
	PortNo           Port
	Enabled          bool
	UplinkEnabled    bool
	MulticastEnabled bool
	BroadcastEnabled bool
	Rate             BandwidthRate
}

func (StormControl) Type() Type { return TypeStormControl }

func (m *StormControl) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, value)
	return e
}
func (m StormControl) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}

type BandwidthLimit struct {
	PortNo  Port
	Enabled bool
	Rate    BandwidthRate
}

func (m *BandwidthLimit) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e
}
func (m BandwidthLimit) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}

type BandwidthLimitIngress struct{ BandwidthLimit }

func (BandwidthLimitIngress) Type() Type { return TypeBandwidthLimitIngress }

type BandwidthLimitEgress struct{ BandwidthLimit }

func (BandwidthLimitEgress) Type() Type { return TypeBandwidthLimitEgress }
