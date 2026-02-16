import Foundation

public enum HandType: String, CaseIterable, Codable {
    case pingHu
    case singleWait
    case bigPairs
    case musclePosition
    case kongBloom
    case sevenPairsOrSingleHook
    case dragonSevenPairs
    case pureSuit

    public var points: Int {
        switch self {
        case .pingHu: return 2
        case .singleWait: return 3
        case .bigPairs, .musclePosition: return 4
        case .kongBloom, .sevenPairsOrSingleHook, .pureSuit: return 5
        case .dragonSevenPairs: return 7
        }
    }

    /// Lower value means higher priority in the V8.0 rules.
    public var priority: Int {
        switch self {
        case .pureSuit: return 0
        case .dragonSevenPairs: return 1
        case .sevenPairsOrSingleHook: return 2
        case .kongBloom: return 3
        case .musclePosition, .bigPairs: return 4
        case .singleWait: return 5
        case .pingHu: return 6
        }
    }
}

public struct HandContext: Codable, Equatable {
    public var candidates: Set<HandType>

    public init(candidates: Set<HandType>) {
        self.candidates = candidates
    }
}

public struct PlayerState: Codable, Equatable {
    public let id: String
    public let isListening: Bool
    public let chickenPoints: Int
    public var kongPointsBeforePenalty: Int

    public init(id: String, isListening: Bool, chickenPoints: Int, kongPointsBeforePenalty: Int = 0) {
        self.id = id
        self.isListening = isListening
        self.chickenPoints = chickenPoints
        self.kongPointsBeforePenalty = kongPointsBeforePenalty
    }
}

public enum WinEvent: Codable, Equatable {
    case selfDraw(winner: String)
    case multiWin(discarder: String, winners: [String])
    case robbingKong(robbedPlayer: String, winners: [String])
}

public struct RoundInput: Codable, Equatable {
    public let players: [PlayerState]
    public let handsByWinner: [String: HandContext]
    public let winEvent: WinEvent?
    public let kongDiscardPenaltyPlayer: String?
    public let isDraw: Bool

    public init(
        players: [PlayerState],
        handsByWinner: [String: HandContext],
        winEvent: WinEvent?,
        kongDiscardPenaltyPlayer: String? = nil,
        isDraw: Bool = false
    ) {
        self.players = players
        self.handsByWinner = handsByWinner
        self.winEvent = winEvent
        self.kongDiscardPenaltyPlayer = kongDiscardPenaltyPlayer
        self.isDraw = isDraw
    }
}

public struct RoundSettlement: Codable, Equatable {
    public let pointTransfers: [String: Int]
    public let selectedHandTypes: [String: HandType]
    public let notes: [String]

    public init(pointTransfers: [String: Int], selectedHandTypes: [String: HandType], notes: [String]) {
        self.pointTransfers = pointTransfers
        self.selectedHandTypes = selectedHandTypes
        self.notes = notes
    }
}
