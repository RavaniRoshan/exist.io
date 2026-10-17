// Package raftcluster provides pure-Go Raft consensus state machine and cluster coordination
// for active-active metadata replication across geographically distributed exist.io nodes.
package raftcluster

// NodeState represents the role of a Raft node in the cluster.
type NodeState string

const (
	StateFollower  NodeState = "follower"
	StateCandidate NodeState = "candidate"
	StateLeader    NodeState = "leader"
)

// CommandType defines the type of state machine mutation proposed through Raft.
type CommandType string

const (
	CmdSetState    CommandType = "set_state"
	CmdCreateRepo  CommandType = "create_repo"
	CmdIssueLease  CommandType = "issue_lease"
	CmdRevokeLease CommandType = "revoke_lease"
)

// Command represents an operational payload replicated via the Raft log.
type Command struct {
	Type      CommandType       `json:"type"`
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp int64             `json:"timestamp"`
}

// LogEntry is an immutable record in the replicated Raft log.
type LogEntry struct {
	Index     uint64  `json:"index"`
	Term      uint64  `json:"term"`
	Command   Command `json:"command"`
	Timestamp int64   `json:"timestamp"`
}

// ClusterStatus exposes telemetry and topology of the Raft cluster node.
type ClusterStatus struct {
	NodeID      string    `json:"nodeId"`
	State       NodeState `json:"state"`
	Term        uint64    `json:"term"`
	LeaderID    string    `json:"leaderId"`
	CommitIndex uint64    `json:"commitIndex"`
	LastApplied uint64    `json:"lastApplied"`
	PeerCount   int       `json:"peerCount"`
	Peers       []string  `json:"peers"`
	IsLeader    bool      `json:"isLeader"`
	Uptime      string    `json:"uptime"`
}

// VoteRequest is sent by candidate nodes to request leader votes.
type VoteRequest struct {
	Term         uint64 `json:"term"`
	CandidateID  string `json:"candidateId"`
	LastLogIndex uint64 `json:"lastLogIndex"`
	LastLogTerm  uint64 `json:"lastLogTerm"`
}

// VoteResponse is returned by nodes evaluating a vote request.
type VoteResponse struct {
	Term        uint64 `json:"term"`
	VoteGranted bool   `json:"voteGranted"`
}

// AppendEntriesRequest is sent by leaders to replicate log entries and heartbeat.
type AppendEntriesRequest struct {
	Term         uint64     `json:"term"`
	LeaderID     string     `json:"leaderId"`
	PrevLogIndex uint64     `json:"prevLogIndex"`
	PrevLogTerm  uint64     `json:"prevLogTerm"`
	Entries      []LogEntry `json:"entries"`
	LeaderCommit uint64     `json:"leaderCommit"`
}

// AppendEntriesResponse is returned by followers after processing AppendEntries.
type AppendEntriesResponse struct {
	Term    uint64 `json:"term"`
	Success bool   `json:"success"`
	Index   uint64 `json:"index"`
}
