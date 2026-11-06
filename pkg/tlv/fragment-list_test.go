package tlv

import (
	"bytes"
	"errors"
	"sync"
	"testing"
)

func TestFragmentListAssignsOutOfOrderFragments(t *testing.T) {
	defragmenter := NewFragmentList()
	if defragmenter.Complete() {
		t.Fatal("new defragmenter is complete")
	}
	if defragmenter.Assign(2, &TLV{Type: 7, Value: []byte("a")}) {
		t.Fatal("defragmenter is complete with missing fragments")
	}
	if defragmenter.Assign(0, &TLV{Type: 7, Value: []byte("c")}) {
		t.Fatal("defragmenter is complete with fragment 1 missing")
	}
	if _, err := defragmenter.Join(); err == nil {
		t.Fatal("Join() succeeded with a missing fragment")
	} else {
		var incomplete IncompleteListError
		if !errors.As(err, &incomplete) {
			t.Fatalf("Join() error = %v, want incomplete fragment", err)
		}
	}
	if !defragmenter.Assign(1, &TLV{Type: 7, Value: []byte("b")}) {
		t.Fatal("defragmenter is incomplete after receiving all fragments")
	}
	if !defragmenter.Complete() {
		t.Fatal("defragmenter is incomplete after receiving all fragments")
	}

	joined, err := defragmenter.Join()
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if joined.Type != 7 || !bytes.Equal(joined.Value, []byte("abc")) {
		t.Errorf("joined TLV = %+v, want type 7 and value %q", joined, "abc")
	}
}

func TestFragmentListJoinErrors(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var defragmenter FragmentList
		_, err := defragmenter.Join()
		var empty EmptyListError
		if !errors.As(err, &empty) {
			t.Errorf("Join() error = %v, want EmptyListError", err)
		}
	})

	t.Run("non-uniform types", func(t *testing.T) {
		defragmenter := NewFragmentList()
		defragmenter.Assign(0, &TLV{Type: 7, Value: []byte("first")})
		defragmenter.Assign(1, &TLV{Type: 8, Value: []byte("second")})
		_, err := defragmenter.Join()
		var mismatch NonUniformListError
		if !errors.As(err, &mismatch) {
			t.Fatalf("Join() error = %v, want NonUniformListError", err)
		}
		if mismatch.index != 1 {
			t.Errorf("mismatched fragment index = %d, want 1", mismatch.index)
		}
	})
}

func TestFragmentListUsesZeroBasedProtocolOffsets(t *testing.T) {
	value := []byte("abcdefghij")
	fragments := Fragment(TLV{Type: 42, Value: value}, 4)
	if got := fragments.Item(2); got == nil || !bytes.Equal(got.Value, []byte("abcd")) {
		t.Fatalf("fragment at offset 2 = %#v, want the first chunk", got)
	}
	if got := fragments.Item(0); got == nil || !bytes.Equal(got.Value, []byte("ij")) {
		t.Fatalf("fragment at offset 0 = %#v, want the final chunk", got)
	}

	received := NewFragmentList()
	for offset := uint16(2); ; offset-- {
		if !received.Assign(offset, fragments.Item(uint(offset))) {
			if offset == 0 {
				t.Fatal("defragmenter is incomplete after assigning final offset 0")
			}
		} else if offset != 0 {
			t.Fatalf("defragmenter completed at offset %d, want completion at offset 0", offset)
		}
		if offset == 0 {
			break
		}
	}
	joined, err := received.Join()
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if !bytes.Equal(joined.Value, value) {
		t.Errorf("reassembled value = %q, want %q", joined.Value, value)
	}
}

func TestFragmentListCopiesAssignedAndReturnedValues(t *testing.T) {
	value := []byte("original")
	fragment := &TLV{Type: 42, Value: value}
	fragments := NewFragmentList()
	if !fragments.Assign(0, fragment) {
		t.Fatal("single-fragment list is incomplete")
	}

	value[0] = 'X'
	fragment.Value[1] = 'X'
	got := fragments.Item(0)
	if got == nil || string(got.Value) != "original" {
		t.Fatalf("stored fragment = %#v, want an isolated copy of the assigned value", got)
	}
	got.Value[2] = 'X'
	if again := fragments.Item(0); again == nil || string(again.Value) != "original" {
		t.Errorf("mutating Item() result changed the stored fragment: %#v", again)
	}
}

func TestFragmentListItemOutOfRange(t *testing.T) {
	fragments := NewFragmentList()
	if got := fragments.Item(0); got != nil {
		t.Errorf("Item(0) = %#v, want nil for an out-of-range index", got)
	}
}

func TestFragmentListConcurrentAccess(t *testing.T) {
	fragments := NewFragmentList()
	var workers sync.WaitGroup
	for offset := uint16(0); offset < 4; offset++ {
		workers.Add(1)
		go func(offset uint16) {
			defer workers.Done()
			for i := 0; i < 100; i++ {
				fragments.Assign(offset, &TLV{Type: 42, Value: []byte{byte(i)}})
				fragments.Item(uint(offset))
				fragments.Length()
				fragments.Complete()
				_, _ = fragments.Join()
			}
		}(offset)
	}
	workers.Wait()

	if !fragments.Complete() {
		t.Fatal("fragment list is incomplete after assigning every offset")
	}
}

func TestFragmenterFragmentsAtChunkBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  map[uint][]byte
	}{
		{
			name:  "exact chunks",
			value: []byte("abcdefgh"),
			want: map[uint][]byte{
				1: []byte("abcd"),
				0: []byte("efgh"),
			},
		},
		{
			name:  "partial final chunk",
			value: []byte("abcdefghij"),
			want: map[uint][]byte{
				2: []byte("abcd"),
				1: []byte("efgh"),
				0: []byte("ij"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Fragment(TLV{Type: 42, Value: test.value}, 4)
			if got.Length() != len(test.want) {
				t.Fatalf("got %d fragments, want %d", got.Length(), len(test.want))
			}

			for index, wantValue := range test.want {
				fragment := got.Item(index)
				if fragment == nil {
					t.Errorf("missing fragment at index %d", index)
					continue
				}
				if fragment.Type != 42 {
					t.Errorf("fragment %d has type %d, want 42", index, fragment.Type)
				}
				if !bytes.Equal(fragment.Value, wantValue) {
					t.Errorf("fragment %d has value %q, want %q", index, fragment.Value, wantValue)
				}
			}
		})
	}
}

func TestFragmentUsesDefaultChunkSize(t *testing.T) {
	value := make([]byte, DefaultChunkSize+1)
	for index := range value {
		value[index] = byte(index)
	}

	fragments := Fragment(TLV{Type: 42, Value: value}, 0)
	if fragments.Length() != 2 {
		t.Fatalf("got %d fragments, want 2", fragments.Length())
	}
	if got := fragments.Item(1).Value; len(got) != DefaultChunkSize || !bytes.Equal(got, value[:DefaultChunkSize]) {
		t.Errorf("first chunk has length %d or unexpected content", len(got))
	}
	if got := fragments.Item(0).Value; len(got) != 1 || !bytes.Equal(got, value[DefaultChunkSize:]) {
		t.Errorf("final chunk has length %d or unexpected content", len(got))
	}
}
