// Copyright 2014 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by a BSD-style
// license found in the LICENSE file.

// Package memkv implements an in-memory key/value store.
package memkv

import (
	"path/filepath"
	"sort"
	"sync"
)

// A Store represents an in-memory key-value store safe for
// concurrent access.
type Store struct {
	sync.RWMutex
	m map[string]Node
}

// New creates and initializes a new Store.
func New() Store {
	return Store{m: make(map[string]Node)}
}

// Get gets the Node associated with key. If there are no Nodes
// associated with key, Get returns Node{}, false.
func (s Store) Get(key string) (Node, bool) {
	s.RLock()
	n, ok := s.m[key]
	s.RUnlock()
	return n, ok
}

// Glob returns a memkv.Node for all nodes with keys matching pattern.
// The syntax of patterns is the same as in filepath.Match.
func (s Store) Glob(pattern string) (Nodes, error) {
	ns := make(Nodes, 0)
	s.RLock()
	defer s.RUnlock()
	for _, n := range s.m {
		m, err := filepath.Match(pattern, n.Key)
		if err != nil {
			return nil, err
		}
		if m {
			ns = append(ns, n)
		}
	}
	sort.Sort(ns)
	return ns, nil
}

// Set sets the node entry associated with key to value. It replaces
// any existing values associated with key.
func (s Store) Set(key string, value string) {
	s.Lock()
	s.m[key] = Node{key, value}
	s.Unlock()
}
