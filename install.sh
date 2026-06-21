#!/usr/bin/env bash
set -e

DOTFILES="$(cd "$(dirname "$0")" && pwd)"
DOTSYNC="$DOTFILES/bin/dotsync"

echo "🚀 Bootstrapping machine..."

# 1. Homebrew — self-guarded
if ! command -v brew &>/dev/null; then
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi
if [ -f /opt/homebrew/bin/brew ]; then
  eval "$(/opt/homebrew/bin/brew shellenv)"
elif [ -f /usr/local/bin/brew ]; then
  eval "$(/usr/local/bin/brew shellenv)"
fi

# 2. Packages
if ! "$DOTSYNC" check brew; then
  brew bundle --verbose --file="$DOTFILES/Brewfile"
  "$DOTSYNC" record brew
fi

# 3. Symlink ~/dotfiles — idempotent
if [ ! -e "$HOME/dotfiles" ]; then
  ln -sf "$DOTFILES" "$HOME/dotfiles"
fi

# 4. Stow — per-package so only changed packages restow
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

# 6. TPM — clone is self-guarded; plugins hashed
if [ ! -d "$HOME/.tmux/plugins/tpm" ]; then
  git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm
fi
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
if [ "$SHELL" != "$(which zsh)" ]; then
  chsh -s "$(which zsh)"
fi

echo "✅ Done. Restart terminal."
