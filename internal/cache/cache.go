// Package cache provides a caching system with TTL and LRU eviction.
package cache

import (
	"container/list"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a cache entry.
type Entry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
	AccessedAt time.Time  `json:"accessed_at"`
	Hits      int64       `json:"hits"`
}

// IsExpired checks if the entry is expired.
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache implements a thread-safe cache with TTL and LRU eviction.
type Cache struct {
	mu          sync.RWMutex
	entries     map[string]*list.Element
	lruList     *list.List
	maxSize     int
	defaultTTL  time.Duration
	persistPath string
	stats       Stats
}

// Stats holds cache statistics.
type Stats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Evictions  int64 `json:"evictions"`
	Expired    int64 `json:"expired"`
	CurrentSize int   `json:"current_size"`
}

// Config holds cache configuration.
type Config struct {
	MaxSize     int           `json:"max_size"`
	DefaultTTL  time.Duration `json:"default_ttl"`
	PersistPath string        `json:"persist_path,omitempty"`
}

// DefaultConfig returns the default cache configuration.
func DefaultConfig() Config {
	return Config{
		MaxSize:    1000,
		DefaultTTL: 15 * time.Minute,
	}
}

// New creates a new cache with the given configuration.
func New(config Config) *Cache {
	if config.MaxSize <= 0 {
		config.MaxSize = 1000
	}
	if config.DefaultTTL <= 0 {
		config.DefaultTTL = 15 * time.Minute
	}

	c := &Cache{
		entries:     make(map[string]*list.Element),
		lruList:     list.New(),
		maxSize:     config.MaxSize,
		defaultTTL:  config.DefaultTTL,
		persistPath: config.PersistPath,
	}

	// Load persisted cache if path is set
	if config.PersistPath != "" {
		c.load()
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves a value from the cache.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.entries[key]
	if !ok {
		c.stats.Misses++
		return nil, false
	}

	entry := element.Value.(*Entry)

	// Check if expired
	if entry.IsExpired() {
		c.remove(element)
		c.stats.Expired++
		c.stats.Misses++
		return nil, false
	}

	// Update access time and move to front
	entry.AccessedAt = time.Now()
	entry.Hits++
	c.lruList.MoveToFront(element)

	c.stats.Hits++
	return entry.Value, true
}

// Set stores a value in the cache with the default TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a specific TTL.
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Check if key already exists
	if element, ok := c.entries[key]; ok {
		entry := element.Value.(*Entry)
		entry.Value = value
		entry.ExpiresAt = now.Add(ttl)
		entry.AccessedAt = now
		c.lruList.MoveToFront(element)
		return
	}

	// Evict if necessary
	for c.lruList.Len() >= c.maxSize {
		c.evictOldest()
	}

	// Create new entry
	entry := &Entry{
		Key:        key,
		Value:      value,
		ExpiresAt:  now.Add(ttl),
		CreatedAt:  now,
		AccessedAt: now,
		Hits:       0,
	}

	element := c.lruList.PushFront(entry)
	c.entries[key] = element
	c.stats.CurrentSize = c.lruList.Len()
}

// Delete removes a value from the cache.
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.entries[key]
	if !ok {
		return false
	}

	c.remove(element)
	return true
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.lruList.Init()
	c.stats.CurrentSize = 0
}

// Size returns the current number of entries in the cache.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lruList.Len()
}

// Stats returns cache statistics.
func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.stats.CurrentSize = c.lruList.Len()
	return c.stats
}

// Keys returns all keys in the cache.
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.entries))
	for key := range c.entries {
		keys = append(keys, key)
	}
	return keys
}

// remove removes an element from the cache (must be called with lock held).
func (c *Cache) remove(element *list.Element) {
	entry := element.Value.(*Entry)
	delete(c.entries, entry.Key)
	c.lruList.Remove(element)
	c.stats.CurrentSize = c.lruList.Len()
}

// evictOldest evicts the least recently used entry (must be called with lock held).
func (c *Cache) evictOldest() {
	element := c.lruList.Back()
	if element != nil {
		c.remove(element)
		c.stats.Evictions++
	}
}

// cleanup periodically removes expired entries.
func (c *Cache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired removes all expired entries.
func (c *Cache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var toRemove []*list.Element

	for element := c.lruList.Front(); element != nil; element = element.Next() {
		entry := element.Value.(*Entry)
		if now.After(entry.ExpiresAt) {
			toRemove = append(toRemove, element)
		}
	}

	for _, element := range toRemove {
		c.remove(element)
		c.stats.Expired++
	}
}

// Persist saves the cache to disk.
func (c *Cache) Persist() error {
	if c.persistPath == "" {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Collect non-expired entries
	var entries []*Entry
	for element := c.lruList.Front(); element != nil; element = element.Next() {
		entry := element.Value.(*Entry)
		if !entry.IsExpired() {
			entries = append(entries, entry)
		}
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(c.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(c.persistPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// load loads the cache from disk.
func (c *Cache) load() error {
	data, err := os.ReadFile(c.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var entries []*Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if now.Before(entry.ExpiresAt) {
			element := c.lruList.PushBack(entry)
			c.entries[entry.Key] = element
		}
	}

	c.stats.CurrentSize = c.lruList.Len()
	return nil
}

// GetOrCompute gets a value from cache or computes it if not present.
func (c *Cache) GetOrCompute(key string, compute func() (interface{}, error)) (interface{}, error) {
	if value, ok := c.Get(key); ok {
		return value, nil
	}

	value, err := compute()
	if err != nil {
		return nil, err
	}

	c.Set(key, value)
	return value, nil
}

// GetOrComputeWithTTL gets a value from cache or computes it if not present.
func (c *Cache) GetOrComputeWithTTL(key string, ttl time.Duration, compute func() (interface{}, error)) (interface{}, error) {
	if value, ok := c.Get(key); ok {
		return value, nil
	}

	value, err := compute()
	if err != nil {
		return nil, err
	}

	c.SetWithTTL(key, value, ttl)
	return value, nil
}
