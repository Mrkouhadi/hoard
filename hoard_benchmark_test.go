package hoard

import (
	"strconv"
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
