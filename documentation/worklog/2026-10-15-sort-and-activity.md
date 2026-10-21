# 2026-10-15 — Sort within groups (`o`) and activity on the row

**Ask (10/13, 2:30am):** "I've got multiple threads going and I don't know
how recent or how much has been logged over time. I almost wish I could cycle
through the most active ones, most recent, or have some better way of sorting
them when I'm doing the shift up and down thing, and shift left and right."

**What existed:** one order, hand-built with K/J; the rings walk it. Recency
was already on every row (the transcript mtime badge); turn counts only in the
stats preview.

**Options walked:** (1) a sort toggle for the list, rings follow; (2) a flat
recent-switcher lens; (3) activity on the row, no reordering; (4) ring order
separate from list order (rejected: list and keys would disagree). Eric: 1 and
3, recent sorts within each group.

## Design

- `GroupTree.SortKey func(*Instance) int64`: when set, `flatten` stable-sorts
  each group's top-level sessions by descending key. Sub-sessions stay under
  their parent; last-in-group flags are computed after the sort; persisted
  `Order` is never touched, so manual restores the hand-built list.
- `sortMode` on Home (manual / recent / busy), `o` cycles, persisted in
  `uiState.sort_mode`. `rebuildFlatItems` sets the tree's key from the mode
  on every rebuild (the tree is recreated on reload).
- Recent key: `GetLastActivityTime().UnixNano()`. Busy key: `RecentTurns()`.
- `RecentTurns`: assistant entries in the last 24h, from `countRecentTurns`,
  a bounded tail read (4 MB) scanned newest-first that stops at the first
  entry older than the window. Refreshed inside `readJSONLTail` (already
  called per tick, size-gated) when the file grew or the count is older than
  ten minutes, so turns age out even when nothing is written.
- Row: ` 12t` badge before the age badge; title fades (`SessionTitleStale`)
  when idle 48h+ and no status or flag overrides it.
- Header: `↓ recent` / `↓ busy` pill when not manual; hint gains `o sort`.

## Tests

`session/sort_test.go`: flatten under a key (order, sub-session attachment,
last-in-group flag, Order untouched, ties, nil restores); the 24h count in
memory, from a file, bounded against a padded multi-MB file, and the
refresh cache on size and age. `ui/sort_test.go`: mode cycle, parse, keys.

## Not done

- No flat recent-switcher lens (option 2). Next step if a cross-group view is
  what he reaches for.
- "Busy" counts only Claude transcripts; other tools read 0.

## 10/21: the row's info moved into right-aligned columns

Eric asked to clean up the shorthand on each row (turns, recency, model).
Three mockups on his real Legal titles: A fixed right-aligned columns, B
exceptions only (model only when unusual, turns as a bar), C activity as
title brightness. He chose A. `sessionColumns` renders model+⚡ (7 cells, the
bolt is two), turns (3, `·` for none), coarse age (4, `columnAge`); snoozed
rows put the wake time across model+turns. `renderSessionList` records the
list width on Home (`listRowWidth`) so the row pads to it and truncates the
title with `…`; the worktree badge drops first. Tests pin every row's column
width and the full row width.

Same day, Eric: age at three characters or less. `columnAge` now caps at
`3w+` (1m–59m, 1h–23h, 1d–6d, 1w–2w, 3w+); the column is 3 wide. The ⚡ stays:
he asked, then kept it once it read as auto mode.

Then: *"I don't need the 5.5 / 5.1, just whether it's Fable or Opus."*
`Instance.ModelFamily()` (first letter of the short name: o, f, s, h, g); the
model column is 3 cells (letter + two-cell bolt). With the narrower columns a
snoozed row's wake time spans all three columns instead of model+turns, so a
date like `10/26 9a` still fits; age is dropped on snoozed rows.
