# TUI Reference

Complete reference for agent-desk Terminal UI features.

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Collapse group / go to parent |
| `l` / `→` / `Tab` | Toggle expand/collapse group |
| `1-9` | Jump to Nth root group |

### Session Actions

| Key | Action |
|-----|--------|
| `Enter` | Attach to session OR toggle group |
| `n` | New session (inherits current group) |
| `r` | Rename session or group |
| `R` | Restart session (reloads MCPs) |
| `K` / `J` | Move item up/down in order |
| `m` | Move session to different group |
| `M` | Sweep idle sessions into an "Inactive" group |
| `d` | Delete session or group |
| `u` | Cycle marker: unread → parked red → parked blue → normal |
| `s` | Snooze: park until a time, then wake as unread |
| `o` | Sort within groups: manual → recent → busy |

### Group Actions

| Key | Action |
|-----|--------|
| `g` | Create group (subgroup if on group) |
| `e` | Rename group (alias for `r`) |

### Search & Filter

| Key | Action |
|-----|--------|
| `/` | Local search (fuzzy) |
| `G` | Global search (all Claude conversations) |
| `Tab` | Switch between local/global search |
| `0` | Clear filter (show all) |
| `!` | Filter: running only (toggle) |
| `@` | Filter: waiting only (toggle) |
| `#` | Filter: idle only (toggle) |
| `$` | Filter: error only (toggle) |

### Global

| Key | Action |
|-----|--------|
| `?` | Help overlay |
| `i` | Import existing tmux sessions |
| `Ctrl+R` | Manual refresh |
| `Ctrl+Q` | Detach (keep tmux running) |
| `q` / `Ctrl+C` | Quit |

## Status Indicators

| Symbol | Status | Color | Meaning |
|--------|--------|-------|---------|
| `●` | Running | Green | Active, content changed in last 2s |
| `◐` | Waiting | Yellow | Stopped, unacknowledged |
| `○` | Idle | Gray | Stopped, acknowledged |
| `✕` | Error | Red | tmux session doesn't exist |
| `⟳` | Starting | Yellow | Session launching |

## Dialogs

### New Session (`n`)

**Fields:**
- Session name (required)
- Project path (required, supports `~/`)
- Command (claude/gemini/opencode/codex/custom)
- Parent group (auto-selected)

**Controls:** `Tab` move fields | `Enter` create | `Esc` cancel

### Sweep idle sessions (`M`)

Moves every idle session into an "Inactive" group in one keypress, creating the
group on first use. Sessions flagged unread (green — "waiting to be read") and
sub-sessions are left where they are. A banner reports how many moved; `Esc`
dismisses it, or it clears itself after 30 seconds.

This is a one-shot sweep, not a live group-by-status view — sessions you restart
out of "Inactive" stay there until you move them or sweep again.

### Sort within groups (`o`)

Cycles manual → recent → busy. Recent orders the sessions inside each group by
last transcript activity, newest first; busy by turns in the last 24 hours.
Groups, nesting and the hand-built order are untouched, so manual brings back
exactly what you had. `Shift+Left`/`Shift+Right` while attached walk the list,
so they follow the sort. The header shows a `↓ recent` / `↓ busy` pill while a
sort is on. Each row ends in three right-aligned columns: model family as one
letter (`o` Opus, `f` Fable, `s` Sonnet; +⚡ in auto mode), turns in the last 24 hours (`·` when none), and age, three characters at most (`40m`, `5h`,
`3d`, `2w`, `3w+`); a snoozed row shows its wake time across all three.
Sessions idle two days or more render faded.

### Snooze a session (`s`)

Parks the session in place until a wake time: it stays in its group, dimmed,
sunk below that group's live sessions (soonest wake first), then brightens as
the unread bookmark (green, "come back to this") in its usual spot. The row shows ⏰ when it wakes today
and 📅 when later, with the wake time after the name (`7p`, `Mon 7a`,
`10/12 9a`). The picker opens on a text field: type a time (`30m`, `2h`,
`7am`, `7:30pm`, `tomorrow 9am`, `mon`, `fri 9am`, `10/12`, `10/12 11:00`;
day-only entries wake at 9am) and the resolved time shows underneath. Below
it, the week's landmarks: Tonight (8pm), Later this week (Thu on Mon/Tue,
Fri on Wed/Thu), This weekend (Sat 9am; Sun on a Saturday), Next Monday,
In 2 weeks, In a month, whichever are still ahead. Typing narrows them;
`↓` picks one. Enter takes the first matching suggestion while any match,
else the typed time. Every pick is counted by row (typed times as "typed");
`agent-desk snooze` prints the tally, the evidence for reshaping the list.

While snoozed the session is skipped by `Shift+Left`/`Shift+Right` like red
parking. `u` or attaching cancels the snooze.

> **Note:** `M` opened an MCP Manager dialog in earlier versions. MCPs are now
> managed from the CLI (`agent-desk mcp list|attach|detach`) — see the CLI
> reference.

### Delete Confirmation (`d`)

**For sessions:** Warning about tmux kill, process termination

**For groups:** Sessions move to default (not deleted)

**Controls:** `y` confirm | `n`/`Esc` cancel

## Search

### Local Search (`/`)

- Fuzzy search session titles and groups
- Max 10 results
- `↑/↓` or `Ctrl+K/J` navigate
- `Enter` select | `Tab` switch to global | `Esc` close

### Global Search (`G`)

- Full content search across `~/.claude/projects/`
- Regex + fuzzy matching
- Recency ranking
- Split view: results + preview
- `[/]` scroll preview
- `Enter` create/jump to session

**Config:**
```toml
[global_search]
enabled = true
recent_days = 30
```

## Preview Pane

- Shows last ~500 lines of session's tmux pane
- Auto-updates every 2 seconds
- Launch animation: 6-15s for Claude/Gemini

## Layout

- **< 50 cols:** List only
- **50-79 cols:** Stacked (list above preview)
- **80+ cols:** Side-by-side (default)

## Tool Icons

| Tool | Icon | Color |
|------|------|-------|
| Claude | 🤖 | Orange |
| Gemini | ✨ | Purple |
| OpenCode | 🌐 | Cyan |
| Codex | 💻 | Cyan |
| Cursor | 📝 | Blue |
| Shell | 🐚 | Default |

## Color Scheme (Tokyo Night)

| Element | Color |
|---------|-------|
| Accent (selection) | #7aa2f7 |
| Running | #9ece6a |
| Waiting | #e0af68 |
| Error | #f7768e |
| Groups | #7dcfff |
| Background | #1a1b26 |
| Surface | #24283b |

## Hidden Features

- **`Ctrl+K/J`:** Vim-style navigation in search
- **Numbers 1-9:** Jump to root groups instantly
- **Status filters are toggles:** Press again to turn off
