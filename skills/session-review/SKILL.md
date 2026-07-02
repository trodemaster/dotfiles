---
name: session-review
description: End-of-session optimization review — analyzes tool use, sandbox escapes, permission prompts, and skill usage from the current conversation, then proposes targeted changes to settings.json to reduce friction in future sessions. Use when asked to "review session", "optimize settings", "session review", "feedback loop", or "tune permissions".
allowed-tools: Read, Bash(cat *), Bash(ls *), Bash(grep *), Bash(find *), Bash(jq *), Bash(git *)
---

# Session Review Skill

Runs at the end of a Claude Code session to analyze what happened and propose targeted config changes that reduce permission prompts, improve sandbox efficiency, and avoid recurring errors in future sessions.

**Never apply changes without explicit confirmation. Always show a diff-style summary and ask.**

---

## Step 1 — Read Current Config

Locate the active settings files. Read in parallel:

- `~/.claude/settings.json` (user-level global config)
- `.claude/settings.json` in the current working directory, if it exists:

```bash
ls .claude/settings.json 2>/dev/null && echo "exists" || echo "no_project_settings"
```

If a project settings file exists, read it too.

---

## Step 2 — Analyze Conversation for Friction Points

Review the conversation history currently in context. For each category below, build a list of findings.

### 2a — Permission Prompts (Bash)

Look for any Bash commands that required user approval or were denied. These are candidates for the `sandbox.excludedCommands` list (runs outside sandbox) or `permissions.allow` (explicit allowlist).

**Hard rules — never propose these for auto-allow:**
- `sudo *`
- `rm -rf *`
- `git push --force*`, `git push -f *`
- `git reset --hard *`
- `git checkout -- *`
- `git clean -f *`
- `gh pr merge *`, `gh pr close *`, `gh release create *`
- `chmod 777 *`, `chmod -R *` under `sudo`
- Database DDL or irreversible migrations (`DROP *`, `TRUNCATE *`, etc.)
- Any command writing to system directories (`/etc/*`, `/usr/*`, `/System/*`)
- Any command exfiltrating credentials (`gh auth token`, `cat ~/.ssh/*`, `cat ~/.aws/*`)

**Safe patterns to consider for auto-allow:**
- Read-only system commands: `date *`, `which *`, `env`, `printenv *`, `uname *`
- Common universal dev tools, if used this session and not already allowed — propose narrowly, only the ones actually seen:
  - SCM: `git *`, `gh *`
  - Language toolchains: `go *`, `cargo *`, `node *`, `npm *`, `npx *`, `yarn *`, `bun *`, `python3 *`, `pip *`, `uv *`
  - Build: `make *`, `cmake *`, `ninja *`
  - Containers/packages: `docker *`, `podman *`, `brew *`, and the platform's native package manager
  - System inspection: `ps *`, `pgrep *`, `lsof *`, `ls *`, `find *`, `grep *`, `df *`, `du *`, `stat *`, `file *`
  - Text processing: `wc *`, `diff *`, `sort *`, `uniq *`, `jq *`, `yq *`, `curl *`
  - Process management: `kill *`, `pkill *`
  - Session tools: `tmux *`, `screen *`
  - Env/utils: `direnv *`, `open *`, `cp *`, `mv *`, `mkdir *`, `which *`
- Project-local file ops that stayed within the repo
- CLI tools that are already in `permissions.allow` but ran unsandboxed
- Tools already in `sandbox.excludedCommands` that work fine

### 2b — Permission Prompts (MCP Tools)

Look for MCP tool calls that required approval. These are candidates for `permissions.allow`.

Check the existing `permissions.allow` list — do not re-add tools already present. Look for new MCP tools used this session that aren't yet in the allowlist.

### 2c — Sandbox Escapes (dangerouslyDisableSandbox)

Note every use of `dangerouslyDisableSandbox: true`. For each:
- What was the command?
- Why did it need to escape? (network access, write to path outside sandbox, etc.)
- Is there a more targeted fix? Options:
  - Add the path to `sandbox.filesystem.allowWrite`
  - Add the command to `sandbox.excludedCommands`
  - Add the network host to `sandbox.network` allowlist
  - Accept the escape was correct (destructive / sensitive operation)

### 2d — Skill Usage Patterns

Look for skills invoked during the session. Note:
- Skills invoked manually that could have had clearer trigger descriptions
- Skills that were invoked but didn't match well (wrong skill used for the task)
- Skills that ran correctly — confirm their `allowed-tools` frontmatter covered all the tools they actually used

### 2e — Tool Efficiency

