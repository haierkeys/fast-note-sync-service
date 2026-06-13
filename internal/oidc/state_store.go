package oidc

import (
	"sync"
	"time"
)

type State struct {
	ProviderID   string
	State        string
	Nonce        string
	CodeVerifier string
	RedirectTo   string
	CreatedAt    time.Time
}

type StateStore struct {
	mu     sync.Mutex
	ttl    time.Duration
	states map[string]State
}

func NewStateStore(ttl time.Duration) *StateStore {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &StateStore{
		ttl:    ttl,
		states: make(map[string]State),
	}
}

func (s *StateStore) Save(state State) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state.CreatedAt.IsZero() {
		state.CreatedAt = time.Now()
	}
	s.states[state.State] = state
}

func (s *StateStore) Consume(value string) (State, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[value]
	if !ok {
		return State{}, false
	}
	delete(s.states, value)
	if time.Since(state.CreatedAt) > s.ttl {
		return State{}, false
	}
	return state, true
}
