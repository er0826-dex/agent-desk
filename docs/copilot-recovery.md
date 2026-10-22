# Copilot conversations after a reboot

Agent Desk's sidebar entries and Copilot's conversation history are persisted
locally. The tmux processes hosting them do not survive a reboot. Restarting an
entry must recreate tmux and resume the correct conversation, not whichever
conversation was used most recently.

## Recover an existing entry

Select a Copilot entry and press **Shift+R**. Enter also starts recovery when its
tmux session is gone.

Older entries that launch `copilot` but were labeled `claude` or `shell` are
recognized as Copilot when loaded. A Claude model selected inside Copilot does
not make it a Claude Code session or change where its history belongs.

If the entry has no stored Copilot conversation UUID, a recovery picker opens.
It shows saved conversation names, UUIDs and update times. Type to filter, use
Up/Down to navigate, and Tab to switch between the entry's folder and all folders.
The all-folders option also covers conversations whose working directory changed.
Press Enter to explicitly link and resume the highlighted conversation. Esc
cancels without changing anything.

The link is saved before restarting. Subsequent restarts use
`copilot --resume=<uuid>`, including after a reboot. Missing history or a lock
error is reported by Copilot; Agent Desk does not silently start another
conversation or delete lock files.

For terminal-based recovery:

```sh
agent-desk session set <entry> copilot-session-id <conversation-uuid>
agent-desk session restart <entry>
agent-desk session show <entry> --json
```

The UUID must identify readable local metadata in
`~/.copilot/session-state/<uuid>/workspace.yaml`. Names and folder matches are
discovery aids, never proof that an entry owns a particular conversation.

## New sessions

For the standard `copilot` command, Agent Desk allocates a UUID and persists it
before launching `copilot --session-id=<uuid>`. This requires a Copilot CLI that
supports `--session-id` (checked before launch). Two entries in the same folder
therefore retain distinct identities.

Custom launch commands are left unchanged. They are not assigned a guessed ID.
Automatic restart refuses to discard or rewrite custom flags; use the standard
`copilot` command for managed recovery, or resume explicitly through the custom
command. Wrappers containing `{command}` continue to wrap managed launch and
resume commands.

The stored link follows a conversation, not an arbitrary interactive tab. If you
switch to a different conversation inside Copilot with `/new`, `/resume`, or
`/fork`, explicitly rebind the entry using `session set ... copilot-session-id`
to change which conversation Agent Desk will restore.

Installing a new Agent Desk binary does not update an already-running TUI.
Exit and reopen Agent Desk to use the updated recovery behavior; no Mac reboot
or deletion of Copilot state is needed.
