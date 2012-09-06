package main

import (
	"testing"
)

// FIXME: Sync tests

func TestBackendsSet(t *testing.T) {
	bs := &Backends{}
	if len(bs.backends) != 0 {
		t.Fatalf("should be empty")
	}

	bs.current = 107
	bs.Set([]string{"a", "b"})
	if len(bs.backends) != 2 {
		t.Fatalf("len should 2 (was %d)", len(bs.backends))
	}

	if bs.current != 0 {
		t.Fatalf("current should be 0 (is %d)", bs.current)
	}
}

func TestBackendsNext(t *testing.T) {
	backend1, backend2 := "localhost:8888", "localhost:8887"
	bs := &Backends{backends: []string{backend1, backend2}}

	for i, expected := range []string{backend2, backend1, backend2} {
		next, _ := bs.Next()
		if next != expected {
			t.Fatalf("backend should be %s at %d (was %s)", expected, i, next)
		}
	}

	bs.Set([]string{})
	_, err := bs.Next()
	if err == nil {
		t.Fatalf("managed to get backend from empty list")
	}
}

func TestBackendsAdd(t *testing.T) {
	bs := &Backends{}
	bs.Add("a")
	if len(bs.backends) != 1 {
		t.Fatalf("not added")
	}
}

func TestBackendsRemove(t *testing.T) {
	count := -1

	bs := &Backends{}
	count = bs.Remove("a")
	if count != 0 {
		t.Fatalf("removed from empty list")
	}

	bs.Set([]string{"a", "b", "a"})
	count = bs.Remove("a")
	if count != 2 {
		t.Fatalf("should remove 2 (was %d)", count)
	}
	if len(bs.backends) != 1 {
		t.Fatalf("should have left one (have %d)", len(bs.backends))
	}

	if bs.backends[0] != "b" {
		t.Fatalf("wrong one left (%s)", bs.backends[0])
	}
}

func TestBackendsString(t *testing.T) {
	bs := &Backends{}
	if bs.String() != "" {
		t.Fatalf("bad empty str: %s", bs.String())
	}

	bs.Set([]string{"a", "b"})
	if bs.String() != "a,b" {
		t.Fatalf("bad str %s (expected a,b)", bs.String())
	}
}
