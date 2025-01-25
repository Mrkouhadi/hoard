package hoard

import (
	"container/list"
	"fmt"
	"hash"
	"hash/fnv"
	"sync"
	"time"
)

// CacheItem represents a piece of data in the cache.
type CacheItem struct {
	Value      []byte
	Expiration int64
	LRUElement *list.Element
}

// CacheShard represents a single shard of the cache.
type CacheShard struct {
	data    sync.Map
	lruList *list.List
	lruMu   sync.Mutex
}

// Cache represents the inmemory cache with sharding.
type Cache struct {
	shards           []*CacheShard
	numShards        int
	maxItemsPerShard int
	cleanupInterval  time.Duration
	hashFn           func() hash.Hash32
}

// NewCache creates a new cache with the specified parameters.
func NewCache(numShards, maxItemsPerShard int, cleanupInterval time.Duration) *Cache {
	if numShards <= 0 {
		panic("number of shards must be greater than 0")
	}
	if maxItemsPerShard <= 0 {
		panic(" maximum items per shard must be greater than 0")
	}
	shards := make([]*CacheShard, numShards)
	for i := range shards {
		shards[i] = &CacheShard{
			lruList: list.New(),
		}
	}

	cache := &Cache{
		shards:           shards,
		numShards:        numShards,
		maxItemsPerShard: maxItemsPerShard,
		cleanupInterval:  cleanupInterval,
		hashFn:           fnv.New32a,
	}

	// start a co-routine in the background for periodic cleanup
	go cache.startCleanup()

	return cache
}

// getShard returns the shard for a given key.
func (c *Cache) getShard(key string) *CacheShard {
	hash := c.hashFn()
	hash.Write([]byte(key))
	shardIdx := hash.Sum32() % uint32(c.numShards)
	return c.shards[shardIdx]
}

// Store inserts a piece of data into the cache.
func (c *Cache) Store(key string, value interface{}, ttl time.Duration) error {
	shard := c.getShard(key)
	expiration := time.Now().Add(ttl).UnixNano()

	// Serialize the value
	serializedValue, err := serialize(value)
	if err != nil {
		return fmt.Errorf("serialization error: %v", err)
	}

	shard.lruMu.Lock()
	defer shard.lruMu.Unlock()

	// Remove the key if it already exists
	if item, exists := shard.data.Load(key); exists {
		shard.lruList.Remove(item.(*CacheItem).LRUElement)
	}

	// Add the new item
	lruElement := shard.lruList.PushFront(key)
	cacheItem := &CacheItem{
		Value:      serializedValue,
		Expiration: expiration,
		LRUElement: lruElement,
	}
	shard.data.Store(key, cacheItem)

	// Evict the least recently used item if the cache is full
	if shard.lruList.Len() > c.maxItemsPerShard {
		oldest := shard.lruList.Back()
		if oldest != nil {
			oldestKey := oldest.Value.(string)
			shard.data.Delete(oldestKey)
			shard.lruList.Remove(oldest)
		}
	}
	return nil
}

// Fetch retrieves an item from the cache.
func (c *Cache) Fetch(key string) (interface{}, bool, error) {
	shard := c.getShard(key)

	item, exists := shard.data.Load(key)
	if !exists {
		return nil, false, nil
	}

	cacheItem := item.(*CacheItem)

	// Check expiration before accessing or returning
	if time.Now().UnixNano() > cacheItem.Expiration {
		// i have to clean up expired items first
		c.removeExpiredItem(shard, key)
		return nil, false, nil
	}

	// Deserialize the value
	value, err := deserialize(cacheItem.Value)
	if err != nil {
		return nil, false, fmt.Errorf("deserialization error: %v", err)
	}

	// Update LRU
	shard.lruMu.Lock()
	shard.lruList.MoveToFront(cacheItem.LRUElement)
	shard.lruMu.Unlock()

	return value, true, nil
}

// removeExpiredItem removes an expired item from a shard.
func (c *Cache) removeExpiredItem(shard *CacheShard, key string) {
	shard.lruMu.Lock()
	defer shard.lruMu.Unlock()

	if item, exists := shard.data.Load(key); exists {
		shard.lruList.Remove(item.(*CacheItem).LRUElement)
		shard.data.Delete(key)
	}
}

// startCleanup starts a background goroutine to clean up expired items.
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		for _, shard := range c.shards {
			c.cleanupShard(shard)
		}
	}
}

// cleanupShard removes expired items from a shard.
func (c *Cache) cleanupShard(shard *CacheShard) {
	shard.lruMu.Lock()
	defer shard.lruMu.Unlock()

	now := time.Now().UnixNano()

	shard.data.Range(func(key, value interface{}) bool {
		cacheItem := value.(*CacheItem)
		if now > cacheItem.Expiration {
			shard.data.Delete(key)
			shard.lruList.Remove(cacheItem.LRUElement)
		}
		return true
	})
}
