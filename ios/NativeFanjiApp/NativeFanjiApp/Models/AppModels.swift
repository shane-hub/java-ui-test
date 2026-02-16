import Foundation

struct User: Codable, Identifiable {
    let id: String
    let username: String
    let online: Bool
}

struct RoomSummary: Codable, Identifiable {
    let id: String
    let name: String
    let currentPeople: Int
    let capacity: Int
    let mode: String
}

struct MatchRecord: Codable, Identifiable {
    let id: String
    let roomId: String
    let participants: [String]
    let winners: [String]
    let settlement: [String: Int]
    let playedAt: Date
}

struct HeadToHeadStats: Codable {
    let me: String
    let opponent: String
    let totalRounds: Int
    let meWins: Int
    let opponentWins: Int
    let draws: Int
    let meWinRate: Double
}

struct HallResponse: Codable {
    let rooms: [RoomSummary]
    let onlineFriends: [User]
}

struct HistoryResponse: Codable {
    let records: [MatchRecord]
}

struct LoginResponse: Codable {
    let user: User
}

struct HandContext: Codable {
    let candidates: [String]
}

struct PlayerState: Codable {
    let id: String
    let isListening: Bool
    let chickenPoints: Int
    let kongPointsBeforePenalty: Int
}

struct WinEvent: Codable {
    let type: String
    let winner: String?
    let discarder: String?
    let robbedPlayer: String?
    let winners: [String]?
}

struct RoundInput: Codable {
    let players: [PlayerState]
    let handsByWinner: [String: HandContext]
    let winEvent: WinEvent?
    let kongDiscardPenaltyPlayer: String?
    let isDraw: Bool
}

struct SettleRequest: Codable {
    let roomId: String
    let round: RoundInput
}

struct SettlementResult: Codable {
    let pointTransfers: [String: Int]
    let selectedHandTypes: [String: String]
    let notes: [String]
}
