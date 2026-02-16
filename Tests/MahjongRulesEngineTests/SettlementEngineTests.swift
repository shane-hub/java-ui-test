import XCTest
@testable import MahjongRulesEngine

final class SettlementEngineTests: XCTestCase {
    private let engine = SettlementEngine()

    func testSelectHandTypeUsesLowestPointsExceptPureSuit() {
        let context = HandContext(candidates: [.dragonSevenPairs, .singleWait, .pingHu])
        let selected = engine.selectHandType(from: context)
        XCTAssertEqual(selected, .pingHu)

        let contextWithPure = HandContext(candidates: [.dragonSevenPairs, .singleWait, .pureSuit])
        XCTAssertEqual(engine.selectHandType(from: contextWithPure), .pureSuit)
    }

    func testRobbingKongMultiWinUsesTripleMultiplier() throws {
        let round = RoundInput(
            players: [
                .init(id: "A", isListening: true, chickenPoints: 0),
                .init(id: "B", isListening: true, chickenPoints: 0),
                .init(id: "C", isListening: true, chickenPoints: 0),
                .init(id: "D", isListening: true, chickenPoints: 0)
            ],
            handsByWinner: [
                "A": .init(candidates: [.bigPairs]),
                "B": .init(candidates: [.singleWait])
            ],
            winEvent: .robbingKong(robbedPlayer: "D", winners: ["A", "B"])
        )

        let result = try engine.settle(round: round)
        XCTAssertEqual(result.pointTransfers["A"], 12)
        XCTAssertEqual(result.pointTransfers["B"], 9)
        XCTAssertEqual(result.pointTransfers["D"], -21)
    }

    func testChickenOffsetAndNonListeningFullPayment() throws {
        let round = RoundInput(
            players: [
                .init(id: "A", isListening: true, chickenPoints: 5),
                .init(id: "B", isListening: true, chickenPoints: 2),
                .init(id: "C", isListening: true, chickenPoints: 0),
                .init(id: "D", isListening: false, chickenPoints: 0)
            ],
            handsByWinner: [
                "A": .init(candidates: [.pingHu])
            ],
            winEvent: .selfDraw(winner: "A")
        )

        let result = try engine.settle(round: round)
        // 胡牌基础：A 从三家各收2 => +6
        // 鸡分：A-B +3, A-C +5, B-C +2, D->A +5, D->B +2
        // 合计：A +13, B +1, C -7, D -7（叠加胡牌后）
        XCTAssertEqual(result.pointTransfers["A"], 19)
        XCTAssertEqual(result.pointTransfers["B"], -1)
        XCTAssertEqual(result.pointTransfers["C"], -9)
        XCTAssertEqual(result.pointTransfers["D"], -9)
    }

    func testKongDiscardPenaltyClearsKongPoints() throws {
        let round = RoundInput(
            players: [
                .init(id: "A", isListening: true, chickenPoints: 0, kongPointsBeforePenalty: 6),
                .init(id: "B", isListening: true, chickenPoints: 0),
                .init(id: "C", isListening: true, chickenPoints: 0),
                .init(id: "D", isListening: true, chickenPoints: 0)
            ],
            handsByWinner: ["B": .init(candidates: [.pingHu])],
            winEvent: .multiWin(discarder: "A", winners: ["B"]),
            kongDiscardPenaltyPlayer: "A"
        )

        let result = try engine.settle(round: round)
        XCTAssertEqual(result.pointTransfers["A"], -8)
        XCTAssertEqual(result.pointTransfers["B"], 2)
    }

    func testDrawReturnsZeroTransfers() throws {
        let round = RoundInput(
            players: [
                .init(id: "A", isListening: true, chickenPoints: 3),
                .init(id: "B", isListening: true, chickenPoints: 1),
                .init(id: "C", isListening: false, chickenPoints: 0),
                .init(id: "D", isListening: false, chickenPoints: 0)
            ],
            handsByWinner: [:],
            winEvent: nil,
            isDraw: true
        )
        let result = try engine.settle(round: round)
        XCTAssertEqual(result.pointTransfers.values.reduce(0, +), 0)
        XCTAssertTrue(result.pointTransfers.values.allSatisfy { $0 == 0 })
    }
}
