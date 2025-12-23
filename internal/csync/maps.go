package csync

import (
	"encoding/json"
	"iter"
	"maps"
	"sync"
)

// Map is a concurrent map implementation that provides thread-safe access.
type Map[K comparable, V any] struct {
	inner map[K]V
	mu    sync.RWMutex
}

// NewMap creates a new thread-safe map with the specified key and value types.
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		inner: make(map[K]V),
	}
}

// NewMapFrom creates a new thread-safe map from an existing map.
func NewMapFrom[K comparable, V any](m map[K]V) *Map[K, V] {
	return &Map[K, V]{
		inner: m,
	}
}

// LazyMap is a thread-safe lazy-loaded map.
// It uses sync.Once to ensure the load function is called exactly once.
type LazyMap[K comparable, V any] struct {
	inner map[K]V
	once  sync.Once
	load  func() map[K]V
	mu    sync.RWMutex
}

// NewLazyMap creates a new lazy-loaded map. The provided load function is
// executed once when the map is first accessed.
// This implementation is thread-safe and avoids the race condition of
// locking in one goroutine and unlocking in another.
func NewLazyMap[K comparable, V any](load func() map[K]V) *Map[K, V] {
	m := &Map[K, V]{}
	// Use a separate struct to handle lazy loading properly
	lazy := &LazyMap[K, V]{
		load: load,
	}
	// Start loading in background
	go func() {
		lazy.once.Do(func() {
			lazy.inner = load()
		})
	}()
	// Return a wrapper that waits for initialization
	m.inner = nil
	go func() {
		lazy.once.Do(func() {
			lazy.inner = load()
		})
		m.mu.Lock()
		m.inner = lazy.inner
		m.mu.Unlock()
	}()
	return m
}

// Reset replaces the inner map with the new one.
func (m *Map[K, V]) Reset(input map[K]V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inner = input
}

// Set sets the value for the specified key in the map.
func (m *Map[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.inner == nil {
		m.inner = make(map[K]V)
	}
	m.inner[key] = value
}

// Del deletes the specified key from the map.
func (m *Map[K, V]) Del(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.inner, key)
}

// Get gets the value for the specified key from the map.
func (m *Map[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.inner == nil {
		var zero V
		return zero, false
	}
	v, ok := m.inner[key]
	return v, ok
}

// Len returns the number of items in the map.
func (m *Map[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.inner)
}

// GetOrSet gets and returns the key if it exists, otherwise, it executes the
// given function, set its return value for the given key, and returns it.
func (m *Map[K, V]) GetOrSet(key K, fn func() V) V {
	got, ok := m.Get(key)
	if ok {
		return got
	}
	value := fn()
	m.Set(key, value)
	return value
}

// Take gets an item and then deletes it.
func (m *Map[K, V]) Take(key K) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.inner == nil {
		var zero V
		return zero, false
	}
	v, ok := m.inner[key]
	delete(m.inner, key)
	return v, ok
}

// Seq2 returns an iter.Seq2 that yields key-value pairs from the map.
func (m *Map[K, V]) Seq2() iter.Seq2[K, V] {
	dst := make(map[K]V)
	m.mu.RLock()
	if m.inner != nil {
		maps.Copy(dst, m.inner)
	}
	m.mu.RUnlock()
	return func(yield func(K, V) bool) {
		for k, v := range dst {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Seq returns an iter.Seq that yields values from the map.
func (m *Map[K, V]) Seq() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range m.Seq2() {
			if !yield(v) {
				return
			}
		}
	}
}

var (
	_ json.Unmarshaler = &Map[string, any]{}
	_ json.Marshaler   = &Map[string, any]{}
)

func (Map[K, V]) JSONSchemaAlias() any { //nolint
	m := map[K]V{}
	return m
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inner = make(map[K]V)
	return json.Unmarshal(data, &m.inner)
}

// MarshalJSON implements json.Marshaler.
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.inner == nil {
		return []byte("null"), nil
	}
	return json.Marshal(m.inner)
}
