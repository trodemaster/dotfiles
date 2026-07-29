import Foundation

struct ParsedArgs {
    var values: [String: String] = [:]
    var flags: Set<String> = []
}

func parseArgs(_ args: [String], boolFlags: Set<String>) throws -> ParsedArgs {
    var result = ParsedArgs()
    var i = 0
    while i < args.count {
        let arg = args[i]
        guard arg.hasPrefix("--") else {
            throw WincapError.usage(message: "Unexpected argument \"\(arg)\".")
        }
        let name = String(arg.dropFirst(2))
        if boolFlags.contains(name) {
            result.flags.insert(name)
            i += 1
            continue
        }
        guard i + 1 < args.count else {
            throw WincapError.usage(message: "Missing value for --\(name).")
        }
        result.values[name] = args[i + 1]
        i += 2
    }
    return result
}

func printPrettyTable(_ windows: [WindowInfo]) {
    if windows.isEmpty {
        print("No windows found.")
        return
    }
    for w in windows {
        let title = w.title ?? "(untitled)"
        let size = "\(Int(w.frame.width))x\(Int(w.frame.height))"
        let pos = "\(Int(w.frame.x)),\(Int(w.frame.y))"
        print("id=\(w.windowID)  app=\"\(w.appName)\"  title=\"\(title)\"  size=\(size)  pos=\(pos)  layer=\(w.layer)\(w.isOnScreen ? "" : "  [offscreen]")")
    }
}

func runList(_ args: [String]) async throws {
    let parsed = try parseArgs(args, boolFlags: ["include-offscreen", "all-layers", "pretty"])
    let onScreenOnly = !parsed.flags.contains("include-offscreen")
    let windows = try await fetchWindows(onScreenOnly: onScreenOnly)

    var filtered = windows
    if let app = parsed.values["app"] {
        let needle = app.lowercased()
        filtered = filtered.filter { $0.appName.lowercased().contains(needle) }
    }
    // Layer 0 is a normal document/app window; anything else is menu-bar status
    // items, HUDs, or system chrome that clutters results and can't usefully be
    // captured as "the app's window". Opt back in with --all-layers.
    if !parsed.flags.contains("all-layers") {
        filtered = filtered.filter { $0.layer == 0 }
    }

    if parsed.flags.contains("pretty") {
        printPrettyTable(filtered)
    } else {
        printJSON(filtered)
    }
}

func runCapture(_ args: [String]) async throws {
    let parsed = try parseArgs(args, boolFlags: [])
    let format = parsed.values["format"] ?? "png"

    let windowID: Int
    let label: String

    if let windowIDStr = parsed.values["window-id"] {
        guard let id = Int(windowIDStr) else {
            throw WincapError.usage(message: "--window-id must be an integer.")
        }
        windowID = id
        label = parsed.values["app"] ?? "window-\(id)"
    } else if let app = parsed.values["app"] {
        let windows = try await fetchWindows(onScreenOnly: true).filter { $0.layer == 0 }
        let match = try resolveWindow(app: app, title: parsed.values["title"], windows: windows)
        windowID = match.windowID
        label = app
    } else {
        throw WincapError.usage(message: "Provide --window-id or --app.")
    }

    let path = resolveOutputPath(explicit: parsed.values["out"], label: label, format: format)
    let (width, height) = try captureWindowToFile(windowID: windowID, path: path, format: format)
    printJSON(CaptureResult(path: path, width: width, height: height))
}

func printUsage() {
    print("""
    wincap - list and capture macOS app windows for agent use (macOS 26+, Apple Silicon only)

    Usage:
      wincap list [--app <name>] [--include-offscreen] [--all-layers] [--pretty]
      wincap capture --window-id <id> [--out <path>] [--format png|jpg|tiff]
      wincap capture --app <name> [--title <substr>] [--out <path>] [--format png|jpg|tiff]

    All output is JSON on stdout unless --pretty is passed to `list`.
    Requires Screen Recording permission for the calling terminal app.
    """)
}

@main
struct WincapMain {
    static func main() async {
        let arguments = Array(CommandLine.arguments.dropFirst())
        guard let command = arguments.first else {
            printUsage()
            exit(1)
        }
        let rest = Array(arguments.dropFirst())

        do {
            switch command {
            case "list":
                try await runList(rest)
            case "capture":
                try await runCapture(rest)
            case "-h", "--help", "help":
                printUsage()
                exit(0)
            default:
                throw WincapError.usage(message: "Unknown command \"\(command)\". Use list, capture, or help.")
            }
            exit(0)
        } catch let err as WincapError {
            printError(err)
            exit(1)
        } catch {
            printError(.capture(message: error.localizedDescription))
            exit(1)
        }
    }
}
