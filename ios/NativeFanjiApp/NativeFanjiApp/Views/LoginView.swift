import SwiftUI

struct LoginView: View {
    @EnvironmentObject var api: APIClient
    @State private var username = ""
    @State private var password = ""
    @State private var loading = false
    @State private var errorText: String?

    var body: some View {
        NavigationStack {
            Form {
                Section("账号") {
                    TextField("用户名", text: $username)
                        .textInputAutocapitalization(.never)
                    SecureField("密码", text: $password)
                }
                Section {
                    Button(loading ? "登录中..." : "登录 / 注册") {
                        Task { await login() }
                    }
                    .disabled(loading || username.isEmpty || password.isEmpty)
                }
                if let errorText {
                    Text(errorText).foregroundStyle(.red)
                }
            }
            .navigationTitle("翻鸡缺一门")
        }
    }

    private func login() async {
        loading = true
        defer { loading = false }
        do {
            try await api.login(username: username, password: password)
            errorText = nil
        } catch {
            errorText = error.localizedDescription
        }
    }
}
