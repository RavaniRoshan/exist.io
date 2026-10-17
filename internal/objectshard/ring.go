package objectshard

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
)

// HashRing implements consistent hashing with virtual nodes.
type HashRing struct {
	mu           sync.RWMutex
	vnodes       int
	ring         []uint32
	ringMap      map[uint32]string
	shards       map[string]ShardInfo
}

// NewHashRing creates a new consistent hash ring.
func NewHashRing(vnodes int) *HashRing {
	if vnodes <= 0 {
		vnodes = 50
	}
	return &HashRing{
		vnodes:  vnodes,
		ringMap: make(map[uint32]string),
		shards:  make(map[string]ShardInfo),
	}
}

// AddShard registers a storage shard on the ring.
func (h *HashRing) AddShard(info ShardInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.shards[info.ID] = info
	for i := 0; i < h.vnodes; i++ {
		vnodeKey := fmt.Sprintf("%s#%d", info.ID, i)
		hash := hashKey(vnodeKey)
		h.ring = append(h.ring, hash)
		h.ringMap[hash] = info.ID
	}
	sort.Slice(h.ring, func(i, j int) bool { return h.ring[i] < h.ring[j] })
}

// RemoveShard removes a storage shard from the ring.
func (h *HashRing) RemoveShard(shardID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.shards, shardID)
	var newRing []uint32
	for _, hash := range h.ring {
		if h.ringMap[hash] == shardID {
			delete(h.ringMap, hash)
		} else {
			newRing = append(newRing, hash)
		}
	}
	h.ring = newRing
}

// Locate determines which shard is responsible for the given object key.
func (h *HashRing) Locate(key string) (string, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.ring) == 0 {
		return "", fmt.Errorf("hashring: no active storage shards available")
	}

	hash := hashKey(key)
	idx := sort.Search(len(h.ring), func(i int) bool { return h.ring[i] >= hash })
	if idx == len(h.ring) {
		idx = 0
	}

	targetHash := h.ring[idx]
	return h.ringMap[targetHash], nil
}

// Shards returns a list of all active shards on the ring.
func (h *HashRing) Shards() []ShardInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	res := make([]ShardInfo, 0, len(h.shards))
	for _, s := range h.shards {
		res = append(res, s)
	}
	return res
}

func hashKey(key string) uint32 {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	sum := hasher.Sum(nil)
	return binary.BigEndian.Uint32(sum[:4])
}
