# dotfiles

Chezmoi-managed dotfiles repo. `chezmoi apply` renders templates and runs scripts to configure the system.

## Chezmoi Conventions

| Prefix/suffix | Effect |
|---|---|
| `dot_` | Maps to `~/.` — e.g. `dot_bash_profile.tmpl` → `~/.bash_profile` |
| `symlink_` | Creates a symlink — e.g. `symlink_settings.json.tmpl` → `~/.claude/settings.json -> ...` |
| `.tmpl` | Rendered as a Go template before writing |
| `run_everytime_` | Shell script run on every `chezmoi apply` |
| `run_onchange_` | Shell script run when its content changes |
| `run_once_` | Shell script run exactly once (tracked in chezmoi state) |

`machine-cfg` is a separate repo at `~/Developer/machine-cfg` with a different upstream per system type (work vs personal). Many dotfile templates pull in machine-cfg content at render time.
When that checkout is absent, `.chezmoiignore` skips its symlink targets, and the skills script still syncs dotfiles-owned skills to Claude and Codex but skips Copilot's missing target.

## Skill System

Claude Code skills live in two places depending on scope:

| Location | Scope | Examples |
|---|---|---|
| `dotfiles/skills/<name>/` | Common — applies on all systems (work and personal) | cross-environment tools |
| `machine-cfg/skills/<name>/` | Work- or personal-focused — machine-cfg has separate upstreams per system type | Adobe-specific, infra tools |

`run_everytime_skills.sh.tmpl` runs on every `chezmoi apply` and syncs both sources into `~/.claude/skills/`, `~/.codex/skills/`, and `~/.copilot/skills/` as symlinks. machine-cfg skills are applied first; dotfiles skills add to but cannot override them. Copilot projects only reviewed local skills that do not require credentials or unavailable hosted MCP connectors.

All cross-agent skill entry files must be named exactly `SKILL.md`, including case. Claude may tolerate other casing, but Copilot does not.

**Chezmoi source vs working repo:** `~/Developer/dotfiles` is the working git repo where edits are made. Chezmoi's actual source directory is `~/.local/share/chezmoi` — a separate copy. `chezmoi apply` reads from `~/.local/share/chezmoi`, not from `~/Developer/dotfiles` directly.

**To add or move a skill:** see the `manage-skills` skill for the correct commit/chezmoi-update/chezmoi-apply workflow — don't manually symlink into `~/.claude/skills/`.

## Git / gh Commands — Always cd First

**Always `cd` into a repo before running `git` or `gh` commands.** direnv loads `.envrc` based on the shell's working directory. `git -C <path>` and `gh -R <repo>` change where git/gh operate but do NOT change the shell's cwd — direnv never fires and the wrong `GH_CONFIG_DIR` (wrong GitHub account) is used.

```bash
# Wrong — direnv won't load dotfiles/.envrc
git -C ~/Developer/dotfiles push

# Correct
cd ~/Developer/dotfiles && git push
```

This applies to all repos that use `.envrc` for credential switching.

## Key Files

- `dot_bash_env.tmpl` → `~/.bash_env` — non-interactive shell env; used by Claude via `BASH_ENV` (must be set in `machine-cfg/claude/settings.json`'s `env.BASH_ENV`); resolves direnv via `direnv exec` (not `direnv export bash`, which is unreliable in cold non-interactive shells)
- `dot_bash_profile.tmpl` → `~/.bash_profile` — interactive shell setup; runs `direnv hook bash`. Only on Linux with `/tmp` mounted `noexec`, wraps `chezmoi` to use a private executable directory under `~/.cache/chezmoi/`; an explicit executable `TMPDIR` is respected. Other systems leave the command unchanged. For the first update before this profile is installed, set `TMPDIR` to an executable directory for that invocation.
- `dot_claude/modify_settings.json.tmpl` → `~/.claude/settings.json` — a real file (not a symlink) rebuilt on every apply from `machine-cfg/<hostname>/claude/settings.json` when that exists, else `machine-cfg/claude/settings.json`, keeping the live file's `model`, `enabledPlugins` and `extraKnownMarketplaces` (see Claude Settings Files)
- `dot_claude/symlink_settings.local.json.tmpl` → `~/.claude/settings.local.json` symlink into `machine-cfg/claude/`

## Claude Settings Files

`~/.claude/settings.local.json` is a **symlink** into machine-cfg. `~/.claude/settings.json` is a **real file**, rebuilt by `modify_settings.json.tmpl` on every `chezmoi apply` so that Claude's runtime writes (`/model`, plugin installs) stay local instead of churning a tracked file. The repo files it is built from:

- `~/Developer/machine-cfg/<hostname>/claude/settings.json` — this machine's Claude settings (permissions, sandbox, model, env), when machine-cfg has one for the host (`chezmoi execute-template '{{ .chezmoi.hostname }}'` gives the name). Machines sharing one file caused problems, so personal machines each get their own.
- `~/Developer/machine-cfg/claude/settings.json` — the shared fallback for hosts without one (e.g. work machines)
- `~/Developer/machine-cfg/claude/settings.local.json` — local overrides (symlink target of `~/.claude/settings.local.json`)

**Merge rules for `~/.claude/settings.json`:** the repo file wins for everything except three runtime-owned keys, which keep the live file's value when it has one (the repo value is only the default and the first-run seed): `model` (live replaces), and `enabledPlugins` / `extraKnownMarketplaces` (merged per entry, live wins, repo adds entries the live file lacks). Anything else changed in the live file, such as a permission added through the UI, is replaced at the next apply, so put it in the repo file. An empty, missing or invalid live file is re-seeded from the repo. If `jq` is missing the live file is left unchanged.

`chezmoi diff ~/.claude/settings.json` shows what the next apply would change. A newly added per-host file takes effect at the next `chezmoi apply`.

**Make lasting settings changes in the repo file** (the machine-cfg path above), commit and push machine-cfg from the terminal, then `chezmoi apply`. Editing `~/.claude/settings.json` directly only lasts for the runtime-owned keys.
