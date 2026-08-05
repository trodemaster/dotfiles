# Resolved issue: `_direnv_resolve: command not found` on sandboxed `Bash` calls

Status: **RESOLVED** (round 4). Root cause was an upstream Claude Code CLI
bug — the shell-snapshot mechanism drops single-leading-underscore function
names ([#55816](https://github.com/anthropics/claude-code/issues/55816),
[#40602](https://github.com/anthropics/claude-code/issues/40602)). Fix:
renamed `_direnv_resolve` to `__direnv_resolve` (double leading underscore)
in `dot_bash_env.tmpl`, confirmed working on both hosts. See "Update (round
4)" near the end for the closing confirmation; everything above it is the
investigation trail that got there — kept intact since the empirical
snapshot-inspection findings (rounds 1–2) are the actual evidence the fix
rests on.

## Symptom

On host `umac.local` (personal Mac), inside Claude Code's `Bash` tool, any
command that goes through the `cd()` wrapper defined in `~/.bash_env`
consistently fails like this:

```
$ cd ~/scratch/direnvtest && echo "FOO=$FOO"
/Users/blake/.claude/shell-snapshots/snapshot-bash-<random>.sh: line 53: _direnv_resolve: command not found
```

Exit code 127. This reproduces on essentially every sandboxed `Bash` tool
call that invokes `cd`, across multiple full session restarts (confirmed by
a new snapshot filename each time).

**On a second host** ("the known-good host", a work Mac) running the
identical Claude Code CLI version (`2.1.222`), the user reports this does
**not** happen — `cd` + direnv resolution works as expected. The user
believes both hosts run the same `~/.bash_env` (from the shared `dotfiles`
repo) and asked us to find what's different.

## What's confirmed NOT to be the problem

**`~/.bash_env` itself is correct.** Sourcing it directly in a clean,
correctly-configured non-interactive shell defines both functions properly:

```
$ env BASH_ENV="$HOME/.bash_env" bash -c 'type _direnv_resolve; type cd'
_direnv_resolve is a function
_direnv_resolve () { ... }
cd is a function
cd () { builtin cd "$@" && _direnv_resolve; }
```

So the bug is not in the template logic (`dot_bash_env.tmpl`) — it's
somewhere in how Claude Code's `Bash` tool sets up the shell for each
invocation on this host specifically.

**settings.json checklist items were already correct / have been fixed.**
Per `dotfiles/skills/direnv/SKILL.md`'s existing checklist, we verified/set
on this host:
- `env.BASH_ENV` → `/Users/blake/.bash_env` (was already correct)
- `sandbox.excludedCommands` includes `"direnv *"` (was **missing**, added
  during this session)
- `permissions.deny` includes `Bash(direnv export*)` / `exec*` / `dump*`
  (was **missing entirely** — no `deny` array existed at all — added during
  this session)

None of these additions fixed the symptom (retested after each).

**`sandbox.enableWeakerNestedSandbox` is not the (sole) cause.** This was
the standout structural difference found when diffing this host's
`sandbox` block against the known-good host's (see full blocks below): the
known-good host has it `true`, this host didn't have it set at all. Added
`"enableWeakerNestedSandbox": true` to this host's settings.json, did a
**full session restart** (`/exit` + reconnect — confirmed via a new
shell-snapshot filename, so this wasn't a stale-config false negative), and
reran the exact same test. Still failed identically. Ruled out.

## What's confirmed TO BE happening — the shell-snapshot mechanism

Claude Code's `Bash` tool restores prior shell state (aliases, functions,
shell options) into each fresh invocation via a generated file at
`~/.claude/shell-snapshots/snapshot-bash-<id>-<random>.sh`. A new one is
generated per session/call — filenames observed across this investigation:
`snapshot-bash-1785905300665-d80mj5.sh`,
`snapshot-bash-1785907416119-gpei4v.sh`,
`snapshot-bash-1785908574419-ltc2pn.sh`.

These files consist almost entirely of lines like:

```bash
eval "$(echo '<base64>' | base64 -d)" > /dev/null 2>&1
```

— one per captured function, each base64-encoding a full function
definition (`declare -f`-style output). Decoding every line across all
three inspected snapshots and listing captured function names:

- **Present, intact, every time:** ~50 bash-completion library functions
  with a **double**-underscore prefix — `__op_handle_word`,
  `__op_handle_flag`, `__limactl_debug`, `__limactl_get_completion_results`,
  `__autopkgcomp`, `__asrcomp`, `__blesscomp`, `__du_comp`, `__hdi_comp`,
  `__launchcomp`, `__logcomp`, etc. — plus a handful of short personal
  functions with no underscore or a name-suffix underscore: `cd`,
  `dequote`, `exitssh`, `killsshhost`, `quote`, `quote_readline`.
- **Absent, every time, without exception:** every **single**-leading-underscore
  function. This includes `_direnv_resolve` (the one we care about) but
  also generic bash-completion primitives that the captured double-underscore
  functions call by name but which are never themselves defined anywhere in
  the snapshot: `_init_completion`, `_get_comp_words_by_ref`, `_filedir`,
  `_comp_dequote`, `_comp_quote_compgen`, `_comp_ltrim_colon_completions`.

Verified with a direct grep-and-decode pass over an entire snapshot file:
zero functions matching `^_[a-zA-Z]` (single underscore + letter) exist
anywhere in it, while `^__[a-zA-Z]` (double underscore) functions are
abundant and fully intact.

Notably, `cd()`'s captured body is byte-for-byte the current
`dot_bash_env.tmpl` definition:

```bash
cd () 
{ 
    builtin cd "$@" && _direnv_resolve
}
```

So the wrapper that *depends on* `_direnv_resolve` survives snapshotting;
the dependency itself does not. This is a clean, deterministic pattern —
not random truncation (the alphabetical block of personal functions
`cd, dequote, exitssh, killsshhost, quote, quote_readline` runs to
completion past where `_direnv_resolve` should sort, so it's not a
size/count cutoff cutting the list off early — if it were, dequote through
quote_readline wouldn't be there either since `_` sorts before all lowercase
letters in ASCII and `_direnv_resolve`/`_comp_*` would sort *before* `cd`).

**Working theory:** Claude Code's snapshot-capture step has a filter (either
an intentional heuristic to avoid dumping bash-completion library internals,
which are conventionally single-underscore-prefixed, or an unrelated bug
that happens to correlate with that naming pattern) that drops any function
whose name starts with exactly one leading underscore. This would be a CLI
implementation detail, not something under our control via `settings.json`.

## A/B test: sandboxed vs `dangerouslyDisableSandbox`

Ran the identical command through the `Bash` tool twice — once through the
normal sandboxed path, once with `dangerouslyDisableSandbox: true`:

```
# Sandboxed (default):
$ cd ~/scratch/direnvtest && type _direnv_resolve >/dev/null 2>&1 && echo DEFINED || echo MISSING
snapshot-bash-....sh: line 53: _direnv_resolve: command not found
SANDBOXED: _direnv_resolve MISSING

# dangerouslyDisableSandbox: true, same command:
UNSANDBOXED: _direnv_resolve DEFINED
```

This is what originally pointed at the sandbox subsystem as implicated
(later narrowed down to *not* being `enableWeakerNestedSandbox`
specifically — see above). It's still true that the failure only manifests
on the sandboxed code path; unsandboxed calls appear to take a different
execution route that sources `~/.bash_env` cleanly without going through
the snapshot-restore step that's dropping the function.

## Where the error is actually thrown from — inconclusive

The error is always attributed to the snapshot file at a specific line
(consistently reported as "line 53" across multiple different snapshot
files/sessions in this investigation). Directly inspecting line 53 of each
implicated snapshot file shows it's just the `eval` statement defining
`exitssh()` — unrelated to `cd` or `_direnv_resolve` on its face. Grepping
each full snapshot file for any literal (non-base64-encoded) `cd`
invocation outside of the function-definition `eval` lines turned up
nothing — the snapshot file itself never calls `cd`, it only defines
functions/aliases/shell options.

This means the actual `cd` invocation that trips the missing-function error
must come from a wrapper *outside* the snapshot file — something that
sources the snapshot, then separately runs `cd <tracked-directory>` to
restore the working directory for this call, invoking the now-broken
snapshot-sourced `cd()` before `~/.bash_env`/`BASH_ENV` has (re-)fired in
that same process. The consistent "line 53" attribution across different
snapshot files/sessions is not yet explained — bash's line-number
attribution for errors inside a function defined via `eval` of a decoded
string is not something we fully traced. Take the "line 53" detail as an
observed constant, not an understood one.

## What the existing `direnv` skill says (and doesn't cover)

`dotfiles/skills/direnv/SKILL.md` documents that:

- Each `Bash` tool call is a fresh, short-lived bash process; only the
  working directory carries over between calls (via wherever the harness
  starts the next process), not shell/function state.
- Within a *single* call, `~/.bash_env` can fire **twice** against the same
  PID — the harness apparently runs an initial setup stage, then `exec`s
  into a final shell, and since `exec` preserves the PID but re-triggers
  bash's own `BASH_ENV` startup handling, the file gets sourced again.
- The direnv-resolution block (`cd()` / `_direnv_resolve()` definitions) is
  deliberately left *outside* the `_BASH_ENV_GUARD`-guarded block so it
  safely re-runs on both firings (function redefinition is idempotent, so
  this was judged harmless).

The skill does **not** mention the `~/.claude/shell-snapshots/` mechanism
at all, nor account for a possible early stage where a *stale, filtered*
snapshot's `cd()` could get invoked before `~/.bash_env` has fired in that
process for the first time. That's the gap this investigation is pointing
at, but it hasn't been proven — only inferred from indirect evidence (the
skill's own documented two-stage/exec model, plus the snapshot contents).

## Proposed fix (NOT YET APPLIED — investigation was paused here)

Make `cd()` defensive so a not-yet-defined `_direnv_resolve` never
surfaces as a crash, regardless of which stage of the process invokes it:

```bash
cd() {
  builtin cd "$@" || return
  declare -F _direnv_resolve >/dev/null && _direnv_resolve
  return 0
}
```

This doesn't fix whatever is stripping `_direnv_resolve` from the
snapshot (still not understood, and possibly outside our control entirely
— a Claude Code CLI implementation detail) — it just makes `cd()` tolerate
that function being transiently undefined instead of erroring. The
expectation is that once `~/.bash_env` does get its "real" firing later in
the same call (per the skill's documented two-stage model), direnv
resolution still happens correctly for the actual user command; this
defensive `cd()` change only silences the spurious failure in the
earlier/stale-snapshot window.

**This has not been tested yet.** The edit was drafted against
`~/.local/share/chezmoi/dot_bash_env.tmpl` (the fast local-iteration copy
per the skill's own debugging tip) but was interrupted before being
applied or verified. A reviewer picking this up should:

1. Apply the `cd()` change above (to `dot_bash_env.tmpl` in this repo, then
   sync to the chezmoi source and `chezmoi apply` for fast iteration).
2. Retest the exact repro (`cd ~/scratch/direnvtest && echo $FOO`, or
   similar) across a few fresh sandboxed `Bash` tool calls.
3. Ideally also get direct confirmation from the known-good host — we were
   mid-way through asking the user to run `type cd` and `type
   _direnv_resolve` there for a byte-for-byte comparison, to rule out (or
   confirm) that its `~/.bash_env` is actually out of sync / structurally
   different rather than the CLI behaving differently per host. That
   comparison was never completed.
4. If the CLI's snapshot-stripping of single-underscore functions really is
   the root mechanism, consider whether it's worth reporting upstream to
   Anthropic (Claude Code CLI issue), since it would affect anyone whose
   dotfiles define single-underscore-prefixed helper functions referenced
   from a wrapped builtin like `cd`.

## Open questions for a reviewer

- Why does the known-good host not exhibit this at all, if the CLI version
  is identical and the snapshot-stripping behavior (if real) should be
  version-driven, not host/config-driven? Unresolved. Candidate factors not
  yet ruled out: total number/size of bash-completion scripts loaded into
  the interactive shell that seeds the snapshot (this host has many:
  1Password `op`, `limactl`, `autopkg`, `dbus-send`, `hdiutil`, `blessed`,
  etc. — possibly more than the work host), in case there's a size-based
  interaction alongside whatever is filtering by name; or the two hosts'
  actual deployed `~/.bash_env` differing more than assumed (unverified,
  see item 3 above).
- Is "single leading underscore, not double" really the filter rule, or is
  that a coincidental correlation given the specific functions available to
  observe? Only two data points contribute the "should be present but
  isn't" signal for single-underscore names beyond `_direnv_resolve`
  itself (the referenced-but-undefined `_init_completion`/`_filedir`/etc.
  completion primitives) — worth stress-testing with a deliberately-added
  throwaway single-underscore function to see if it reliably reproduces.
- Is the "line 53" attribution meaningful/debuggable further, or a red
  herring from how bash reports errors inside `eval`-defined function
  bodies? Not resolved.

## Relevant `sandbox` config from both hosts, for comparison

### This host (umac.local / personal, currently broken) — after this
session's fixes (`enableWeakerNestedSandbox` added and ruled out;
`excludedCommands` and `permissions.deny` direnv entries added per the
skill's checklist, also insufficient alone)

```json
"permissions": {
  "deny": [
    "Bash(direnv export*)",
    "Bash(direnv exec*)",
    "Bash(direnv dump*)"
  ]
},
"sandbox": {
  "enabled": true,
  "autoAllowBashIfSandboxed": true,
  "enableWeakerNestedSandbox": true,
  "network": {
    "allowedDomains": [
      "github.com",
      "objects.githubusercontent.com",
      "raw.githubusercontent.com",
      "api.github.com",
      "crates.io",
      "static.crates.io",
      "distfiles.macports.org",
      "ports.macports.org",
      "proxy.golang.org",
      "sum.golang.org",
      "storage.googleapis.com",
      "registry.npmjs.org",
      "registry.yarnpkg.com",
      "codeload.github.com",
      "release-assets.githubusercontent.com"
    ],
    "allowUnixSockets": [
      "/Users/blake/.lima/default/ha.sock",
      "/Users/blake/.lima/default/ssh.sock",
      "/Users/blake/.lima/default/default_ep.sock",
      "/Users/blake/.lima/default/default_fd.sock",
      "/Users/blake/.lima/weewx-dev/ha.sock",
      "/Users/blake/.lima/weewx-dev/ssh.sock",
      "/Users/blake/.lima/weewx-dev/default_ep.sock",
      "/Users/blake/.lima/weewx-dev/default_fd.sock"
    ],
    "allowLocalBinding": true
  },
  "filesystem": {
    "allowWrite": [
      "/opt/local",
      "/tmp",
      "/private/tmp",
      "/Users/blake/Developer/blakeports",
      "/Users/blake/Developer/machine-cfg",
      "/Users/blake/.lima"
    ],
    "denyWrite": [
      "/Users/blake/.lima/_config/user"
    ],
    "denyRead": [
      "/Users/blake/Developer/machine-cfg/tools/certool/output"
    ]
  },
  "enableWeakerNetworkIsolation": true,
  "allowAppleEvents": true,
  "excludedCommands": [
    "gh",
    "ekctl",
    "amail",
    "lima",
    "limactl",
    "diskutil",
    "hdiutil",
    "sudo port",
    "direnv *"
  ]
}
```

### Known-good host (work Mac, direnv works fine) — as reported by the user

```json
"permissions": {
  "deny": [
    "Bash(secrets keychain get *)",
    "Bash(secrets keychain export *)",
    "Bash(secrets 1pass get *)",
    "Bash(secrets 1pass export *)",
    "Bash(secrets vault-token-helper get)",
    "Bash(vault token lookup*)",
    "Bash(direnv export*)",
    "Bash(direnv exec*)",
    "Bash(direnv dump*)",
    "Read(*.tfvars)",
    "Read(*.auto.tfvars)",
    "Read(~/.azure/msal_token_cache.json)",
    "Read(~/.azure/msal_http_cache.bin)",
    "Read(~/.azure/config)"
  ]
},
"model": "claude-opus-4-8[1m]",
"sandbox": {
  "enabled": true,
  "failIfUnavailable": true,
  "autoAllowBashIfSandboxed": true,
  "allowUnsandboxedCommands": true,
  "network": {
    "allowedDomains": [
      "graph.microsoft.com",
      "login.microsoftonline.com",
      "management.azure.com",
      "pre-signed-firefly-prod.s3-accelerate.amazonaws.com",
      "inside.corp.adobe.com"
    ],
    "allowMachLookup": [
      "com.apple.trustd",
      "com.apple.SystemConfiguration.configd"
    ]
  },
  "filesystem": {
    "allowWrite": [
      "$TMPDIR",
      "/tmp/claude",
      "/private/tmp/claude",
      "~/.claude/skills",
      "~/.azure",
      "~/Library/Caches/golangci-lint",
      "~/Library/Caches/go-build"
    ]
  },
  "enableWeakerNestedSandbox": true,
  "enableWeakerNetworkIsolation": true,
  "excludedCommands": [
    "chezmoi *",
    "docker-compose *",
    "docker *",
    "gh *",
    "git *",
    "/Applications/Obsidian.app/Contents/MacOS/Obsidian *",
    "osascript *",
    "open *",
    "vault *",
    "az *",
    "direnv *",
    "terraform *",
    "dev-browser *",
    "nc *",
    "limactl *",
    "lima *",
    "ssh *",
    "make *",
    "go *",
    "vpn *",
    "curl *"
  ]
},
"tui": "fullscreen"
```

Key remaining differences after this session's changes: known-good host's
`excludedCommands` is far broader (git, curl, make, go, ssh, and many more
run fully unsandboxed there, vs. only a handful on this host); known-good
host sets `failIfUnavailable`, `allowUnsandboxedCommands` explicitly, `tui:
fullscreen`, and a non-default `model`; different `network.allowedDomains`
and `filesystem.allowWrite` sets (expected — those are legitimately
environment-specific). None of these have been individually tested as the
differentiator except `enableWeakerNestedSandbox`, which was ruled out.

## Both hosts' CLI version (confirmed identical)

```
$ claude --version
2.1.222 (Claude Code)
```

## Update (round 2): confirmed as an upstream Claude Code CLI bug — settings.json avenue mostly ruled out

### Confirmed: this is an already-filed upstream bug, not a local misconfiguration

[GitHub issue #55816](https://github.com/anthropics/claude-code/issues/55816)
and [#40602](https://github.com/anthropics/claude-code/issues/40602)
independently document exactly the pattern found in this file's round 1:
Claude Code's shell-snapshot builder
(`~/.claude/shell-snapshots/snapshot-bash-*.sh`) systematically drops
functions with a single leading underscore (`_direnv_resolve`, `_encode`,
`_lazy_load_chruby`, etc.) while preserving double-underscore and
plain-named functions — including the `cd()` wrapper that calls
`_direnv_resolve`. This is a defect in the CLI itself, confirmed via
official docs/issue tracker research, not just inferred locally.

### Settings.json hypotheses tested this round — all ruled out (or backwards)

Before accepting the round-1 diff of the two hosts' `sandbox` blocks as the
explanation, pulled authoritative behavior for the specific flags involved
(via Claude Code's own `sandboxing.md` docs) rather than continuing to
guess from the config diff alone:

- **`sandbox.excludedCommands` missing wildcards** (round 1 flagged that the
  broken host's `"gh"`, `"lima"`, `"limactl"`, etc. lack the trailing ` *`
  that the known-good host's equivalent entries have). Weaker evidence than
  first presented — bare tokens likely still prefix-match in practice per
  behavior reported in
  [anthropics/claude-code#40831](https://github.com/anthropics/claude-code/issues/40831),
  and this setting's matching semantics are largely undocumented/buggy in
  general
  ([#53012](https://github.com/anthropics/claude-code/issues/53012)).
  More importantly: `"direnv *"` — the one pattern that actually matters
  for *this* bug — is correctly wildcarded and present on **both** hosts
  already, so wildcard-ness was never the differentiator for this specific
  symptom regardless of the general matching-semantics question.
- **`sandbox.allowUnsandboxedCommands`**: defaults to `true` (confirmed via
  official docs). It gates the `dangerouslyDisableSandbox` retry-escape-hatch,
  not whether `excludedCommands` has any effect at all (that was an
  incorrect guess made mid-investigation). Both hosts are effectively `true`
  either way. Ruled out.
- **`sandbox.failIfUnavailable`**: defaults to `false`. When false/unset, a
  sandbox-init failure causes Claude Code to silently fall back to
  **unsandboxed** execution — which, per this file's own round-1 A/B test,
  is the condition that does *not* exhibit the bug. So this flag being unset
  on the broken host would point away from causing the failure, not toward
  it, if it were relevant at all. Ruled out.

**Conclusion: stop pursuing settings.json reconfiguration as the primary
fix.** The mechanism is a CLI defect, not a config gap on either host. The
open question is now specifically **why one host triggers it and the other
doesn't**, given both run the identical CLI build (see below).

### Both hosts confirmed to be running the exact same build

Not just the same version string — checked `claude doctor` on both:

```
Running: native (2.1.222)
Commit: fbf49312c284
Platform: darwin-arm64
Config install method: native
Auto-update channel: latest
Last update attempt: success → 2.1.222 (2026-08-04)
```

Identical on both hosts, including the commit hash and last-update date.
Rules out "different npm package / build channel" as an explanation —
this is genuinely the same compiled binary on both machines.

### Repro re-verified NOT to occur on the known-good host (nextbook, work Mac)

Ran the exact repro from this file's round 1, freshly, fully sandboxed
(no `dangerouslyDisableSandbox`):

```
$ mkdir -p ~/scratch/direnvtest && printf 'export FOO=bar\n' > ~/scratch/direnvtest/.envrc
   # (required disabling sandbox just for this one file-write step —
   # ~/scratch isn't in this host's sandbox.filesystem.allowWrite)
$ cd ~/scratch/direnvtest && direnv allow
   # separate tool call — standard direnv-allow-then-retry-in-a-new-call pattern
$ cd ~/scratch/direnvtest && echo "FOO=$FOO"
   # separate, fully sandboxed tool call — the actual repro
FOO=bar
```

No "command not found," `_direnv_resolve` fired correctly both times
(gracefully warning on the first `cd` since `.envrc` wasn't allowed yet,
then correctly resolving `FOO=bar` on the second). Confirms, again, this
host isn't affected.

**New observation, possibly relevant:** both `cd` calls into
`~/scratch/direnvtest` produced an extra line of tool output that no
command in the chain actually printed:

```
Shell cwd was reset to /Users/blake/Developer/dotfiles
```

This appears to be the Claude Code harness itself (not bash, not direnv)
reporting that it moved this session's tracked working directory back to
the primary project directory after the call. This sits structurally close
to the suspected mechanism (a harness-level cd/state-restore step adjacent
to the shell-snapshot restore) — but it fired here with **no error**, so
its mere presence isn't sufficient on its own to trigger the bug, at
minimum. Worth checking whether this same message appears in the broken
host's repro output, and if so, its exact timing relative to the crash.

### SSH-based checks on the broken host (umac.local) — config/baseline confirmed current, but can't reproduce the actual bug this way

Connected via `ssh umac` to check live state without going through an
actual Claude Code session there (important caveat: **plain SSH command
execution does not go through Claude Code's Bash-tool sandbox/snapshot
machinery at all** — it can only verify static state, not reproduce the
bug itself):

- **`settings.json` matches what's pasted earlier in this file** —
  `sandbox.allowUnsandboxedCommands: true` and the rest of the `sandbox`
  block are confirmed current on disk right now, not a stale snapshot from
  round 1.
- **Correction to a side-finding above, verified directly on `umac` in a
  live session (not over SSH) immediately after round 2 landed:** the claim
  that "this host has no such deny rule at all" is wrong as of right now —
  `permissions.deny` on this host already has
  `Bash(direnv export*)` / `exec*` / `dump*`, added during round 1 of this
  same investigation (it's the block pasted earlier in this file, under
  "Relevant sandbox config from both hosts"). The blanket
  `permissions.allow: "Bash(direnv *)"` entry does still coexist alongside
  it, which is fine — `deny` takes precedence over `allow` in Claude Code's
  permission model, so the narrower `deny` rules already block the literal
  `direnv export/exec/dump` invocations regardless of the broader `allow`
  entry. Either the SSH check above ran before round 1's edit had been
  saved, or it mis-read the block. Not a real gap; no action needed here.
  Separately, this session also independently re-confirmed round 2's
  `allowUnsandboxedCommands` conclusion: added it explicitly, did a full
  session restart (new shell-snapshot filename
  `snapshot-bash-1785909876813-43xrjv.sh`), verified it was `true` in the
  live resolved config, and the bug still reproduced identically — same
  "ruled out" result round 2 reached via docs research.
- **Baseline `~/.bash_env` confirmed correct outside Claude Code entirely**
  (matches round 1's finding exactly, re-verified fresh):
  ```
  $ env BASH_ENV=$HOME/.bash_env bash -c 'type _direnv_resolve; type cd'
  _direnv_resolve is a function
  ...
  cd is a function
  cd () { builtin cd "$@" && _direnv_resolve }
  ```
  The template/rendered file itself has never been the problem.
- `~/scratch/direnvtest/.envrc` from round 1 is still present and intact.

**Dead end, don't retry: headless remote repro via `claude -p` over SSH.**
Attempted `ssh umac "claude -p '...'"` to spawn a one-shot headless Claude
Code session and drive the actual Bash-tool/sandbox/snapshot path remotely
without an interactive session. Failed immediately: `Not logged in ·
Please run /login`. Headless mode on that host isn't authenticated
independently of whatever session/credentials the interactive app uses —
this isn't a viable path for remote reproduction and shouldn't be
re-attempted the same way.

### Where this leaves things for whoever picks this up on `umac` directly

1. Re-run the exact repro (`cd ~/scratch/direnvtest && echo $FOO`, fresh
   session/new shell-snapshot, no `dangerouslyDisableSandbox`) and confirm
   it still fails the same way, to rule out anything having changed since
   round 1.
2. Check whether the `Shell cwd was reset to ...` harness message (see
   above) also appears in this host's repro output, and if so, look at its
   exact ordering relative to the `command not found` error — before,
   after, or interleaved with the crash?
3. Given the root mechanism is now a confirmed upstream CLI bug (not
   fixable via settings.json), the open question is specifically **why
   this host triggers it and the known-good host doesn't**, despite
   identical CLI build. Candidate angles not yet tested:
   - Total number/size of bash-completion scripts sourced into the
     interactive shell that seeds the snapshot (this host has many — 1Password
     `op`, `limactl`, `autopkg`, `dbus-send`, `hdiutil`, etc. — possibly
     enough to hit a size-based cutoff or ordering issue alongside the
     name-based filter).
   - Whether the two hosts' actual live `~/.bash_env` are truly identical
     right now — confirm with a live diff/hash between the two hosts, not
     an assumption based on "should be, per the shared dotfiles repo."
   - Whether frequency/pattern of `cd` usage in this host's typical
     workflow exercises the buggy early-snapshot path more than it does on
     the known-good host.
4. If root-caused further, consider commenting on
   [anthropics/claude-code#55816](https://github.com/anthropics/claude-code/issues/55816)
   or [#40602](https://github.com/anthropics/claude-code/issues/40602) with
   these findings — the sandboxed-vs-`dangerouslyDisableSandbox` A/B test
   and the cross-host non-reproduction from this investigation is genuinely
   useful data that doesn't seem to be on either issue yet.
5. The defensive `cd()` fix drafted in round 1 (guard the call with
   `declare -F _direnv_resolve`) is still on the table as a workaround if
   root-causing further doesn't pan out — but the current direction is to
   prefer understanding/reporting the actual cause over a host-specific
   code workaround if at all avoidable.

## Update (round 3): renamed `_direnv_resolve` → `__direnv_resolve` (double leading underscore) — tested on the known-good host, needs testing on `umac`

Since the snapshot-capture filter provably preserves double-underscore
functions (round 2's confirmed upstream issues, plus round 1's own snapshot
inventory — `__op_handle_word`, `__limactl_debug`, etc. all survive every
time) while dropping single-underscore ones, the direct test is simply: do
the double-underscore variant actually survive too, and does that make the
`cd()`-triggered failure go away? Renamed `_direnv_resolve` to
`__direnv_resolve` everywhere in `dot_bash_env.tmpl` (the function
definition, the `cd()` wrapper's call, and the standalone call at the end
of the block) and in `SKILL.md`'s matching code/prose.

**Tested on the known-good host (nextbook) — passed, with one important
caveat about *when* to retest.** First attempt was invalidated by session
staleness, not a real result: edited the file mid-session (via the
fast-iteration `~/.local/share/chezmoi/dot_bash_env.tmpl` + `chezmoi apply`
loop), then immediately reran the repro — it failed with `_direnv_resolve:
command not found` (note: the *old* single-underscore name in the error).
That's because this session's shell-snapshot was captured *before* the
edit, so its cached `cd()` still called the now-nonexistent old name. This
is itself a useful, independent confirmation of the "stale/early snapshot
`cd()` invoked before `~/.bash_env`'s real per-call firing" mechanism —
same failure shape as the upstream bug, just caused by a self-inflicted
mid-session mismatch instead of the snapshot filter.

After a full session restart (fresh shell-snapshot, new filename), reran
the exact repro:

```
$ cd ~/scratch/direnvtest && echo "FOO=$FOO"
FOO=bar
```

No error. Then directly confirmed (not just inferred from the absence of an
error) that `__direnv_resolve` is actually present in the fresh snapshot,
by extracting and base64-decoding every captured function from
`~/.claude/shell-snapshots/snapshot-bash-<latest>.sh`:

```
__direnv_resolve
__expand_tilde_by_ref
__limactl_debug
... (all double-underscore completion functions)
cd
dequote
exitssh
killsshhost
quote
quote_readline
```

`__direnv_resolve` sits right alongside the other double-underscore
functions that have survived every snapshot inspected across both rounds
of this investigation — matching the theory exactly.

**Important: this is not proof the rename fixes the bug on `umac`.** This
host never reproduced the bug in the first place (round 1 and round 2 both
confirmed that), so this test only shows the rename doesn't break anything
here and behaves as expected given the confirmed filter rule. **The actual
test that matters is rerunning the exact repro on `umac`, in a fresh
session (new shell-snapshot — `/exit` and reconnect, or equivalent), after
pulling this rename.** If `__direnv_resolve` survives the snapshot there
too (it should, per the same filter rule, assuming the rule is really
"exactly one leading underscore" and not something more specific to this
particular function/host), the `command not found` error should simply
stop occurring — no defensive `cd()` code needed, no settings.json changes
needed, just the rename.

**If it works on `umac`:** this becomes the actual fix, not a workaround —
`_direnv_resolve` and `_BASH_ENV_GUARD`-adjacent naming conventions
elsewhere in `dot_bash_env.tmpl` should probably be reviewed too in case
anything else picks up a single-underscore name later (currently nothing
else does — `_BASH_ENV_GUARD` is a variable, not a function, and isn't
subject to this filter).

**If it doesn't work on `umac`:** that would mean the filter rule inferred
from round 1/round 2 isn't the whole story (e.g. maybe it's positional,
size-based, or specific to something about how `_direnv_resolve` in
particular got captured), and the investigation should go back to the open
questions in round 2's write-up rather than assuming double-underscore is
a universal escape hatch.

## Update (round 4): confirmed fixed on `umac` — status: RESOLVED

Pulled the rename via `chezmoi update` (already current — `~/.local/share/chezmoi`
was at `6e82a25`), confirmed `~/.bash_env` re-rendered with `__direnv_resolve`
throughout (`grep direnv_resolve ~/.bash_env` shows the function def, the
`cd()` call, and the standalone call all using the double-underscore name).

After a full session restart, reran the exact repro fully sandboxed (no
`dangerouslyDisableSandbox`), twice, with different commands chained after
the `cd`:

```
$ cd ~/scratch/direnvtest && echo "FOO=$FOO"
FOO=bar

$ cd /Users/blake/Developer/dotfiles && git log --oneline -1
6e82a25 direnv: rename _direnv_resolve to __direnv_resolve (double underscore)
```

No `command not found` error either time. This is the first clean sandboxed
`cd` on `umac` across the entire investigation (rounds 1–3 reproduced the
crash on essentially every sandboxed `cd`-involving call).

**Status: resolved.** The rename is the actual fix, confirmed on both
hosts now — not a workaround, not something requiring settings.json
changes. Per round 3's note, `dot_bash_env.tmpl` has no other
single-leading-underscore function names left to worry about
(`_BASH_ENV_GUARD` is a variable, not a function, and isn't subject to
this filter). No further action needed on this file unless the bug
resurfaces or someone wants to file the corroborating cross-host A/B
evidence from this investigation against the upstream issues
([#55816](https://github.com/anthropics/claude-code/issues/55816),
[#40602](https://github.com/anthropics/claude-code/issues/40602)).
