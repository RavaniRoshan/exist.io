// Package replication provides high-availability disaster recovery, WAL checkpointing,
// snapshot replication, and automated failover control for exist.io clusters.
package replication

import (
	"time"
)

// Role represents the high-availability role of an exist.io cluster node.
type Role string

const (
	// RolePrimary represents the active read-write primary node.
	RolePrimary Role = "primary"
	// RoleStandby represents a hot read-only standby replica node.
	RoleStandby Role = "standby"
)

// ReplicationStatus captures the real-time synchronization state of the cluster node.
type ReplicationStatus struct {
	NodeID                string    `json:"nodeId"`
	Role                  Role      `json:"role"`
	Health                string    `json:"health"`
	PrimaryEndpoint       string    `json:"primaryEndpoint,omitempty"`
	LastCheckpoint        time.Time `json:"lastCheckpoint"`
	LastSnapshotTimestamp time.Time `json:"lastSnapshotTimestamp"`
	LastSnapshotSeq       uint64    `json:"lastSnapshotSeq"`
	ReplicationLagSeconds float64   `json:"replicationLagSeconds"`
	WALBytesSynced        int64     `json:"walBytesSynced"`
	TotalSnapshots        int       `json:"totalSnapshots"`
	ActiveReplicasCount   int       `json:"activeReplicasCount"`
}

// SnapshotMetadata describes an immutable point-in-time database snapshot.
type SnapshotMetadata struct {
	Seq            uint64    `json:"seq"`
	Timestamp      time.Time `json:"timestamp"`
	ChecksumSHA256 string    `json:"checksumSha256"`
	SizeBytes      int64     `json:"sizeBytes"`
	SourceDB       string    `json:"sourceDb"`
	Filename       string    `json:"filename"`
}
