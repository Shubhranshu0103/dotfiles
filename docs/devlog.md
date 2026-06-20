# Dotfiles DevLog

Running log of design decisions, alternatives considered, and reasoning. Append new entries at the bottom.

---

## 2026-06-20 — Initial Setup

### Repo as source of truth
**Decision:** All config lives in the repo. Never edit via the application's own UI or settings panel. Edit in `~/dotfiles`, then sync.  
**Why:** Prevents config drift across machines. The repo is always the canonical state; the system is a derived artifact of it.

### GNU Stow for symlinking
**Decision:** Use Stow to symlink each tool's config directory into `$HOME`.  
**Alternatives considered:** Copying files on sync (requires diffing to detect changes), chezmoi (replaces the entire paradigm), yadm (Git-based but no symlink management).  
**Why:** Stow is transparent — it does one thing (symlink trees), leaves files editable in place, and is trivially reversible (`stow --delete`). No magic, no new syntax to learn.

### One Stow package per tool
**Decision:** Each top-level directory is one Stow package (`zsh/`, `nvim/`, `tmux/`, etc.).  
**Why:** Granularity matches how configs are actually changed — you update your Neovim config independently of your tmux config. Finer granularity also enables per-package incremental sync (see 2026-06-21).

### Homebrew as sole package manager
**Decision:** Everything installed via Homebrew. `Brewfile` is the single source of truth for installed tools and GUI apps. Nothing installed ad hoc.  
**Why:** Ad-hoc installs are invisible to the repo. A fresh machine bootstrap becomes `brew bundle` instead of a series of manual steps. `brew bundle check` gives a quick health check.

### Mise for runtime versions
**Decision:** Use `mise` to pin Node, Python, and Go versions. Rust deferred to `rustup`.  
**Alternatives considered:** `nvm` + `pyenv` + individual version managers per language.  
**Why:** One tool, one config file (`mise/.config/mise/config.toml`), activated once in `.zshrc`. `nvm` alone adds ~100ms to shell startup. Rust is excluded because `rustup` provides richer toolchain control (nightly, targets, components) that `mise` doesn't replicate well.

### `just` for task running
**Decision:** `Justfile` at repo root for all dotfiles operations (`sync`, `update`, `push`, `add`, `nuke`, `check`).  
**Alternatives considered:** Plain shell aliases, a custom CLI script, a Makefile.  
**Why:** `just --list` is self-documenting. Justfile recipes are readable shell with no Make-specific syntax. The `dots` alias (`just --justfile ~/dotfiles/Justfile --working-directory ~/dotfiles`) makes every recipe available from any directory.

### Zsh without a plugin manager
**Decision:** Source zsh plugins directly from Homebrew paths. No oh-my-zsh, zinit, or zplug.  
**Why:** Plugin managers add startup overhead and their own config surface. Homebrew already manages the plugin files; sourcing them directly is two lines in `.zshrc` and fully transparent. Shell startup stays fast.

### Catppuccin Mocha everywhere
**Decision:** Single theme applied across Ghostty, tmux, Neovim, VS Code, and Starship.  
**Why:** Visual consistency across tools reduces cognitive noise. Catppuccin Mocha has official plugin support in all five tools, so there's no manual color-matching required.

### Ghostty as terminal
**Decision:** Ghostty over WezTerm or iTerm2.  
**Why:** Native Metal renderer, built-in Catppuccin theme support, zero Lua config required (unlike WezTerm). Config is a simple `.ini`-style file that fits in ~15 lines.

### VS Code managed via profiles + symlinks
**Decision:** Four profiles (ML, Java, WebDev, Rust) exported as `.code-profile` files. `settings.json` and `keybindings.json` symlinked directly into `~/Library/Application Support/Code/User/`.  
**Why:** Profiles isolate extension sets per context. Symlinking global settings means UI changes write back to the repo file directly — no export step needed for global settings. Profile-specific settings require a manual export (see `dots vscode-export`, added 2026-06-21).

### Neovim: hand-rolled Lua with lazy.nvim
**Decision:** No pre-built distro (no LazyVim, AstroNvim). Four files: `options.lua`, `keymaps.lua`, `plugins.lua`, `lsp.lua`. lazy.nvim as plugin manager.  
**Why:** Distros obscure what's actually configured, making debugging hard. A hand-rolled config is fully understood and owned. lazy.nvim is the de facto standard for hand-rolled configs; bootstrapped automatically so the repo is self-contained.

---

## 2026-06-21 — Incremental Sync (`dotsync`)

### Problem
`install.sh` and `just sync` re-run every step unconditionally. `brew bundle` alone takes 5–30 seconds when nothing changed. On an already-bootstrapped machine, a re-run should be nearly instant.

