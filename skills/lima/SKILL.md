---
name: lima
description: Troubleshoot everyday Lima VM problems — starting, stopping, and diagnosing a VM that hangs or fails on `limactl start`/`stop`, plus Claude Code sandbox configuration for `limactl`. Use before running ANY `limactl` or `lima` command — including `list`, `shell`, `start`, `stop` — since a sandboxed Claude Code session needs the sandbox settings documented here or those commands fail with permission errors. Also use when a VM is slow to start, won't stop cleanly, or otherwise misbehaves.
---

# Lima Skill

Everyday use of [Lima](https://github.com/lima-vm/lima) VMs: diagnosing start/stop hangs and configuring Claude Code's sandbox so `limactl` works from an agent session.

---

## Troubleshooting VM Start/Stop Hangs

### Method: diff against a known-good baseline
Capture debug logs for both a healthy VM and the problem VM, then diff:
```bash
limactl start --debug <vm> 2>&1 | tee lima-start-<vm>.log
limactl stop  --debug <vm> 2>&1 | tee lima-stop-<vm>.log
```
A healthy `vz` boot: VM state → running in ~1s, port 22 reachable in ~5s, all three essential
requirements (`ssh`, `user session is ready for ssh`, `ControlMaster`) satisfied within another
second or two, `boot scripts must have finished` passes on the first or second poll. A healthy
stop: SIGINT/SIGTERM → SSH ControlMaster exit → `VZ vm state change: stopped` in ~2s. Anything
taking much longer than that means something is genuinely stuck, not just "VMs are slow."

### Failure mode: orphaned hostagent processes pile up
Symptom: `start` is slow/contended, or `stop` sends SIGINT/SIGTERM but the CLI eventually fatals
with `did not receive an event with the "exiting" status` — the VM still shows `Running` after.

Cause: each `limactl start` spawns a `hostagent` process bound to that instance's `ha.sock` /
`ha.pid` / `ssh.sock`. If a prior start/stop attempt didn't clean up (killed terminal, crashed
session, etc.), the old hostagent process leaks and keeps running — often spinning at 100% CPU
indefinitely. `limactl stop` only signals whatever PID is currently in `~/.lima/<vm>/ha.pid`
(the "canonical" one); orphans never receive the signal and hold the sockets forever, causing
contention on the next start and making stop look broken.

Diagnose:
```bash
cat ~/.lima/<vm>/ha.pid                        # the canonical PID
ps aux | grep -E 'hostagent|ssh.*ssh\.sock' | grep -v grep   # everything actually running
```
If more than one `hostagent` process references the same `--socket .../ha.sock`, the extras are
orphans. Fix: `kill -TERM` the orphans first, confirm they die, then deal with the canonical PID
(escalate to `kill -9` if it's stuck on shutdown — safe for the `vz` driver since that just
hard-powers-off the guest; not destructive to disk state as long as the guest had already finished
booting). Verify with `limactl list` (should show `Stopped`) and check `~/.lima/<vm>/*.sock`,
`*.pid` are gone.

### Failure mode: cloud-init network-wait fails on a masked systemd unit
Symptom: `start` doesn't fatal outright but takes several minutes longer than the ~10s baseline,
or does fatal with `did not receive an event with the "running" status` after ~10 min, even though
the guest is actually healthy and reachable (`limactl shell <vm>` still works — hostagent keeps
running and answering even after the CLI gives up).

Diagnose from inside the guest:
```bash
limactl shell <vm>
cloud-init status --long          # look for an ERROR block, not just "status: done" (non-fatal errors still show status done)
systemctl status systemd-networkd-wait-online.service   # "Loaded: masked" is the tell
networkctl status                 # confirms the network is actually fine — this is a false-negative readiness check, not real breakage
```
Seen concretely on an Ubuntu 26.04 (Resolute) cloud image: it ships `systemd-networkd-wait-online`
masked, but cloud-init's network module still calls `systemctl start` on it every boot. That call
fails immediately ("Unit is masked"); cloud-init's retry/backoff around the failure is what eats
several minutes, even though the network is already up by the time it happens.

Fix — unmask in a `mode: system` provision script in the instance's lima yaml (runs as root every
boot; the unmask persists across reboots since it edits `/etc/systemd/system` symlinks on disk, so
this only has to actually fire once):
```bash
systemctl unmask systemd-networkd-wait-online.service
systemctl enable systemd-networkd-wait-online.service
```
Caveat: lima's `system`-mode provisioning runs after cloud-init's own network module, so a brand
new VM's very first boot will still eat the delay once. Every boot after that is fast because the
unmask already landed on disk.

