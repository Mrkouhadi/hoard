package hoard

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// testing storing and fetching items from the cache.
func TestStoreAndFetch(t *testing.T) {
	cache := NewCache(4, 1000, time.Second)

	// Store an item
	err := cache.Store("bakr", "kouhadi", time.Second*10)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Fetch the item
	value, exists, err := cache.Fetch("bakr")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !exists {
		t.Fatal("Expected item to exist in the cache")
	}
	if value != "kouhadi" {
		t.Fatalf("Expected value 'kouhadi', got '%v'", value)
	}
}

// testing that items expire after their TTL.
func TestExpiration(t *testing.T) {
	cache := NewCache(4, 1000, time.Second)

	// Store an item with a short TTL
	err := cache.Store("aboubakr", "kouhadi", time.Second*2)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Fetch the item immediately (should exist)
	value, exists, err := cache.Fetch("aboubakr")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !exists {
		t.Fatal("Expected item to exist in the cache")
	}
	if value != "kouhadi" {
		t.Fatalf("Expected value 'bar', got '%v'", value)
	}

	// Wait for the item to expire
	time.Sleep(3 * time.Second)

	// Fetch the item again (should not exist)
	_, exists, err = cache.Fetch("aboubakr")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if exists {
		t.Fatal("Expected item to be expired")
	}
}

// testing that the least recently used item is evicted when the cache is full.
func TestLRUEviction(t *testing.T) {
	cache := NewCache(1, 2, time.Second) // 1 shard, max 2 items per shard

	// Store two items
	err := cache.Store("aboubakr", "kouhadi", time.Second*10)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	err = cache.Store("baz", 42, time.Second*10)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Fetch "aboubakr" to make it recently used
	_, _, err = cache.Fetch("aboubakr")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Store a third item (should evict "baz" as it is the least recently used)
	err = cache.Store("qux", 3.14, time.Second*10)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Fetch "baz" (should not exist)
	_, exists, err := cache.Fetch("baz")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if exists {
		t.Fatal("Expected 'baz' to be evicted")
	}

	// Fetch "aboubakr" (should still exist)
	value, exists, err := cache.Fetch("aboubakr")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !exists {
		t.Fatal("Expected 'aboubakr' to exist in the cache")
	}
	if value != "kouhadi" {
		t.Fatalf("Expected value 'bar', got '%v'", value)
	}

	// Fetch "qux" (should exist)
	value, exists, err = cache.Fetch("qux")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !exists {
		t.Fatal("Expected 'qux' to exist in the cache")
	}
	if value != 3.14 {
		t.Fatalf("Expected value 3.14, got '%v'", value)
	}
}

// testing  that expired items are removed by the cleanup goroutine.
func TestCleanup(t *testing.T) {
	cache := NewCache(4, 1000, time.Second)

	// Store an item with a short TTL
	err := cache.Store("aboubakr", "kouhadi", time.Second*2)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Wait for the item to expire and the cleanup goroutine to run
	time.Sleep(3 * time.Second)

	// Fetch the item (should not exist)
	_, exists, err := cache.Fetch("aboubakr")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if exists {
		t.Fatal("Expected item to be expired and removed by cleanup")
	}
}

// testing concurrent access to the cache.
func TestConcurrentAccess(t *testing.T) {
	cache := NewCache(4, 1000, time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)

			// Store the item
			err := cache.Store(key, value, time.Second*10)
			if err != nil {
				t.Errorf("Store failed: %v", err)
				return
			}

			// Fetch the item
			fetchedValue, exists, err := cache.Fetch(key)
			if err != nil {
				t.Errorf("Fetch failed: %v", err)
				return
			}
			if !exists {
				t.Errorf("Expected item %s to exist", key)
				return
			}
			if fetchedValue != value {
				t.Errorf("Expected value %s, got %v", value, fetchedValue)
			}
		}(i)
	}
	wg.Wait()
}
