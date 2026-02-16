package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/example/fanji-backend/internal/domain"
	"github.com/example/fanji-backend/internal/messaging"
	"github.com/example/fanji-backend/internal/realtime"
	"github.com/example/fanji-backend/internal/settlement"
	"github.com/example/fanji-backend/internal/storage"
)

type Server struct {
	settlement *settlement.Engine
	hub        *realtime.Hub
	redis      *storage.RedisStore
	mysql      *storage.MySQLStore
	publisher  messaging.EventPublisher

	sessionMu sync.RWMutex
	sessions  map[string]domain.User
	rooms     []domain.RoomSummary
}

func NewServer(redis *storage.RedisStore, mysql *storage.MySQLStore, publisher messaging.EventPublisher) *Server {
	return &Server{
		settlement: settlement.NewEngine(),
		hub:        realtime.NewHub(),
		redis:      redis,
		mysql:      mysql,
		publisher:  publisher,
		sessions:   map[string]domain.User{},
		rooms: []domain.RoomSummary{
			{ID: "r-1001", Name: "好友房-西安茶馆", CurrentPeople: 2, Capacity: 4, Mode: "friend"},
			{ID: "r-2001", Name: "金币初级场", CurrentPeople: 4, Capacity: 4, Mode: "ranked"},
			{ID: "r-2002", Name: "金币进阶场", CurrentPeople: 3, Capacity: 4, Mode: "ranked"},
		},
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/logout", s.withAuth(s.handleLogout))
	mux.HandleFunc("/api/hall", s.withAuth(s.handleHall))
	mux.HandleFunc("/api/history", s.withAuth(s.handleHistory))
	mux.HandleFunc("/api/stats/headtohead", s.withAuth(s.handleHeadToHead))
	mux.HandleFunc("/api/settle", s.withAuth(s.handleSettle))
	mux.HandleFunc("/ws/rooms/", s.withAuth(s.handleRoomWS))
	mux.HandleFunc("/", s.handleIndex)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		http.Error(w, "username/password required", http.StatusBadRequest)
		return
	}
	user, err := s.mysql.UpsertUser(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	token, err := newToken()
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	s.sessionMu.Lock()
	s.sessions[token] = user
	s.sessionMu.Unlock()

	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	_ = json.NewEncoder(w).Encode(map[string]any{"user": user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request, user domain.User) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		s.sessionMu.Lock()
		delete(s.sessions, cookie.Value)
		s.sessionMu.Unlock()
	}
	s.mysql.SetUserOnline(r.Context(), user.ID, false)
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHall(w http.ResponseWriter, r *http.Request, _ domain.User) {
	resp := map[string]any{
		"rooms":         s.rooms,
		"onlineFriends": s.mysql.ListOnlineUsers(r.Context()),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request, user domain.User) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	history := s.mysql.UserHistory(r.Context(), user.ID, limit)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"records": history})
}

func (s *Server) handleHeadToHead(w http.ResponseWriter, r *http.Request, user domain.User) {
	opponent := r.URL.Query().Get("opponent")
	if opponent == "" {
		http.Error(w, "opponent is required", http.StatusBadRequest)
		return
	}
	stats := s.mysql.HeadToHead(r.Context(), user.ID, opponent)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleSettle(w http.ResponseWriter, r *http.Request, user domain.User) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		RoomID string            `json:"roomId"`
		Round  domain.RoundInput `json:"round"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := s.settlement.Settle(req.Round)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload, _ := json.Marshal(result)

	ctx := r.Context()
	participants := playerIDs(req.Round.Players)
	winners := winnerIDs(req.Round.WinEvent)
	if len(winners) == 0 {
		winners = deriveWinnersBySettlement(result.PointTransfers)
	}
	if req.RoomID != "" {
		if err = s.redis.SetRoomState(ctx, req.RoomID, string(payload)); err != nil {
			log.Printf("redis set state failed: %v", err)
		}
		if err = s.mysql.SaveRound(ctx, req.RoomID, payload); err != nil {
			log.Printf("mysql save replay failed: %v", err)
		}
		s.mysql.SaveMatch(ctx, req.RoomID, participants, winners, result.PointTransfers)
		s.hub.BroadcastJSON(ctx, req.RoomID, map[string]any{"type": "round_settled", "by": user.Username, "payload": result})
		if err = s.publisher.Publish(ctx, req.RoomID, payload); err != nil {
			log.Printf("kafka publish failed: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleRoomWS(w http.ResponseWriter, r *http.Request, _ domain.User) {
	path := strings.TrimPrefix(r.URL.Path, "/ws/rooms/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "connect" || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	s.hub.ServeWS(w, r, parts[0])
}

func (s *Server) withAuth(next func(http.ResponseWriter, *http.Request, domain.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.currentUser(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r, user)
	}
}

func (s *Server) currentUser(r *http.Request) (domain.User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return domain.User{}, err
	}
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()
	u, ok := s.sessions[cookie.Value]
	if !ok {
		return domain.User{}, errors.New("session not found")
	}
	return u, nil
}

func newToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func playerIDs(ps []domain.PlayerState) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.ID)
	}
	return out
}

func winnerIDs(ev *domain.WinEvent) []string {
	if ev == nil {
		return nil
	}
	switch ev.Type {
	case domain.WinSelfDraw:
		if ev.Winner != "" {
			return []string{ev.Winner}
		}
	case domain.WinMultiWin, domain.WinRobbingKong:
		return append([]string(nil), ev.Winners...)
	}
	return nil
}

func deriveWinnersBySettlement(points map[string]int) []string {
	max := -1 << 30
	for _, p := range points {
		if p > max {
			max = p
		}
	}
	if max <= 0 {
		return nil
	}
	out := []string{}
	for id, p := range points {
		if p == max {
			out = append(out, id)
		}
	}
	return out
}

func (s *Server) Close(_ context.Context) error {
	if s.publisher != nil {
		_ = s.publisher.Close()
	}
	return nil
}
