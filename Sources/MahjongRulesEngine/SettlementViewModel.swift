import Foundation

#if canImport(SwiftUI)
import SwiftUI

@MainActor
public final class SettlementViewModel: ObservableObject {
    @Published public var inputJSON: String
    @Published public private(set) var outputJSON: String = ""
    @Published public private(set) var errorMessage: String?

    private let engine = SettlementEngine()

    public init(inputJSON: String = "") {
        self.inputJSON = inputJSON
    }

    public func runSettlement() {
        do {
            let data = Data(inputJSON.utf8)
            let input = try JSONDecoder().decode(RoundInput.self, from: data)
            let result = try engine.settle(round: input)
            let output = try JSONEncoder.pretty.encode(result)
            outputJSON = String(decoding: output, as: UTF8.self)
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
            outputJSON = ""
        }
    }
}

public struct SettlementSandboxView: View {
    @StateObject private var viewModel = SettlementViewModel(inputJSON: Self.seedJSON)

    public init() {}

    public var body: some View {
        NavigationStack {
            VStack(spacing: 12) {
                Text("V8.0 规则结算沙盒")
                    .font(.title3.weight(.semibold))
                TextEditor(text: $viewModel.inputJSON)
                    .border(.gray)
                    .frame(minHeight: 220)
                Button("执行结算") {
                    viewModel.runSettlement()
                }
                .buttonStyle(.borderedProminent)

                if let errorMessage = viewModel.errorMessage {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }

                TextEditor(text: .constant(viewModel.outputJSON))
                    .border(.gray)
                    .frame(minHeight: 180)
            }
            .padding()
            .navigationTitle("翻鸡缺一门")
        }
    }

    private static let seedJSON = """
    {
      "players" : [
        {
          "id" : "A",
          "isListening" : true,
          "chickenPoints" : 5,
          "kongPointsBeforePenalty" : 0
        },
        {
          "id" : "B",
          "isListening" : true,
          "chickenPoints" : 2,
          "kongPointsBeforePenalty" : 0
        },
        {
          "id" : "C",
          "isListening" : true,
          "chickenPoints" : 0,
          "kongPointsBeforePenalty" : 0
        },
        {
          "id" : "D",
          "isListening" : false,
          "chickenPoints" : 0,
          "kongPointsBeforePenalty" : 0
        }
      ],
      "handsByWinner" : {
        "A" : {
          "candidates" : [
            "pingHu",
            "singleWait"
          ]
        }
      },
      "winEvent" : {
        "selfDraw" : {
          "winner" : "A"
        }
      },
      "isDraw" : false
    }
    """
}

private extension JSONEncoder {
    static var pretty: JSONEncoder {
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        return encoder
    }
}
#endif
