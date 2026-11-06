package tlv

import (
	"fmt"
	"slices"
	"sync"
)

const DefaultChunkSize = 1000

type EmptyListError struct{}

func (EmptyListError) Error() string {
	return fmt.Sprintf("list of %T is empty", TLV{})
}

type IncompleteListError struct{}

func (icl IncompleteListError) Error() string {
	return fmt.Sprintf("list of %T is incomplete", TLV{})
}

type NonUniformListError struct {
	base  any
	add   any
	index int
}

func (icl NonUniformListError) Error() string {
	return fmt.Sprintf("inconsistent %T type can't add %v to %v at position %d", TLV{}, icl.add, icl.base, icl.index)
}

type FragmentList struct {
	mux       sync.RWMutex
	fragments []*TLV
}

// Assign stores a fragment at its zero-based protocol offset.
func (f *FragmentList) Assign(index uint16, fragment *TLV) (complete bool) {
	if f == nil {
		return false
	}
	f.mux.Lock()
	defer f.mux.Unlock()

	fragments := f.fragments
	if grow := int(index) - (len(fragments) - 1); grow > 0 {
		fragments = append(fragments, make([]*TLV, grow)...)
	}

	fragments[index] = cloneTLV(fragment)
	complete = slices.Index(fragments, nil) == -1 //no nil item found
	f.fragments = fragments
	return
}

func (f *FragmentList) Length() int {
	if f == nil {
		return 0
	}
	f.mux.RLock()
	defer f.mux.RUnlock()
	return len(f.fragments)
}

// Item returns a copy of the fragment at its zero-based protocol offset.
func (f *FragmentList) Item(index uint) *TLV {
	if f == nil {
		return nil
	}
	f.mux.RLock()
	defer f.mux.RUnlock()
	if index >= uint(len(f.fragments)) {
		return nil
	}
	return cloneTLV(f.fragments[index])
}

func (f *FragmentList) Complete() bool {
	if f == nil {
		return false
	}
	f.mux.RLock()
	defer f.mux.RUnlock()
	return len(f.fragments) > 0 && slices.Index(f.fragments, nil) == -1
}

func (fl *FragmentList) Join() (TLV, error) {
	if fl == nil {
		return TLV{}, EmptyListError{}
	}

	fl.mux.RLock()
	if len(fl.fragments) == 0 {
		fl.mux.RUnlock()
		return TLV{}, EmptyListError{}
	}
	if firstEmpty := slices.Index(fl.fragments, nil); firstEmpty != -1 {
		fl.mux.RUnlock()
		return TLV{}, IncompleteListError{}
	}
	fragments := make([]*TLV, len(fl.fragments))
	for i, fragment := range fl.fragments {
		fragments[i] = cloneTLV(fragment)
	}
	fl.mux.RUnlock()

	result := TLV{
		Type: fragments[0].Type,
	}

	for i := len(fragments); i > 0; i-- {
		if ft := fragments[i-1].Type; ft != result.Type {
			return TLV{}, NonUniformListError{base: result.Type, add: ft, index: i - 1}
		} else {
			result.Value = append(result.Value, fragments[i-1].Value...)
		}
	}
	return result, nil
}

func cloneTLV(fragment *TLV) *TLV {
	if fragment == nil {
		return nil
	}
	return &TLV{
		Type:  fragment.Type,
		Value: append([]byte(nil), fragment.Value...),
	}
}

func NewFragmentList() *FragmentList {
	return &FragmentList{
		fragments: make([]*TLV, 0),
	}
}

// Fragment splits item.Value into chunks. Fragment slots are zero-based protocol
// offsets: the first chunk is at the highest offset and the last chunk at zero.
func Fragment(item TLV, customChunksize uint16) *FragmentList {
	chunkSize := DefaultChunkSize
	if customChunksize != 0 {
		chunkSize = int(customChunksize)
	}
	chunkCount := len(item.Value) / chunkSize
	if len(item.Value)%chunkSize != 0 {
		chunkCount += 1
	}
	fragments := make([]*TLV, chunkCount)
	for i := 0; i < chunkCount; i++ {
		end := chunkSize * (i + 1)
		if end > len(item.Value) {
			end = len(item.Value)
		}
		fragments[uint16(chunkCount-i-1)] = &TLV{
			Type:  item.Type,
			Value: append([]byte(nil), item.Value[chunkSize*i:end]...),
		}
	}
	return &FragmentList{
		fragments: fragments,
	}
}
