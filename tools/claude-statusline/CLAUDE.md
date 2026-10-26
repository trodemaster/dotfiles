# claude-statusline

Custom two-line status line renderer for Claude Code, read from stdin as the
`statusLine` hook JSON payload (model, workspace, cost, context window, rate
limits, etc.) and printed to stdout.

## Build / Install

Standard `go install` convention — no external dependencies.

```bash
make build    # produces ./claude-statusline
make install  # go install . -> $GOPATH/bin/claude-statusline (usually ~/go/bin)
make clean    # remove local build artifact
go test ./... # validate rendering and optional payload fields
```

## Configuring Claude Code to use it

In `~/.claude/settings.json` (or the machine-cfg-managed source for it):

```json
"statusLine": {
  "type": "command",
  "command": "/absolute/path/to/go/bin/claude-statusline"
}
```

Run `make install`, then confirm the path matches your `$GOPATH/bin` (check
with `go env GOPATH`).

## Notes

- Pure stdlib (`encoding/json`, `os/exec` for `git branch --show-current`,
  etc.) — no third-party dependencies, no machine- or org-specific config.
- Line 1: session name, compact model group (display name and version, effort level,
  thinking indicator, and `fast` when enabled), short session identifier, cwd, project dir
  (if different from cwd), added dirs, git branch. The identifier is the
  derived peer-messaging agent name (e.g. `@daneel-cf`, addressable via
  `SendMessage`) looked up from `~/.claude/sessions/*.json` by matching
  `session_id`; falls back to `#<first 8 chars of session_id>` if no match
  is found.
- Line 2: context-usage thermometer and used/capacity tokens, cost, session
  duration, cache token counts, 200k+ warning, 5h/7d rate-limit
  thermometers.
- Model details come from `model.display_name` (falling back to `model.id`),
  `effort.level`, `thinking.enabled`, and `fast_mode`. Optional or disabled
  effort/mode indicators are omitted; the exact model ID is not shown when
  a display name is available. Bare family names (`Opus`, `Sonnet`, `Haiku`)
  include the version parsed from a matching model ID, for example
  `claude-opus-5-5` becomes `Opus 5.5`. Display names already containing a
  version are preserved; date and context suffixes are not treated as versions.
- Context capacity comes from `context_window.context_window_size`, not the
  model name. Usage sums current input and cache tokens, falling back to
  `total_input_tokens` when available. Output tokens are excluded to match
  Claude's input-only `used_percentage`. Before usage is available (including
  immediately after compaction), only capacity is added; missing capacity
  preserves the percentage-only display. Token counts use `k`/`M` units.
- Payload reference: [Claude Code statusline data](https://code.claude.com/docs/en/statusline#available-data).
