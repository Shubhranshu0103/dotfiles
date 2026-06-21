# Incremental Sync — Design Spec

**Date:** 2026-06-21  
**Status:** Approved  
**Scope:** `install.sh` first, `just sync` second

---

## 1. Problem

`install.sh` and `just sync` re-run every step unconditionally on every invocation. `brew bundle` alone takes 5–30 seconds even when nothing changed. On a machine that has already been bootstrapped, re-running `install.sh` should be nearly instant if nothing in the repo has changed.

---

## 2. Approach

A Go binary (`dotsync`) acts as a **state oracle**. It maintains a persistent, human-readable record of what each install step looked like last time it ran successfully. On each invocation, `install.sh` asks the binary whether a step is stale before running it and records a new hash after it succeeds.

The binary does not own execution — `install.sh` and the `Justfile` own that. The binary only answers "skip or run?" and records state.

---

## 3. Binary

### Location

```
tools/dotsync/
  main.go
  go.mod
  go.sum
bin/
  dotsync          ← committed darwin/arm64 binary, no runtime needed on fresh machine
```

Build command (run on any machine with Go after source changes):

```bash
cd tools/dotsync && GOOS=darwin GOARCH=arm64 go build -o ../../bin/dotsync .
```

A `just build-dotsync` recipe wraps this so it is not forgotten.

### Subcommands

| Subcommand | Behaviour |
|---|---|
| `dotsync check <step>` | Exit `0` (up-to-date, skip) or `1` (stale, run). If the step is unknown to the state file, exits `1` (treat as never run). |
| `dotsync record <step>` | Compute current hash + per-file manifest for the step's inputs; write to `.dotfiles-state`. Only called after the step command exits `0`. |
| `dotsync status` | Pretty-print a table of all steps: status, last run, commit, and drift (see §7). |
| `dotsync rollback <commit>` | Restore a snapshot from `.dotfiles-snapshots` as the current state, so subsequent `check` calls treat that snapshot's hashes as ground truth. |
| `dotsync gc` | Remove orphaned step entries from `.dotfiles-state` that no longer exist in `dotsync.toml`. |

The binary finds the repo root by walking up from `$PWD` until it finds `.dotfiles-state` or `dotsync.toml`, the same way git finds `.git`.

### Failure safety

`install.sh` runs under `set -e`. If a step command fails, the script exits before `dotsync record` is ever reached. A hash in `.dotfiles-state` therefore always means that step last ran successfully. A failed step leaves state unchanged — the next run retries it.

---

## 4. Step Definitions

Steps are declared in **`dotsync.toml`** (committed, versioned). The binary reads this to know what to hash for each step. Adding a step requires only two changes: a new entry here and a new guard in `install.sh`. No binary rebuild needed.

```toml
# dotsync.toml

[[steps]]
name   = "brew"
inputs = ["Brewfile"]

[[steps]]
name   = "stow_zsh"
inputs = ["zsh/"]

[[steps]]
name   = "stow_tmux"
inputs = ["tmux/"]

[[steps]]
name   = "stow_nvim"
inputs = ["nvim/"]

[[steps]]
name   = "stow_git"
inputs = ["git/"]

[[steps]]
name   = "stow_ghostty"
inputs = ["ghostty/"]

[[steps]]
name   = "stow_starship"
inputs = ["starship/"]

[[steps]]
name   = "stow_mise"
inputs = ["mise/"]

[[steps]]
name   = "mise_runtimes"
inputs = ["mise/.config/mise/config.toml"]

[[steps]]
name   = "tpm"
inputs = ["tmux/.tmux.conf"]

[[steps]]
name        = "vscode_profiles"
inputs      = ["vscode/profiles/", "vscode/extensions-common.txt"]
drift_check = "vscode"

[[steps]]
name   = "vscode_settings"
inputs = ["vscode/settings.json", "vscode/keybindings.json"]

[[steps]]
name   = "macos"
inputs = ["macos/defaults.sh"]
```

**Input hashing rules:**
- Entries ending in `/` are walked recursively; each file is hashed individually.
- File entries are hashed directly.
- The step's top-level `hash` is a SHA256 of all individual file hashes concatenated in sorted path order.

**Adding a step:** add entry to `dotsync.toml`, add guard in `install.sh`. Done.  
**Removing a step:** delete from `dotsync.toml` and `install.sh`. Orphaned entries in `.dotfiles-state` are ignored. `dotsync gc` prunes them.  
**Unknown step name in `install.sh`:** treated as `never run` — `check` returns stale, step runs, `record` adds it to state. No crash.

