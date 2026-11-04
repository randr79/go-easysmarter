package rc4

import "fmt"

type InvalidKeyError [3]int

func (k InvalidKeyError) Error() string {
	return fmt.Sprintf("easysmart/crypt: invalid key, length must be  %d <= `%d` <= %d", k[0], k[1], k[2])
}

type Rc4Key struct {
	state []uint32
}

// this rc variant does not maintain state between encoding of multiple XORKey calls and
// uses a variable sized state (len(key)) instead of concatenating the input key up to 256 bytes
func NewRc4Key(key []byte) (*Rc4Key, error) {
	keyLen := uint32(len(key))

	if keyLen < 1 || keyLen > 256 {
		return nil, InvalidKeyError{int(keyLen), 1, 256}
	}

	state := make([]uint32, keyLen)
	for index := uint32(0); index < keyLen; index++ {
		state[index] = index
	}

	var keyIndex uint32
	for index := uint32(0); index < keyLen; index++ {
		keyIndex += state[index] + uint32(key[index])
		state[index], state[keyIndex%keyLen] = state[keyIndex%keyLen], state[index]
	}
	return &Rc4Key{state: state}, nil
}

func (s Rc4Key) XORKey(data []byte) []byte {
	state := append([]uint32{}, s.state...) //always start from a copy
	keyLen := uint32(len(state))

	output := make([]byte, len(data))
	var stateIndex uint32
	var keyIndex uint32
	for index, value := range data {
		stateIndex = (stateIndex + 1) % keyLen
		keyIndex = (keyIndex + state[stateIndex]) % keyLen
		state[stateIndex], state[keyIndex] = state[keyIndex], state[stateIndex]
		streamIndex := (state[stateIndex] + state[keyIndex]) % keyLen
		output[index] = value ^ byte(state[streamIndex])
	}

	return output
}
