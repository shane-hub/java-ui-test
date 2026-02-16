import SwiftUI

struct HallView: View {
    @EnvironmentObject var api: APIClient

    @State private var hall: HallResponse = .init(rooms: [], onlineFriends: [])
    @State private var history: [MatchRecord] = []
    @State private var opponentID = ""
    @State private var stats: HeadToHeadStats?
    @State private var settlementResult: SettlementResult?
    @State private var loading = false
    @State private var errorText: String?

    var body: some View {
        NavigationStack {
            List {
                Section("大厅房间") {
                    ForEach(hall.rooms) { room in
                        VStack(alignment: .leading, spacing: 4) {
                            Text(room.name).font(.headline)
                            Text("模式: \(room.mode) · 人数: \(room.currentPeople)/\(room.capacity)")
                                .font(.footnote)
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                Section("在线好友") {
                    ForEach(hall.onlineFriends) { user in
                        Text("\(user.username) (\(user.id))")
                    }
                }

                Section("历史记录") {
                    ForEach(history) { record in
                        VStack(alignment: .leading, spacing: 4) {
                            Text("\(record.roomId) · \(record.id)").font(.subheadline)
                            Text("赢家: \(record.winners.joined(separator: ","))")
                                .font(.footnote)
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                Section("咱俩胜率") {
                    TextField("输入对手ID（如 u-2）", text: $opponentID)
                    Button("查询") { Task { await queryHeadToHead() } }
                    if let stats {
                        Text("总局数: \(stats.totalRounds)")
                        Text("我方胜局: \(stats.meWins)，对方胜局: \(stats.opponentWins)，平局: \(stats.draws)")
                        Text(String(format: "我方胜率: %.2f%%", stats.meWinRate * 100))
                    }
                }

                Section("结算调试（联调）") {
                    Button("提交一局示例结算") { Task { await settleSampleRound() } }
                    if let settlementResult {
                        Text("转移: \(settlementResult.pointTransfers.description)")
                            .font(.footnote)
                        ForEach(settlementResult.notes, id: \.self) { note in
                            Text(note).font(.footnote).foregroundStyle(.secondary)
                        }
                    }
                }

                if let errorText {
                    Section {
                        Text(errorText).foregroundStyle(.red)
                    }
                }
            }
            .overlay {
                if loading { ProgressView("加载中...") }
            }
            .navigationTitle("原生大厅")
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button("刷新") { Task { await refreshAll() } }
                }
                ToolbarItem(placement: .topBarTrailing) {
                    Button("退出") { Task { await logout() } }
                }
            }
            .task { await refreshAll() }
        }
    }

    private func refreshAll() async {
        loading = true
        defer { loading = false }
        do {
            async let hallData = api.loadHall()
            async let historyData = api.loadHistory(limit: 20)
            hall = try await hallData
            history = try await historyData.records
            errorText = nil
        } catch {
            errorText = error.localizedDescription
        }
    }

    private func queryHeadToHead() async {
        guard !opponentID.isEmpty else { return }
        do {
            stats = try await api.loadHeadToHead(opponent: opponentID)
            errorText = nil
        } catch {
            errorText = error.localizedDescription
        }
    }

    private func settleSampleRound() async {
        guard let me = api.currentUser else { return }
        let req = SettleRequest(
            roomId: "r-1001",
            round: RoundInput(
                players: [
                    .init(id: me.id, isListening: true, chickenPoints: 2, kongPointsBeforePenalty: 0),
                    .init(id: "u-2", isListening: true, chickenPoints: 1, kongPointsBeforePenalty: 0),
                    .init(id: "u-3", isListening: false, chickenPoints: 0, kongPointsBeforePenalty: 0),
                    .init(id: "u-4", isListening: false, chickenPoints: 0, kongPointsBeforePenalty: 0)
                ],
                handsByWinner: [me.id: .init(candidates: ["pingHu"])],
                winEvent: .init(type: "selfDraw", winner: me.id, discarder: nil, robbedPlayer: nil, winners: nil),
                kongDiscardPenaltyPlayer: nil,
                isDraw: false
            )
        )
        do {
            settlementResult = try await api.settle(req)
            await refreshAll()
            errorText = nil
        } catch {
            errorText = error.localizedDescription
        }
    }

    private func logout() async {
        do {
            try await api.logout()
        } catch {
            errorText = error.localizedDescription
        }
    }
}
