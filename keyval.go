package main

import "sync"

// KV is the inner hashMap we are using for our inMem data store.
type KV struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewKeyVal creates an inMemory data store.
func NewKeyVal() *KV {
	return &KV{
		data: map[string][]byte{},
	}
}

// Set sets a key and a value into the store.
func (kv *KV) Set(key, value []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data[string(key)] = []byte(value)

	return nil
}

// Get gets the value associated with the key from the store.
func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	val, ok := kv.data[string(key)]

	return val, ok
}
