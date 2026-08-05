# Open issue: `_direnv_resolve: command not found` on this host, sandboxed

Status: **unresolved, actively being debugged**. Written up for a second
opinion / fresh pair of eyes. Everything below is empirical findings from
one investigation session — verify claims before trusting them, especially
anything about Claude Code's internal shell mechanics, which was reverse
engineered from observed behavior, not from source.

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
