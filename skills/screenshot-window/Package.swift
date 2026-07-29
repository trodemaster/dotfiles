// swift-tools-version:6.2
import PackageDescription

let package = Package(
    name: "wincap",
    platforms: [.macOS(.v26)],
    targets: [
        .executableTarget(
            name: "wincap",
            path: "Sources/wincap"
        )
    ]
)
