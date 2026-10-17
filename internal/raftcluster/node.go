package raftcluster

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Transport defines the communication interface between Raft nodes.
type Transport interface {
	SendVote(ctx context.Context, target string, req VoteRequest) (*VoteResponse, error)
	SendAppendEntries(ctx context.Context, target string, req AppendEntriesRequest) (*AppendEntriesResponse, error)
}

// Config configures a Raft cluster node.
type Config struct {
	NodeID            string
	Peers             []string
	ElectionTimeoutMin time.Duration
	ElectionTimeoutMax time.Duration
	HeartbeatInterval  time.Duration
	Transport         Transport
}

// Node coordinates Raft consensus, leader election, log replication, and FSM commits.
type Node struct {
	cfg Config
	fsm *FSM

	mu          sync.RWMutex
	state       NodeState
	currentTerm uint64
	votedFor    string
	leaderID    string

	log         []LogEntry
	commitIndex uint64
	lastApplied uint64

	peers map[string]struct{}

	started   atomic.Bool
	stopCh    chan struct{}
	resetTimer chan struct{}
	startTime time.Time
}

// NewNode instantiates a new Raft consensus node.
func NewNode(cfg Config, fsm *FSM) *Node {
	if cfg.NodeID == "" {
		cfg.NodeID = fmt.Sprintf("raft-%d", time.Now().UnixNano()%10000)
	}
	if cfg.ElectionTimeoutMin <= 0 {
		cfg.ElectionTimeoutMin = 150 * time.Millisecond
	}
	if cfg.ElectionTimeoutMax <= 0 {
		cfg.ElectionTimeoutMax = 300 * time.Millisecond
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = 50 * time.Millisecond
	}
	if fsm == nil {
		fsm = NewFSM()
	}

	peerMap := make(map[string]struct{})
	for _, p := range cfg.Peers {
		if p != cfg.NodeID {
			peerMap[p] = struct{}{}
		}
	}

	return &Node{
		cfg:        cfg,
		fsm:        fsm,
		state:      StateFollower,
		peers:      peerMap,
		log:        []LogEntry{{Index: 0, Term: 0}}, // 1-based log entry indexing
		stopCh:     make(chan struct{}),
		resetTimer: make(chan struct{}, 1),
		startTime:  time.Now(),
	}
}

// Start begins background election timeouts and leader heartbeats.
func (n *Node) Start() error {
	if n.started.Swap(true) {
		return nil
	}
	go n.runLoop()
	return nil
}

// Stop shuts down the consensus node.
func (n *Node) Stop() {
	if !n.started.Swap(false) {
		return
	}
	close(n.stopCh)
}

// Status returns cluster telemetry.
func (n *Node) Status() ClusterStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var peerList []string
	for p := range n.peers {
		peerList = append(peerList, p)
	}

	return ClusterStatus{
		NodeID:      n.cfg.NodeID,
		State:       n.state,
		Term:        n.currentTerm,
		LeaderID:    n.leaderID,
		CommitIndex: n.commitIndex,
		LastApplied: n.lastApplied,
		PeerCount:   len(n.peers),
		Peers:       peerList,
		IsLeader:    n.state == StateLeader,
		Uptime:      time.Since(n.startTime).Round(time.Second).String(),
	}
}

// FSM returns the underlying state machine.
func (n *Node) FSM() *FSM {
	return n.fsm
}

