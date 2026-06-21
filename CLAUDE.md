# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

A single Git repository that fully describes a personal macOS development environment. Cloning and running `./install.sh` on a fresh machine reproduces the complete setup — terminal, shell, editor, toolchains, and system preferences — with no manual steps afterward.

**The repo is always the source of truth.** Config files are never edited via their applications — always edit in the repo, then sync to the system.

## Core workflow commands

The `dots` alias (defined in `.zshrc`) points `just` at this repo from any directory:

```sh
dots sync              # restow all packages — apply repo edits to $HOME
dots update            # git pull + sync + brew bundle
dots push "message"    # git add -A + commit + push
dots add <file> <pkg>  # adopt a loose $HOME file into a stow package
just check             # lint scripts + simulate stow (no system changes)
```

`just check` is the pre-flight for CI: runs `shellcheck` on `.sh` files and `stow --simulate` to catch conflicts before they land.

## Architecture

**GNU Stow** manages all symlinks. Each top-level directory is a stow package; `stow --restow --target=$HOME <package>` mirrors its tree into `$HOME`. The package list is defined in `Justfile:1` (`PACKAGES`) and must stay in sync with `install.sh`.

```
zsh/       → ~/.zshrc, ~/.zprofile
tmux/      → ~/.tmux.conf
nvim/      → ~/.config/nvim/
git/       → ~/.gitconfig, ~/.gitignore_global
ghostty/   → ~/.config/ghostty/config
starship/  → ~/.config/starship.toml
mise/      → ~/.config/mise/config.toml
```

**VS Code** is not stowed — `install.sh` symlinks `vscode/settings.json` and `vscode/keybindings.json` directly into `~/Library/Application Support/Code/User/` because that path is outside `$HOME`'s stow tree.

**Homebrew** (`Brewfile`) is the sole package manager. Nothing is installed ad hoc.

**Mise** pins language runtimes (`mise/.config/mise/config.toml`): Node 22, Python 3.12, Go 1.23. Rust defers to rustup.

## Neovim config structure

`nvim/.config/nvim/` uses lazy.nvim as the plugin manager, bootstrapped in `plugins.lua`.

- `init.lua` — entry point; loads `options`, `keymaps`, `plugins` in order
- `lua/options.lua` — vim options
- `lua/keymaps.lua` — keybindings (`<leader>` = space)
- `lua/plugins.lua` — all plugins declared with lazy.nvim; LSP is triggered by Mason's config callback which calls `require("lsp")`
- `lua/lsp.lua` — mason-lspconfig setup and per-server handlers

Key plugin keybindings: `<leader>e` (nvim-tree), `<leader>ff/fg/fb` (Telescope).

## CI

`.github/workflows/check.yml` runs on every push/PR to `main`:
1. `shellcheck` on `install.sh`, `macos/defaults.sh`, `vscode/install-profiles.sh`
2. `stow --simulate` for all packages

A PR should pass `just check` locally before pushing.

## Theme

Catppuccin Mocha is applied everywhere: Ghostty (built-in), tmux (`catppuccin/tmux` plugin), Neovim (`catppuccin/nvim`), VS Code (extension), Starship (`bold mauve`). Keep new tooling consistent with this theme.

## Key constraints

- `macos/defaults.sh` is **not stowed** — it writes directly to macOS preference files and is run once during bootstrap.
- Never commit secrets, SSH keys, or tokens. The `.gitignore_global` covers common cases but is not exhaustive.
- `vscode/` profiles are exported `.code-profile` files. To update a profile, export it from VS Code and overwrite the file in the repo.

## dotsync

`bin/dotsync` is a committed Go binary that tracks step hashes for `install.sh` and `just sync`. Source lives in `tools/dotsync/`.

- `dotsync check <step>` — exits 0 (skip) or 1 (stale/run)
- `dotsync record <step>` — records current hash after a step succeeds
- `dotsync status` — shows all steps, staleness, and VS Code drift
- `dotsync rollback <commit>` — restores a previous state snapshot
- `dotsync gc` — prunes orphaned state entries

Step definitions live in `dotsync.toml`. To add a step: add an entry there and a guard in `install.sh`. To rebuild the binary after source changes: `just build-dotsync`.

State files (`.dotfiles-state`, `.dotfiles-snapshots`) are gitignored.
