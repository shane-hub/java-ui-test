package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/example/fanji-backend/internal/domain"
	"github.com/example/fanji-backend/internal/messaging"
	"github.com/example/fanji-backend/internal/storage"
)

func TestLoginHallHistoryAndHeadToHead(t *testing.T) {
	redis := storage.NewRedisStore("", "")
	mysql, _ := storage.NewMySQLStore("inmemory")
	pub, _ := messaging.NewKafkaPublisher(nil, "")
	srv := NewServer(redis, mysql, pub)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	newClient := func() *http.Client {
		jar, _ := cookiejar.New(nil)
		return &http.Client{Jar: jar}
	}
	c1 := newClient()
	c2 := newClient()

	u1 := login(t, c1, ts.URL, "alice", "123")
	u2 := login(t, c2, ts.URL, "bob", "123")
	if u1.ID == "" || u2.ID == "" {
		t.Fatalf("expect user ids")
	}

	resp, err := c1.Get(ts.URL + "/api/hall")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("hall failed: %v status=%d", err, resp.StatusCode)
	}

	round := map[string]any{
		"roomId": "r-1001",
		"round": map[string]any{
			"players": []map[string]any{
				{"id": u1.ID, "isListening": true, "chickenPoints": 2, "kongPointsBeforePenalty": 0},
				{"id": u2.ID, "isListening": true, "chickenPoints": 0, "kongPointsBeforePenalty": 0},
				{"id": "u-3", "isListening": false, "chickenPoints": 0, "kongPointsBeforePenalty": 0},
				{"id": "u-4", "isListening": false, "chickenPoints": 0, "kongPointsBeforePenalty": 0},
			},
			"handsByWinner": map[string]any{u1.ID: map[string]any{"candidates": []string{"pingHu"}}},
			"winEvent":      map[string]any{"type": "selfDraw", "winner": u1.ID},
			"isDraw":        false,
		},
	}
	body, _ := json.Marshal(round)
	resp, err = c1.Post(ts.URL+"/api/settle", "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("settle failed: %v status=%d", err, resp.StatusCode)
	}

	resp, err = c1.Get(ts.URL + "/api/history")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("history failed: %v status=%d", err, resp.StatusCode)
	}
	var hist struct {
		Records []domain.MatchRecord `json:"records"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&hist)
	if len(hist.Records) == 0 {
		t.Fatalf("expected history records")
	}

	resp, err = c1.Get(ts.URL + "/api/stats/headtohead?opponent=" + u2.ID)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("h2h failed: %v status=%d", err, resp.StatusCode)
	}
	var stats domain.HeadToHeadStats
	_ = json.NewDecoder(resp.Body).Decode(&stats)
	if stats.TotalRounds < 1 || stats.MeWins < 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func login(t *testing.T, c *http.Client, base, username, password string) domain.User {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := c.Post(base+"/api/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status=%d", resp.StatusCode)
	}
	var out struct {
		User domain.User `json:"user"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out.User
}
