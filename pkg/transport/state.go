package transport

import (
	"encoding/binary"
	"sync"

	"github.com/randr79/go-easysmarter/pkg/crypt/rc4"
)

const wireHeaderSize = 32

type wireHeader struct {
	Version   byte    // 1 bytes
	Method    byte    // 1 bytes
	SwitchMac [6]byte // 6 bytes
	ClientMac [6]byte // 6 bytes
	RequestId uint16  // 2 bytes
	Status    uint32  // 4 bytes
	Length    uint16  // 2 bytes; total message length
	Fragment  uint16  // 2 bytes
	Flags     uint16  // 2 bytes; reserved; always zero
	TokenId   uint16  // 2 bytes
	CRC       uint32  // 4 bytes; reserved; always zero
} //                    32 bytes

type stateTable = syncMap[MACAddress, *State]

func NewStateTable(defaultKey []byte) (*stateTable, error) {
	if rc4key, err := rc4.NewRc4Key(defaultKey); err != nil {
		return nil, err
	} else {
		return &stateTable{
			items: map[MACAddress]*State{
				{0, 0, 0, 0, 0, 0}: {
					Encryption: rc4key,
				},
			},
		}, nil
	}
}

type State struct {
	mux        sync.RWMutex
	Encryption *rc4.Rc4Key
	TokenId    uint16
}

func (c *State) decode(data []byte) (*wireHeader, []byte, error) {
	var header wireHeader
	decoded := c.Encryption.XORKey(data)
	if used, err := binary.Decode(decoded, binary.BigEndian, &header); err != nil {
		return nil, nil, err
	} else if used != wireHeaderSize {
		return nil, nil, InvalidWireHeaderSizeError{Got: used, Want: wireHeaderSize}
	} else if header.Version != 1 || int(header.Length) > len(data) || header.Method > 4 || header.CRC != 0 || header.Flags != 0 {
		return nil, nil, nil
	} else {
		return &header, decoded[wireHeaderSize:], nil
	}
}

func (s *State) encode(head wireHeader, body []byte) ([]byte, error) {
	if cl := len(body); cl > maxContentLength {
		return nil, ContentLengthExceededError(cl)
	} else {
		head.Length = uint16(cl + 32)
	}

	if data, err := binary.Append(nil, binary.BigEndian, head); err != nil {
		return nil, err
	} else {
		return s.Encryption.XORKey(append(data, body...)), nil
	}
}
