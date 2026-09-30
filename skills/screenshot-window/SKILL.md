---
name: screenshot-window
description: "Capture and inspect a specific macOS app window. Use when asked to screenshot, show, capture, view, or analyze an app window, including requests like 'screenshot X', 'show me the X window', or 'what does the Z window look like'."
compatibility: Requires macOS 26 or later on Apple Silicon and Screen Recording access.
allowed-tools: Bash(make -C ~/.claude/skills/screenshot-window *), Bash(grep -F * ~/Developer/machine-cfg/claude/settings.json), Bash(realpath ~/.claude/skills/screenshot-window/bin/wincap), Bash(~/.claude/skills/screenshot-window/bin/wincap *), Read
---

# Screenshot Window Skill

Captures a specific on-screen window by name using `wincap`, a native Swift/ScreenCaptureKit CLI
built for this skill. **macOS 26+, Apple Silicon (arm64) only.**

**Binary:** `~/.claude/skills/screenshot-window/bin/wincap` (source in `Sources/wincap/`,
`Package.swift`). The binary is **not committed to git** — it's built on demand and gitignored.

Everything is JSON on stdout by default — parse it directly, don't ask the user to read raw output.

---

## Build fallback

Invoke `wincap` directly without checking whether the binary exists first. Only if the shell reports
that the binary is missing or not executable, build it and retry the original command once:

```bash
make -C ~/.claude/skills/screenshot-window build
```

This runs `swift build -c release` and copies the result into `bin/wincap`. Takes a few seconds.
If it fails with sandbox-style "Operation not permitted" errors writing to Swift's module cache,
retry with the sandbox disabled.

**Rebuilding invalidates wincap's own Screen Recording grant unless signed with a stable identity.**
`swift build` ad-hoc-signs the binary (required just to run on Apple Silicon), and an ad-hoc
signature embeds a hash of the binary's own bytes as part of its code identity — so every rebuild
looks like a brand-new app to TCC, and any grant tied to the old build goes stale. To avoid
re-granting after every `make build`, export `PERSONAL_CODESIGN_IDENTITY` to a stable signing identity
before building — a Developer ID Application identity you own, or a local self-signed Code Signing
certificate created once via Keychain Access ("Certificate Assistant → Create a Certificate…", type
Code Signing). Never hardcode an actual identity string in this repo; the env var is the only place
it should ever live, and it defaults to ad-hoc (`-`) if unset.

**Do not run `wincap` in a sandbox.** `list` (and `capture`, which calls it internally) talks to
`tccd` over XPC via `SCShareableContent`. Sandboxes that block this XPC call cause an `"error":
"timeout"` response. Use the client-specific mechanism for running the command outside its sandbox.

For Claude Code, `wincap` must be listed in `sandbox.excludedCommands` in `settings.json`:

```bash
grep -F '~/.claude/skills/screenshot-window/bin/wincap' ~/Developer/machine-cfg/claude/settings.json
```

It must match exactly `"~/.claude/skills/screenshot-window/bin/wincap *"` — a different literal
invocation string (a relative path like `./bin/wincap`, `cd`-then-relative, or the
`~/Developer/dotfiles/...` working-copy path instead of `~/.claude/skills/...`) will **not** match
this pattern and will run inside the sandbox. If the exclusion is absent, use Claude Code's
per-command sandbox bypass rather than guessing.

Two separate **Screen Recording** grants are required (System Settings → Privacy & Security →
Screen Recording):
- **wincap itself** — `SCShareableContent` needs the grant on the calling binary, not just the
  parent app. Resolve the symlink first (`realpath ~/.claude/skills/screenshot-window/bin/wincap`,
  since TCC keys off the real file) and add that path with **+**.
- **The terminal app** (Ghostty, Terminal, etc.) — pixel capture shells out to Apple's
  `screencapture`, which inherits the grant from the parent terminal.

If `wincap capture` returns `"error": "no_matching_window"` even though the window is visibly open,
the app name likely doesn't match — app names are matched case-insensitively as a substring against
the name shown in the menu bar / Activity Monitor.

**Always run `wincap` directly in the foreground. Never use shell backgrounding or an external
watchdog.**
`wincap` already has its own internal 10s timeout on the `SCShareableContent` call — it cannot hang
your shell. Backgrounding the process can prevent `SCShareableContent`'s async completion from being
delivered, guaranteeing that the internal timeout fires. If a foreground invocation returns
`"error": "timeout"`, retry it once unchanged.

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

Use the available image/file reading tool on the returned `path` to view the window contents.

---

## Agent workflow

```
1. Ask user which app (and optionally which window title/label), if not already clear
2. Try `capture --app <name> [--title <substr>]` directly without a preliminary existence check
3. If the shell says wincap is missing or not executable, build it and retry the command once
4. If ambiguous_match comes back, show the candidate titles and either ask the user
   or pick the best match by title, then retry with --window-id
5. If no_matching_window comes back, run `list --app <name>` to sanity-check the app name/spelling
6. Read the returned image path
7. Describe / analyze the window contents
```

If any command returns `"error": "capture_failed"` mentioning permissions, stop and tell the user
to grant Screen Recording access as described in Prerequisites.
