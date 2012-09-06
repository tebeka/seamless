package main

import (
	"fmt"
	"strings"
	"sync"
)

// List of backends, support (thread safe) adding, removing and getting next in list (circular)
type Backends struct {
	backends []string
	current  int
	lock     sync.Mutex
}

// Set sets the current list of backends
func (bs *Backends) Set(backends []string) {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	bs.backends = backends
	bs.current = 0
}

// Next returns the next back in circular fashion
func (bs *Backends) Next() (string, error) {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	if len(bs.backends) == 0 {
		return "", fmt.Errorf("empty backends")
	}

	// We advance first to make sure we're in bounds
	bs.current = (bs.current + 1) % len(bs.backends)
	backend := bs.backends[bs.current]
	return backend, nil
}

// Add adds a new backend
func (bs *Backends) Add(backend string) {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	bs.backends = append(bs.backends, backend)
}

// Remove removes all occurrences of backend from list of backends, returns the number of items removed
func (bs *Backends) Remove(backend string) int {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	i, count := 0, 0
	for i < len(bs.backends) {
		if bs.backends[i] == backend {
			count++
			bs.backends = append(bs.backends[:i], bs.backends[i+1:]...)
		} else {
			i++
		}
	}

	return count
}

// String is string representation of backends
func (bs *Backends) String() string {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	return strings.Join(bs.backends, ",")
}
