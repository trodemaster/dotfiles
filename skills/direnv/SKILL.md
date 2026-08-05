---
name: direnv
description: Reference for how direnv is wired into Claude Code's bash shell via BASH_ENV and chezmoi. Use when asked about .envrc files, BASH_ENV, direnv behavior in Claude sessions, environment variable inheritance in Bash tool calls, or why Claude picks up different env vars in different repos. Also use when debugging missing env vars, credential switches not taking effect, settings.json's sandbox/env config for direnv, or any question about how the shell environment works inside Claude.
---

# direnv + Claude Code Integration

## One Process Per Tool Call — But `BASH_ENV` Can Fire Twice Within It

Each `Bash` tool call spawns a **fresh bash process** — no state persists from one call to the next (variables, `_BASH_ENV_GUARD`, etc. all reset). The one thing that *does* persist across calls is the working directory: the harness remembers the last directory a `cd` landed in and starts the next call's process there.

**Within a single call, `~/.bash_env` can be sourced twice by the same PID.** Verified by adding a `pid=$$` debug line to the file: both firings reported the identical PID. The harness apparently execs into the final shell after an earlier setup stage, and since `exec` preserves the PID but re-triggers bash's own `BASH_ENV` startup handling, the file gets sourced again. `_BASH_ENV_GUARD` correctly stops the guarded block from running twice — but **anything left unguarded above/below the guarded block still runs on both firings**. This bit us for real: the file used to look like

```bash
export PATH="..."
export GH_CONFIG_DIR="/Users/blake/.config/gh-work"   # baseline default, unguarded
export VAULT_USER=blake
if [[ -z "$_BASH_ENV_GUARD" ]]; then
  export _BASH_ENV_GUARD=1
  eval "$(direnv export bash)"    # correctly resolves GH_CONFIG_DIR on firing #1
fi
```

Firing #1 ran the whole file and correctly set `GH_CONFIG_DIR` via direnv. Firing #2 (same PID) hit the guard and skipped the `direnv` line — but the **unguarded baseline `export GH_CONFIG_DIR=...gh-work` above it ran again**, silently clobbering the correct value firing #1 had just computed. Net effect: `GH_CONFIG_DIR` (and any other per-repo override) was *always* wrong at the end of the call, no matter what the direnv logic did, because the last write always won and the last write was always the unguarded default.

**The fix: put the entire file's body inside the guard**, not just the direnv call. See `dot_bash_env.tmpl` — `PATH`, the machine-cfg baseline vars, and the direnv resolution are all inside one `if [[ -z "$_BASH_ENV_GUARD" ]]; then ... fi` block now. A second same-PID sourcing is then a complete no-op, and whatever the first sourcing computed survives.

## The Core Mechanism

Claude Code sets `BASH_ENV=~/.bash_env` in `machine-cfg/claude/settings.json`:

```json
"env": {
  "BASH_ENV": "/Users/<you>/.bash_env"
}
```

Bash reads `$BASH_ENV` at startup for **every non-interactive shell**, including every `Bash` tool call Claude makes, at whatever the shell's actual starting cwd is (see below). This is distinct from the interactive hook — the full chain is:

| Context | Mechanism | File |
|---------|-----------|------|
| Claude Bash tool calls | `BASH_ENV` read at shell start | `~/.bash_env` |
| Interactive terminal sessions | `direnv hook bash` in `PROMPT_COMMAND` | `~/.bash_profile` |

**This `env.BASH_ENV` setting lives in `machine-cfg/claude/settings.json`, not in dotfiles.** `machine-cfg` has a separate upstream per system type (work vs. personal) — see `dotfiles/CLAUDE.md`. Each fork's `settings.json` needs this wiring independently; it does not propagate between forks. See "Settings.json Requirements" below for the full checklist to replicate on a new fork.

## `direnv export bash` Is Unreliable in a Cold, One-Shot Shell — Use `direnv exec` Instead

`direnv export bash` computes an *incremental diff* against its own prior invocation (tracked via the `DIRENV_DIR`/`DIRENV_FILE`/`DIRENV_DIFF`/`DIRENV_WATCHES` env vars it leaves behind). This makes it a great fit for an interactive shell, where `direnv hook bash` re-invokes it on every prompt and it converges quickly. It is a bad fit for a one-shot non-interactive shell:

- Verified experimentally that calling it once, twice, or even in a loop-until-empty-diff (up to 4 iterations) in a freshly-sourced `BASH_ENV` context could *still* leave a direnv-set var unresolved — the number of passes needed to converge was not consistent, and sometimes never converged within a handful of tries.
- Unsetting the inherited `DIRENV_*` tracking vars first did not reliably fix it either.

