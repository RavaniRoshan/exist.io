package raftcluster

import (
	"encoding/json"
	"fmt"
	"sync"
)

// FSM is the in-memory deterministic finite state machine replicated by Raft.
type FSM struct {
	mu    sync.RWMutex
	state map[string]string
	repos map[string]string
}

// NewFSM creates a fresh finite state machine.
func NewFSM() *FSM {
	return &FSM{
		state: make(map[string]string),
		repos: make(map[string]string),
	}
}

// Apply executes a committed Raft command against the state machine.
func (f *FSM) Apply(cmd Command) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Type {
	case CmdSetState:
		f.state[cmd.Key] = cmd.Value
	case CmdCreateRepo:
		f.repos[cmd.Key] = cmd.Value
	case CmdIssueLease:
		f.state["lease:"+cmd.Key] = cmd.Value
	case CmdRevokeLease:
		delete(f.state, "lease:"+cmd.Key)
	default:
		return fmt.Errorf("fsm: unknown command type: %s", cmd.Type)
	}

	return nil
}

// Get retrieves a key from the state machine.
func (f *FSM) Get(key string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	val, ok := f.state[key]
	return val, ok
}

// Snapshot serializes the current state machine state for checkpointing.
func (f *FSM) Snapshot() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	payload := struct {
		State map[string]string `json:"state"`
		Repos map[string]string `json:"repos"`
	}{
		State: f.state,
		Repos: f.repos,
	}

	return json.Marshal(payload)
}

// Restore resets the state machine to a previously serialized snapshot.
func (f *FSM) Restore(data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var payload struct {
		State map[string]string `json:"state"`
		Repos map[string]string `json:"repos"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("fsm: restore snapshot failed: %w", err)
	}

	f.state = payload.State
	if f.state == nil {
		f.state = make(map[string]string)
	}
	f.repos = payload.Repos
	if f.repos == nil {
		f.repos = make(map[string]string)
	}

	return nil
}
