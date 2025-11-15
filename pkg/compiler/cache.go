package compiler

import (
	"sync"
)

// Cache provides thread-safe caching of compiled WASM modules
type Cache struct {
	mu    sync.RWMutex
	cache map[string][]byte // hash -> wasm bytes
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
	return &Cache{
		cache: make(map[string][]byte),
	}
}

// Get retrieves a compiled WASM module from cache by source code
// Returns the WASM bytes and true if found, or nil and false if not cached
func (c *Cache) Get(sourceCode string) ([]byte, bool) {
	hash := ComputeHash(sourceCode)

	c.mu.RLock()
	defer c.mu.RUnlock()

	wasmBytes, found := c.cache[hash]
	return wasmBytes, found
}

// Set stores a compiled WASM module in the cache
func (c *Cache) Set(sourceCode string, wasmBytes []byte) {
	hash := ComputeHash(sourceCode)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[hash] = wasmBytes
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string][]byte)
}

// Size returns the number of cached modules
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// Has checks if a source code is in the cache
func (c *Cache) Has(sourceCode string) bool {
	hash := ComputeHash(sourceCode)

	c.mu.RLock()
	defer c.mu.RUnlock()

	_, found := c.cache[hash]
	return found
}
