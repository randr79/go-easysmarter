package typed

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"slices"
)

type IgmpSnoopingEnabled bool

func (IgmpSnoopingEnabled) Type() Type { return TypeIgmpSnoopingEnabled }

func (m IgmpSnoopingEnabled) Encode() ([]byte, error) { return binary.Append(nil, binary.BigEndian, m) }
func (m *IgmpSnoopingEnabled) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type IgmpReportSuppressionEnabled bool

func (IgmpReportSuppressionEnabled) Type() Type { return TypeIgmpReportSuppressionEnabled }

func (m IgmpReportSuppressionEnabled) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *IgmpReportSuppressionEnabled) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type IgmpMulticastIp struct {
	MulticastIP net.IP //4
	VlanID      uint16 //2
	Ports       Ports  //4
}

func (a IgmpMulticastIp) Encode() ([]byte, error) {
	var buffer []byte
	switch len(a.MulticastIP) {
	default:
		return nil, errors.New("only ipv4 is supported")
	case 0:
		buffer = make([]byte, 4)
	case 4:
		buffer = append([]byte{}, a.MulticastIP...)
	}
	buffer = binary.BigEndian.AppendUint16(buffer, a.VlanID)
	return binary.BigEndian.AppendUint32(buffer, uint32(a.Ports)), nil
}
func (a *IgmpMulticastIp) Decode(value []byte) error {
	if vl := len(value); vl != 10 {
		return fmt.Errorf("input needs to be a 10 bytes, got %d", vl)
	} else {
		*a = IgmpMulticastIp{
			MulticastIP: value[0:4],
			VlanID:      binary.BigEndian.Uint16(value[4:6]),
			Ports:       Ports(binary.BigEndian.Uint32(value[6:10])),
		}
	}
	return nil
}

type IgmpMulticastIps []IgmpMulticastIp

func (IgmpMulticastIps) Type() Type { return TypeIgmpMulticastIps }

func (a IgmpMulticastIps) Encode() ([]byte, error) {
	var all []byte
	var errs []error
	for _, m := range a {
		if buffer, err := m.Encode(); err != nil {
			errs = append(errs, err)
		} else {
			all = append(all, buffer...)
		}
	}
	if len(errs) == 0 {
		return all, nil
	} else {
		return nil, errors.Join(errs...)
	}
}
func (m *IgmpMulticastIps) Decode(value []byte) error {
	var ips []IgmpMulticastIp
	var errs []error
	if vl := len(value); vl%10 != 0 {
		return fmt.Errorf("input needs to be a multiple of 10 bytes, got %d", vl)
	} else {
		ips = make([]IgmpMulticastIp, 0, vl/10)
	}
	for c := range slices.Chunk(value, 10) {
		ip := IgmpMulticastIp{}
		if err := ip.Decode(c); err != nil {
			errs = append(errs, err)
		} else {
			ips = append(ips, ip)
		}
	}
	*m = ips
	return nil
}
