package transport

import (
	"slices"
	"testing"
)

func TestSyncMapOperations(t *testing.T) {
	items := &syncMap[string, int]{items: make(map[string]int)}

	if got := items.GetDefault("one", 1); got != 1 {
		t.Fatalf("GetDefault() = %d, want 1", got)
	}
	if got := items.GetDefault("one", 2); got != 1 {
		t.Fatalf("GetDefault() on existing key = %d, want 1", got)
	}
	if got := items.Has("one"); got != 1 {
		t.Fatalf("Has() = %d, want 1", got)
	}

	items.Set("two", 2)
	if got, ok := items.Get("two"); !ok || got != 2 {
		t.Fatalf("Get() = (%d, %t), want (2, true)", got, ok)
	}
	if _, ok := items.Get("missing"); ok {
		t.Fatal("Get() reported a missing key as present")
	}

	if got := items.Keys(); len(got) != 2 || !slices.Contains(got, "one") || !slices.Contains(got, "two") {
		t.Fatalf("Keys() = %v, want [one two]", got)
	}
	snapshot := items.Items()
	delete(snapshot, "one")
	if got := items.Has("one"); got != 1 {
		t.Fatalf("mutating Items() result changed original map; Has(one) = %d", got)
	}

	items.Apply(func(key string, value int) int {
		if key == "one" {
			return value + 10
		}
		return value * 2
	})
	if got := items.Has("one"); got != 11 {
		t.Errorf("Apply() value for one = %d, want 11", got)
	}
	if got := items.Has("two"); got != 4 {
		t.Errorf("Apply() value for two = %d, want 4", got)
	}

	if !items.Expire("one", func(value int) bool { return value == 11 }) {
		t.Error("Expire() returned false for an existing key")
	}
	if _, ok := items.Get("one"); ok {
		t.Error("Expire() did not remove the key when callback returned true")
	}
	if !items.Expire("two", func(int) bool { return false }) {
		t.Error("Expire() returned false for an existing key")
	}
	if got := items.Has("two"); got != 4 {
		t.Errorf("Expire() removed key when callback returned false; got %d", got)
	}
	if items.Expire("missing", func(int) bool { return true }) {
		t.Error("Expire() returned true for a missing key")
	}

	if got, ok := items.Pop("two"); !ok || got != 4 {
		t.Errorf("Pop() = (%d, %t), want (4, true)", got, ok)
	}
	if _, ok := items.Pop("two"); ok {
		t.Error("Pop() reported a removed key as present")
	}
	items.Del("missing")
	items.Set("three", 3)
	items.Del("three")
	if _, ok := items.Get("three"); ok {
		t.Error("Del() did not remove the key")
	}
}
