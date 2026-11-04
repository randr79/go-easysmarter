package transport

import (
	"maps"
	"slices"
	"sync"
)

type syncMap[K comparable, V any] struct {
	mux   sync.RWMutex
	items map[K]V
}

func (s *syncMap[K, V]) GetDefault(key K, defaultValue V) V {
	defer s.mux.Unlock()
	s.mux.Lock()
	v, ok := s.items[key]
	if !ok {
		s.items[key] = defaultValue
		v = defaultValue
	}
	return v
}

func (s *syncMap[K, V]) Get(key K) (V, bool) {
	defer s.mux.RUnlock()
	s.mux.RLock()
	v, ok := s.items[key]
	return v, ok
}

func (s *syncMap[K, V]) Expire(key K, expired func(v V) bool) bool {
	defer s.mux.Unlock()
	s.mux.Lock()
	if v, ok := s.items[key]; !ok {
		return false
	} else if expired(v) {
		delete(s.items, key)
	}
	return true

}

func (s *syncMap[K, V]) Pop(key K) (V, bool) {
	defer s.mux.Unlock()
	s.mux.Lock()
	v, ok := s.items[key]
	delete(s.items, key)
	return v, ok
}

func (s *syncMap[K, V]) Has(key K) V {
	defer s.mux.RUnlock()
	s.mux.RLock()
	return s.items[key]
}

func (s *syncMap[K, V]) Set(key K, value V) {
	defer s.mux.Unlock()
	s.mux.Lock()
	s.items[key] = value
}

func (s *syncMap[K, V]) Del(key K) {
	defer s.mux.Unlock()
	s.mux.Lock()
	delete(s.items, key)
}

func (s *syncMap[K, V]) Keys() []K {
	defer s.mux.RUnlock()
	s.mux.RLock()
	return slices.Collect(maps.Keys(s.items))
}

func (s *syncMap[K, V]) Apply(fun func(K, V) V) {
	defer s.mux.Unlock()
	s.mux.Lock()

	keys := slices.Collect(maps.Keys(s.items))
	for _, k := range keys {
		s.items[k] = fun(k, s.items[k])
	}
}

func (s *syncMap[K, V]) Items() map[K]V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return maps.Clone(s.items)
}
