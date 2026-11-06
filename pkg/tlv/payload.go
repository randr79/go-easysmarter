package tlv

import "fmt"

// MaxPayloadSize is the maximum sum of TLV value bytes sent without fragmentation.
const MaxPayloadSize = 1400

// EncodedFragment pairs a transport fragment offset with its encoded TLV body.
type EncodedFragment struct {
	Offset uint16
	Body   []byte
}

// EncodeFragments encodes TLVs as one or more bodies. Oversized payloads must
// contain exactly one TLV, which is split into the default 1000-byte chunks.
func EncodeFragments(items []TLV) ([]EncodedFragment, error) {
	payloadSize := 0
	for _, item := range items {
		if item.Length() > maxLen {
			return nil, InvalidPayloadError{
				format: "maximum tlv size is %d, got %d",
				params: []any{maxLen, item.Length()},
			}
		}
		payloadSize += item.Length()
	}

	if payloadSize <= MaxPayloadSize {
		body, err := Encode(items)
		if err != nil {
			return nil, err
		}
		return []EncodedFragment{{Body: body}}, nil
	}

	if len(items) != 1 {
		return nil, fmt.Errorf(
			"cannot fragment payload of %d bytes containing %d TLVs; only a single TLV can be fragmented",
			payloadSize,
			len(items),
		)
	}

	fragments := Fragment(items[0], 0)
	encoded := make([]EncodedFragment, 0, fragments.Length())
	for offset := fragments.Length() - 1; offset >= 0; offset-- {
		body, err := Encode([]TLV{*fragments.Item(uint(offset))})
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, EncodedFragment{
			Offset: uint16(offset),
			Body:   body,
		})
	}
	return encoded, nil
}