---

## 5. State Files

Both files are gitignored.

### `.dotfiles-state` — current state (TOML, human-readable)

```toml
# Managed by dotsync — do not edit manually.
# Run `dotsync status` for a pretty view.

[meta]
current_commit    = "76392e9"
current_commit_at = "2026-06-21T10:00:00Z"
dotsync_version   = "0.1.0"

[steps.brew]
hash        = "a3f9c1..."
recorded_at = "2026-06-21T10:00:00Z"
commit      = "76392e9"
status      = "ok"           # ok | never_run

[steps.brew.files]
"Brewfile" = "a3f9c1..."

[steps.stow_nvim]
hash        = "7bc19f..."
recorded_at = "2026-06-21T10:01:00Z"
commit      = "76392e9"
status      = "ok"

[steps.stow_nvim.files]
"lua/plugins.lua" = "c8a2f1..."
"lua/lsp.lua"     = "90b3d4..."
"init.lua"        = "551ea9..."
```

### `.dotfiles-snapshots` — rollback history (JSON, capped at 10 entries)

Appended to on every `dotsync record` call that advances the commit. Oldest entry pruned when cap is exceeded.

```json
[
  {
    "commit": "76392e9",
    "recorded_at": "2026-06-21T10:00:00Z",
    "steps": {
      "brew":      { "hash": "a3f9c1...", "files": { "Brewfile": "a3f9c1..." } },
      "stow_nvim": { "hash": "7bc19f...", "files": { "lua/plugins.lua": "c8a2f1...", "init.lua": "551ea9..." } }
    }
  },
  {
    "commit": "0542b37",
    "recorded_at": "2026-06-20T09:00:00Z",
    "steps": { ... }
  }
]
```

**`dotsync rollback <commit>`:** finds the matching entry, writes its steps back into `.dotfiles-state` as the current state, updates `[meta]`. Subsequent `check` calls treat those hashes as ground truth — skipping steps that already matched at that commit without re-running them.

---

## 6. `install.sh` Integration

Steps already self-guarded (Homebrew binary check, `~/dotfiles` symlink, TPM directory clone, Claude Code binary check, `chsh`) are left as-is. Only steps with real file inputs get hash guards.

```bash
#!/usr/bin/env bash
set -e

DOTFILES="$(cd "$(dirname "$0")" && pwd)"
DOTSYNC="$DOTFILES/bin/dotsync"

echo "🚀 Bootstrapping machine..."

# 1. Homebrew — self-guarded
if ! command -v brew &>/dev/null; then
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi
eval "$(/opt/homebrew/bin/brew shellenv 2>/dev/null || /usr/local/bin/brew shellenv)"

# 2. Packages
if ! "$DOTSYNC" check brew; then
  brew bundle --verbose --file="$DOTFILES/Brewfile"
  "$DOTSYNC" record brew
fi

# 3. Symlink ~/dotfiles — idempotent
[ ! -e "$HOME/dotfiles" ] && ln -sf "$DOTFILES" "$HOME/dotfiles"

# 4. Stow — per-package
for pkg in zsh tmux nvim git ghostty starship mise; do
  if ! "$DOTSYNC" check "stow_${pkg}"; then
    stow --restow --target="$HOME" "$pkg"
    "$DOTSYNC" record "stow_${pkg}"
  fi
done

# 5. Mise runtimes
if ! "$DOTSYNC" check mise_runtimes; then
  mise install
  "$DOTSYNC" record mise_runtimes
fi

# 6. TPM — clone is self-guarded; plugins are hashed
[ ! -d "$HOME/.tmux/plugins/tpm" ] && \
  git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm
if ! "$DOTSYNC" check tpm; then
  TMUX_PLUGIN_MANAGER_PATH="$HOME/.tmux/plugins/" \
    ~/.tmux/plugins/tpm/scripts/install_plugins.sh
  "$DOTSYNC" record tpm
fi

# 7. VS Code profiles
if command -v code &>/dev/null && ! "$DOTSYNC" check vscode_profiles; then
  bash "$DOTFILES/vscode/install-profiles.sh"
  "$DOTSYNC" record vscode_profiles
fi

# 8. VS Code settings symlinks
if ! "$DOTSYNC" check vscode_settings; then
  VSCODE_DIR="$HOME/Library/Application Support/Code/User"
  mkdir -p "$VSCODE_DIR"
  ln -sf "$DOTFILES/vscode/settings.json" "$VSCODE_DIR/settings.json"
  ln -sf "$DOTFILES/vscode/keybindings.json" "$VSCODE_DIR/keybindings.json"
  "$DOTSYNC" record vscode_settings
fi

# 9. macOS defaults
if ! "$DOTSYNC" check macos; then
  bash "$DOTFILES/macos/defaults.sh"
  "$DOTSYNC" record macos
fi

# 10. Claude Code — self-guarded
if ! command -v claude &>/dev/null; then
  curl -fsSL https://claude.ai/install.sh | bash
fi

# 11. Default shell — self-guarded
[ "$SHELL" != "$(which zsh)" ] && chsh -s "$(which zsh)"

echo "✅ Done. Restart terminal."
```

