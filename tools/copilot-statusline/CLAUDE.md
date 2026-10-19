# copilot-statusline

Custom two-line status line renderer for GitHub Copilot CLI. It reads the
`statusLine` command JSON payload from stdin and prints terminal-formatted
session information to stdout.

## Build / Install

Standard `go install` convention with no external dependencies.

```bash
make build
make test
make install
make clean
```

## Configuring Copilot CLI to use it

Run `/statusline` inside Copilot CLI and select a custom command, or add this
to `~/.copilot/settings.json`:

```json
"statusLine": {
  "type": "command",
  "command": "/absolute/path/to/go/bin/copilot-statusline"
}
```

Run `make install`, then confirm the path matches `$GOPATH/bin` with
`go env GOPATH`.

## Output

- Line 1: session name, compact model segment, cwd, remote task indicator,
  and git branch. The model segment has no brackets and groups model name,
  effort, and context capacity with bullets:
  `Claude Opus 5.5 • ⚡high • 1.1M`. Capacity uses `displayed_context_limit`,
  falling back to `context_window_size`, with `k`/`M` units and no "context"
  label. Missing effort or capacity is omitted without leaving extra bullets.
- Line 2: live context-usage thermometer and percentage, cache tokens
  (`🧠 ↓read ↑write`), reasoning tokens (`💭`), then GitHub AI Credits
  (amount followed by `💰` instead of the "AI" label) and premium requests
  (`💵` followed by the request count, not a dollar amount) together at the end.
  Cache counts appear when either is nonzero; reasoning tokens appear when
  positive. Cache totals fall back to current API-call counts when both
  totals are zero.

The context bar retains its original width, colors, and live-usage precedence.
The remote indicator remains conditional on an active remote connection.
Short session IDs, session duration, and changed-line counts are not displayed.

Copilot does not expose Claude Code's USD cost or 5-hour/7-day rate-limit
windows in the status-line payload, so those segments are intentionally
omitted.
