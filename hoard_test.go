package hoard

import (
	"fmt"
	"strconv"
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
	err = cache.Store("kouhadi", 42, time.Second*10)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Fetch "aboubakr" to make it recently used
	_, _, err = cache.Fetch("aboubakr")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Store a third item (should evict "kouhadi" as it is the least recently used)
	err = cache.Store("qux", 3.14, time.Second*10)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Fetch "kouhadi" (should not exist)
	_, exists, err := cache.Fetch("kouhadi")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if exists {
		t.Fatal("Expected 'kouahdi' to be evicted")
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

// testing the update of a piece of data
func TestUpdate(t *testing.T) {
	cache := NewCache(10, 1000, time.Minute)

	// Store a value
	err := cache.Store("haroun", 30, time.Minute)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Update the value
	err = cache.Update("haroun", "kouhadi", time.Minute)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Fetch the updated value
	value, exists, err := cache.Fetch("haroun")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !exists {
		t.Error("Expected key 'haroun' to exist, but it does not")
	}
	if value != "kouhadi" {
		t.Errorf("Expected value 'kouhadi', got '%v'", value)
	}

	// Test updating a non-existent key
	err = cache.Update("nonexistent", "value", time.Minute)
	if err == nil {
		t.Error("Expected error when updating a non-existent key, but got nil")
	}
}

// testing the deleting of a piece of data
func TestDelete(t *testing.T) {
	cache := NewCache(10, 1000, time.Minute)

	// Store a value
	err := cache.Store("haroun", 30, time.Minute)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Delete the value
	cache.Delete("haroun")

	// Fetch the deleted value
	value, exists, err := cache.Fetch("haroun")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if exists {
		t.Error("Expected key 'haroun' to be deleted, but it still exists")
	}
	if value != nil {
		t.Errorf("Expected value to be nil, got '%v'", value)
	}

	// Test deleting a non-existent key (should not panic or error)
	cache.Delete("nonexistent")
}

// testing the concurrent update and delete
func TestConcurrentUpdateDelete(t *testing.T) {
	cache := NewCache(10, 1000, time.Minute)

	// Store a value
	err := cache.Store("haroun", 30, time.Minute)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrently update and delete the same key
	go func() {
		defer wg.Done()
		err := cache.Update("haroun", "kouhadi", time.Minute)
		if err != nil {
			t.Logf("Update failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		cache.Delete("haroun")
	}()

	wg.Wait()

	// Fetch the key to check the final state
	value, exists, err := cache.Fetch("haroun")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	t.Logf("Final state: value=%v, exists=%v", value, exists)
}

// testing the cleaning up of cache
func TestCleanupAllParallel(t *testing.T) {
	cache := NewCache(10, 1000, time.Minute)

	// Store some values
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		err := cache.Store(key, "value", time.Minute)
		if err != nil {
			t.Fatalf("Store failed: %v", err)
		}
	}

	// Cleanup all data
	cache.CleanupAll()

	// Verify that the cache is empty
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		value, exists, err := cache.Fetch(key)
		if err != nil {
			t.Fatalf("Fetch failed: %v", err)
		}
		if exists || value != nil {
			t.Errorf("Expected cache to be empty after cleanup, but key '%s' still exists", key)
		}
	}
}

// testing Iterate
// TestIterate ensures that Iterate visits all items in the cache.
func TestIterate(t *testing.T) {
	cache := NewCache(10, 10000, time.Minute)

	// Pre-fill cache
	numItems := 1000
	for i := 0; i < numItems; i++ {
		key := "key" + strconv.Itoa(i)
		value := []byte("value" + strconv.Itoa(i))
		_ = cache.Store(key, value, time.Minute)
	}

	visited := make(map[string]bool)
	var mu sync.Mutex

	cache.Iterate(func(key string, value []byte) {
		mu.Lock()
		defer mu.Unlock()
		visited[key] = true
	})

	if len(visited) != numItems {
		t.Fatalf("Expected %d items, got %d", numItems, len(visited))
	}
}
