---
name: manage-skills
description: Correct workflow for adding or moving a Claude Code skill in the dotfiles/machine-cfg system (chezmoi-managed). Use when asked to "add a skill", "create a new skill", "move a skill", or when a new skill needs to end up in ~/.claude/skills/. Covers the commit/chezmoi-update/chezmoi-apply sequence and why manually symlinking into ~/.claude/skills/ is wrong.
---

# Adding or moving a skill

See the dotfiles `CLAUDE.md` "Skill System" section for background on where skills live (`dotfiles/skills/<name>/` for cross-environment tools, `machine-cfg/skills/<name>/` for work- or personal-specific ones) and how `run_everytime_skills.sh.tmpl` syncs them.

**Correct workflow:**
1. Create/move the skill directory (`SKILL.md` inside) in the appropriate source repo.
2. **Commit and push** from `~/Developer/dotfiles` (as two separate commands — see the dotfiles `CLAUDE.md` git section).
3. Run `chezmoi update` — pulls from the remote and refreshes `~/.local/share/chezmoi`.
4. Run `chezmoi apply` — `run_everytime_skills.sh.tmpl` creates the symlink in `~/.claude/skills/`.

Do not manually create symlinks in `~/.claude/skills/` — chezmoi owns that directory.
