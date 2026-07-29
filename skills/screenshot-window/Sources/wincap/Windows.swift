import Foundation
import ScreenCaptureKit

func fetchWindows(onScreenOnly: Bool) async throws -> [WindowInfo] {
    let content: SCShareableContent
    do {
        content = try await SCShareableContent.excludingDesktopWindows(true, onScreenWindowsOnly: onScreenOnly)
    } catch {
        throw WincapError.permission(underlying: error)
    }

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
