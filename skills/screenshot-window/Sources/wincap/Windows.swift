import Foundation
import ScreenCaptureKit

// SCShareableContent's async bridging occasionally leaks its own continuation
// instead of resuming it (a known ScreenCaptureKit flake, not something in our
// control -- surfaces as "SWIFT TASK CONTINUATION MISUSE" on stderr and an
// indefinite hang). A withThrowingTaskGroup-based timeout does NOT help here:
// structured concurrency waits for every child task to finish before the group
// itself returns, even after cancelAll(), and a task stuck on a continuation
// that will never resume never observes cancellation either -- so the "timeout"
// task group would hang right alongside the leak. Instead, race two detached,
// never-awaited Tasks against a single continuation we control directly: as
// soon as either resumes it, withTimeout returns -- the other, if genuinely
// stuck, is simply abandoned rather than joined.
let wincapDebugEnabled = ProcessInfo.processInfo.environment["WINCAP_DEBUG"] != nil

func debugLog(_ message: @autoclosure () -> String) {
    guard wincapDebugEnabled else { return }
    FileHandle.standardError.write("[wincap debug +\(String(format: "%.3f", ProcessInfo.processInfo.systemUptime))s] \(message())\n".data(using: .utf8)!)
}

private final class ResumeOnce<T: Sendable>: @unchecked Sendable {
    private let lock = NSLock()
    private var didResume = false
    private let continuation: CheckedContinuation<T, Error>

    init(_ continuation: CheckedContinuation<T, Error>) {
        self.continuation = continuation
    }

    func resume(returning value: T) {
        lock.lock(); defer { lock.unlock() }
        guard !didResume else {
            debugLog("ResumeOnce: ignored LATE returning-resume (already resumed)")
            return
        }
        didResume = true
        debugLog("ResumeOnce: resuming with SUCCESS")
        continuation.resume(returning: value)
    }

    func resume(throwing error: Error) {
        lock.lock(); defer { lock.unlock() }
        guard !didResume else {
            debugLog("ResumeOnce: ignored LATE throwing-resume (already resumed): \(error)")
            return
        }
        didResume = true
        debugLog("ResumeOnce: resuming with ERROR: \(error)")
        continuation.resume(throwing: error)
    }
}

func withTimeout<T: Sendable>(seconds: Double, operation: @escaping @Sendable () async throws -> T) async throws -> T {
    debugLog("withTimeout: entered, seconds=\(seconds)")
    return try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<T, Error>) in
        let once = ResumeOnce(continuation)
        debugLog("withTimeout: spawning operation task")
        Task {
            do {
                debugLog("operation task: calling operation()")
                let value = try await operation()
                debugLog("operation task: operation() returned successfully")
                once.resume(returning: value)
            } catch {
                debugLog("operation task: operation() threw: \(error)")
                once.resume(throwing: error)
            }
        }
        debugLog("withTimeout: spawning timer task")
        Task {
            try? await Task.sleep(nanoseconds: UInt64(seconds * 1_000_000_000))
            debugLog("timer task: sleep elapsed, resuming with timeout")
            once.resume(throwing: WincapError.timeout(seconds: seconds))
        }
    }
}

func fetchWindows(onScreenOnly: Bool) async throws -> [WindowInfo] {
    do {
        return try await withTimeout(seconds: 10) {
            let content = try await SCShareableContent.excludingDesktopWindows(true, onScreenWindowsOnly: onScreenOnly)
            return content.windows.map { window in
                let app = window.owningApplication
                return WindowInfo(
                    windowID: Int(window.windowID),
                    title: (window.title?.isEmpty ?? true) ? nil : window.title,
                    appName: app?.applicationName ?? "",
                    bundleIdentifier: app?.bundleIdentifier,
                    pid: app?.processID ?? 0,
                    frame: Frame(
                        x: window.frame.origin.x,
                        y: window.frame.origin.y,
                        width: window.frame.size.width,
                        height: window.frame.size.height
                    ),
                    layer: window.windowLayer,
                    isOnScreen: window.isOnScreen,
                    isActive: window.isActive
                )
            }
        }
    } catch let error as WincapError {
        throw error
    } catch {
        throw WincapError.permission(underlying: error)
    }
}

func matchWindows(_ windows: [WindowInfo], app: String, title: String?) -> [WindowInfo] {
    let needleApp = app.lowercased()
    let needleTitle = title?.lowercased()
    return windows.filter { window in
        guard window.appName.lowercased().contains(needleApp) else { return false }
        if let needleTitle, !needleTitle.isEmpty {
            return (window.title ?? "").lowercased().contains(needleTitle)
        }
        return true
    }
}

func resolveWindow(app: String, title: String?, windows: [WindowInfo]) throws -> WindowInfo {
    let matches = matchWindows(windows, app: app, title: title)
    if matches.isEmpty {
        throw WincapError.noMatch(app: app, title: title)
    }
    if matches.count > 1 {
        throw WincapError.ambiguousMatch(candidates: matches)
    }
    return matches[0]
}
