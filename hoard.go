package hoard

import (
	"container/list"
	"fmt"
	"hash"
	"hash/fnv"
	"sync"
	"time"
)

type CacheItem struct {
	Value      []byte
	Expiration int64
	LRUElement *list.Element
}

type CacheShard struct {
	mu      sync.RWMutex
	data    map[string]*CacheItem
	lruList *list.List
}

type Cache struct {
	shards           []*CacheShard
	numShards        int
	maxItemsPerShard int
	cleanupInterval  time.Duration
	hashFn           func() hash.Hash32
}

var cacheItemPool = sync.Pool{
	New: func() interface{} { return &CacheItem{} },
}

func NewCache(numShards, maxItemsPerShard int, cleanupInterval time.Duration) *Cache {
	if numShards <= 0 || maxItemsPerShard <= 0 {
		panic("invalid shard or maxItemsPerShard")
	}
	shards := make([]*CacheShard, numShards)
	for i := range shards {
		shards[i] = &CacheShard{
			data:    make(map[string]*CacheItem),
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
	go cache.startCleanup()
	return cache
}

func (c *Cache) getShard(key string) *CacheShard {
	h := c.hashFn()
	h.Write([]byte(key))
	return c.shards[h.Sum32()%uint32(c.numShards)]
}

//Store / Fetch

func (c *Cache) Store(key string, value interface{}, ttl time.Duration) error {
	shard := c.getShard(key)
	exp := time.Now().Add(ttl).UnixNano()

	val, err := serialize(value)
	if err != nil {
		return err
	}

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Remove existing
	if existing, ok := shard.data[key]; ok {
		shard.lruList.Remove(existing.LRUElement)
	}

	item := cacheItemPool.Get().(*CacheItem)
	item.Value = val
	item.Expiration = exp
	item.LRUElement = shard.lruList.PushFront(key)
	shard.data[key] = item

	// Evict LRU if over capacity
	if len(shard.data) > c.maxItemsPerShard {
		oldest := shard.lruList.Back()
		if oldest != nil {
			oldKey := oldest.Value.(string)
			delete(shard.data, oldKey)
			shard.lruList.Remove(oldest)
		}
	}
	return nil
}

func (c *Cache) Fetch(key string) (interface{}, bool, error) {
	shard := c.getShard(key)

	shard.mu.Lock() // combine RLock + Lock into a single Lock
	defer shard.mu.Unlock()

	item, ok := shard.data[key]
	if !ok {
		return nil, false, nil
	}

	if time.Now().UnixNano() > item.Expiration {
		// Remove expired item immediately
		shard.lruList.Remove(item.LRUElement)
		delete(shard.data, key)
		return nil, false, nil
	}

	// Move to front in the same lock
	shard.lruList.MoveToFront(item.LRUElement)

	// Deserialize outside the lock to reduce contention if value is large
	value, err := deserialize(item.Value)
	return value, true, err
}

func (c *Cache) Update(key string, value interface{}, ttl time.Duration) error {
	shard := c.getShard(key)
	exp := time.Now().Add(ttl).UnixNano()

	val, err := serialize(value)
	if err != nil {
		return err
	}

	shard.mu.Lock()
	defer shard.mu.Unlock()

	item, ok := shard.data[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	item.Value = val
	item.Expiration = exp
	shard.lruList.MoveToFront(item.LRUElement)
	return nil
}

func (c *Cache) Delete(key string) {
	shard := c.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if item, ok := shard.data[key]; ok {
		shard.lruList.Remove(item.LRUElement)
		delete(shard.data, key)
		cacheItemPool.Put(item)
	}
}
// Iterate
func (c *Cache) Iterate(fn func(key string, value []byte)) {
	now := time.Now().UnixNano()
	var wg sync.WaitGroup
	wg.Add(len(c.shards))

	for _, shard := range c.shards {
		go func(s *CacheShard) {
			defer wg.Done()
			s.mu.RLock()
			for k, item := range s.data {
				if now <= item.Expiration {
					fn(k, item.Value)
				}
			}
			s.mu.RUnlock()
		}(shard)
	}
	wg.Wait()
}
//  Cleanup
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		for _, shard := range c.shards {
			c.cleanupShard(shard)
		}
	}
}

func (c *Cache) cleanupShard(shard *CacheShard) {
	shard.mu.Lock()
	defer shard.mu.Unlock()
	for key, item := range shard.data {
		if time.Now().UnixNano() > item.Expiration {
			shard.lruList.Remove(item.LRUElement)
			delete(shard.data, key)
			cacheItemPool.Put(item)
		}
	}
}



//  CleanupAll

func (c *Cache) CleanupAll() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		for key, item := range shard.data {
			shard.lruList.Remove(item.LRUElement)
			cacheItemPool.Put(item)
			delete(shard.data, key)
		}
		shard.mu.Unlock()
	}
}
