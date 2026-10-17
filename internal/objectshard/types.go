// Package objectshard provides distributed Git packfile and object sharding with
// consistent ring hashing and local high-performance LRU read caching.
package objectshard

import "time"

// ShardInfo describes a distributed storage bucket or backend node.
type ShardInfo struct {
	ID       string `json:"id"`
	Endpoint string `json:"endpoint"`
	Capacity int64  `json:"capacityBytes"`
	Active   bool   `json:"active"`
}

// ShardStats reports telemetry across the distributed object sharding layer.
type ShardStats struct {
	TotalObjectsStored int64   `json:"totalObjectsStored"`
	TotalBytesStored   int64   `json:"totalBytesStored"`
	CacheHitCount      int64   `json:"cacheHitCount"`
	CacheMissCount     int64   `json:"cacheMissCount"`
	CacheHitRatio      float64 `json:"cacheHitRatio"`
	ActiveShardsCount  int     `json:"activeShardsCount"`
	Shards             []ShardInfo `json:"shards"`
}

// ObjectMetadata records metadata for an individual sharded Git object.
type ObjectMetadata struct {
	SHA        string    `json:"sha"`
	ShardID    string    `json:"shardId"`
	SizeBytes  int64     `json:"sizeBytes"`
	CreatedAt  time.Time `json:"createdAt"`
	Compressed bool      `json:"compressed"`
}
