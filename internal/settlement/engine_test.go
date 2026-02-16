package settlement

import (
	"testing"

	"github.com/example/fanji-backend/internal/domain"
)

func TestSelectHandTypeLowestPointsExceptPureSuit(t *testing.T) {
	eng := NewEngine()
	got := eng.SelectHandType(domain.HandContext{Candidates: []domain.HandType{
		domain.HandDragonSevenPairs,
		domain.HandSingleWait,
		domain.HandPingHu,
	}})
	if got != domain.HandPingHu {
		t.Fatalf("expected pingHu, got %s", got)
	}
	got = eng.SelectHandType(domain.HandContext{Candidates: []domain.HandType{
		domain.HandSingleWait,
		domain.HandPureSuit,
	}})
	if got != domain.HandPureSuit {
		t.Fatalf("expected pureSuit, got %s", got)
	}
}

func TestRobbingKongMultiWinTriple(t *testing.T) {
	eng := NewEngine()
	res, err := eng.Settle(domain.RoundInput{
		Players: []domain.PlayerState{{ID: "A", IsListening: true}, {ID: "B", IsListening: true}, {ID: "C", IsListening: true}, {ID: "D", IsListening: true}},
		HandsByWinner: map[string]domain.HandContext{
			"A": {Candidates: []domain.HandType{domain.HandBigPairs}},
			"B": {Candidates: []domain.HandType{domain.HandSingleWait}},
		},
		WinEvent: &domain.WinEvent{Type: domain.WinRobbingKong, RobbedPlayer: "D", Winners: []string{"A", "B"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.PointTransfers["A"] != 12 || res.PointTransfers["B"] != 9 || res.PointTransfers["D"] != -21 {
		t.Fatalf("unexpected transfers: %#v", res.PointTransfers)
	}
}

func TestChickenSettlementAndPenalty(t *testing.T) {
	eng := NewEngine()
	res, err := eng.Settle(domain.RoundInput{
		Players: []domain.PlayerState{
			{ID: "A", IsListening: true, ChickenPoints: 5, KongPointsBeforePenalty: 6},
			{ID: "B", IsListening: true, ChickenPoints: 2},
			{ID: "C", IsListening: true, ChickenPoints: 0},
			{ID: "D", IsListening: false, ChickenPoints: 0},
		},
		HandsByWinner:            map[string]domain.HandContext{"A": {Candidates: []domain.HandType{domain.HandPingHu}}},
		WinEvent:                 &domain.WinEvent{Type: domain.WinSelfDraw, Winner: "A"},
		KongDiscardPenaltyPlayer: "A",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.PointTransfers["A"] != 13 {
		t.Fatalf("expected A=13 got %d", res.PointTransfers["A"])
	}
}

func TestDrawNoTransfers(t *testing.T) {
	eng := NewEngine()
	res, err := eng.Settle(domain.RoundInput{
		Players: []domain.PlayerState{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}},
		IsDraw:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for id, v := range res.PointTransfers {
		if v != 0 {
			t.Fatalf("expected %s zero, got %d", id, v)
		}
	}
}
