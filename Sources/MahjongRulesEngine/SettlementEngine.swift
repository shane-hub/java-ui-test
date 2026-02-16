import Foundation

public enum RuleEngineError: Error, LocalizedError {
    case missingWinnerHand(String)
    case invalidPlayers

    public var errorDescription: String? {
        switch self {
        case .missingWinnerHand(let player):
            return "Missing hand context for winner: \(player)"
        case .invalidPlayers:
            return "Round requires exactly 4 players."
        }
    }
}

public final class SettlementEngine {
    public init() {}

    public func settle(round: RoundInput) throws -> RoundSettlement {
        guard round.players.count == 4 else {
            throw RuleEngineError.invalidPlayers
        }

        var transfers = Dictionary(uniqueKeysWithValues: round.players.map { ($0.id, 0) })
        var notes: [String] = []
        var selectedTypes: [String: HandType] = [:]

        if round.isDraw {
            notes.append("流局：所有杠分归零，不翻鸡。")
            return RoundSettlement(pointTransfers: transfers, selectedHandTypes: selectedTypes, notes: notes)
        }

        if let event = round.winEvent {
            try applyWinSettlement(event, round: round, transfers: &transfers, selectedTypes: &selectedTypes, notes: &notes)
            applyChickenSettlement(round: round, transfers: &transfers, notes: &notes)
        }

        if let penaltyPlayer = round.kongDiscardPenaltyPlayer {
            if let preKong = round.players.first(where: { $0.id == penaltyPlayer })?.kongPointsBeforePenalty, preKong > 0 {
                transfers[penaltyPlayer, default: 0] -= preKong
                notes.append("杠后点炮惩罚：\(penaltyPlayer) 的杠分 \(preKong) 已作废。")
            }
        }

        return RoundSettlement(pointTransfers: transfers, selectedHandTypes: selectedTypes, notes: notes)
    }

    public func selectHandType(from context: HandContext) -> HandType {
        let candidates = context.candidates
        if candidates.contains(.pureSuit) {
            return .pureSuit
        }

        return candidates.min(by: { lhs, rhs in
            if lhs.points == rhs.points {
                return lhs.priority < rhs.priority
            }
            return lhs.points < rhs.points
        }) ?? .pingHu
    }

    private func applyWinSettlement(
        _ event: WinEvent,
        round: RoundInput,
        transfers: inout [String: Int],
        selectedTypes: inout [String: HandType],
        notes: inout [String]
    ) throws {
        switch event {
        case .selfDraw(let winner):
            let handType = try resolvedHandType(for: winner, round: round)
            selectedTypes[winner] = handType
            let points = handType.points

            for player in round.players where player.id != winner {
                transfers[player.id, default: 0] -= points
                transfers[winner, default: 0] += points
            }
            notes.append("自摸：\(winner) 按 \(handType.rawValue) 收取三家。")

        case .multiWin(let discarder, let winners):
            for winner in winners {
                let handType = try resolvedHandType(for: winner, round: round)
                selectedTypes[winner] = handType
                let points = handType.points
                transfers[discarder, default: 0] -= points
                transfers[winner, default: 0] += points
            }
            notes.append("一炮多响：\(discarder) 分别向赢家支付。")

        case .robbingKong(let robbedPlayer, let winners):
            for winner in winners {
                let handType = try resolvedHandType(for: winner, round: round)
                selectedTypes[winner] = handType
                let points = handType.points * 3
                transfers[robbedPlayer, default: 0] -= points
                transfers[winner, default: 0] += points
            }
            notes.append("抢杠胡多响：\(robbedPlayer) 单人全包（3 倍）。")
        }
    }

    private func applyChickenSettlement(
        round: RoundInput,
        transfers: inout [String: Int],
        notes: inout [String]
    ) {
        let listeners = round.players.filter { $0.isListening }
        let nonListeners = round.players.filter { !$0.isListening }

        // 对冲：听牌者之间做差值结算
        for i in 0..<listeners.count {
            for j in (i + 1)..<listeners.count {
                let a = listeners[i]
                let b = listeners[j]
                let delta = a.chickenPoints - b.chickenPoints
                if delta > 0 {
                    transfers[a.id, default: 0] += delta
                    transfers[b.id, default: 0] -= delta
                } else if delta < 0 {
                    transfers[b.id, default: 0] += -delta
                    transfers[a.id, default: 0] -= -delta
                }
            }
        }

        // 未听牌向所有有鸡分听牌者全额赔付
        let listenersWithChicken = listeners.filter { $0.chickenPoints > 0 }
        for nonListener in nonListeners {
            for listener in listenersWithChicken {
                transfers[nonListener.id, default: 0] -= listener.chickenPoints
                transfers[listener.id, default: 0] += listener.chickenPoints
            }
        }

        if !listeners.isEmpty {
            notes.append("鸡分结算已执行：听牌对冲 + 未听牌全赔。")
        }
    }

    private func resolvedHandType(for winner: String, round: RoundInput) throws -> HandType {
        guard let context = round.handsByWinner[winner] else {
            throw RuleEngineError.missingWinnerHand(winner)
        }
        return selectHandType(from: context)
    }
}
