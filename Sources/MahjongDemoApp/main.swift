import Foundation
import MahjongRulesEngine

let engine = SettlementEngine()

let sample = RoundInput(
    players: [
        PlayerState(id: "A", isListening: true, chickenPoints: 5, kongPointsBeforePenalty: 2),
        PlayerState(id: "B", isListening: true, chickenPoints: 2),
        PlayerState(id: "C", isListening: true, chickenPoints: 0),
        PlayerState(id: "D", isListening: false, chickenPoints: 0)
    ],
    handsByWinner: [
        "A": HandContext(candidates: [.pingHu, .singleWait])
    ],
    winEvent: .selfDraw(winner: "A"),
    kongDiscardPenaltyPlayer: nil,
    isDraw: false
)

do {
    let result = try engine.settle(round: sample)
    print("--- Settlement Demo ---")
    print("Transfers: \(result.pointTransfers)")
    print("Selected hand types: \(result.selectedHandTypes)")
    for note in result.notes {
        print("- \(note)")
    }
} catch {
    fputs("Settlement failed: \(error)\n", stderr)
    exit(1)
}
