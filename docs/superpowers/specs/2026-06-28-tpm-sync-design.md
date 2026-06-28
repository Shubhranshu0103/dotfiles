# TPM Sync in `dots sync` — Design Spec

## Goal

Make `dots sync` automatically run TPM's plugin install script when `tmux/.tmux.conf` has changed, so `dots status` no longer shows `tpm` as stale after a routine sync.

## Background

The `tpm` step in `dotsync.toml` tracks `tmux/.tmux.conf` as its input. Currently, `dots sync` loops over stow packages and records `stow_<pkg>` steps, but never touches the `tpm` step. When `.tmux.conf` changes (new plugin, keybinding, color), `stow_tmux` gets updated but `tpm` stays stale until the user manually runs `install.sh`.

## Design

Add a `tpm` block to the `sync` recipe's shebang block, immediately after the stow loop. Use the same `check → run → record` pattern already established for stow packages.

```bash
if ! ./bin/dotsync check tpm; then
  if [ -f "$HOME/.tmux/plugins/tpm/scripts/install_plugins.sh" ]; then
    TMUX_PLUGIN_MANAGER_PATH="$HOME/.tmux/plugins/" \
      ~/.tmux/plugins/tpm/scripts/install_plugins.sh
    ./bin/dotsync record tpm
  else
    echo "⚠️  TPM not installed — skipping plugin sync (run install.sh first)"
  fi
fi
```

## Key Decisions

- **Any `.tmux.conf` change triggers TPM** — no plugin-line parsing. TPM's install script is idempotent and fast; running it on non-plugin config changes is harmless.
- **Guard on TPM existence** — `dots sync` is a day-to-day command. On a fresh machine before `install.sh` has run, TPM won't be cloned yet. The guard warns and skips rather than erroring out.
- **No new state files or fields** — reuses `dotsync check` / `dotsync record` on the existing `tpm` step.
- **`TMUX_PLUGIN_MANAGER_PATH` is required** — TPM's install script uses this env var to locate the plugins directory; omitting it causes a silent no-op.

## Files Changed

- Modify: `Justfile` — add tpm block inside `sync` recipe's shebang block