**Fix: don't use `direnv export bash` at all in this file.** Use `direnv exec . env -0` instead — it fully resolves `.envrc` (handling parent-directory discovery and the allow check exactly like `export` would) and executes the given command in that environment, dumping the *complete resulting environment* rather than an incremental diff. This has been reliable in every test. Re-export it into the current shell:

```bash
if command -v direnv &>/dev/null; then
  while IFS= read -r -d '' _direnv_kv; do
    _direnv_key=${_direnv_kv%%=*}
    [[ $_direnv_key =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] && export "$_direnv_kv"
  done < <(direnv exec . env -0 2>/dev/null)
  unset _direnv_kv _direnv_key
fi
```

Note the identifier check on the key: some system daemons set malformed env vars with literal quote characters baked into the name/value (e.g. `'SOME_VAR'='...'` as the literal string). Re-exporting those blindly throws a "not a valid identifier" error from bash's `export`; skipping non-identifier keys avoids that cosmetically, harmlessly.

Use process substitution (`< <(...)`) for the read loop, not a pipe (`... | while read`) — a pipe puts the loop in a subshell, so `export` inside it would not affect the calling shell at all.

## Handling `direnv allow`

**`.envrc` files must be explicitly allowed before direnv will load them.** This is a one-time per-machine step required after creating or editing an `.envrc`. If a directory has an `.envrc` that hasn't been allowed, direnv writes an error to stderr:

```
direnv: error /path/to/.envrc is blocked. Run `direnv allow` to approve its content.
```

Run `direnv allow` in one tool call, then retry the original command in a **separate, subsequent** tool call — `direnv allow` only writes a trust record to disk, it does not reload the environment in the current shell process:

```bash
# Tool call 1: allow it — nothing else, don't chain the retry here
cd ~/Developer/some-repo && direnv allow

# Tool call 2: a brand-new shell process, BASH_ENV fires fresh and now loads it
cd ~/Developer/some-repo && <original command>
```

**Caveat learned the hard way:** a wrong env var value (e.g. `GH_CONFIG_DIR`) is *not* always an allow problem. It can equally be the double-sourcing/unguarded-clobber bug described above, which has nothing to do with `direnv allow`. Check `direnv status` for an explicit "not allowed" signal before assuming that's the cause — don't reflexively reach for `direnv allow` as the fix for every wrong-env-var symptom.

## `cd` and Credentials — What Actually Works Now

`BASH_ENV` fires **at the shell's actual starting cwd** — the directory the harness starts the process in, which is wherever the *previous* tool call's `cd` left off (or the primary project directory, for the first call in a session). Now that the whole file is guarded and uses `direnv exec` correctly, **a tool call that starts in a repo with a credential-switching `.envrc` will automatically pick up the right `GH_CONFIG_DIR`, `VAULT_ADDR`, etc. with zero manual intervention.** Verified: pushing to a personal-fork repo (with its own `gh` profile) now just works with a plain `git push`, no workaround needed.

**What still doesn't work, and why:** a single command that does `cd other_dir && some_command` will *not* pick up `other_dir`'s `.envrc` for `some_command` — `BASH_ENV` already fired (at the *previous* starting cwd) before that `cd` runs, same as always. This is a real, structural limit, not a bug to chase further.

**The correct pattern when you need to switch repos:** issue the `cd` as its **own tool call first**:

```bash
# Tool call 1
cd ~/Developer/machine-cfg

# Tool call 2 — starts with cwd already = machine-cfg, BASH_ENV resolves its .envrc correctly
git push
```

This is strictly better than the old advice of asking the user to run the push themselves — it's fully automatic now, it just needs the `cd` to be its own call rather than chained with `&&` into the credentialed command.

**Never use `eval "$(direnv export bash)")` manually to "fix" a session.** Besides being unreliable (see above), it captures env var values — including secrets — as a string in Claude's tool output, landing in the conversation transcript. If `~/.bash_env` is doing its job, this should never be necessary.

## The Chezmoi Chain

```
dotfiles/dot_bash_env.tmpl          ← edit here
  └─ commit + push to GitHub
       └─ chezmoi update             ← pulls remote into ~/.local/share/chezmoi, then applies
            └─ renders → ~/.bash_env
                 └─ single `if [[ -z "$_BASH_ENV_GUARD" ]]; then ... fi` block containing:
                      ├─ PATH
                      ├─ machine-specific vars (from machine-cfg/bash_env, inlined at render time)
                      └─ direnv exec . env -0 resolution

machine-cfg/claude/settings.json
  └─ "env": { "BASH_ENV": "~/.bash_env" }
       └─ Claude Code injects BASH_ENV before every Bash tool call
```

**Important:** `~/Developer/dotfiles` is the working git repo. Chezmoi's actual source is `~/.local/share/chezmoi` — a separate copy. Editing files in `~/Developer/dotfiles` has no effect until committed, pushed, and pulled via `chezmoi update`. Then `chezmoi apply` (which `chezmoi update` also runs) renders templates from `~/.local/share/chezmoi`.

