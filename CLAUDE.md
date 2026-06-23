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

**To add or move a skill — correct workflow:**
1. Create/move the skill directory (`SKILL.md` inside) in the appropriate source repo
2. **Commit and push that repo** — chezmoi's source is the git state; applying before committing leaves the repos out of sync and won't propagate to other machines via `chezmoi update`
3. Run `chezmoi apply` — `run_everytime_skills.sh.tmpl` creates the symlink in `~/.claude/skills/`

Do not manually create symlinks in `~/.claude/skills/` — chezmoi owns that directory. If a symlink was created manually, run `chezmoi apply` after committing to have chezmoi take over management of it.

## Key Files

- `dot_bash_env.tmpl` → `~/.bash_env` — non-interactive shell env; used by Claude via `BASH_ENV`; runs `direnv export bash`
- `dot_bash_profile.tmpl` → `~/.bash_profile` — interactive shell setup; runs `direnv hook bash`
- `dot_claude/symlink_settings.json.tmpl` → `~/.claude/settings.json` symlink into `machine-cfg/claude/`
- `dot_claude/symlink_settings.local.json.tmpl` → `~/.claude/settings.local.json` symlink into `machine-cfg/claude/`
