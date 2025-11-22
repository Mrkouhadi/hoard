package hoard

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

// Generate a large set of keys and values for heavy load
const (
	NumKeys     = 1_000_000
	Concurrency = 64
	ValueSize   = 256 // bytes
)

// Helper to generate a random value of fixed size
func randomValue(size int) string {
	bytes := make([]byte, size)
	for i := range bytes {
		bytes[i] = byte('a' + rand.Intn(26))
	}
	return string(bytes)
}

// Benchmark storing heavy data
func BenchmarkStoreHeavy(b *testing.B) {
	cache := NewCache(16, NumKeys, time.Minute)

	keys := make([]string, NumKeys)
	values := make([]string, NumKeys)
	for i := 0; i < NumKeys; i++ {
		keys[i] = "key_" + strconv.Itoa(i)
		values[i] = randomValue(ValueSize)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			idx := rnd.Intn(NumKeys)
			cache.Store(keys[idx], values[idx], time.Minute)
		}
	})
}

// Benchmark fetching heavy data
func BenchmarkFetchDataHeavy(b *testing.B) {
	cache := NewCache(16, NumKeys, time.Minute)

	keys := make([]string, NumKeys)
	values := make([]string, NumKeys)
	for i := 0; i < NumKeys; i++ {
		keys[i] = "key_" + strconv.Itoa(i)
		values[i] = randomValue(ValueSize)
		cache.Store(keys[i], values[i], time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			idx := rnd.Intn(NumKeys)
			cache.FetchData(keys[idx])
		}
	})
}
func BenchmarkFetchBytesDataHeavy(b *testing.B) {
	cache := NewCache(16, NumKeys, time.Minute)

	keys := make([]string, NumKeys)
	values := make([]string, NumKeys)
	for i := 0; i < NumKeys; i++ {
		keys[i] = "key_" + strconv.Itoa(i)
		values[i] = randomValue(ValueSize)
		cache.Store(keys[i], values[i], time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			idx := rnd.Intn(NumKeys)
			cache.FetchBytesData(keys[idx])
		}
	})
}

// Benchmark mixed concurrent Store + FetchData + FetchBytesData + Delete + Update
func BenchmarkConcurrentHeavy(b *testing.B) {
	cache := NewCache(16, NumKeys, time.Minute)

	keys := make([]string, NumKeys)
	values := make([]string, NumKeys)
	for i := 0; i < NumKeys; i++ {
		keys[i] = "key_" + strconv.Itoa(i)
		values[i] = randomValue(ValueSize)
		cache.Store(keys[i], values[i], time.Minute)
	}

	b.ResetTimer()
	var wg sync.WaitGroup
	wg.Add(Concurrency)

	for i := 0; i < Concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
			for j := 0; j < b.N; j++ {
				idx := rnd.Intn(NumKeys)
				op := rnd.Intn(5)
				switch op {
				case 0:
					cache.Store(keys[idx], values[idx], time.Minute)
				case 1:
					cache.FetchData(keys[idx])
				case 2:
					cache.FetchBytesData(keys[idx])
				case 3:
					cache.Update(keys[idx], values[idx], time.Minute)
				case 4:
					cache.Delete(keys[idx])
				}
			}
		}(i)
	}

	wg.Wait()
}

// BenchmarkIterateHeavy benchmarks Iterate under heavy load with 1M items.
func BenchmarkIterateHeavy(b *testing.B) {
	cache := NewCache(10, 100000, time.Minute)

	// Pre-fill cache with 1M items
	numItems := 100000
	for i := 0; i < numItems; i++ {
		key := fmt.Sprintf("key%d", i)
		value := []byte(fmt.Sprintf("value%d", i))
		_ = cache.Store(key, value, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Iterate(func(key string, value []byte) {
			// no-op, just iterate
		})
	}
}

// Benchmark cleanup of all items under heavy load
func BenchmarkCleanupAllHeavy(b *testing.B) {
	cache := NewCache(16, NumKeys, time.Minute)

	for i := 0; i < NumKeys; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := randomValue(ValueSize)
		cache.Store(key, value, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.CleanupAll()
	}
}