`machine-cfg/bash_env` provides machine-specific baseline vars (e.g. a default `GH_CONFIG_DIR`). These are inlined into `~/.bash_env` at chezmoi render time as static defaults — `.envrc` files in repos override them for that directory via the `direnv exec` resolution. Since `machine-cfg` is fork-specific (see below), this baseline file's content is independent per fork too.

**Debugging tip:** when chasing a `dot_bash_env.tmpl` bug, editing `~/.local/share/chezmoi/dot_bash_env.tmpl` directly + `chezmoi apply` is a fast iteration loop — no commit/push/update round-trip needed for local experiments. Sync the working version back to `~/Developer/dotfiles/dot_bash_env.tmpl` before committing for real.

## `.envrc` Files

`.envrc` files are **always git-ignored** — never committed to any repo. They are machine-local and must be created manually after cloning.

Each repo that needs non-default env vars has a `.envrc`. The most common pattern is switching the active GitHub account:

```bash
# a repo tied to your personal GitHub account
export GH_CONFIG_DIR="/Users/<you>/.config/gh-personal"

# a repo tied to your work GitHub account
export GH_CONFIG_DIR="/Users/<you>/.config/gh-work"
```

Other repos may use `.envrc` for Vault addresses, namespaces, or other infra-specific vars.

After creating or editing an `.envrc`, run `direnv allow` once to approve it.

## Settings.json Requirements (per machine-cfg fork)

`dot_bash_env.tmpl` is common (lives in dotfiles), but the settings that make Claude actually invoke it correctly live in **`machine-cfg/claude/settings.json`**, which is fork-specific — a work checkout and a personal checkout are separate repos with separate upstreams (see `dotfiles/CLAUDE.md`). Nothing here propagates automatically between forks; each one needs this configured independently. Checklist for any new/other fork:

1. **`env.BASH_ENV`** must point at the rendered file:
   ```json
   "env": { "BASH_ENV": "/Users/<you>/.bash_env" }
   ```
   Without this, Claude's Bash tool never sources `~/.bash_env` at all — no PATH additions, no machine-cfg baseline vars, no direnv resolution, and (critically) nothing in that file ever runs, so this specific symptom wouldn't be a "command not found" error — it'd simply be a bare, uncustomized shell with none of the expected env vars.

2. **`sandbox.excludedCommands`** should include `"direnv *"`:
   ```json
   "sandbox": {
     "excludedCommands": ["direnv *", "..."]
   }
   ```
   Intended to let direnv's own filesystem access bypass the sandbox's stricter allowlists. Note this only actually applies when the *literal* Bash tool command starts with `direnv` — it doesn't retroactively cover the nested `direnv exec` call inside `~/.bash_env`'s auto-sourcing for an unrelated top-level command. In practice this hasn't been a blocker for simple `.envrc` files (they just `export` a few vars), but keep it in mind if a fork's `.envrc` files do more than that.

3. **`permissions.deny`** should include:
   ```json
   "Bash(direnv export*)",
   "Bash(direnv exec*)",
   "Bash(direnv dump*)"
   ```
   This stops Claude from ever invoking direnv directly as a literal Bash tool call (which would dump resolved secrets into the transcript). It does not block `~/.bash_env`'s own internal `direnv exec` call, since that one runs via sourcing, not as a literal tool-call command string.

If direnv-backed env vars ("_direnv_resolve: command not found", a wrong `GH_CONFIG_DIR`, etc.) work on one fork/box but not another, check this checklist against that fork's `settings.json` before assuming the `dot_bash_env.tmpl` logic itself is at fault — a missing/incomplete `env.BASH_ENV` or a settings.json that predates one of these fixes is a more likely culprit than a logic bug in a file that's identical across forks.

## direnv.toml Whitelist

`~/.config/direnv/direnv.toml` can have a `[whitelist] prefix` entry restricting which directory trees direnv will even consider. If it's stale (e.g. references an old project root that no longer matches where your repos actually live), direnv silently behaves as if nothing is whitelisted for your real project directories — this doesn't block per-directory `direnv allow` from working, but it's worth checking if direnv seems to be ignoring an otherwise-correct `.envrc`.

## Verifying the Active Environment

```bash
echo $GH_CONFIG_DIR        # which gh account is active
echo $VAULT_ADDR            # which Vault cluster
direnv status               # whether .envrc is loaded and allowed for current dir
pwd                          # confirm this call's actual starting cwd
```

If a variable isn't what you expect: first confirm `pwd` is actually the repo you think it is (remember `cd other_dir && command` in one call won't pick it up — see above), then check `direnv status` for an allow problem, then check the settings.json checklist above, and only then suspect the `dot_bash_env.tmpl` logic itself.
