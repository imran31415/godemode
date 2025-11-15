package compiler

import (
	"testing"
)

func TestComputeHash(t *testing.T) {
	code1 := "package main\nfunc main() {}"
	code2 := "package main\nfunc main() {}"
	code3 := "package main\nfunc main() { println() }"

	hash1 := ComputeHash(code1)
	hash2 := ComputeHash(code2)
	hash3 := ComputeHash(code3)

	if hash1 != hash2 {
		t.Error("Same code should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("Different code should produce different hash")
	}

	if len(hash1) != 64 {
		t.Errorf("Hash should be 64 characters (SHA256 hex), got %d", len(hash1))
	}
}

func TestCache(t *testing.T) {
	cache := NewCache()

	code := "test code"
	wasm := []byte{0x00, 0x61, 0x73, 0x6d} // WASM magic number

	// Test cache miss
	if _, found := cache.Get(code); found {
		t.Error("Should not find uncached code")
	}

	// Test cache set
	cache.Set(code, wasm)

	// Test cache hit
	cached, found := cache.Get(code)
	if !found {
		t.Error("Should find cached code")
	}

	if len(cached) != len(wasm) {
		t.Error("Cached WASM should match original")
	}

	// Test cache size
	if cache.Size() != 1 {
		t.Errorf("Cache size should be 1, got %d", cache.Size())
	}

	// Test clear
	cache.Clear()
	if cache.Size() != 0 {
		t.Error("Cache should be empty after clear")
	}
}

func TestCheckTinyGo(t *testing.T) {
	// This test will only pass if TinyGo is installed
	// We'll make it a skip if not found
	err := CheckTinyGo()
	if err != nil {
		t.Skipf("TinyGo not installed: %v", err)
	}
}
