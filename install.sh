#!/usr/bin/env bash
set -e

DOTFILES="$HOME/dotfiles"

echo "🚀 Bootstrapping machine..."

# 1. Homebrew
if ! command -v brew &>/dev/null; then
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

# Add Homebrew to PATH for the remainder of this script
if [ -f /opt/homebrew/bin/brew ]; then
  eval "$(/opt/homebrew/bin/brew shellenv)"
elif [ -f /usr/local/bin/brew ]; then
  eval "$(/usr/local/bin/brew shellenv)"
fi

# 2. Packages
brew bundle --file="$DOTFILES/Brewfile"

# 3. Stow all packages
cd "$DOTFILES"
stow --restow zsh tmux nvim git ghostty starship mise

# 4. Tmux plugins
if [ ! -d "$HOME/.tmux/plugins/tpm" ]; then
  git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm
fi
~/.tmux/plugins/tpm/scripts/install_plugins.sh

# 5. VS Code profiles and extensions
if command -v code &>/dev/null; then
  bash "$DOTFILES/vscode/install-profiles.sh"
fi

# 6. VS Code settings symlink
VSCODE_DIR="$HOME/Library/Application Support/Code/User"
mkdir -p "$VSCODE_DIR"
ln -sf "$DOTFILES/vscode/settings.json" "$VSCODE_DIR/settings.json"
ln -sf "$DOTFILES/vscode/keybindings.json" "$VSCODE_DIR/keybindings.json"

# 7. macOS defaults
bash "$DOTFILES/macos/defaults.sh"

# 8. Claude Code
if ! command -v claude &>/dev/null; then
  curl -fsSL https://claude.ai/install.sh | bash
fi

# 9. Default shell
if [ "$SHELL" != "$(which zsh)" ]; then
  chsh -s "$(which zsh)"
fi

echo "✅ Done. Restart terminal."
