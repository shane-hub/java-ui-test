package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/example/fanji-backend/internal/domain"
)

type MySQLStore struct {
	mu         sync.Mutex
	replays    map[string][][]byte
	users      map[string]domain.User
	userByID   map[string]string
	matches    []domain.MatchRecord
	idSequence int64
}

func NewMySQLStore(_ string) (*MySQLStore, error) {
	return &MySQLStore{
		replays:  map[string][][]byte{},
		users:    map[string]domain.User{},
		userByID: map[string]string{},
		matches:  []domain.MatchRecord{},
	}, nil
}

func (s *MySQLStore) SaveRound(_ context.Context, roomID string, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.replays[roomID] = append(s.replays[roomID], append([]byte(nil), payload...))
	return nil
}

func (s *MySQLStore) UpsertUser(_ context.Context, username, password string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.users[username]; ok {
		if u.Password != password {
			return domain.User{}, errors.New("invalid password")
		}
		u.Online = true
		s.users[username] = u
		s.userByID[u.ID] = username
		return u, nil
	}
	s.idSequence++
	u := domain.User{ID: fmt.Sprintf("u-%d", s.idSequence), Username: username, Password: password, Online: true}
	s.users[username] = u
	s.userByID[u.ID] = username
	return u, nil
}

func (s *MySQLStore) SetUserOnline(_ context.Context, userID string, online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	username, ok := s.userByID[userID]
	if !ok {
		return
	}
	u := s.users[username]
	u.Online = online
	s.users[username] = u
}

func (s *MySQLStore) ListOnlineUsers(_ context.Context) []domain.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.User, 0, len(s.users))
	for _, u := range s.users {
		if u.Online {
			u.Password = ""
			out = append(out, u)
		}
	}
	return out
}

func (s *MySQLStore) SaveMatch(_ context.Context, roomID string, participants, winners []string, settlement map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idSequence++
	cp := map[string]int{}
	for k, v := range settlement {
		cp[k] = v
	}
	rec := domain.MatchRecord{
		ID:           fmt.Sprintf("m-%d", s.idSequence),
		RoomID:       roomID,
		Participants: append([]string(nil), participants...),
		Winners:      append([]string(nil), winners...),
		Settlement:   cp,
		PlayedAt:     time.Now(),
	}
	s.matches = append([]domain.MatchRecord{rec}, s.matches...)
}

func (s *MySQLStore) UserHistory(_ context.Context, userID string, limit int) []domain.MatchRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit <= 0 {
		limit = 20
	}
	out := make([]domain.MatchRecord, 0, limit)
	for _, m := range s.matches {
		found := false
		for _, p := range m.Participants {
			if p == userID {
				found = true
				break
			}
		}
		if found {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}

func (s *MySQLStore) HeadToHead(_ context.Context, me, opponent string) domain.HeadToHeadStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	stats := domain.HeadToHeadStats{Me: me, Opponent: opponent}
	for _, m := range s.matches {
		hasMe, hasOpp := false, false
		for _, p := range m.Participants {
			if p == me {
				hasMe = true
			}
			if p == opponent {
				hasOpp = true
			}
		}
		if !hasMe || !hasOpp {
			continue
		}
		stats.TotalRounds++
		meWin := contains(m.Winners, me)
		oppWin := contains(m.Winners, opponent)
		switch {
		case meWin && !oppWin:
			stats.MeWins++
		case oppWin && !meWin:
			stats.OpponentWins++
		default:
			stats.Draws++
		}
	}
	if stats.TotalRounds > 0 {
		stats.MeWinRate = float64(stats.MeWins) / float64(stats.TotalRounds)
	}
	return stats
}

func contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}
