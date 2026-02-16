package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: map[string]map[chan []byte]struct{}{}}
}

// ServeWS uses SSE transport under websocket-compatible route.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, roomID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	msgCh := make(chan []byte, 16)
	h.addConn(roomID, msgCh)
	defer h.removeConn(roomID, msgCh)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgCh:
			_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (h *Hub) BroadcastJSON(_ context.Context, roomID string, event any) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	chs := h.rooms[roomID]
	h.mu.RUnlock()
	for ch := range chs {
		select {
		case ch <- data:
		default:
		}
	}
}

func (h *Hub) addConn(roomID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = map[chan []byte]struct{}{}
	}
	h.rooms[roomID][ch] = struct{}{}
}

func (h *Hub) removeConn(roomID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[roomID]; ok {
		delete(room, ch)
		close(ch)
		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
	}
}
