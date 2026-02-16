package settlement

import (
	"errors"
	"fmt"
	"sort"

	"github.com/example/fanji-backend/internal/domain"
)

var (
	ErrInvalidPlayers = errors.New("round requires exactly 4 players")
)

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Settle(round domain.RoundInput) (domain.RoundSettlement, error) {
	if len(round.Players) != 4 {
		return domain.RoundSettlement{}, ErrInvalidPlayers
	}

	transfers := map[string]int{}
	for _, p := range round.Players {
		transfers[p.ID] = 0
	}
	selected := map[string]domain.HandType{}
	notes := []string{}

	if round.IsDraw {
		notes = append(notes, "流局：杠分归零，不翻鸡，不结算鸡钱。")
		return domain.RoundSettlement{PointTransfers: transfers, SelectedHandTypes: selected, Notes: notes}, nil
	}

	if round.WinEvent != nil {
		if err := e.applyWinSettlement(*round.WinEvent, round, transfers, selected, &notes); err != nil {
			return domain.RoundSettlement{}, err
		}
		e.applyChickenSettlement(round, transfers, &notes)
	}

	if round.KongDiscardPenaltyPlayer != "" {
		for _, p := range round.Players {
			if p.ID == round.KongDiscardPenaltyPlayer && p.KongPointsBeforePenalty > 0 {
				transfers[p.ID] -= p.KongPointsBeforePenalty
				notes = append(notes, fmt.Sprintf("杠后点炮惩罚：%s 杠分 %d 作废。", p.ID, p.KongPointsBeforePenalty))
			}
		}
	}

	return domain.RoundSettlement{PointTransfers: transfers, SelectedHandTypes: selected, Notes: notes}, nil
}

func (e *Engine) SelectHandType(ctx domain.HandContext) domain.HandType {
	if len(ctx.Candidates) == 0 {
		return domain.HandPingHu
	}
	for _, c := range ctx.Candidates {
		if c == domain.HandPureSuit {
			return domain.HandPureSuit
		}
	}

	candidates := append([]domain.HandType(nil), ctx.Candidates...)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Points() == candidates[j].Points() {
			return candidates[i].Priority() < candidates[j].Priority()
		}
		return candidates[i].Points() < candidates[j].Points()
	})
	return candidates[0]
}

func (e *Engine) applyWinSettlement(ev domain.WinEvent, round domain.RoundInput, transfers map[string]int, selected map[string]domain.HandType, notes *[]string) error {
	switch ev.Type {
	case domain.WinSelfDraw:
		hand, err := e.resolveWinnerHand(ev.Winner, round)
		if err != nil {
			return err
		}
		selected[ev.Winner] = hand
		points := hand.Points()
		for _, p := range round.Players {
			if p.ID == ev.Winner {
				continue
			}
			transfers[p.ID] -= points
			transfers[ev.Winner] += points
		}
		*notes = append(*notes, fmt.Sprintf("自摸：%s 收三家。", ev.Winner))
	case domain.WinMultiWin:
		for _, w := range ev.Winners {
			hand, err := e.resolveWinnerHand(w, round)
			if err != nil {
				return err
			}
			selected[w] = hand
			points := hand.Points()
			transfers[ev.Discarder] -= points
			transfers[w] += points
		}
		*notes = append(*notes, fmt.Sprintf("一炮多响：%s 分别支付。", ev.Discarder))
	case domain.WinRobbingKong:
		for _, w := range ev.Winners {
			hand, err := e.resolveWinnerHand(w, round)
			if err != nil {
				return err
			}
			selected[w] = hand
			points := hand.Points() * 3
			transfers[ev.RobbedPlayer] -= points
			transfers[w] += points
		}
		*notes = append(*notes, fmt.Sprintf("抢杠胡：%s 单人全包 3 倍。", ev.RobbedPlayer))
	}
	return nil
}

func (e *Engine) applyChickenSettlement(round domain.RoundInput, transfers map[string]int, notes *[]string) {
	listeners := make([]domain.PlayerState, 0, 4)
	nonListeners := make([]domain.PlayerState, 0, 4)
	for _, p := range round.Players {
		if p.IsListening {
			listeners = append(listeners, p)
		} else {
			nonListeners = append(nonListeners, p)
		}
	}

	for i := 0; i < len(listeners); i++ {
		for j := i + 1; j < len(listeners); j++ {
			a, b := listeners[i], listeners[j]
			d := a.ChickenPoints - b.ChickenPoints
			if d > 0 {
				transfers[a.ID] += d
				transfers[b.ID] -= d
			} else if d < 0 {
				transfers[b.ID] += -d
				transfers[a.ID] -= -d
			}
		}
	}

	for _, n := range nonListeners {
		for _, l := range listeners {
			if l.ChickenPoints <= 0 {
				continue
			}
			transfers[n.ID] -= l.ChickenPoints
			transfers[l.ID] += l.ChickenPoints
		}
	}

	if len(listeners) > 0 {
		*notes = append(*notes, "鸡分结算：听牌对冲 + 未听牌全赔。")
	}
}

func (e *Engine) resolveWinnerHand(winner string, round domain.RoundInput) (domain.HandType, error) {
	ctx, ok := round.HandsByWinner[winner]
	if !ok {
		return "", fmt.Errorf("missing hand context for winner: %s", winner)
	}
	return e.SelectHandType(ctx), nil
}
