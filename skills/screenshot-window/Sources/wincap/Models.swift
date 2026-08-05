import Foundation

struct Frame: Codable, Sendable {
    let x: Double
    let y: Double
    let width: Double
    let height: Double
}

struct WindowInfo: Codable, Sendable {
    let windowID: Int
    let title: String?
    let appName: String
    let bundleIdentifier: String?
    let pid: Int32
    let frame: Frame
    let layer: Int
    let isOnScreen: Bool
    let isActive: Bool
}

struct Candidate: Codable {
    let windowID: Int
    let title: String?
}

struct ErrorOutput: Codable {
    let error: String
    let message: String
    let candidates: [Candidate]?

    init(error: String, message: String, candidates: [Candidate]? = nil) {
        self.error = error
        self.message = message
        self.candidates = candidates
    }
}

struct CaptureResult: Codable {
    let path: String
    let width: Int
    let height: Int
}

enum WincapError: Error {
    case permission(underlying: Error)
    case noMatch(app: String, title: String?)
    case ambiguousMatch(candidates: [WindowInfo])
    case capture(message: String)
    case usage(message: String)
    case timeout(seconds: Double)
}

let jsonEncoder: JSONEncoder = {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
    return encoder
}()

func printJSON<T: Encodable>(_ value: T) {
    guard let data = try? jsonEncoder.encode(value), let text = String(data: data, encoding: .utf8) else {
        print("{\"error\": \"encoding_failed\", \"message\": \"Failed to encode JSON output\"}")
        return
    }
    print(text)
}

func printError(_ error: WincapError) {
    let output: ErrorOutput
    switch error {
    case .permission(let underlying):
        output = ErrorOutput(
            error: "screen_recording_permission_required",
            message: "Screen Recording permission is required. Grant it to your terminal app in System Settings \u{2192} Privacy & Security \u{2192} Screen Recording, then restart the terminal. (\(underlying.localizedDescription))"
        )
    case .noMatch(let app, let title):
        let titlePart = title.map { " with title containing \"\($0)\"" } ?? ""
        output = ErrorOutput(
            error: "no_matching_window",
            message: "No on-screen window found for app \"\(app)\"\(titlePart)."
        )
    case .ambiguousMatch(let candidates):
        output = ErrorOutput(
            error: "ambiguous_match",
            message: "Multiple windows matched; retry with --window-id to disambiguate.",
            candidates: candidates.map { Candidate(windowID: $0.windowID, title: $0.title) }
        )
    case .capture(let message):
        output = ErrorOutput(error: "capture_failed", message: message)
    case .usage(let message):
        output = ErrorOutput(error: "usage_error", message: message)
    case .timeout(let seconds):
        output = ErrorOutput(
            error: "timeout",
            message: "Timed out after \(Int(seconds))s waiting for the system window-sharing service (SCShareableContent). This is a known ScreenCaptureKit flake (Apple's own async bridging occasionally leaks its continuation instead of hanging up cleanly) rather than a wincap bug -- just retry."
        )
    }
    printJSON(output)
}
