package gocash

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"reflect"
)

//
// Some of this is borrowed from https://github.com/timtadh/data-structures/blob/master/hashtable/hashtable.go
//

type entry struct {
	key   string
	value interface{}
	next  *entry
}

func (e *entry) Put(key string, value interface{}) (*entry, bool) {
	if e == nil {
		return &entry{key, value, nil}, true
	}
	if reflect.DeepEqual(e.key, key) {
		e.value = value
		return e, false
	}
	var appended bool
	e.next, appended = e.next.Put(key, value)
	return e, appended
}

func (e *entry) Get(key string) (bool, interface{}) {
	if e == nil {
		return false, nil
	} else if reflect.DeepEqual(e.key, key) {
		return true, e.value
	}
	return e.next.Get(key)
}

func (e *entry) Remove(key string) *entry {
	if e == nil {
		panic("Not found in bucket")
	}
	if reflect.DeepEqual(e.key, key) {
		return e.next
	}
	e.next = e.next.Remove(key)
	return e
}

// NewHash returns a pointer to a new Hash object
func NewHash() *Hash {
	hash := &Hash{
		hasher:     fnv.New32a(),
		bucketSize: uint32(64),
	}
	hash.snapshots = make([][]*entry, 0, 4)
	hash.table = make([]*entry, hash.bucketSize)
	for i := range hash.table {
		hash.table[i] = &entry{}
	}
	hash.current = make([]*entry, hash.bucketSize)
	return hash
}

// Hash is a copy on write Hash Tree
type Hash struct {
	snapshots  [][]*entry
	table      []*entry
	current    []*entry
	bucketSize uint32
	size       uint32
	hasher     hash.Hash
}

func (h *Hash) bucket(key string) uint32 {
	defer h.hasher.Reset()
	h.hasher.Write([]byte(key))
	return binary.LittleEndian.Uint32(h.hasher.Sum(nil)) % h.bucketSize
}

func (h *Hash) copy(e *entry) *entry {
	if e == nil {
		return nil
	}
	return &entry{
		key:   e.key,
		value: e.value,
		next:  h.copy(e.next),
	}
}

// copyIfSnapshot copy's data to current table if bucket refers to a snapshot
func (h *Hash) copyIfSnapshot(bucket uint32) {
	if h.current[bucket] == nil {
		h.current[bucket] = h.copy(h.table[bucket])
		h.table[bucket] = h.current[bucket]
	}
}

// Put inserts or updates `key` with `value` of the Hash Tree
func (h *Hash) Put(key string, value interface{}) error {
	bucket := h.bucket(key)
	h.copyIfSnapshot(bucket)
	var appended bool
	h.table[bucket], appended = (*h.table[bucket]).Put(key, value)
	if appended {
		h.size++
	}
	return nil
}

// Get the value of `key`
func (h *Hash) Get(key string) (interface{}, error) {
	bucket := h.bucket(key)
	if has, value := (*h.table[bucket]).Get(key); has {
		return value, nil
	}
	return nil, fmt.Errorf("Key not found: %s", key)
}

// Remove the entry of `key`
func (h *Hash) Remove(key string) (interface{}, error) {
	bucket := h.bucket(key)
	h.copyIfSnapshot(bucket)
	has, value := (*h.table[bucket]).Get(key)
	if !has {
		return nil, fmt.Errorf("Key not found %s", key)
	}
	h.table[bucket] = (*h.table[bucket]).Remove(key)
	h.size--
	return value, nil
}

// Snapshot creates a RO snapshot
func (h *Hash) Snapshot() {
	h.snapshots = append(h.snapshots, h.table)
	h.table = make([]*entry, h.bucketSize)
	h.current = make([]*entry, h.bucketSize)

	snap := h.snapshots[len(h.snapshots)-1]
	for i, value := range snap {
		h.table[i] = value
	}
}
