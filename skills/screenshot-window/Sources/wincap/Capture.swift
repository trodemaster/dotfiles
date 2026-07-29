import Foundation
import ImageIO

// ScreenCaptureKit's SCScreenshotManager requires a Screen Recording TCC grant
// tied to the calling binary's own code identity, and a bare unbundled SwiftPM
// executable never gets a chance to register for one -- CGRequestScreenCaptureAccess()
// just hangs waiting for a decision that TCC never surfaces a prompt for. Apple's
// own `screencapture` binary, however, already carries the Screen Recording grant
// through the parent terminal app (the same mechanism the previous JXA-based
// version of this skill relied on), so pixel capture is delegated to it while
// `list` keeps using ScreenCaptureKit for metadata.
func captureWindowToFile(windowID: Int, path: String, format: String) throws -> (width: Int, height: Int) {
    let normalizedFormat: String
    switch format.lowercased() {
    case "png": normalizedFormat = "png"
    case "jpg", "jpeg": normalizedFormat = "jpg"
    case "tiff", "tif": normalizedFormat = "tiff"
    default:
        throw WincapError.usage(message: "Unsupported format \"\(format)\". Use png, jpg, or tiff.")
    }

    let task = Process()
    task.executableURL = URL(fileURLWithPath: "/usr/sbin/screencapture")
    task.arguments = ["-x", "-t", normalizedFormat, "-l\(windowID)", path]
    let errPipe = Pipe()
    task.standardError = errPipe

    do {
        try task.run()
        task.waitUntilExit()
    } catch {
        throw WincapError.capture(message: "Failed to launch screencapture: \(error.localizedDescription)")
    }

    guard task.terminationStatus == 0, FileManager.default.fileExists(atPath: path) else {
        let errData = errPipe.fileHandleForReading.readDataToEndOfFile()
        let errText = String(data: errData, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        throw WincapError.capture(message: "screencapture failed (exit \(task.terminationStatus)) for window \(windowID). It may have closed; run `list` again. \(errText.isEmpty ? "Screen Recording permission may be required for the calling terminal app." : errText)")
    }

    guard let source = CGImageSourceCreateWithURL(URL(fileURLWithPath: path) as CFURL, nil),
          let properties = CGImageSourceCopyPropertiesAtIndex(source, 0, nil) as? [CFString: Any],
          let width = properties[kCGImagePropertyPixelWidth] as? Int,
          let height = properties[kCGImagePropertyPixelHeight] as? Int else {
        throw WincapError.capture(message: "Captured file at \(path) but could not read its dimensions.")
    }
    return (width, height)
}

func defaultScreenshotDirectory() -> String {
    let home = FileManager.default.homeDirectoryForCurrentUser.path
    let task = Process()
    task.executableURL = URL(fileURLWithPath: "/usr/bin/defaults")
    task.arguments = ["read", "com.apple.screencapture", "location"]
    let outPipe = Pipe()
    task.standardOutput = outPipe
    task.standardError = Pipe()

    do {
        try task.run()
        task.waitUntilExit()
        let data = outPipe.fileHandleForReading.readDataToEndOfFile()
        if let raw = String(data: data, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines),
           !raw.isEmpty {
            var path = raw
            if path.hasPrefix("~") {
                path = home + path.dropFirst(1)
            }
            var isDir: ObjCBool = false
            if FileManager.default.fileExists(atPath: path, isDirectory: &isDir), isDir.boolValue {
                return path
            }
        }
    } catch {
        // fall through to default below
    }
    return home + "/Desktop"
}

func timestampedFilename(label: String, ext: String) -> String {
    let formatter = DateFormatter()
    formatter.dateFormat = "yyyy-MM-dd 'at' hh.mm.ss a"
    let ts = formatter.string(from: Date())
    let safeLabel = String(label.map { c -> Character in
        (c.isLetter || c.isNumber || c == "_" || c == "-" || c == "." || c == " ") ? c : "_"
    })
    return "Screenshot \(ts) - \(safeLabel).\(ext)"
}

func resolveOutputPath(explicit: String?, label: String, format: String) -> String {
    if let explicit {
        return (explicit as NSString).expandingTildeInPath
    }
    let dir = defaultScreenshotDirectory()
    let filename = timestampedFilename(label: label, ext: format.lowercased() == "jpeg" ? "jpg" : format.lowercased())
    return dir + "/" + filename
}
