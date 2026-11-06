package typed

import (
	"encoding/binary"
	"fmt"
	"slices"
)

type FactoryReset struct{}

func (FactoryReset) Type() Type { return TypeFactoryReset }

func (FactoryReset) Encode() ([]byte, error)      { return nil, nil }
func (m *FactoryReset) Decode(value []byte) error { return noContent(value, m) }

type Reboot struct {
	PersistConfig bool
}

func (Reboot) Type() Type { return TypeReboot }

func (m Reboot) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *Reboot) Decode(value []byte) error {
	if i, err := binary.Decode(value, binary.BigEndian, m); err != nil {
		return err
	} else if vl := len(value); vl != i {
		return fmt.Errorf("value needs to be %d bytes, got %d", i, vl)
	} else {
		return nil
	}
}

type ConfigData []byte

func (m ConfigData) Encode() ([]byte, error)    { return m, nil }
func (m *ConfigData) Decode(value []byte) error { *m = value; return nil }

type ReadConfig struct{ ConfigData }

func (ReadConfig) Type() Type { return TypeReadConfig }

type WriteConfig struct{ ConfigData }

func (WriteConfig) Type() Type { return TypeWriteConfig }

type RebootToFirmwareUpgradeMode struct{}

func (RebootToFirmwareUpgradeMode) Type() Type { return TypeRebootToFirmwareUpgradeMode }

func (m RebootToFirmwareUpgradeMode) Encode() ([]byte, error)    { return nil, nil }
func (m *RebootToFirmwareUpgradeMode) Decode(value []byte) error { return noContent(value, m) }

type CableTestResult byte

const (
	CableTestResultNoCable   CableTestResult = 0
	CableTestResultNormal    CableTestResult = 1
	CableTestResultOpen      CableTestResult = 2
	CableTestResultShort     CableTestResult = 3
	CableTestResultOpenShort CableTestResult = 4
	CableTestResultCrossTalk CableTestResult = 5
)

type CableTest struct {
	PortNo   Port
	Result   CableTestResult
	Distance uint32
}

func (CableTest) Type() Type { return TypeCableTest }

func (m CableTest) Encode() ([]byte, error) {
	return binary.Append(nil, binary.BigEndian, m)
}
func (m *CableTest) Decode(value []byte) error {
	_, e := binary.Decode(value, binary.BigEndian, m)
	return e
}

type GetToken struct{}

func (GetToken) Type() Type { return TypeGetToken }

func (m *GetToken) Encode() ([]byte, error)   { return nil, nil }
func (m *GetToken) Decode(value []byte) error { return noContent(value, m) }

type PersistConfig struct{}

func (PersistConfig) Type() Type { return TypePersistConfig }

func (PersistConfig) Encode() ([]byte, error)      { return nil, nil }
func (m *PersistConfig) Decode(value []byte) error { return noContent(value, m) }

type RsaChallenge struct {
	Limbs []uint16
}

func (RsaChallenge) Type() Type { return TypeRsaChallenge }

func (m RsaChallenge) Encode() ([]byte, error) {
	uint16count := len(m.Limbs) + 1
	res := make([]byte, uint16count*2)
	binary.BigEndian.PutUint16(res[0:2], uint16(uint16count))
	for i, limb := range m.Limbs {
		start := (i + 1) * 2
		binary.BigEndian.PutUint16(res[start:start+2], limb)
	}
	return res, nil
}

func (m *RsaChallenge) Decode(value []byte) error {
	if vl := len(value); vl&1 == 1 || vl < 4 {
		return fmt.Errorf("incorrect number of bytes, we need an even number of bytes and at least 4 bytes, got %d", vl)
	}
	parts := slices.Chunk(value, 2)
	uints := make([]uint16, 0, len(value)>>1)
	for i := range parts {
		uints = append(uints, binary.BigEndian.Uint16(i))
	}
	if pl, al := int(uints[0]), len(uints); pl != al {
		return fmt.Errorf("prefixed length %d does not equal the number of items (%d)", pl, al)
	}

	m.Limbs = uints[1:]
	return nil
}
