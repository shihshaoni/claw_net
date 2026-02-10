package eventstore

import (
	"sync"
)

type Store interface {
	Append(runID string, e Event) (Event, error)
	List(runID string) ([]Event, error)
}

type InMemoryStore struct {
	mu     sync.Mutex
	events map[string][]Event
	seq    map[string]int64
}

func NewInMemory() *InMemoryStore {
	return &InMemoryStore{
		events: map[string][]Event{},
		seq:    map[string]int64{},
	}
}

func (s *InMemoryStore) Append(runID string, e Event) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq[runID]++
	e.SeqNo = s.seq[runID]
	s.events[runID] = append(s.events[runID], e)
	return e, nil
}

func (s *InMemoryStore) List(runID string) ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Event, len(s.events[runID]))
	copy(out, s.events[runID])
	return out, nil
}
