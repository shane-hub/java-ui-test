// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "MahjongRulesEngine",
    platforms: [
        .iOS(.v17),
        .macOS(.v14)
    ],
    products: [
        .library(name: "MahjongRulesEngine", targets: ["MahjongRulesEngine"]),
        .executable(name: "MahjongDemoApp", targets: ["MahjongDemoApp"])
    ],
    targets: [
        .target(
            name: "MahjongRulesEngine"
        ),
        .executableTarget(
            name: "MahjongDemoApp",
            dependencies: ["MahjongRulesEngine"]
        ),
        .testTarget(
            name: "MahjongRulesEngineTests",
            dependencies: ["MahjongRulesEngine"]
        )
    ]
)
