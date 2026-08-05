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
- Line 1: session name, model, cwd, project dir (if different from cwd),
  added dirs, git branch.
- Line 2: context-usage thermometer, cost, session duration, effort level,
  thinking indicator, cache token counts, 200k+ warning, 5h/7d rate-limit
  thermometers.
