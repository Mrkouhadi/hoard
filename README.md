# Hoard 🗃️

**Hoard** is a high-performance, in-memory caching library for Go, designed to handle high-throughput workloads with minimal latency. It features **sharding**, **LRU eviction**, and **time-to-live (TTL)** support, making it ideal for applications that require fast and efficient caching.

---

## Features ✨

- **Sharding**: Distributes cache data across multiple shards to reduce lock contention and improve performance.
- **LRU Eviction**: Automatically evicts the least recently used items when the cache reaches its capacity.
- **TTL Support**: Allows setting a time-to-live (TTL) for each cache item, ensuring stale data is automatically removed.
- **Thread-Safe**: Built with `sync.Map` and `sync.Mutex` to ensure safe concurrent access.
- **High Performance**: Optimized for low latency and high throughput, with benchmarks showing **500 ns/op for Fetch** and **1.5 µs/op for Store**.
- **Lightweight**: Minimal dependencies and efficient memory usage.

---

## Installation 📦

To install `hoard`, use `go get`:

```bash
go get github.com/mrkouhadi/hoard
```

---

## Usage 🚀

### Basic Example

```go
package main

import (
	"fmt"
	"time"
	"github.com/mrkouhadi/hoard"
)

func main() {
	// Create a new cache with 5 shards, max 10000 items per shard, and a cleanup interval of 10 minute
	cache := hoard.NewCache(5, 10000, time.Minute*10)

	// Store an item with a TTL of 10 seconds
	cache.Store("foo", "bar", 10*time.Second)

	// Fetch the item
	value, exists, err := cache.Fetch("foo")
	if err != nil {
		fmt.Println("Error fetching item:", err)
	} else if exists {
		fmt.Println("Fetched value:", value) // Output: Fetched value: bar
	} else {
		fmt.Println("Item not found or expired")
	}

	// Delete an item
	cache.Delete("foo")

	// Clear the cache
	cache.Clear()
}
```

---

## Performance 🚀

Hoard is optimized for **low latency** and **high throughput**. Here are the benchmark results on an **Apple M1**:

## Benchmarks

Here are the latest benchmark results for the caching library:

| Benchmark         | Time per Operation | Memory Allocation | Allocations per Operation |
| ----------------- | ------------------ | ----------------- | ------------------------- |
| **Store**         | 1.47 µs/op         | 430 B/op          | 16 allocs/op              |
| **Fetch**         | 509 ns/op          | 104 B/op          | 6 allocs/op               |
| **StoreAndFetch** | 1.26 µs/op         | 469 B/op          | 20 allocs/op              |
| **EvictLRU**      | 1.27 µs/op         | 423 B/op          | 16 allocs/op              |

---

## Advanced Usage 🛠️

### Custom Hash Function

You can provide a custom hash function when creating the cache:

```go
cache := hoard.NewCacheWithHash(4, 1000, time.Minute, func() hash.Hash32 {
	return fnv.New32a() // Use FNV-1a hash
})
```

### Concurrent Access

Hoard is designed for concurrent use. You can safely call `Store`, `Fetch`, and `Delete` from multiple goroutines:

```go
go func() {
	cache.Store("key1", "value1", time.Second*10)
}()

go func() {
	value, exists, _ := cache.Fetch("key1")
	if exists {
		fmt.Println("Fetched value:", value)
	}
}()
```

---

## Benchmarks 📊

To run the benchmarks yourself, use the following command:

```bash
go test -bench=. -benchmem
```

---

## Contributing 🤝

Contributions are welcome! If you find a bug or have a feature request, please open an issue. If you'd like to contribute code, fork the repository and submit a pull request.

---

## License 📄

Hoard is licensed under the MIT License. See [LICENSE](LICENSE) for details.

---

## Acknowledgments 🙏

- Uses [MessagePack](https://github.com/vmihailenco/msgpack) for efficient serialization.

---

## Why Hoard? 🧐

- **Fast**: Benchmarks show **500 ns/op for Fetch** and **1.5 µs/op for Store**.
- **Scalable**: Sharding and LRU eviction ensure the cache scales with your workload.
- **Simple**: Easy-to-use API with minimal configuration.

---

## Get Started Today! 🎉

```bash
go get github.com/mrkouhadi/hoard
```

---

Happy caching! 🚀
