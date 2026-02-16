package storage

import (
	"context"
	"sync"
)

type RedisStore struct {
	mu     sync.Mutex
	states map[string]string
}

func NewRedisStore(_ string, _ string) *RedisStore {
	return &RedisStore{states: map[string]string{}}
}

func (r *RedisStore) SetRoomState(_ context.Context, roomID string, state string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[roomID] = state
	return nil
}
