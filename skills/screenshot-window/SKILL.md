---
name: screenshot-window
description: Capture a screenshot of a specific named app window by listing its windows, targeting one by ID, and reading the result. Use when asked to "screenshot X", "show me the X window", "capture the Y app", or "what does the Z window look like".
allowed-tools: Bash(~/.claude/skills/screenshot-window/bin/wincap *), Bash(make -C ~/.claude/skills/screenshot-window *), Read
---

# Screenshot Window Skill

Captures a specific on-screen window by name using `wincap`, a native Swift/ScreenCaptureKit CLI
built for this skill. **macOS 26+, Apple Silicon (arm64) only.**

**Binary:** `~/.claude/skills/screenshot-window/bin/wincap` (source in `Sources/wincap/`,
`Package.swift`). The binary is **not committed to git** — it's built on demand and gitignored.

Everything is JSON on stdout by default — parse it directly, don't ask the user to read raw output.

---

## Prerequisites

**Binary present?** Before the first command below, check:

```bash
test -x ~/.claude/skills/screenshot-window/bin/wincap
```

If missing (fresh checkout, or after a source update), build it:

```bash
make -C ~/.claude/skills/screenshot-window build
```

This runs `swift build -c release` and copies the result into `bin/wincap`. Takes a few seconds.
If it fails with sandbox-style "Operation not permitted" errors writing to Swift's module cache,
that's a Claude Code sandbox restriction, not a real error — retry with the sandbox disabled.

The calling terminal app also needs **Screen Recording** permission (actual pixel capture shells out
to Apple's `screencapture`, which inherits the grant from the parent terminal — the same requirement
the previous version of this skill had).

Check/fix: **System Settings → Privacy & Security → Screen Recording** → enable your terminal
(Ghostty, Terminal, etc.).

If `wincap capture` returns `"error": "no_matching_window"` even though the window is visibly open,
the app name likely doesn't match — app names are matched case-insensitively as a substring against
the name shown in the menu bar / Activity Monitor.

---

## Step 1 — List windows

```bash
~/.claude/skills/screenshot-window/bin/wincap list --app "App Name"
```

Returns a JSON array, one object per window:

```json
[
  {
    "windowID": 6668,
    "title": "rebuild-screenshot-window-swift",
    "appName": "Ghostty",
    "bundleIdentifier": "com.mitchellh.ghostty",
    "pid": 1234,
    "frame": { "x": 3556, "y": 859, "width": 1120, "height": 949 },
    "layer": 0,
    "isOnScreen": true,
    "isActive": false
  }
]
```

- `windowID` is what you pass to `capture --window-id`.
- Only `layer == 0` (normal document/app windows) are returned by default — menu-bar status items,
  HUDs, and other system chrome are filtered out. Pass `--all-layers` to see everything.
- Only on-screen windows are returned by default (minimized windows / other Spaces are excluded).
  Pass `--include-offscreen` to include them, though `capture` may still fail for a window that
  isn't actually visible.
- `title` is omitted (not `null`) when the window has no title.
- Add `--pretty` for a human-readable table instead of JSON.

**App name must match** what's shown in the menu bar / Activity Monitor (substring, case-insensitive)
— e.g. `"Google Chrome"`, `"Notes"`, `"Obsidian"`, `"Ghostty"`, `"Microsoft Outlook"`.

---

## Step 2 — Screenshot a window

By window ID (unambiguous, preferred once you've run `list`):

```bash
~/.claude/skills/screenshot-window/bin/wincap capture --window-id 6668
```

Or resolve directly by app name (and optionally a title substring) without a separate `list` call:

```bash
~/.claude/skills/screenshot-window/bin/wincap capture --app "Ghostty" --title "rebuild-screenshot"
```

- If the app/title match is ambiguous, `capture` returns `"error": "ambiguous_match"` with a
  `candidates` array (`windowID` + `title`) — retry with `--window-id` using one of those.
- If nothing matches, returns `"error": "no_matching_window"`.
- `--out <path>` overrides the default location (the user's configured screenshot directory, or
  `~/Desktop`, with a timestamped filename — same convention as macOS's own `screencapture`).
- `--format png|jpg|tiff` (default `png`).

On success:

```json
{ "path": "/Users/blake/Desktop/Screenshot 2026-07-29 at 02.15.00 PM - Ghostty.png", "width": 2376, "height": 2034 }
```

---

## Step 3 — Read the image

Use the `Read` tool on the returned `path` to view the window contents.

---

## Agent workflow

```
0. If bin/wincap is missing, run `make -C ~/.claude/skills/screenshot-window build` first
1. Ask user which app (and optionally which window title/label), if not already clear
2. Try `capture --app <name> [--title <substr>]` directly
3. If ambiguous_match comes back, show the candidate titles and either ask the user
   or pick the best match by title, then retry with --window-id
4. If no_matching_window comes back, run `list --app <name>` to sanity-check the app name/spelling
5. Read the returned path with the Read tool
6. Describe / analyze the window contents
```

If any command returns `"error": "capture_failed"` mentioning permissions, stop and tell the user
to grant Screen Recording access as described in Prerequisites.
