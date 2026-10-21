# Changelog

All notable changes to this fork of Agent Desk are documented here.

This is a [Millwright Software](https://github.com/millwright-software) fork of
[`asheshgoplani/agent-deck`](https://github.com/asheshgoplani/agent-deck). The pre-fork release history lives
upstream; this file tracks changes made in the fork. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Releases are tagged and published at
[millwright-software/agent-desk/releases](https://github.com/millwright-software/agent-desk/releases);
the in-app updater downloads the `agent-desk_<version>_<os>_<arch>.tar.gz` asset attached to the latest one.

## [Unreleased]

### Changed
- **Session rows: right-aligned columns.** Model family as one letter (`o` Opus, `f` Fable, `s` Sonnet,
  `h` Haiku, `g` Gemini; versions stay in the preview) with ⚡ for auto mode, turns in the last 24 hours (`·` when
  none) and age now sit in fixed columns at the right edge of the list instead of trailing the title, so
  they read straight down. Age is three characters at most (`40m`, `5h`, `3d`, `2w`, then `3w+`); the exact
  time stays in the preview. A snoozed row shows its wake time across all three columns. Long titles are truncated
  with `…` to keep the columns in place; a worktree branch badge drops first when space is short.
- **Snooze in place.** A snoozed session now stays in its own group instead of moving to a global "Snoozed"
  group: dimmed, with the ⏰/📅 marker and wake time, sunk below that group's live sessions (soonest wake
  first, under any sort), and skipped by the Shift rings as before. On wake it brightens as unread in its
  manual place. Sessions left in the old Snoozed group go back to their home group on first launch, still
  snoozed, and the group is removed.

## [1.0.22] - 2026-10-16

### Added
- **Sort within groups (`o`).** Cycles manual → recent → busy. Recent orders each group's sessions by last
  transcript activity, newest first; busy by turns in the last 24 hours, most first. Groups, nesting and the
  hand-built order are untouched, so manual restores exactly what you had. The attached `Shift+Left`/
  `Shift+Right` ring walks the list, so it follows the sort. A pill in the header shows the mode when it is
  not manual; it persists with the other view state.
- **Activity on the row.** A `12t` badge is turns in the last 24 hours (assistant turns, from a bounded tail
  read of the transcript, 4 MB max, recomputed when the file grows or every ten minutes). Sessions idle for
  two days or more render faded, so a long list reads at a glance.

### Changed
- **Snooze picker: type first, landmarks below.** The text field comes first with focus, so typing starts
  immediately and the resolved time shows underneath as you type. The fixed presets (1h, 3h, …) are
  replaced by the week's landmarks: Tonight (8pm), Later this week (Thu on Mon/Tue, Fri on Wed/Thu), This
  weekend (Sat 9am; Sun on a Saturday), Next Monday, In 2 weeks, In a month, whichever are still at least
  30 min ahead. Typing narrows them; `↓` walks them. Enter takes the first matching suggestion while any
  match, else the typed time. Day-only entries (`tomorrow`, `mon`, `10/12`) now wake at 9am (was 7am).
- **`agent-desk snooze` prints the pick tally.** Every snooze counts its row (typed times as "typed") in the
  state database, so the ladder can be reshaped from what actually gets used.

## [1.0.21] - 2026-10-03

### Added
- **Snooze a session (`s`).** Moves it into a "Snoozed" group pinned at the bottom of the list until a
  wake time, then returns it to the group it came from as the unread bookmark (green, "come back to
  this"). The row shows ⏰ when it wakes today and 📅 when later, a dim title, and the wake time (`7p`,
  `Mon 7a`, `10/12 7a`); the two-cell emoji drops the separator after it so titles stay aligned. The
  picker offers 1 hour, 3 hours, tonight, tomorrow morning, Monday morning, next week, or a typed time
  (`30m`, `2h`, `7am`, `7:30pm`, `tomorrow 9am`, `mon`, `fri 9am`, `10/12`, `10/12 11:00`; day-only
  entries wake at 7am). Attached `Shift+Left`/`Shift+Right` skip a snoozed session like red parking.
  `u` or attaching cancels the snooze and sends it home. The Snoozed group is removed when its last
  session leaves; a snoozed session moved out of it by hand stays put on wake. Wake time and home
  group persist in `tool_data` (`snooze_until`, `snooze_from_group`), so a snooze survives restarts;
  the home tick checks every 2 s. Fifth flag value `FlagSnoozed` (appended, stable order kept).

## [1.0.20] - 2026-10-03

### Fixed
- **Acknowledged ("you've looked at this") has one owner.** It was a flag inside the tmux status engine with
  a SQLite mirror, and the mirror was wrong in both directions: the column was only ever written as `true`, every
  full save (`INSERT OR REPLACE`, column not in the row) silently reset it to `0`, and every status tick
  re-applied a `true` from the column onto memory. Visible result: a session that had been acknowledged once
  could go straight to gray after its next finished turn instead of orange, with no bell. Now the instance owns
  the flag: `Acknowledge` / `ResetAcknowledged` set it, write the column (both values, only on change), and
  inform the engine; the column rides along in the row so saves preserve it; other TUIs' changes are adopted in
  both directions. The engine keeps a copy because pattern-based detection (Gemini, Copilot, stale hooks) is
  the only thing that can see "new output since you looked"; its verdict is adopted and persisted after each
  pattern read. Second slice of consolidating status into table lookups, after the hook inbox in 1.0.19.
  On first run with this version, sessions saved as idle count as acknowledged so nothing turns orange just
  because the column had never been written.

## [1.0.19] - 2026-10-03

### Changed
- **Claude Code hook events land in one SQLite table instead of one JSON file per session.** The hook handler
  upserts a row in `~/.agent-desk/hooks.db` (instance, status, event, Claude session ID, time); the TUI reads
  the table once per status tick. Gone: the `~/.agent-desk/hooks/` file-per-instance side channel, the fsnotify
  watcher, and its debounce. Existing files are imported once at startup (newest wins) and left in place;
  `agent-desk hooks status` reports both. Rows for instances deleted more than a day ago are pruned hourly.
  First slice of consolidating status into table lookups; the acknowledged flag is next.
  ⚠️ The inbox is its own file, not a table in the profile's `state.db`: the handler only knows the instance
  ID, not the profile, and instance IDs are globally unique. It also keeps hook writers off `state.db`'s lock.
  Contention was measured with the same driver before committing to this: 30 concurrent handler processes,
  zero lock errors, typical write 0.5 ms, worst 10 ms; 100 processes, worst 98 ms. The hook is async in Claude
  Code, so none of that is on your path. If the table cannot be written the handler falls back to the old file.
  A TUI still running from before this release keeps reading files and will go stale for sessions started
  after the upgrade: restart it.

## [1.0.18] - 2026-10-03

### Fixed
- **A session sitting on a question or permission prompt is "waiting" even if you had opened it before.**
  `Shift+Up` / `Shift+Down` walked straight past a session with an AskUserQuestion dialog on screen, and the
  sidebar showed it gray. Cause: the acknowledgment rule ("you attached, so it is seen") was applied to every
  hook "waiting", and a prompt arrives as the same "waiting" as a finished turn. A prompt blocks Claude until you
  answer it, so it now stays waiting regardless of acknowledgment, and a new prompt clears the acknowledgment
  (once, not every tick) so the eventual Stop reads as unseen too. `PermissionRequest` and the matched
  `Notification` events count as prompts; `Stop` keeps the old rule.
- **`PostToolUse` hook added (status "running").** It is what ends a prompt: you approved or answered, the tool
  ran, Claude is going again. Without it nothing fired until Stop, so an answered question read as waiting for
  the rest of the turn. Existing installs pick it up at the next launch (hooks are re-installed when any are
  missing and you had accepted them) or with `agent-desk hooks install`.
- **`Shift+Up` / `Shift+Down` reach sessions inside collapsed groups.** A session waiting for you is still
  waiting when its group is folded. Landing on one expands the group so the cursor has a row to return to.
  `Shift+Left` / `Shift+Right` still follow the sidebar as shown.

## [1.0.17] - 2026-10-03

### Fixed
- **Resuming a session after the tmux server died no longer adopts a sibling instance's conversation.**
  Restart rescanned the project directory for the most recently modified transcript so a `/clear` inside a
  live pane would be picked up. With the server gone there were no live panes to exclude, so instances that
  share a project path each resumed whichever sibling's transcript had been touched last (three vault
  instances came back inside three other instances' conversations on 2026-10-03). The rescan now only runs
  while the pane still exists; with it gone, a stored session ID whose transcript has conversation data is
  kept. Empty or zombie IDs still rescan. Record: `documentation/worklog/2026-10-03-restart-stole-sibling-sessions.md`.

## [1.0.16] - 2026-10-02

### Changed
- **`Shift+Up` / `Shift+Down` with nothing waiting now stay in the session and flash a "Nothing waiting" card**
  instead of dropping you to the sidebar. The attach loop asks whether there is anywhere to go before it
  detaches; if not, the key is swallowed and the card shows for a second (or until the next keystroke), same
  mechanics as the landing card.

## [1.0.15] - 2026-10-02

### Added
- **`Shift+Down` / `Shift+Up` while attached jump to the next or previous session that is waiting for you**,
  in sidebar order, wrapping at the ends. "Waiting" is the green dot: Claude asked a question or wants a
  permission. The session you are leaving is usually not waiting any more (attaching acknowledged it), so the
  walk goes by its position in the list rather than its place in the ring: down really is the next one below,
  up the next one above. With nothing waiting you land back in the sidebar with a notice instead of the key
  looking like it did nothing. Parked-red sessions are skipped, as with `Shift+Left` / `Shift+Right`.
  Idea borrowed from Juggler's "jump to next session awaiting you" command (see `JUGGLER-PORT-CANDIDATES.md`).

### Changed
- `Shift+Right` / `Shift+Left` from a session that died while attached now continue from its position in the
  list instead of jumping to the first live session.

## [1.0.14] - 2026-09-18

### Changed
- **`Shift+Right` / `Shift+Left` skip sessions parked red with `u`.** Red parking means "shelved, skip this";
  landing on one while flipping through live work was exactly what the marker was set to avoid. Blue parking
  and the unread bookmark are not skips. Opening a parked session from the list still works, and clears the
  marker as before.

## [1.0.13] - 2026-09-11

### Changed
- **The switch card stays up for one second again.** Two seconds was a workaround for the early dismissal
  fixed in 1.0.12; with that gone, one second is enough to read the name and two felt like a wait.

## [1.0.12] - 2026-09-11

### Fixed
- **The switch card was still being dismissed almost at once, by terminal replies split across reads.** 1.0.10
  stopped treating a chunk that *starts* with ESC as a keystroke, but the proxy reads stdin 32 bytes at a time
  and the burst of replies to tmux's attach-time queries (device attributes, XTVERSION, colours) is longer than
  that, so the second read began mid-sequence with an ordinary byte and looked like typing. The log confirmed
  it: every close was exit 129, tmux's code for "closed by `display-popup -C`", which only our own dismiss sends.
  Typed-or-not is now decided by a small parser (`InputClassifier`) that keeps state across reads: it tracks
  CSI / OSC / DCS sequences to their terminator, so replies, focus events and mouse reports never count, while
  letters, control keys, arrows, function keys and Alt-combinations do — even when a key's own sequence is split.
  The card process uses the same parser for what tmux forwards to it. The proxy read buffer is 256 bytes now.
  Each dismissal is logged to `~/.agent-desk/debug.log` with the exact input that caused it
  (`attach_banner_dismissed`), alongside `attach_banner_open` / `attach_banner_closed` with timings.

## [1.0.11] - 2026-09-11

### Changed
- **The switch card stays up for two seconds** instead of one. One second read as a flash.
- Terminal focus-in / focus-out events no longer count as a keystroke to the card, so a window focus change
  during the two seconds does not dismiss it.

## [1.0.10] - 2026-09-11

### Changed
- **The switch card is bigger and draws the session title in block letters** (a 5-row block font, uppercase),
  with the group underneath. When the title is too wide for the letters at the current terminal size the card
  falls back to the one-line bold version from 1.0.9, so a long title never gets clipped.

### Fixed
- **The switch card no longer flashes and vanishes.** tmux queries the terminal on attach (device attributes,
  colours), and the replies land in the input stream tens of milliseconds later — after the 50ms window in
  which attach-time control sequences are discarded. They were being treated as the first keystroke, which
  dismissed the card almost as soon as it was drawn. Only plainly typed input dismisses it now; ESC-led input is
  left to the card process, which tells a reply from a key and re-sends the key to the session if it was one.

## [1.0.9] - 2026-09-11

### Added
- **A name card pops up for a second when you switch sessions with `Shift+Right` / `Shift+Left`**, so you know
  where you landed: the session title in bold, and its group (or project folder) underneath. It goes away on its
  own after a second, or the instant you press a key — the key still reaches the session. It only appears on a
  switch; opening a session from the list you already know what you picked.
  ⚠️ It is a tmux popup (`display-popup`), not something drawn into the proxied byte stream: tmux composites it
  over the pane and repaints what was underneath when it closes, which a hand-drawn overlay over a session that
  is still scrolling could not do. The popup runs the `agent-desk` binary itself (hidden `attach-banner` verb)
  with the text passed as environment, so a title with quotes or a `;` never meets a shell. Needs tmux 3.3+ for
  the rounded border; on older tmux the card is skipped and the switch works as before.

## [1.0.8] - 2026-09-09

### Added
- **Switch sessions from inside an attached session** with `Shift+Right` / `Shift+Left`. The ring follows the
  order shown in the sidebar — including group ordering you arranged yourself — skips sessions whose tmux
  session is gone, crosses group boundaries rather than stopping at them, and wraps both ways. With only one
  live session it returns you to the list instead of silently re-attaching, which would look like the key did
  nothing.
  ⚠️ Not `Shift+]` / `Shift+[`: in a terminal those *are* `}` and `{` (bytes `0x7D`/`0x7B`), and agent-desk
  proxies the PTY, so binding them would swallow every brace typed into every session.

### Fixed
- **`ProjectPath` is normalized on the way in, so the database stops storing dirty paths.** It used to be stored
  exactly as it arrived and cleaned only where someone remembered to compare it — `isDuplicateSession` trimmed a
  trailing slash before comparing, which means the value was known to be wrong at rest and every other reader was
  left to rediscover that. The ones that didn't were silently wrong: the `.claude.json` lookup in
  `GetClaudeSessionID`, the CLI's `inst.ProjectPath == identifier`, and experiment path matching all compare
  exact strings. Normalizing at rest also repairs what is already stored without a migration — rows are cleaned
  as they load, so the next save writes the clean value back. **11 of 26 rows in a real profile carried a
  trailing slash**, and they heal on their own once this build has run.
- **Session discovery no longer misses `~/.claude/projects` directories for paths with a trailing slash.**
  `ConvertToClaudeDirName` maps every non-alphanumeric to `-`, so `/Users/me/proj/` encoded to `-Users-me-proj-`,
  a directory that does not exist next to the real one. Those instances showed a stale transcript forever. One
  `ClaudeProjectDirName` (EvalSymlinks + `filepath.Clean`) is now the only way such a name is built.

### Changed
- **The help screen documents group keys that already worked.** `d` (delete a group, with a confirm) was listed
  nowhere at all, and `Shift+Up`/`Shift+Down` (or `K`/`J`) to reorder a group or session appeared only under
  general movement. Both are now under GROUPS, alongside a new WHILE ATTACHED section.
- **`go fmt` pass** — struct field alignment across nine pre-existing files. Kept as its own commit so the
  release diff stays readable.


## [1.0.7] - 2026-08-25

### Removed
- **Deleted the session-forking feature entirely** (~2,200 lines). It had been dormant/unreachable since the
  keybindings were unwired, and v1.0.6 stopped advertising it; this removes the code: the fork dialog, all
  `home.go` fork handlers and animation state, the session-package primitives (`CanFork`, `Fork*`,
  `CreateForkedInstance*`, `ForkOpenCode*`, `writeOpenCodeForkScript`), the `ToArgsForFork` option methods and
  their fork-only transient fields, the `session fork` CLI subcommand, and all fork tests. No behavior change —
  fork was already unusable.

## [1.0.6] - 2026-08-24

### Removed
- **Fork is no longer advertised.** The fork keybindings were already unwired in this fork (dead code), but the
  help screen and README still listed `f`/`F`, a `session fork` CLI example, and a "Fork sessions" section.
  Removed those so the docs match reality. The dormant fork code remains in the tree.

### Fixed
- **Corrected the README status table.** It claimed Running=green / Waiting=yellow with a `◐` symbol; the actual
  UI is Running=**yellow** `●`, Waiting=**green** `●`. Also documented the dimmer-green unread bookmark (`u`) vs
  the bright-green waiting state.

## [1.0.5] - 2026-08-22

### Fixed
- **Worktrees no longer nest inside worktrees.** Creating a new session (or forking) while the base directory
  was itself a git worktree produced `.worktrees/` inside a worktree, because the base resolved to the
  worktree's own root. New worktrees now always base off the **main repo root** (de-nested via
  `GetWorktreeBaseRoot`).

### Changed
- **New Session worktree UX is a single name.** Removed the separate branch field: the session name is the one
  input, and it fills both the branch and the worktree folder — identical, sanitized, and **lowercased**
  (e.g. "Auth Refactor" → folder `auth-refactor`, branch `auth-refactor`, no prefix).
- **The dialog now shows the full resolved worktree path** (de-nested, `~`-collapsed) and the derived branch,
  live as you type — so you can see exactly where the worktree will land before creating it. Says so when the
  selected path isn't a git repo.

## [1.0.4] - 2026-08-22

### Added
- **Attention bell.** Rings the terminal bell once when any session transitions into "waiting" (needs your
  input) — so you hear it even while attached to another session or looking away. One ring per poll regardless
  of how many sessions transition. On by default; disable with `bell = false` under `[notifications]` in
  `~/.agent-desk/config.toml`.

### Changed
- **New sessions are inserted directly below the highlighted row** instead of at the bottom of the group. The
  position persists (renormalized `Order`), and falls back to appending when there's no same-group anchor.
  Forking is unchanged (forks already nest under their parent).
- **Distinct greens for the two "green" states.** "Needs attention / not yet looked at" (waiting) stays bright
  green; the `u` unread bookmark ("seen it, reply still owed") is now a dimmer green — they were previously the
  same color and indistinguishable.

## [1.0.3] - 2026-08-07

### Added
- **Light color schemes** (`Paper`, `Daylight`, `Linen`) selectable with `c`, alongside the existing dark presets.

### Fixed
- **Copilot text unreadable (dark-on-dark).** Copilot CLI mis-detects a dark terminal — its background-detection
  query is swallowed inside tmux — so it renders as if on a light terminal and prints dark text. Rather than
  fight its detection, new Copilot sessions now default to a **light pane** (`Paper`), so Copilot's dark text is
  readable. Change it per session with `c`; for an existing Copilot session, press `c` and pick a light scheme.

## [1.0.0] - 2026-08-06

First tagged release of the Millwright Software fork.

### Added
- **GitHub Copilot CLI** as a first-class tool: selectable in the new-session dialog and setup wizard, with its
  own icon, name detection, `copilot --continue` resume on restart, and status-detection patterns adapted from
  upstream agent-deck's real Copilot CLI transcripts.
- Manual session flags: press `u` to cycle a session through unread → parked (red) → parked (blue) → normal;
  the flag clears automatically when you attach.
- Model and mode (auto/manual) surfaced in the navigation rows, with shorthand model names (e.g. `o4.8`, `g2.5p`).
- Name-forward new-session path picker: rows lead with the folder name, collapse `$HOME` to `~`, and mark recent
  projects with a star.

### Changed
- Session saves are now upsert-only, so multiple concurrent Agent Desk windows no longer delete each other's
  sessions (previous full-table-sweep saves were a data-loss vector).
- "Last active" time is derived from session-log activity and shown as minute-rounded shorthand.
- The attached tmux status bar now shows the session's own title.
- Claude model detection refreshes after a `/model` switch, even for idle sessions.
- The numbered quick-switch / notification bar defaults off.

### Removed
- The **conductor** meta-agent orchestration subsystem (persistent orchestrator sessions + Telegram/Slack
  bridges) — an agent-deck feature this fork doesn't use (~5,000 lines).
- The per-session context-note feature.
- Stale upstream marketing site, SEO files, and the `llms-full.txt` dump.

### Known limitations
- Orphaned `tmux -C` control clients left by crashed/killed TUIs are not yet swept, so on very heavy long-term
  use they could accumulate toward the OS pty cap. Upstream's fix is a ~300-line platform-specific port and is
  deferred; clean exits don't orphan, and the primary "can't create sessions" cause (concurrent-window save
  sweeps) is already fixed above.