---

## 7. VS Code Drift Detection

VS Code profile-specific settings and extensions are stored in VS Code's internal directory (`~/Library/Application Support/Code/User/profiles/<hash>/`), not in the symlinked files. When you change a setting or install an extension through the VS Code UI while in a profile, the change is written there — not to the repo.

This creates two distinct change directions:

| Direction | Term | Action |
|---|---|---|
| Repo → system | **stale** | Run the install step |
| System → repo | **drifted** | Run `dots vscode-export` |

`dotsync status` shows both. For steps with `drift_check = "vscode"`, the binary exports each profile to a temp file and compares `settings` and `extensions` fields (semantic JSON comparison, ignoring metadata noise like timestamps and VS Code version). Byte comparison would produce false positives.

The drift check is skipped gracefully if `code` CLI is not available. Drift state is computed on demand — nothing is added to `.dotfiles-state`.

**`dotsync status` output:**

```
Step              Status         Last Run              Commit     Drift
───────────────────────────────────────────────────────────────────────
brew              ✓ up-to-date   2026-06-21 10:00      76392e9    —
stow_zsh          ✓ up-to-date   2026-06-21 10:01      76392e9    —
stow_nvim         ✗ stale        2026-06-21 10:01      0542b37    —
  └ lua/plugins.lua modified
stow_tmux         ✓ up-to-date   2026-06-21 10:01      76392e9    —
stow_git          ✓ up-to-date   2026-06-21 10:01      76392e9    —
stow_ghostty      ✓ up-to-date   2026-06-21 10:01      76392e9    —
stow_starship     ✓ up-to-date   2026-06-21 10:01      76392e9    —
stow_mise         ✓ up-to-date   2026-06-21 10:01      76392e9    —
mise_runtimes     ✓ up-to-date   2026-06-21 10:02      76392e9    —
tpm               ✓ up-to-date   2026-06-21 10:03      76392e9    —
vscode_profiles   ✓ up-to-date   2026-06-21 10:04      76392e9    ⚠ drifted
  └ ml.code-profile has UI changes (run: dots vscode-export)
vscode_settings   ✓ up-to-date   2026-06-21 10:04      76392e9    —
macos             ✓ up-to-date   2026-06-20 09:00      0542b37    —

Snapshots: 2 stored  →  rollback available to 0542b37
```

**Per-step status values:**

| Symbol | Meaning |
|---|---|
| `✓ up-to-date` | Hash matches last recorded; last run succeeded |
| `✗ stale` | Inputs changed since last record |
| `— never run` | No entry in state file |

---

## 8. Justfile Changes

```just
# Build the dotsync binary (run after editing tools/dotsync/)
build-dotsync:
    cd tools/dotsync && GOOS=darwin GOARCH=arm64 go build -o ../../bin/dotsync .
    @echo "✅ bin/dotsync built"

# Restow changed packages only
sync:
    #!/usr/bin/env bash
    for pkg in zsh tmux nvim git ghostty starship mise; do
      if ! ./bin/dotsync check "stow_${pkg}"; then
        stow --restow --target=$HOME "$pkg"
        ./bin/dotsync record "stow_${pkg}"
      fi
    done
    @echo "✅ Sync done"

# Show what needs to run (and VS Code drift)
status:
    ./bin/dotsync status

# Rollback to a previous snapshot
rollback commit:
    ./bin/dotsync rollback {{commit}}

# Export current VS Code profiles back to repo (run after UI changes)
vscode-export:
    #!/usr/bin/env bash
    for profile in ML Java WebDev Rust; do
      code --profile "$profile" \
        --export "$(pwd)/vscode/profiles/${profile,,}.code-profile"
    done
    echo "✅ Profiles exported — review with git diff, then dots push"
```

---

## 9. Out of Scope

- Linux / non-macOS support (darwin/arm64 binary only for now)
- Windows support
- Multi-architecture binary distribution
- Automated VS Code profile export on file-system events
- SSH config tracking (never commit keys)