### `serialv.log` is not useful for boot diagnostics
`~/.lima/<vm>/serialv.log` only ever shows the `agetty` login banner (`<vm> login:`), on both
healthy and stuck instances — cloud-init output doesn't route to that console. Don't spend time
reading it; use `cloud-init status --long` / `journalctl -b -u cloud-init*` via a live shell
instead, or `~/.lima/<vm>/ha.stderr.log` for the hostagent's own debug trail (same content as
`limactl start --debug`, useful after the fact if you didn't capture it live).

---

## Claude Code sandbox configuration for `limactl`

Add per-VM instance to `sandbox` in `settings.json` (repeat the four socket lines and nothing else
per additional instance — everything below is the complete, trimmed set actually required):

```json
"sandbox": {
  "network": {
    "allowUnixSockets": [
      "~/.lima/<vm>/ha.sock",
      "~/.lima/<vm>/ssh.sock",
      "~/.lima/<vm>/default_ep.sock",
      "~/.lima/<vm>/default_fd.sock"
    ],
    "allowLocalBinding": true
  },
  "filesystem": {
    "allowWrite": ["~/.lima"],
    "denyWrite": ["~/.lima/_config/user"]
  },
  "excludedCommands": ["lima", "limactl"]
}
```

Note: `settings.json` does not expand `~` — substitute the actual absolute home directory path when
writing the real entries.

| Field | What it's doing here |
|---|---|
| `network.allowUnixSockets` | macOS-only exact-path allowlist (no wildcards — verified against the settings schema, which documents wildcard support elsewhere, e.g. `allowMachLookup`, but not here). `ha.sock`/`ssh.sock` need `connect`; the vzNAT network-helper sockets `default_ep.sock`/`default_fd.sock` (recreated fresh every boot) need `bind`. Without these: `list` shows a live instance as falsely `Broken`, and `start` fails with `bind: ...: operation not permitted`. |
| `network.allowLocalBinding` | Undocumented in the settings schema beyond its `boolean` type, but required alongside `allowUnixSockets` for the vzNAT helper's `bind()` calls on `default_ep.sock`/`default_fd.sock` to succeed — without it those binds fail even with the sockets allowlisted. |
| `filesystem.allowWrite` | hostagent unlinks/rewrites files under `~/.lima/<vm>/` (logs, sockets, pid file) every start. Without it: `unlinkat ...: operation not permitted`. Scoped to all of `~/.lima` rather than per-instance since that directory holds no other secrets. |
| `filesystem.denyWrite` | Carve-out for the one real secret under `~/.lima`: `_config/user` is lima's own SSH private key, shared across all instances. Everything else in `~/.lima` (yaml, disk images, sockets, logs, EFI vars) isn't credential material. |
| `excludedCommands: ["lima", "limactl"]` | Present in a working config but **does not fix** `start`/`stop` — see below. Kept because it's harmless and may help a future subcommand that hits a different, genuinely sandbox-classifiable failure. Don't rely on it for `start`/`stop`. |

### What works sandboxed once configured vs. what never will

Confirmed by direct testing, not inferred:

| Command | Sandboxed? |
|---|---|
| `limactl list` | **Works**, with the socket entries above |
| `limactl shell <vm> -- <cmd>` (instance already running) | **Works**, with the socket entries above — it's just an SSH connection over `ssh.sock`, no VZ or signal involved |
| `limactl start` | **Never** — Virtualization.framework itself refuses to run sandboxed, failing instantly with `Error Domain=VZErrorDomain Code=2 ... "Virtualization is not available on this hardware"` even though the hardware is fully capable. No settings.json field reaches this; it's a macOS entitlement boundary. |
| `limactl stop` | **Never** — sends SIGINT to the hostagent PID, which the sandbox blocks with `operation not permitted` since that process isn't a descendant of the sandboxed call. The CLI doesn't detect the failed signal and hangs polling forever; kill the stuck background task and retry. This holds regardless of whether the VM was started sandboxed or not. |

For `start`/`stop`, pass `dangerouslyDisableSandbox: true` on that specific Bash call — there's no
way to avoid it structurally.

**`sandbox.excludedCommands` is not a fix.** It reads like a bypass list but isn't one: it's a
fallback classifier — the command still runs sandboxed first, and only auto-retries unsandboxed if
the failure is recognized as sandbox-specific. `limactl`/`lima` were listed there throughout this
investigation and it prevented none of the failures above (the VZ entitlement error in particular
doesn't trigger the fallback). There is no settings.json field that marks a command as "always run
unsandboxed" — per-call `dangerouslyDisableSandbox: true` is the only mechanism.
