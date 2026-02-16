import SwiftUI

@main
struct NativeFanjiApp: App {
    @StateObject private var api = APIClient()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(api)
        }
    }
}