// Propose applies a state command through Raft consensus.
func (n *Node) Propose(ctx context.Context, cmd Command) error {
	n.mu.Lock()
	if n.state != StateLeader {
		leader := n.leaderID
		n.mu.Unlock()
		return fmt.Errorf("raft: not cluster leader (current leader is %s)", leader)
	}

	cmd.Timestamp = time.Now().UnixNano()
	newIndex := uint64(len(n.log))
	entry := LogEntry{
		Index:     newIndex,
		Term:      n.currentTerm,
		Command:   cmd,
		Timestamp: cmd.Timestamp,
	}
	n.log = append(n.log, entry)
	n.mu.Unlock()

	// If single node cluster, commit immediately
	if len(n.peers) == 0 {
		n.mu.Lock()
		n.commitIndex = newIndex
		_ = n.applyEntriesLocked()
		n.mu.Unlock()
		return nil
	}

	// Replicate to quorum
	successCount := 1 // self
	quorum := (len(n.peers) + 1)/2 + 1
	var replMu sync.Mutex

	var wg sync.WaitGroup
	for peer := range n.peers {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			if n.cfg.Transport == nil {
				return
			}
			n.mu.RLock()
			req := AppendEntriesRequest{
				Term:         n.currentTerm,
				LeaderID:     n.cfg.NodeID,
				PrevLogIndex: newIndex - 1,
				PrevLogTerm:  n.log[newIndex-1].Term,
				Entries:      []LogEntry{entry},
				LeaderCommit: n.commitIndex,
			}
			n.mu.RUnlock()

			resp, err := n.cfg.Transport.SendAppendEntries(ctx, target, req)
			if err == nil && resp.Success {
				replMu.Lock()
				successCount++
				replMu.Unlock()
			}
		}(peer)
	}
	wg.Wait()

	if successCount >= quorum {
		n.mu.Lock()
		if newIndex > n.commitIndex {
			n.commitIndex = newIndex
			_ = n.applyEntriesLocked()
		}
		n.mu.Unlock()
		return nil
	}

	return errors.New("raft: failed to replicate to quorum")
}

func (n *Node) applyEntriesLocked() error {
	for n.commitIndex > n.lastApplied {
		n.lastApplied++
		entry := n.log[n.lastApplied]
		if err := n.fsm.Apply(entry.Command); err != nil {
			return err
		}
	}
	return nil
}

// RequestVote handles an incoming vote request from a candidate.
func (n *Node) RequestVote(req VoteRequest) VoteResponse {
	n.mu.Lock()
	defer n.mu.Unlock()

	if req.Term > n.currentTerm {
		n.currentTerm = req.Term
		n.state = StateFollower
		n.votedFor = ""
	}

	resp := VoteResponse{Term: n.currentTerm, VoteGranted: false}

	if req.Term == n.currentTerm && (n.votedFor == "" || n.votedFor == req.CandidateID) {
		lastLogIndex := uint64(len(n.log) - 1)
		lastLogTerm := n.log[lastLogIndex].Term

		if req.LastLogTerm > lastLogTerm || (req.LastLogTerm == lastLogTerm && req.LastLogIndex >= lastLogIndex) {
			resp.VoteGranted = true
			n.votedFor = req.CandidateID
			n.triggerResetTimer()
		}
	}

	return resp
}

// AppendEntries handles log replication and leader heartbeats.
func (n *Node) AppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	n.mu.Lock()
	defer n.mu.Unlock()

	resp := AppendEntriesResponse{Term: n.currentTerm, Success: false}

	if req.Term < n.currentTerm {
		return resp
	}

	if req.Term > n.currentTerm || n.state == StateCandidate {
		n.currentTerm = req.Term
		n.state = StateFollower
		n.votedFor = ""
	}

	n.leaderID = req.LeaderID
	n.triggerResetTimer()

	// Log continuity check
	if req.PrevLogIndex >= uint64(len(n.log)) || n.log[req.PrevLogIndex].Term != req.PrevLogTerm {
		return resp
	}

	// Append new entries
	for i, entry := range req.Entries {
		idx := req.PrevLogIndex + 1 + uint64(i)
		if idx < uint64(len(n.log)) {
			if n.log[idx].Term != entry.Term {
				n.log = n.log[:idx]
				n.log = append(n.log, entry)
			}
		} else {
			n.log = append(n.log, entry)
		}
	}

	// Advance commit index
	if req.LeaderCommit > n.commitIndex {
		lastIdx := uint64(len(n.log) - 1)
		if req.LeaderCommit < lastIdx {
			n.commitIndex = req.LeaderCommit
		} else {
			n.commitIndex = lastIdx
		}
		_ = n.applyEntriesLocked()
	}

	resp.Success = true
	resp.Index = uint64(len(n.log) - 1)
	return resp
}

