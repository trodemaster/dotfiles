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

## Skill System

Claude Code skills live in two places depending on scope:

| Location | Scope | Examples |
|---|---|---|
| `dotfiles/skills/<name>/` | Common — applies on all systems (work and personal) | cross-environment tools |
| `machine-cfg/skills/<name>/` | Work- or personal-focused — machine-cfg has separate upstreams per system type | Adobe-specific, infra tools |

`run_everytime_skills.sh.tmpl` runs on every `chezmoi apply` and syncs both sources into `~/.claude/skills/` and `~/.codex/skills/` as symlinks. machine-cfg skills are applied first; dotfiles skills add to but cannot override them.

**Chezmoi source vs working repo:** `~/Developer/dotfiles` is the working git repo where edits are made. Chezmoi's actual source directory is `~/.local/share/chezmoi` — a separate copy. `chezmoi apply` reads from `~/.local/share/chezmoi`, not from `~/Developer/dotfiles` directly.

**To add or move a skill — correct workflow:**
1. Create/move the skill directory (`SKILL.md` inside) in the appropriate source repo
2. **Commit and push** from `~/Developer/dotfiles` (as two separate commands — see git section below)
3. Run `chezmoi update` — pulls from the remote and refreshes `~/.local/share/chezmoi`
4. Run `chezmoi apply` — `run_everytime_skills.sh.tmpl` creates the symlink in `~/.claude/skills/`

Do not manually create symlinks in `~/.claude/skills/` — chezmoi owns that directory.

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

- `dot_bash_env.tmpl` → `~/.bash_env` — non-interactive shell env; used by Claude via `BASH_ENV`; runs `direnv export bash`
- `dot_bash_profile.tmpl` → `~/.bash_profile` — interactive shell setup; runs `direnv hook bash`
- `dot_claude/symlink_settings.json.tmpl` → `~/.claude/settings.json` symlink into `machine-cfg/claude/`
- `dot_claude/symlink_settings.local.json.tmpl` → `~/.claude/settings.local.json` symlink into `machine-cfg/claude/`

## Claude Settings Files

`~/.claude/settings.json` and `~/.claude/settings.local.json` are **symlinks** managed by chezmoi. The real files are:

- `~/Developer/machine-cfg/claude/settings.json` — global Claude settings (permissions, sandbox, model, env)
- `~/Developer/machine-cfg/claude/settings.local.json` — local overrides

**Always edit the real paths directly** — the Edit tool refuses to write through symlinks. After editing, commit and push machine-cfg from the terminal.