Look for patterns of unnecessary overhead:
- Sequential tool calls that could have been parallel (Read + Read, Bash + Bash with no dependency)
- Repeated reads of the same file
- Bash commands that failed and were retried with minor changes (argument errors, missing flags)
- Use of Bash for tasks that have a dedicated tool (cat instead of Read, echo > instead of Write)

### 2f — Filesystem Access Patterns

Look for write operations to paths that failed due to sandbox restrictions. These are candidates for `sandbox.filesystem.allowWrite`. Only propose paths that are:
- Within the user's home directory
- Project-local working directories
- Not system paths (`/etc`, `/usr`, `/System`, `/Library`)

**Principle of least privilege — secrets carve-out (mandatory for every `allowWrite` addition):**

Don't reject an otherwise-legitimate `allowWrite` path just because it *might* contain secrets somewhere inside it — and don't grant it blindly either. Scan the directory first:

```bash
find <path> -maxdepth 4 \( \
  -iname "*.pem" -o -iname "*.key" -o -iname "id_rsa*" -o -iname "id_ed25519*" \
  -o -iname ".env" -o -iname ".env.*" -o -iname ".netrc" -o -iname ".npmrc" \
  -o -iname "credentials*" -o -iname "*.credentials" -o -iname ".git-credentials" \
  -o -iname "*token*" -o -iname "*secret*" -o -iname "*.p12" -o -iname "*.pfx" \
  -o -ipath "*/vault/certs/*" -o -ipath "*/vault/data/*" -o -ipath "*/.ssh/*" \
  -o -ipath "*/.aws/credentials" -o -ipath "*/.gnupg/*" \
\) 2>/dev/null
```

