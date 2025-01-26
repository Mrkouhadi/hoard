package hoard

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

// measuring the performance of the Store method.
func BenchmarkStore(b *testing.B) {
	cache := NewCache(5, 10000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "kouhadi" + strconv.Itoa(i)
		value := "aboubakr-essaddik" + strconv.Itoa(i)
		_ = cache.Store(key, value, time.Second*10)
	}
}

// measuring the performance of the Fetch method.
func BenchmarkFetch(b *testing.B) {
	cache := NewCache(5, 10000, time.Minute)

	// inset some data
	for i := 0; i < 1000; i++ {
		key := "kouhadi" + strconv.Itoa(i)
		value := "aboubakr-essaddik" + strconv.Itoa(i)
		_ = cache.Store(key, value, time.Second*10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "kouhadi" + strconv.Itoa(i%1000) // Cycle through the keys
		_, _, _ = cache.Fetch(key)
	}
}

// measuring the performance of concurrent Store and Fetch operations.
func BenchmarkStoreAndFetch(b *testing.B) {
	cache := NewCache(5, 10000, time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "kouhadi" + strconv.Itoa(i)
			value := "aboubakr-essaddik" + strconv.Itoa(i)
			_ = cache.Store(key, value, time.Second*10)
			_, _, _ = cache.Fetch(key)
			i++
		}
	})
}

// measuring the performance of the LRU eviction logic.
func BenchmarkEvictLRU(b *testing.B) {
	// 1 shard, max 10000 items per shard
	cache := NewCache(1, 10000, time.Minute)

	// insert in the cache some data
	for i := 0; i < 1000; i++ {
		key := "kouhadi" + strconv.Itoa(i)
		value := "aboubakr-essaddik" + strconv.Itoa(i)
		_ = cache.Store(key, value, time.Second*10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "kouhadi" + strconv.Itoa(1000+i) // Add some new data to trigger eviction :)
		value := "aboubakr-essaddik" + strconv.Itoa(1000+i)
		_ = cache.Store(key, value, time.Second*10)
	}
}

// measuring the update of data
func BenchmarkUpdate(b *testing.B) {
	cache := NewCache(10, 1000, time.Minute)

	// Store a value to update later
	err := cache.Store("foo", "bar", time.Minute)
	if err != nil {
		b.Fatalf("Store failed: %v", err)
	}

	b.ResetTimer() // Reset the timer to exclude setup time

	for i := 0; i < b.N; i++ {
		err := cache.Update("foo", "baz", time.Minute)
		if err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

// measuring the deletion of data

func BenchmarkDelete(b *testing.B) {
	cache := NewCache(10, 1000, time.Minute)

	// Pre-fill the cache with keys to delete
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		err := cache.Store(keys[i], "value", time.Minute)
		if err != nil {
			b.Fatalf("Store failed: %v", err)
		}
	}

	b.ResetTimer() // Reset the timer to exclude setup time

	for i := 0; i < b.N; i++ {
		cache.Delete(keys[i])
	}
}

// measuring the concurrent deletion and update of data
func BenchmarkConcurrentUpdateDelete(b *testing.B) {
	cache := NewCache(10, 1000, time.Minute)

	// Pre-fill the cache with keys
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		err := cache.Store(keys[i], "value", time.Minute)
		if err != nil {
			b.Fatalf("Store failed: %v", err)
		}
	}

	b.ResetTimer() // Reset the timer to exclude setup time

	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrently update and delete keys
	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			err := cache.Update(keys[i], "newvalue", time.Minute)
			if err != nil {
				b.Logf("Update failed: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			cache.Delete(keys[i])
		}
	}()

	wg.Wait()
}

// measuring the cleaning up of cache
func BenchmarkCleanupAll(b *testing.B) {
	cache := NewCache(10, 1000, time.Minute)

	// Pre-fill the cache with keys
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		err := cache.Store(key, "value", time.Minute)
		if err != nil {
			b.Fatalf("Store failed: %v", err)
		}
	}

	b.ResetTimer() // Reset the timer to exclude setup time

	for i := 0; i < b.N; i++ {
		cache.CleanupAll()
	}
}

// measuring fetch all data at once
func BenchmarkFetchAll(b *testing.B) {
	cache := NewCache(10, 1000, time.Minute)

	// Pre-fill the cache with keys
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		err := cache.Store(key, "value", time.Minute)
		if err != nil {
			b.Fatalf("Store failed: %v", err)
		}
	}

	b.ResetTimer() // Reset the timer to exclude setup time

	for i := 0; i < b.N; i++ {
		cache.FetchAll()
	}
}
