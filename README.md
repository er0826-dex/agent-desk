<div align="center">

<img src="site/logo.svg" alt="Agent Desk" width="96">

# Agent Desk

**Terminal command center for AI coding agents.**

Run and switch between Claude Code, Copilot, Gemini, and other terminal AI tools —
each in its own tmux session — from one keyboard-driven view.

[![Release](https://img.shields.io/github/v/release/millwright-software/agent-desk?style=flat-square&color=e0af68&labelColor=1a1b26)](https://github.com/millwright-software/agent-desk/releases)
[![License](https://img.shields.io/badge/License-MIT-9ece6a?style=flat-square&labelColor=1a1b26)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go&labelColor=1a1b26)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20WSL-7aa2f7?style=flat-square&labelColor=1a1b26)](#install)

</div>

> **A Millwright Software fork of [`asheshgoplani/agent-deck`](https://github.com/asheshgoplani/agent-deck)** by Ashesh Goplani.
> Maintained independently under the original MIT license (see [`LICENSE`](LICENSE)); upstream authorship is
> preserved in the commit history. This fork drops the conductor layer and the web dashboard. Upstream is
> bigger and actively maintained.

## Install

Requires `tmux`. macOS, Linux, or Windows via [WSL](https://learn.microsoft.com/en-us/windows/wsl/install).

```bash
curl -fsSL https://raw.githubusercontent.com/millwright-software/agent-desk/main/install.sh | bash
```

<details>
<summary>Other ways to install</summary>

```bash
go install github.com/millwright-software/agent-desk/cmd/agent-desk@latest   # needs Go 1.24+

git clone https://github.com/millwright-software/agent-desk.git              # from source
cd agent-desk && make install
```

</details>

The app checks for new releases and can update itself. Or re-run the installer.

## Use

```bash
agent-desk                        # launch the TUI
agent-desk add . -c claude        # add the current dir as a Claude session
```

| Key | Action |
|-----|--------|
| `Enter` | Attach to session |
| `Shift+←` / `Shift+→` | Previous / next live session, while attached |
| `Shift+↑` / `Shift+↓` | Previous / next session waiting for you, while attached |
| `Ctrl+Q` | Detach back to the list |
| `n` · `d` | New session · delete |
| `g` · `G` | New group · new top-level group |
| `m` · `e` | Move session to a group · rename group |
| `s` · `u` | Snooze until a time · cycle unread / parked marker |
| `o` | Sort within groups: manual · recent · busy |
| `Shift+↑` / `Shift+↓` | Move a group or session up/down |
| `Tab` · `1`–`9` | Expand/collapse group · jump to group |
| `?` | Every shortcut |

Each session shows what its agent is doing:

| Status | Symbol | Meaning |
|--------|--------|---------|
| Running | `●` yellow | Actively working |
| Waiting | `●` green | Needs your input |
| Idle | `○` gray | Ready for commands |
| Error | `✕` red | Something went wrong |

Claude sessions read the session hook and transcript, so the status is exact and includes the current model.
Other tools fall back to output patterns and tmux activity.

## Worktrees

Each session can run in its own git worktree and branch.

```bash
agent-desk add . -c claude --worktree feature/a --new-branch
agent-desk worktree finish "My Session"    # merge, remove worktree, delete session
agent-desk worktree cleanup                # remove orphaned worktrees
```

Worktrees go in a sibling directory by default; `[worktree] default_location` in `~/.agent-desk/config.toml`
changes that.

## Tools

| Tool | Support |
|------|---------|
| Claude Code | Status, resume, model detection |
| Gemini CLI | Status, resume |
| GitHub Copilot CLI | Status, resume (`--continue`) |
| OpenCode / Codex / Cursor | Status detection |
| Anything else in a terminal | Add it under `[tools.*]` in `config.toml` |

## Documentation

| Guide | Contents |
|-------|----------|
| [CLI Reference](skills/agent-desk/references/cli-reference.md) | Commands, flags, scripting |
| [Configuration](skills/agent-desk/references/config-reference.md) | `config.toml`, custom tools, MCP |
| [TUI Reference](skills/agent-desk/references/tui-reference.md) | Shortcuts, status indicators, navigation |
| [Troubleshooting](skills/agent-desk/references/troubleshooting.md) | Common issues, recovery, uninstalling |

Agent Desk makes its own tmux sessions, prefixed `agentdesk_*`. Existing ones are untouched.

## Development

```bash
make build
make test
make lint
```

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE). Original work © 2025 Ashesh Goplani; fork modifications © 2026 Millwright Software.

---

<div align="center">

Built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [tmux](https://github.com/tmux/tmux).

**[Docs](skills/agent-desk/references/) · [Issues](https://github.com/millwright-software/agent-desk/issues) · [Discussions](https://github.com/millwright-software/agent-desk/discussions)**

</div>