func (n *Node) runLoop() {
	for {
		n.mu.RLock()
		st := n.state
		n.mu.RUnlock()

		switch st {
		case StateFollower, StateCandidate:
			timeout := n.randomElectionTimeout()
			select {
			case <-n.stopCh:
				return
			case <-n.resetTimer:
				continue
			case <-time.After(timeout):
				n.startElection()
			}
		case StateLeader:
			select {
			case <-n.stopCh:
				return
			case <-time.After(n.cfg.HeartbeatInterval):
				n.sendHeartbeats()
			}
		}
	}
}

func (n *Node) triggerResetTimer() {
	select {
	case n.resetTimer <- struct{}{}:
	default:
	}
}

func (n *Node) randomElectionTimeout() time.Duration {
	diff := n.cfg.ElectionTimeoutMax - n.cfg.ElectionTimeoutMin
	if diff <= 0 {
		return n.cfg.ElectionTimeoutMin
	}
	extra := time.Duration(rand.Int63n(int64(diff)))
	return n.cfg.ElectionTimeoutMin + extra
}

func (n *Node) startElection() {
	n.mu.Lock()
	n.state = StateCandidate
	n.currentTerm++
	n.votedFor = n.cfg.NodeID
	term := n.currentTerm
	lastLogIndex := uint64(len(n.log) - 1)
	lastLogTerm := n.log[lastLogIndex].Term
	peers := make([]string, 0, len(n.peers))
	for p := range n.peers {
		peers = append(peers, p)
	}
	n.mu.Unlock()

	// If single node cluster, immediately become leader
	if len(peers) == 0 {
		n.mu.Lock()
		n.state = StateLeader
		n.leaderID = n.cfg.NodeID
		n.mu.Unlock()
		return
	}

	votesReceived := 1 // self
	quorum := (len(peers) + 1)/2 + 1
	var voteMu sync.Mutex

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), n.cfg.ElectionTimeoutMin)
	defer cancel()

	for _, peer := range peers {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			if n.cfg.Transport == nil {
				return
			}
			req := VoteRequest{
				Term:         term,
				CandidateID:  n.cfg.NodeID,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}
			resp, err := n.cfg.Transport.SendVote(ctx, target, req)
			if err == nil && resp.VoteGranted {
				voteMu.Lock()
				votesReceived++
				voteMu.Unlock()
			}
		}(peer)
	}
	wg.Wait()

	n.mu.Lock()
	defer n.mu.Unlock()
	if n.state == StateCandidate && n.currentTerm == term && votesReceived >= quorum {
		n.state = StateLeader
		n.leaderID = n.cfg.NodeID
	}
}

func (n *Node) sendHeartbeats() {
	n.mu.RLock()
	if n.state != StateLeader {
		n.mu.RUnlock()
		return
	}
	term := n.currentTerm
	commit := n.commitIndex
	lastLogIdx := uint64(len(n.log) - 1)
	lastLogTerm := n.log[lastLogIdx].Term
	peers := make([]string, 0, len(n.peers))
	for p := range n.peers {
		peers = append(peers, p)
	}
	n.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), n.cfg.HeartbeatInterval)
	defer cancel()

	for _, peer := range peers {
		go func(target string) {
			if n.cfg.Transport == nil {
				return
			}
			req := AppendEntriesRequest{
				Term:         term,
				LeaderID:     n.cfg.NodeID,
				PrevLogIndex: lastLogIdx,
				PrevLogTerm:  lastLogTerm,
				Entries:      nil,
				LeaderCommit: commit,
			}
			_, _ = n.cfg.Transport.SendAppendEntries(ctx, target, req)
		}(peer)
	}
}
