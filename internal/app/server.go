package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

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
}

func NewServer(redis *storage.RedisStore, mysql *storage.MySQLStore, publisher messaging.EventPublisher) *Server {
	return &Server{
		settlement: settlement.NewEngine(),
		hub:        realtime.NewHub(),
		redis:      redis,
		mysql:      mysql,
		publisher:  publisher,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/settle", s.handleSettle)
	mux.HandleFunc("/ws/rooms/", s.handleRoomWS)
	return mux
}

func (s *Server) handleSettle(w http.ResponseWriter, r *http.Request) {
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
	if req.RoomID != "" {
		if err = s.redis.SetRoomState(ctx, req.RoomID, string(payload)); err != nil {
			log.Printf("redis set state failed: %v", err)
		}
		if err = s.mysql.SaveRound(ctx, req.RoomID, payload); err != nil {
			log.Printf("mysql save replay failed: %v", err)
		}
		s.hub.BroadcastJSON(ctx, req.RoomID, map[string]any{"type": "round_settled", "payload": result})
		if err = s.publisher.Publish(ctx, req.RoomID, payload); err != nil {
			log.Printf("kafka publish failed: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleRoomWS(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/ws/rooms/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "connect" || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	s.hub.ServeWS(w, r, parts[0])
}

func (s *Server) Close(_ context.Context) error {
	if s.publisher != nil {
		_ = s.publisher.Close()
	}
	return nil
}
