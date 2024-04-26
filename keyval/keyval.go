package keyval

import (
	"fmt"
	"strconv"
	"sync"
)

// KV is the inner hashMap we are using for our inMem data store.
type KV struct {
	mu       sync.RWMutex
	data     map[string][]byte
	slices   map[string][]string
}

// NewKeyVal creates an inMemory data store.
func NewKeyVal() *KV {
	return &KV{
		data:   map[string][]byte{},
		slices: map[string][]string{},
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

// NOTE: We are not returning anything because redis a key is ignored in case it doesn't exists
func (kv *KV) Del(key []byte) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	delete(kv.data, string(key))
}

func (kv *KV) Incr(key []byte) (int, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	m, ok := kv.data[string(key)]
	if !ok {
		return 0, fmt.Errorf("sorry but this key doesn't exists")
	}

	intValue, err := strconv.Atoi(string(m))
	if err != nil {
		return 0, err
	}

	intValue += 1
	kv.data[string(key)] = []byte(strconv.Itoa(intValue))

	return intValue, nil
}

// TODO: This has some code duplication remove it later to something shared (Pay attention to concurrency)
// And also the mutex usage here is a bit expensive we can think of doing this with atomics depends on the benchmarks
func (kv *KV) Decr(key []byte) (int, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	m, ok := kv.data[string(key)]
	if !ok {
		return 0, fmt.Errorf("sorry but this key doesn't exists")
	}

	intValue, err := strconv.Atoi(string(m))
	if err != nil {
		return 0, err
	}

	intValue -= 1
	kv.data[string(key)] = []byte(strconv.Itoa(intValue))

	return intValue, nil
}

func (kv *KV) Push(key string, value []string) (int, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.slices[key] = append(kv.slices[key], value...)

	return len(kv.slices[key]), nil
}
