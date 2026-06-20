PACKAGES := "zsh tmux nvim git ghostty starship mise"

# Show all available commands
default:
    @just --list

# Restow all packages — apply repo changes to system
sync:
    @echo "🔄 Syncing..."
    stow --restow --target=$HOME {{PACKAGES}}
    @echo "✅ Done"

# Pull remote + sync + brew bundle
update:
    @echo "⬇️  Pulling latest..."
    git pull origin main
    just sync
    brew bundle --file=Brewfile
    @echo "✅ Updated"

# Commit all changes and push: dots push "my message"
push message:
    git add -A
    git commit -m "{{message}}"
    git push origin main
    @echo "✅ Pushed"

# Git status
status:
    git status

# Open dotfiles in $EDITOR
edit:
    $EDITOR .

# List stow packages
list:
    @echo "{{PACKAGES}}"

# Lint scripts and simulate stow — no changes made to system
check:
    @echo "🔍 Running shellcheck..."
    shellcheck install.sh macos/defaults.sh vscode/install-profiles.sh
    @echo "📦 Simulating stow..."
    stow --simulate --restow --target=$HOME {{PACKAGES}}
    @echo "✅ All checks passed"

# Adopt a loose file into a package: dots add ~/.somerc zsh
add file package:
    #!/usr/bin/env bash
    rel="${{file}#$HOME/}"
    dest="$(pwd)/{{package}}/$rel"
    mkdir -p "$(dirname "$dest")"
    mv "{{file}}" "$dest"
    stow --restow --target=$HOME {{package}}
    echo "✅ Adopted {{file}}"