- **Nothing found** — propose the `allowWrite` addition as-is.
- **Something found** — still propose the `allowWrite` addition (the directory has a legitimate reason to need writes, e.g. it's a git repo the user pushes from), but pair it with a `sandbox.filesystem.denyWrite` (and consider `denyRead`) entry for each specific secret-bearing path found, scoped as narrowly as possible. Precedent: `~/.lima` is in `allowWrite` for hostagent's routine file churn, with `~/.lima/_config/user` (Lima's own SSH private key) carved out via `denyWrite` — the directory isn't blocked wholesale, only the one file that actually holds a secret.
- Always show both halves of the pair together in the proposal (Step 4) — never propose `allowWrite` without having run this scan.

### 2g — Other Session Optimizations (report-only, no settings.json change)

These don't map to a specific settings.json key — surface them as suggestions in the EFFICIENCY NOTES section of the proposal (Step 4), not as line-item config changes.

- **Stop hook for long tasks** — if this session had long-running steps, suggest a `Stop` hook so the user gets notified when Claude finishes without watching the terminal. Example (macOS):
  ```json
  "hooks": {
    "Stop": [{ "type": "command", "command": "afplay /System/Library/Sounds/Glass.aiff" }]
  }
  ```
- **Custom slash commands** — if the same multi-step workflow was spelled out in the prompt more than once this session (or across recent sessions), suggest capturing it as a project-local command under `.claude/commands/`.
- **Effort level tuning** — if this was routine, low-stakes work (note management, status checks) under a global `effortLevel: "high"`, suggest a project-local `effortLevel: "medium"` override; leave high effort for genuine engineering work.
- **Context hygiene** — note if the session would have benefited from `/clear` at a task-domain switch, from putting stable content (file paths, templates) before variable content to hit the prompt cache, or if CLAUDE.md looked stale relative to the current codebase.
- **Model selection** — if the session's tasks were routine/low-complexity, note that a cheaper model (e.g. via `/config` → Model) would likely have been sufficient, reserving higher-tier models for engineering work.

---

## Step 3 — Classify Each Finding

For each finding, classify it as:

**Global** — relevant across all Claude Code sessions, not specific to this project:
- Goes into `~/.claude/settings.json`
- Examples: common CLI tools (date, which, env), widely-used MCP tools, system-level commands

**Project-specific** — only relevant when working in this repo:
- Goes into `.claude/settings.json` (create if it doesn't exist)
- Examples: project-specific paths, repo-local scripts, dev server commands

**Discard** — not safe or not worth adding:
- Anything matching the hard-rule blocklist in Step 2a
- One-off commands unlikely to recur
- Commands that escaped correctly (sensitive operations)

---

## Step 4 — Build Proposed Changes

Summarize findings into a structured proposal. Format:

```
SESSION REVIEW — Proposed Changes
==================================

GLOBAL CONFIG (~/.claude/settings.json)
----------------------------------------
permissions.allow additions:
  + "Bash(date *)"
  + "mcp__some_tool__read"

sandbox.excludedCommands additions:
  + "some-cli *"

sandbox.filesystem.allowWrite additions:
  + "~/some/path"

PROJECT CONFIG (.claude/settings.json)
--------------------------------------
permissions.allow additions:
  + "Bash(./scripts/build.sh *)"

SKIPPED / NOT PROPOSED
-----------------------
- "sudo npm install" — sudo is on the blocklist
- "rm -rf node_modules" — destructive, one-off

EFFICIENCY NOTES (no config change needed)
-----------------------------------------
- Steps 3 and 4 read the same file twice; could be parallelized or cached
- Used Bash(cat ...) where Read tool would have been cleaner
```

---

## Step 5 — Ask for Confirmation

Present the proposal and ask:

> Here are the proposed changes based on this session. Review each section:
> - **Accept all** — apply everything as proposed
> - **Accept global only** — skip project-level changes
> - **Review line by line** — go through each addition individually
> - **Skip** — no changes this time
>
> What would you like to do?

Do not proceed until the user responds.

---

## Step 6 — Apply Confirmed Changes

For each confirmed change, use the **update-config skill** to apply the change. Do not edit settings.json files directly with Edit/Write — delegate to update-config so it handles merging correctly.

If the project `.claude/settings.json` does not exist and project-specific changes were confirmed, note this and ask:
> No `.claude/settings.json` exists for this project. Create it?

If confirmed, create a minimal valid file before delegating to update-config:
```json
{
  "permissions": {
    "allow": []
  }
}
```

After applying, re-read both settings files and confirm the changes are present.

---

## Step 7 — Git Hygiene: Commit and Push All Changes

**Goal: no work lost to a dead laptop; every session resumable from any machine.**

### 7a — Identify repos with uncommitted changes

Check all repos touched during this session. At minimum check the current working directory:

```bash
git -C . status --short
```

Also check any other repos that had changes applied during the session (e.g. config repos, dotfiles repos). Run each check in parallel if there are multiple.

If a repo shows output it has uncommitted changes. Skip repos that are clean.

Also check for untracked files that look intentional (new skill files, scripts, configs) — stage them too if they belong to the session's work.

### 7b — Determine branch strategy for each dirty repo

For each repo with changes:

```bash
git -C <repo> branch --show-current
git -C <repo> remote -v
```

Apply these rules:

| Current branch | Has `upstream` remote? | Action |
|---|---|---|
| Not `main`/`master` | Either | Commit and push to current branch |
| `main`/`master` | Yes (fork) | Pushing main is safe — commit and push |
| `main`/`master` | No (not a fork) | **Create a WIP branch first**, then commit and push |

To create a WIP branch:
```bash
git -C <repo> checkout -b wip/session-$(date +%Y-%m-%d)
```

### 7c — Stage, commit, and push

For each repo, stage specific files rather than `git add -A` to avoid accidentally capturing sensitive files (`.env`, credentials):

```bash
git -C <repo> add <specific files changed this session>
git -C <repo> status   # confirm staged set before committing
```

Commit with a message that describes the session work. Examples:
- `wip: <brief description of work done>`
- `session-review: sandbox and permissions updates`
- `add session-review skill`

Then push:
```bash
git -C <repo> commit -m "$(cat <<'EOF'
<message>
EOF
)"
git -C <repo> push -u origin HEAD
```

### 7d — Confirm and report

After pushing each repo, verify:
```bash
git -C <repo> log --oneline -3
```

Report back: repo path, branch pushed, commit SHA. If any push fails, say so and explain why rather than silently skipping.

---

## Safety Checklist

Before proposing any change, verify it does not match any of:
- [ ] `sudo *`
- [ ] Credential access (`gh auth token`, `cat ~/.ssh/*`, `cat ~/.aws/*`, `cat ~/.gnupg/*`)
- [ ] Destructive file ops (`rm -rf *`, `git reset --hard *`, `git push --force*`)
- [ ] System directory writes (`/etc/*`, `/usr/*`, `/System/*`, `/Library/*`)
- [ ] Network access to non-work hosts (check against existing sandbox.network allowlist)

If any proposed change would match, remove it and note it in the SKIPPED section.

Separately, for every `sandbox.filesystem.allowWrite` addition specifically:
- [ ] Ran the secrets scan from Step 2f against the proposed path
- [ ] Any secret-bearing file found has a matching `denyWrite` (and considered `denyRead`) entry in the same proposal
