package objectshard

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// LRUCache provides a thread-safe least-recently-used object cache.
type LRUCache struct {
	mu         sync.Mutex
	capacity   int
	items      map[string]*list.Element
	evictList  *list.List
}

type cacheEntry struct {
	key   string
	value []byte
}

// NewLRUCache creates a new bounded LRU cache.
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 500
	}
	return &LRUCache{
		capacity:  capacity,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

// Get retrieves an item from the cache.
func (c *LRUCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return nil, false
}

// Put adds an item to the cache, evicting the oldest if capacity is exceeded.
func (c *LRUCache) Put(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return
	}

	entry := &cacheEntry{key: key, value: value}
	elem := c.evictList.PushFront(entry)
	c.items[key] = elem

	if c.evictList.Len() > c.capacity {
		oldest := c.evictList.Back()
		if oldest != nil {
			c.evictList.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).key)
		}
	}
}

// Manager coordinates distributed Git object sharding and local caching.
type Manager struct {
	ring        *HashRing
	cache       *LRUCache
	shardsStore map[string]map[string][]byte
	mu          sync.RWMutex

	totalObjects int64
	totalBytes   int64
	cacheHits    int64
	cacheMisses  int64
}

// Config configures the distributed object manager.
type Config struct {
	Shards        []ShardInfo
	CacheCapacity int
}

// NewManager creates an instance of the distributed object manager.
func NewManager(cfg Config) *Manager {
	ring := NewHashRing(50)
	for _, s := range cfg.Shards {
		ring.AddShard(s)
	}

	return &Manager{
		ring:        ring,
		cache:       NewLRUCache(cfg.CacheCapacity),
		shardsStore: make(map[string]map[string][]byte),
	}
}

// PutObject stores a Git object onto its hashed destination shard.
func (m *Manager) PutObject(ctx context.Context, sha string, data []byte) (string, error) {
	shardID, err := m.ring.Locate(sha)
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	if _, ok := m.shardsStore[shardID]; !ok {
		m.shardsStore[shardID] = make(map[string][]byte)
	}
	m.shardsStore[shardID][sha] = data
	m.mu.Unlock()

	// Update local LRU cache
	m.cache.Put(sha, data)

	atomic.AddInt64(&m.totalObjects, 1)
	atomic.AddInt64(&m.totalBytes, int64(len(data)))

	return shardID, nil
}

// GetObject retrieves a Git object, checking the local LRU cache first.
func (m *Manager) GetObject(ctx context.Context, sha string) ([]byte, error) {
	// Check LRU cache
	if val, ok := m.cache.Get(sha); ok {
		atomic.AddInt64(&m.cacheHits, 1)
		return val, nil
	}

	atomic.AddInt64(&m.cacheMisses, 1)

	shardID, err := m.ring.Locate(sha)
	if err != nil {
		return nil, err
	}

	m.mu.RLock()
	shardData, ok := m.shardsStore[shardID]
	if !ok {
		m.mu.RUnlock()
		return nil, errors.New("objectshard: shard store not found")
	}
	data, exists := shardData[sha]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("objectshard: object %s not found on shard %s", sha, shardID)
	}

	// Warm the cache on miss
	m.cache.Put(sha, data)
	return data, nil
}

// HasObject returns whether the object exists on the cluster.
func (m *Manager) HasObject(ctx context.Context, sha string) bool {
	if _, ok := m.cache.Get(sha); ok {
		return true
	}

	shardID, err := m.ring.Locate(sha)
	if err != nil {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	if shardData, ok := m.shardsStore[shardID]; ok {
		_, exists := shardData[sha]
		return exists
	}
	return false
}

// Stats returns real-time sharding and cache performance metrics.
func (m *Manager) Stats() ShardStats {
	hits := atomic.LoadInt64(&m.cacheHits)
	misses := atomic.LoadInt64(&m.cacheMisses)
	totalReqs := hits + misses

	var hitRatio float64
	if totalReqs > 0 {
		hitRatio = float64(hits) / float64(totalReqs)
	}

	return ShardStats{
		TotalObjectsStored: atomic.LoadInt64(&m.totalObjects),
		TotalBytesStored:   atomic.LoadInt64(&m.totalBytes),
		CacheHitCount:      hits,
		CacheMissCount:     misses,
		CacheHitRatio:      hitRatio,
		ActiveShardsCount:  len(m.ring.Shards()),
		Shards:             m.ring.Shards(),
	}
}
