package domain

type HandType string

const (
	HandPingHu                 HandType = "pingHu"
	HandSingleWait             HandType = "singleWait"
	HandBigPairs               HandType = "bigPairs"
	HandMusclePosition         HandType = "musclePosition"
	HandKongBloom              HandType = "kongBloom"
	HandSevenPairsOrSingleHook HandType = "sevenPairsOrSingleHook"
	HandDragonSevenPairs       HandType = "dragonSevenPairs"
	HandPureSuit               HandType = "pureSuit"
)

func (h HandType) Points() int {
	switch h {
	case HandPingHu:
		return 2
	case HandSingleWait:
		return 3
	case HandBigPairs, HandMusclePosition:
		return 4
	case HandKongBloom, HandSevenPairsOrSingleHook, HandPureSuit:
		return 5
	case HandDragonSevenPairs:
		return 7
	default:
		return 2
	}
}

func (h HandType) Priority() int {
	switch h {
	case HandPureSuit:
		return 0
	case HandDragonSevenPairs:
		return 1
	case HandSevenPairsOrSingleHook:
		return 2
	case HandKongBloom:
		return 3
	case HandMusclePosition, HandBigPairs:
		return 4
	case HandSingleWait:
		return 5
	default:
		return 6
	}
}

type HandContext struct {
	Candidates []HandType `json:"candidates"`
}

type PlayerState struct {
	ID                      string `json:"id"`
	IsListening             bool   `json:"isListening"`
	ChickenPoints           int    `json:"chickenPoints"`
	KongPointsBeforePenalty int    `json:"kongPointsBeforePenalty"`
}

type WinEventType string

const (
	WinSelfDraw    WinEventType = "selfDraw"
	WinMultiWin    WinEventType = "multiWin"
	WinRobbingKong WinEventType = "robbingKong"
)

type WinEvent struct {
	Type         WinEventType `json:"type"`
	Winner       string       `json:"winner,omitempty"`
	Discarder    string       `json:"discarder,omitempty"`
	RobbedPlayer string       `json:"robbedPlayer,omitempty"`
	Winners      []string     `json:"winners,omitempty"`
}

type RoundInput struct {
	Players                  []PlayerState          `json:"players"`
	HandsByWinner            map[string]HandContext `json:"handsByWinner"`
	WinEvent                 *WinEvent              `json:"winEvent,omitempty"`
	KongDiscardPenaltyPlayer string                 `json:"kongDiscardPenaltyPlayer,omitempty"`
	IsDraw                   bool                   `json:"isDraw"`
}

type RoundSettlement struct {
	PointTransfers    map[string]int      `json:"pointTransfers"`
	SelectedHandTypes map[string]HandType `json:"selectedHandTypes"`
	Notes             []string            `json:"notes"`
}
