import Foundation

@MainActor
final class APIClient: ObservableObject {
    @Published var currentUser: User?

    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init(baseURL: URL = URL(string: "http://127.0.0.1:8080")!) {
        self.baseURL = baseURL

        let config = URLSessionConfiguration.default
        config.httpCookieStorage = .shared
        config.httpShouldSetCookies = true
        config.requestCachePolicy = .reloadIgnoringLocalAndRemoteCacheData

        self.session = URLSession(configuration: config)

        self.decoder = JSONDecoder()
        self.decoder.dateDecodingStrategy = .iso8601
        self.encoder = JSONEncoder()
    }

    func login(username: String, password: String) async throws {
        let reqBody = ["username": username, "password": password]
        var request = URLRequest(url: baseURL.appendingPathComponent("api/login"))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: reqBody)

        let (data, response) = try await session.data(for: request)
        try validate(response: response, data: data)
        let login = try decoder.decode(LoginResponse.self, from: data)
        currentUser = login.user
    }

    func logout() async throws {
        var request = URLRequest(url: baseURL.appendingPathComponent("api/logout"))
        request.httpMethod = "POST"
        let (data, response) = try await session.data(for: request)
        try validate(response: response, data: data, allowNoContent: true)
        currentUser = nil
    }

    func loadHall() async throws -> HallResponse {
        try await get(path: "api/hall")
    }

    func loadHistory(limit: Int = 20) async throws -> HistoryResponse {
        try await get(path: "api/history?limit=\(limit)")
    }

    func loadHeadToHead(opponent: String) async throws -> HeadToHeadStats {
        try await get(path: "api/stats/headtohead?opponent=\(opponent.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? "")")
    }

    func settle(_ req: SettleRequest) async throws -> SettlementResult {
        var request = URLRequest(url: baseURL.appendingPathComponent("api/settle"))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try encoder.encode(req)

        let (data, response) = try await session.data(for: request)
        try validate(response: response, data: data)
        return try decoder.decode(SettlementResult.self, from: data)
    }

    private func get<T: Decodable>(path: String) async throws -> T {
        let url = URL(string: path, relativeTo: baseURL)!
        let (data, response) = try await session.data(from: url)
        try validate(response: response, data: data)
        return try decoder.decode(T.self, from: data)
    }

    private func validate(response: URLResponse, data: Data, allowNoContent: Bool = false) throws {
        guard let http = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }
        if allowNoContent, http.statusCode == 204 { return }
        guard (200...299).contains(http.statusCode) else {
            let msg = String(data: data, encoding: .utf8) ?? "Request failed"
            throw NSError(domain: "APIError", code: http.statusCode, userInfo: [NSLocalizedDescriptionKey: msg])
        }
    }
}
