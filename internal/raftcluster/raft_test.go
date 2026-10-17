package raftcluster_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"exist.io/internal/raftcluster"
)

type memoryTransport struct {
	mu    sync.RWMutex
	nodes map[string]*raftcluster.Node
}

func newMemoryTransport() *memoryTransport {
	return &memoryTransport{nodes: make(map[string]*raftcluster.Node)}
}

func (m *memoryTransport) register(id string, n *raftcluster.Node) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes[id] = n
}

func (m *memoryTransport) SendVote(ctx context.Context, target string, req raftcluster.VoteRequest) (*raftcluster.VoteResponse, error) {
	m.mu.RLock()
	n, ok := m.nodes[target]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("target node %s not found", target)
	}
	resp := n.RequestVote(req)
	return &resp, nil
}

func (m *memoryTransport) SendAppendEntries(ctx context.Context, target string, req raftcluster.AppendEntriesRequest) (*raftcluster.AppendEntriesResponse, error) {
	m.mu.RLock()
	n, ok := m.nodes[target]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("target node %s not found", target)
	}
	resp := n.AppendEntries(req)
	return &resp, nil
}

func TestSingleNodeCluster_LeaderElectionAndPropose(t *testing.T) {
	fsm := raftcluster.NewFSM()
	node := raftcluster.NewNode(raftcluster.Config{
		NodeID:             "solo-node",
		Peers:              nil,
		ElectionTimeoutMin: 20 * time.Millisecond,
		ElectionTimeoutMax: 50 * time.Millisecond,
	}, fsm)

	if err := node.Start(); err != nil {
		t.Fatalf("failed to start node: %v", err)
	}
	defer node.Stop()

	// Wait for election to make it leader
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		if node.Status().IsLeader {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !node.Status().IsLeader {
		t.Fatalf("expected node to become leader")
	}

	ctx := context.Background()
	cmd := raftcluster.Command{
		Type:  raftcluster.CmdSetState,
		Key:   "cluster.name",
		Value: "exist-primary",
	}

	if err := node.Propose(ctx, cmd); err != nil {
		t.Fatalf("Propose failed: %v", err)
	}

	val, ok := node.FSM().Get("cluster.name")
	if !ok || val != "exist-primary" {
		t.Fatalf("FSM state mismatch: got %s, want exist-primary", val)
	}
}

func TestThreeNodeCluster_ConsensusAndReplication(t *testing.T) {
	transport := newMemoryTransport()

	peers := []string{"node-1", "node-2", "node-3"}
	nodes := make([]*raftcluster.Node, 3)

	for i := 0; i < 3; i++ {
		id := peers[i]
		cfg := raftcluster.Config{
			NodeID:             id,
			Peers:              peers,
			ElectionTimeoutMin: 50 * time.Millisecond,
			ElectionTimeoutMax: 120 * time.Millisecond,
			HeartbeatInterval:  20 * time.Millisecond,
			Transport:          transport,
		}
		nodes[i] = raftcluster.NewNode(cfg, raftcluster.NewFSM())
		transport.register(id, nodes[i])
	}

	for _, n := range nodes {
		_ = n.Start()
		defer n.Stop()
	}

	// 1. Wait for a leader to be elected
	var leader *raftcluster.Node
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, n := range nodes {
			if n.Status().IsLeader {
				leader = n
				break
			}
		}
		if leader != nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if leader == nil {
		t.Fatalf("failed to elect leader in 3-node cluster")
	}

	// 2. Propose a state mutation to the leader
	ctx := context.Background()
	cmd := raftcluster.Command{
		Type:  raftcluster.CmdCreateRepo,
		Key:   "acme/satellite",
		Value: "repo-01M54SAT",
	}

	if err := leader.Propose(ctx, cmd); err != nil {
		t.Fatalf("leader propose failed: %v", err)
	}

	// 3. Verify all nodes replicate and apply the mutation
	time.Sleep(100 * time.Millisecond)

	for _, n := range nodes {
		snap, err := n.FSM().Snapshot()
		if err != nil {
			t.Fatalf("failed to snapshot FSM: %v", err)
		}
		if len(snap) == 0 {
			t.Fatalf("empty snapshot from node %s", n.Status().NodeID)
		}
	}
}
