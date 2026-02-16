package storage

import (
	"context"
	"sync"
)

type MySQLStore struct {
	mu      sync.Mutex
	replays map[string][][]byte
}

func NewMySQLStore(_ string) (*MySQLStore, error) {
	return &MySQLStore{replays: map[string][][]byte{}}, nil
}

func (s *MySQLStore) SaveRound(_ context.Context, roomID string, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.replays[roomID] = append(s.replays[roomID], append([]byte(nil), payload...))
	return nil
}
