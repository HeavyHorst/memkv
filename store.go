package memkv

import (
	"path"
	"sort"
	"strings"
	"sync"

	radix "github.com/armon/go-radix"
)

// A Store represents an in-memory key-value store safe for
// concurrent access.
type Store struct {
	FuncMap map[string]interface{}
	sync.RWMutex
	t *radix.Tree
}

func New() *Store {
	s := &Store{
		t:       radix.New(),
		RWMutex: sync.RWMutex{},
	}
	s.FuncMap = map[string]interface{}{
		"exists":    s.Exists,
		"ls":        s.List,
		"lsdir":     s.ListDir,
		"get":       s.Get,
		"gets":      s.GetAll,
		"getallkvs": s.GetAllKVs,
		"getv":      s.GetValue,
		"getvs":     s.GetAllValues,
	}
	return s
}

// Del deletes the KVPair associated with key.
func (s *Store) Del(key string) {
	s.Lock()
	s.t.Delete(key)
	s.Unlock()
}

// Exists checks for the existence of key in the store.
func (s *Store) Exists(key string) bool {
	s.RLock()
	_, ok := s.t.Get(key)
	s.RUnlock()
	return ok
}

// Get gets the KVPair associated with key. If there is no KVPair
// associated with key, Get returns KVPair{}.
func (s *Store) Get(key string) KVPair {
	s.RLock()
	defer s.RUnlock()
	data, ok := s.t.Get(key)
	if !ok {
		return KVPair{}
	}
	return data.(KVPair)
}

// GetAll returns a KVPair for all nodes with keys matching pattern.
// The syntax of patterns is the same as in path.Match.
func (s *Store) GetAll(pattern string) KVPairs {
	ks := make(KVPairs, 0)
	s.RLock()
	defer s.RUnlock()

	s.t.Walk(func(key string, value interface{}) bool {
		m, err := path.Match(pattern, key)
		if err != nil {
			return true
		}
		if m {
			ks = append(ks, value.(KVPair))
		}
		return false
	})

	if len(ks) == 0 {
		return nil
	}
	return ks
}

func (s *Store) GetAllValues(pattern string) []string {
	var vs []string
	for _, kv := range s.GetAll(pattern) {
		vs = append(vs, kv.Value)
	}
	sort.Strings(vs)
	return vs
}

// GetAllKVs returns all KV-Pairs
func (s *Store) GetAllKVs() KVPairs {
	ks := make(KVPairs, 0)
	s.RLock()
	defer s.RUnlock()

	s.t.Walk(func(key string, value interface{}) bool {
		ks = append(ks, value.(KVPair))
		return false
	})
	return ks
}

// GetValue gets the value associated with key. If there are no values
// associated with key, GetValue returns "".
func (s *Store) GetValue(key string, v ...string) string {
	defaultValue := ""
	if len(v) > 0 {
		defaultValue = v[0]
	}
	kv := s.Get(key)
	if kv.Key == "" {
		return defaultValue
	}
	return kv.Value
}

func (s *Store) list(filePath string, dir bool) []string {
	var vs []string
	m := make(map[string]bool)
	// The prefix search should only return dirs
	filePath = path.Clean(filePath) + "/"
	s.RLock()
	defer s.RUnlock()

	s.t.WalkPrefix(filePath, func(key string, value interface{}) bool {
		items := strings.Split(stripKey(key, filePath), "/")
		if dir {
			if len(items) < 2 {
				return false
			}
		}
		m[items[0]] = true
		return false
	})

	for k := range m {
		vs = append(vs, k)
	}
	sort.Strings(vs)
	return vs
}

//List returns all keys prefixed with filePath.
func (s *Store) List(filePath string) []string {
	return s.list(filePath, false)
}

//ListDir returns all directories prefixed with filePath.
func (s *Store) ListDir(filePath string) []string {
	return s.list(filePath, true)
}

// Set sets the KVPair entry associated with key to value.
func (s *Store) Set(key string, value string) {
	s.Lock()
	s.t.Insert(key, KVPair{key, value})
	s.Unlock()
}

//Purge removes all keys from the store.
func (s *Store) Purge() {
	s.Lock()
	s.t = radix.New()
	s.Unlock()
}

func stripKey(key, prefix string) string {
	return strings.TrimPrefix(strings.TrimPrefix(key, prefix), "/")
}
