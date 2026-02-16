import SwiftUI

struct ContentView: View {
    @EnvironmentObject var api: APIClient

    var body: some View {
        Group {
            if api.currentUser == nil {
                LoginView()
            } else {
                HallView()
            }
        }
    }
}
