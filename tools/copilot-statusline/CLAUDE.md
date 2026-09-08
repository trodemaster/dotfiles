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

- Line 1: session name, model, short session identifier, cwd, remote task
  indicator, and git branch.
- Line 2: live context-usage thermometer, GitHub AI Credits, session duration,
  thinking effort, cache and reasoning token counts, changed lines, and
  premium requests.

Copilot does not expose Claude Code's USD cost or 5-hour/7-day rate-limit
windows in the status-line payload, so those segments are intentionally
omitted.
