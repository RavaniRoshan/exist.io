package objectshard_test

import (
	"context"
	"testing"

	"exist.io/internal/objectshard"
)

func TestHashRing_DeterminismAndLocate(t *testing.T) {
	ring := objectshard.NewHashRing(20)

	shards := []objectshard.ShardInfo{
		{ID: "shard-us-east-1", Endpoint: "s3.us-east-1.amazonaws.com", Active: true},
		{ID: "shard-eu-central-1", Endpoint: "s3.eu-central-1.amazonaws.com", Active: true},
		{ID: "shard-ap-southeast-1", Endpoint: "s3.ap-southeast-1.amazonaws.com", Active: true},
	}

	for _, s := range shards {
		ring.AddShard(s)
	}

	if len(ring.Shards()) != 3 {
		t.Fatalf("expected 3 shards, got %d", len(ring.Shards()))
	}

	key := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
	shard1, err := ring.Locate(key)
	if err != nil {
		t.Fatalf("locate failed: %v", err)
	}

	// Must be deterministic
	shard2, err := ring.Locate(key)
	if err != nil || shard1 != shard2 {
		t.Fatalf("expected deterministic locate, got %s vs %s", shard1, shard2)
	}
}

func TestManager_PutGetAndLRUCache(t *testing.T) {
	ctx := context.Background()

	mgr := objectshard.NewManager(objectshard.Config{
		Shards: []objectshard.ShardInfo{
			{ID: "bucket-alpha", Endpoint: "minio-1:9000", Active: true},
			{ID: "bucket-beta", Endpoint: "minio-2:9000", Active: true},
		},
		CacheCapacity: 2, // Tiny capacity to test eviction
	})

	data1 := []byte("blob 12\x00Hello World!")
	sha1 := "557db03de997c86a4a028e1ebd3a1ceb225be238"

	shardID, err := mgr.PutObject(ctx, sha1, data1)
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}
	if shardID == "" {
		t.Fatalf("expected valid shardID")
	}

	if !mgr.HasObject(ctx, sha1) {
		t.Fatalf("HasObject expected true")
	}

	// First Get -> Hits LRU cache because it was Put
	got1, err := mgr.GetObject(ctx, sha1)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if string(got1) != string(data1) {
		t.Fatalf("got %s, want %s", got1, data1)
	}

	stats := mgr.Stats()
	if stats.CacheHitCount < 1 {
		t.Errorf("expected at least 1 cache hit, got %d", stats.CacheHitCount)
	}

	// Put 2 more objects to exceed cache capacity 2 and force eviction
	data2 := []byte("blob 10\x00Commit One")
	sha2 := "0123456789abcdef0123456789abcdef01234567"
	_, _ = mgr.PutObject(ctx, sha2, data2)

	data3 := []byte("blob 10\x00Commit Two")
	sha3 := "abcdef0123456789abcdef0123456789abcdef01"
	_, _ = mgr.PutObject(ctx, sha3, data3)

	// Now sha1 should have been evicted from LRU cache, but still retrievable from shard store
	gotEvicted, err := mgr.GetObject(ctx, sha1)
	if err != nil {
		t.Fatalf("failed to retrieve evicted object from shard: %v", err)
	}
	if string(gotEvicted) != string(data1) {
		t.Fatalf("data mismatch after cache miss: %s", gotEvicted)
	}

	statsAfter := mgr.Stats()
	if statsAfter.CacheMissCount < 1 {
		t.Errorf("expected at least 1 cache miss, got %d", statsAfter.CacheMissCount)
	}
}
