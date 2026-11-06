package tlv

import (
	"bytes"
	"testing"
)

func TestEncodeFragmentsAtPayloadLimit(t *testing.T) {
	item := TLV{Type: 42, Value: bytes.Repeat([]byte{1}, MaxPayloadSize)}
	fragments, err := EncodeFragments([]TLV{item})
	if err != nil {
		t.Fatalf("EncodeFragments() error = %v", err)
	}
	if len(fragments) != 1 {
		t.Fatalf("EncodeFragments() returned %d fragments, want 1", len(fragments))
	}
	if fragments[0].Offset != 0 {
		t.Errorf("fragment offset = %d, want 0", fragments[0].Offset)
	}
	if got, want := len(fragments[0].Body), MaxPayloadSize+8; got != want {
		t.Errorf("body length = %d, want %d (value plus TLV header and terminator)", got, want)
	}
}

func TestEncodeFragmentsAbovePayloadLimit(t *testing.T) {
	item := TLV{Type: 42, Value: bytes.Repeat([]byte{1}, MaxPayloadSize+1)}
	fragments, err := EncodeFragments([]TLV{item})
	if err != nil {
		t.Fatalf("EncodeFragments() error = %v", err)
	}
	if len(fragments) != 2 {
		t.Fatalf("EncodeFragments() returned %d fragments, want 2", len(fragments))
	}

	defragmenter := NewFragmentList()
	for _, fragment := range fragments {
		items, err := Parse(fragment.Body)
		if err != nil {
			t.Fatalf("Parse(fragment %d) error = %v", fragment.Offset, err)
		}
		if len(items) != 1 {
			t.Fatalf("fragment %d has %d TLVs, want 1", fragment.Offset, len(items))
		}
		defragmenter.Assign(fragment.Offset, &items[0])
	}
	if !defragmenter.Complete() {
		t.Fatal("defragmenter is incomplete after receiving all encoded fragments")
	}
	joined, err := defragmenter.Join()
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if joined.Type != item.Type || !bytes.Equal(joined.Value, item.Value) {
		t.Errorf("reassembled TLV = %#v, want type %d with original value", joined, item.Type)
	}

	for _, fragment := range fragments {
		if len(fragment.Body) > DefaultChunkSize+8 {
			t.Errorf("fragment %d body length = %d, exceeds chunk, TLV header, and terminator",
				fragment.Offset, len(fragment.Body))
		}
	}
}

func TestEncodeFragmentsRejectsOversizedMultipleTLVs(t *testing.T) {
	items := []TLV{
		{Type: 1, Value: bytes.Repeat([]byte{1}, 701)},
		{Type: 2, Value: bytes.Repeat([]byte{2}, 701)},
	}
	if _, err := EncodeFragments(items); err == nil {
		t.Fatal("EncodeFragments() succeeded for an oversized multi-TLV payload, want an error")
	}
}

func TestEncodeFragmentsExcludesTLVHeadersAndTerminatorFromLimit(t *testing.T) {
	item := TLV{Type: 42, Value: bytes.Repeat([]byte{1}, MaxPayloadSize)}
	fragments, err := EncodeFragments([]TLV{item})
	if err != nil {
		t.Fatalf("EncodeFragments() error = %v", err)
	}
	if len(fragments) != 1 {
		t.Fatalf("EncodeFragments() returned %d fragments, want 1; TLV headers and terminator are outside the 1400-byte value threshold", len(fragments))
	}
}
