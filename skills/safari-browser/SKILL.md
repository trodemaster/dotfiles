---
name: safari-browser
description: General-purpose browser automation via the safari-mcp-stp MCP server (Safari Technology Preview) — navigate URLs, read page content, click/type/fill forms, take screenshots, and debug page loads via network requests and console messages. Reach for this any time an actual browser would settle the question, not just when the user says "browser": checking whether a URL or link is reachable/valid, verifying a UI or frontend change actually renders and works before reporting it done, reproducing a bug the user describes seeing on a page, reading/scraping a live webpage, filling out or submitting a form, or debugging a blank page, failed request, redirect loop, or JS console error.
---

# Safari Browser Skill

Drives Safari (via the `safari-mcp-stp` MCP server / WebKit Inspector) for general browser automation and for debugging why a page isn't loading or behaving right. Every claim and code path below was exercised directly against live pages (example.com, httpbin.org, a nonexistent domain) in this session — not inferred from docs.

## Tool inventory

| Tool | Purpose |
|---|---|
| `list_tabs` / `create_tab` / `switch_tab` / `close_tab` | Tab lifecycle |
| `navigate_to_url` | Load a URL in a tab (auto-creates one if none open) |
| `page_info` | Current URL, title, load state — the fast way to sanity-check where you are |
| `get_page_content` | **Preferred** way to read a page — WebKit text extraction, returns `uid=N` node references |
| `page_interactions` | Batched click/type/scroll/hover/selectMenuItem/keyPress, by `node` uid or `text` find-in-page |
| `evaluate_javascript` | Run JS; use `$uid(N)` to reference nodes from `get_page_content`. Last resort for reading — prefer `get_page_content` |
| `screenshot` | PNG saved to disk (path returned, never inline base64) |
| `set_viewport_size` / `set_emulated_media` | Window size / CSS media emulation |
| `list_network_requests` / `get_network_request` | Request list (filterable by method/status/URL substring) and full detail (headers, body, timing) for one request |
| `browser_console_messages` | Buffered console logs (`debug`/`info`/`warn`/`error`), **drains the buffer by default** |
| `browser_dialogs` | List/accept/dismiss JS `alert`/`confirm`/`prompt` |
| `wait_for_navigation` | Block until in-flight navigation finishes |

## If the MCP tools aren't responding

`safari-mcp-stp` needs Safari Technology Preview actually running to have anything to attach to. If a call to one of the tools above errors out or times out as if the server were offline, launch the app with the Bash tool, then retry the same MCP call:

```bash
open "/Applications/Safari Technology Preview.app"
```

The path contains a space, so it must be double-quoted (not backslash-escaped) — this is a plain `open` invocation, not an MCP tool call. Give the app a moment to finish launching before retrying; if the first retry still fails, the MCP server itself (not just the app) may need attention — say so rather than silently giving up.

## Standard workflow

```
create_tab(url)                      # or navigate_to_url if a tab already exists
set_viewport_size(1440, 900)         # see "Viewport" below — do this early, once
get_page_content()                   # read; note the uid=N on elements you need
page_interactions([...])             # click/type using those uids, batched
get_page_content()                   # re-read if the diff wasn't enough context
close_tab(handle)                    # clean up when done
```

Reach for `evaluate_javascript` or `screenshot` only when text extraction can't express what you need (a computed value, visual layout, a chart). `get_page_content` is cheaper and more reliable for reading.

## Viewport — set it early

The default viewport is cramped (measured **1024×641** via `window.innerWidth/innerHeight` on a fresh tab) — small enough that responsive layouts render in a mobile/tablet breakpoint you probably don't want. Call `set_viewport_size` once, right after opening the first tab:

```
set_viewport_size(width: 1440, height: 900)
```

This is a **browser-window-level** setting, not per-tab — verified by creating a second tab after setting it and finding the new tab already at the wider width. Set it once per session, before your first real interaction. Note the effective `innerHeight` will be somewhat less than what you request (browser chrome eats ~50–130px in testing) — that's normal, not a bug.

## Debugging a URL that won't load

This is the part that isn't obvious from the tool descriptions: **`navigate_to_url` does not error or throw on a failed load.** DNS failures, unreachable hosts, etc. all "succeed" as a tool call and silently return Safari's own error page as content. Confirmed by navigating to a nonexistent domain:

```
navigate_to_url("https://this-domain-does-not-exist-abc123xyz.invalid")
→ {"content": "root\n\t'Safari Can't Find the Server'\n\t...", "url": "safari-resource:/ErrorPage.html", "title": ""}
```

**Detect a failed load with `page_info`, not the navigation result.** `page_info` after a failed navigation reliably reports:

```
{"title": "Failed to open page", "url": "https://this-domain-does-not-exist-abc123xyz.invalid/"}
```

`page_info.title === "Failed to open page"` is the check. Don't match on URL — `get_page_content`'s reported `url` flips to the internal `safari-resource:/ErrorPage.html`, but `page_info`'s `url` field usefully keeps the original requested address.

For loads that "succeed" but render wrong (blank page, error banner, wrong content), work through these in order:

1. **`page_info`** — confirm you're actually on the URL/title you expect.
2. **`list_network_requests`** (optionally `filter: {status_min: 400}`) — find failed/redirected requests. Each entry has method, status, mime type, timing.
3. **`get_network_request(request_id)`** — full headers + response body for one request from that list. This is how you see an API's actual error payload, not just its status code.
4. **`browser_console_messages`** — JS errors and browser-logged resource failures (e.g. resource 404s auto-log as `"Failed to load resource: the server responded with a status of 404 ()"` at `level: "error"`, no explicit `console.error` call needed).
5. **`screenshot`** — only once the above hasn't explained it; confirms *what it looks like*, not *why*.

## Gotchas

- **`browser_console_messages` clears on read by default.** `clear` defaults to `true`, so a call drains the buffer — a second call (even with a different `level_filter`) right after will come back empty. If you need to inspect the same batch multiple ways, pass `clear: false` on the read(s) you don't want to be final, and let the last one clear.
- **`list_network_requests` misses the navigation triggered by `create_tab`'s `url` argument.** Requests only start accumulating partway through that first load — tested empty (`count: 0`) immediately after `create_tab(url: ...)` even though the page rendered fully. If you need the *full* request list for a page's initial load, `create_tab()` bare, then `navigate_to_url()` to it — the buffer captures navigations that happen after the tab exists.
- **`get_page_content`'s node `uid`s are call-scoped** — they come from the most recent `get_page_content`/`page_interactions` response. Re-fetch content before reusing a uid from an older read.
- **Screenshots are files, not inline data.** The tool returns a path; `Read` that path to actually view it.
- **`page_interactions` batches** — pass the whole sequence of clicks/types in one call rather than one call per step; it runs them with a 400ms settle between each and returns a single diff (or full text if >50% of the page changed).
