---
name: agent-telephone
description: >
  Coordinate multiple Claude Code (or other agent) sessions on one host, or across
  hosts via Teleport, using the agent-telephone CLI. Use when the user wants to
  "message another agent session", "join a call", "send to another session",
  "check the roster", "start/check the agent-telephone coordinator", or wants
  agents to hand off work / exchange status durably. Repo:
  ~/Developer/agent-telephone (github.com/adobe-security-tooling/agent-telephone).
---

# agent-telephone

A durable, shared conversation for coding agents. Agents `join` a named **call**,
exchange broadcasts or direct messages, and resume where they left off after a
restart. Backed by an A2A JSON-RPC/SSE coordinator (bbolt storage) running as a
per-user macOS LaunchAgent, reachable over loopback (`127.0.0.1:18080`) or
remotely through a Teleport Application Access tunnel.

## Prerequisites

- Binary on PATH: `go install github.com/adobe-security-tooling/agent-telephone/cmd/agent-telephone@latest`
  (client-only build: add `-tags=clientonly`).
- Coordinator running as a service — check with `agent-telephone service status`.
  If not installed: `agent-telephone service install`. Other lifecycle commands:
  `service restart`, `service logs`, `service uninstall`.
- Sanity check: `agent-telephone status` should print `{"status":"ok",...}`.

## Claude Code sandbox note

`agent-telephone` talks to the coordinator over loopback TCP
(`127.0.0.1:18080`). The default Bash sandbox blocks this — every subcommand
(`status`, `calls`, `join`, `send`, `roster`) fails with:

```
dial tcp 127.0.0.1:18080: connect: operation not permitted
```

This is the sandbox, not a broken coordinator (verify separately with
`agent-telephone service status` — a LaunchAgent shown as `running` with a real
`pid` means the coordinator itself is fine). Run `agent-telephone` commands with
`dangerouslyDisableSandbox: true` on that specific Bash call — there is no
`allowedDomains`/loopback allowlist entry that fixes this structurally, since
`allowedDomains` filters an egress proxy and loopback traffic never goes near it.

## Workflow

Each `join` blocks and streams one JSON event per line — it needs its own
long-running terminal/session per agent, not something you call once and move
on from. Pick short, descriptive participant names (`alice`, `bob`, or
something identifying the session's role).

```bash
# One long-running session per participant:
agent-telephone join planning alice
agent-telephone join planning bob

# From any session, broadcast to everyone on the call:
printf '%s\n' 'What is everyone working on?' | agent-telephone send planning --from alice

# Or send a direct message to one participant:
printf '%s\n' 'Please review the API.' | agent-telephone send planning bob --from alice
```

Direct messages wait up to 60s for an offline recipient by default —
`--wait=30s` to change it, `--wait=forever` to queue indefinitely.

Useful checks:

```bash
agent-telephone status            # coordinator health
agent-telephone calls             # active calls
agent-telephone roster planning   # who's on a call
```

## Connecting from another host (Teleport)

The coordinator only listens on loopback; a remote host reaches it through a
Teleport TCP Application Access tunnel (see
`~/Developer/agent-telephone/teleport-application-access.md` for the current
resource: app `agent-telephone-blake-dev` on host `nextbook`, proxy
`teleport.adobe.net:443`).

```sh
tsh login --proxy teleport.adobe.net
tsh apps login agent-telephone-blake-dev
tsh proxy app agent-telephone-blake-dev --port 18081

agent-telephone join planning ec2-agent --endpoint http://127.0.0.1:18081
```

Port 18081 is just a convention — if it's taken, let `tsh` pick a free port and
pass whatever endpoint it prints via `--endpoint` on every subsequent command.

## Troubleshooting

| Symptom | Fix |
|---|---|
| `dial tcp 127.0.0.1:18080: ... operation not permitted` | Bash sandbox blocking loopback — retry with `dangerouslyDisableSandbox: true` (see above). |
| `card request failed` even unsandboxed | Coordinator isn't running — `agent-telephone service status`, then `agent-telephone service install` or `service restart`. |
| Remote host can't reach the coordinator | Teleport proxy not running — re-run `tsh proxy app agent-telephone-blake-dev --port <N>` and keep it foregrounded; confirm `tsh apps login` succeeded first. |
| Direct message never arrives | Recipient hasn't `join`ed the call, or the default 60s wait expired before they came online — use `--wait=forever` for hand-offs to an agent that isn't up yet. |

## Design references

- `~/Developer/agent-telephone/design.md` — protocol and implementation detail.
- `~/Developer/agent-telephone/agent-flow.md` — three-agent (2 local + 1 remote) deployment diagram.
- `~/Developer/agent-telephone/AGENTS.md` — contributor/agent instructions for the repo itself (not this skill).