### Considered existing tools
- **Taskfile (go-task):** Has `method: checksum` for skip-unchanged steps. Would cover the basic case but: (a) requires migrating from `just` to Taskfile format, (b) has no rollback/snapshot support, (c) state is opaque binary files in `.task/`, not human-readable.
- **Chezmoi:** Replaces stow entirely. Too large a paradigm shift.
- **GNU Make:** mtime-based, not content-hash-based. Renames and moves evade it.  
**Conclusion:** Existing tools cover the skip-unchanged case but not rollback. Worth building a custom binary for the combination.

### Language: Go with committed binary
**Decision:** Go binary at `tools/dotsync/`, compiled output committed to `bin/dotsync`.  
**Alternatives considered:** Bash helper script (not strongly typed, hard to test), Swift (pre-installed on macOS, viable but fewer dotfiles-world Go→Swift contributors), TypeScript via Bun (requires adding Bun to Brewfile solely for tooling).  
**Why:** Go compiles to a static binary with zero runtime dependency. Committed to the repo means it's available immediately on a fresh clone — before `mise install` has run Node or Python. Build once on any existing machine; commit; all machines get it.  
**Build target:** `GOOS=darwin GOARCH=arm64` (M-series Mac).

### Binary as state oracle, not runner
**Decision:** `dotsync` only answers "skip or run?" and records state. `install.sh` owns execution.  
**Alternatives considered:** Wrapper pattern (`dotsync run brew -- brew bundle`) where the binary decides whether to run a command and records on success.  
**Why:** Guard pattern keeps `install.sh` readable — you can see exactly what command runs without reading the binary source. Easier to audit, debug, and override.

### Failure safety via `set -e`
**Decision:** Rely on `set -e` in `install.sh` to ensure `dotsync record` is never called after a failed step. No explicit failure state written to `.dotfiles-state`.  
**Why:** A hash in the state file means the step last succeeded. A missing or stale hash means "run it." No third "failed" state needed — the next run just retries.

### Step definitions in `dotsync.toml`
**Decision:** Steps (name + input paths) declared in a committed `dotsync.toml`. Binary reads this; no rebuild needed to add a step.  
**Alternatives considered:** Hardcoding steps in the Go source (requires rebuild on every new step).  
**Why:** Adding a step is a two-file change: `dotsync.toml` + `install.sh`. No Go knowledge needed for routine maintenance. Unknown step names (in `install.sh` but not in `dotsync.toml`) are treated as `never run` — safe fallback, no crash.

### State file format: TOML + JSON split
**Decision:** Two gitignored files. `.dotfiles-state` is TOML (current state, human-readable). `.dotfiles-snapshots` is JSON (rollback history, machine-oriented).  
**Why:** TOML reads cleanly as a file (`cat .dotfiles-state` gives an immediate picture of what's installed and when). JSON handles nested arrays better for the snapshot history. Keeping them separate means the file you'd actually open stays lean.

### Hash structure: top-level hash + per-file manifest
**Decision:** Each step stores one top-level SHA256 (for fast `check`) and a per-file hash map (for informative `status` output).  
**Alternatives considered:** Flat hash only (fast but opaque — can't tell which file changed), per-file only (N comparisons per check, verbose state file).  
**Why:** `check` compares one value (O(1)). When stale, `status` walks the file table and reports exactly which file changed. Negligible storage cost.

### Rollback via snapshot history
**Decision:** `.dotfiles-snapshots` keeps the last 10 commit snapshots. `dotsync rollback <commit>` restores a snapshot as the current state, making subsequent `check` calls treat those hashes as ground truth — no re-running of steps that already matched at that commit.  
**Why:** Without snapshots, rollback means re-running all steps with old config files (git reset + install.sh). With snapshots, steps that already matched at the target commit are skipped — only genuinely changed steps re-run. Faster and safer.  
**Cap at 10:** Enough for practical rollback use (covers recent commits), bounded file size.

### VS Code: two distinct change directions
**Decision:** Distinguish "stale" (repo changed → system needs updating) from "drifted" (VS Code UI changed → repo needs updating). `dotsync status` shows both.  
**Why:** Global `settings.json` and `keybindings.json` are symlinked — UI changes write back to the repo directly (no tracking needed). Profile-specific settings and extensions are written to VS Code's internal directory, not to the repo — these can drift silently. Without explicit drift detection, changes made in the UI would be lost on the next `dots vscode-export` overwrite.

### VS Code drift detection: semantic comparison
**Decision:** Export current profiles to temp files, compare `settings` and `extensions` JSON fields against repo files. Ignore metadata fields (timestamps, VS Code version).  
**Why:** Byte comparison produces false positives — VS Code injects version metadata and may reorder keys on export. Semantic field comparison gives a clean signal. Skipped gracefully if `code` CLI is unavailable (fresh machine).

### `vscode_profiles` split into two steps
**Decision:** `vscode_profiles` (profile files + extensions) and `vscode_settings` (settings.json + keybindings.json) are separate steps.  
**Why:** Re-importing a profile is slow (extension downloads, VS Code restart). Re-symlinking settings is instant. Splitting means a settings change doesn't trigger a full profile reimport.
