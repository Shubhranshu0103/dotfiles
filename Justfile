PACKAGES := "zsh tmux nvim git ghostty starship mise"

# Show all available commands
default:
    @just --list

# Build the dotsync binary (run after editing tools/dotsync/)
build-dotsync:
    cd tools/dotsync && GOOS=darwin GOARCH=arm64 go build -o ../../bin/dotsync .
    @echo "✅ bin/dotsync built"

# Restow changed packages only
sync:
    #!/usr/bin/env bash
    set -e
    for pkg in {{PACKAGES}}; do
      if ! ./bin/dotsync check "stow_${pkg}"; then
        stow --restow --target=$HOME "$pkg"
        ./bin/dotsync record "stow_${pkg}"
      fi
    done
    echo "✅ Sync done"

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

# Show step status and VS Code drift
status:
    ./bin/dotsync status

# Rollback state to a previous commit snapshot
rollback commit:
    ./bin/dotsync rollback {{commit}}

# Remove orphaned state entries
gc:
    ./bin/dotsync gc

# Export current VS Code profiles back to repo (run after UI changes)
vscode-export:
    #!/usr/bin/env bash
    for profile in ML Java WebDev Rust; do
      code --profile "$profile" \
        --export-profile "$(pwd)/vscode/profiles/${profile,,}.code-profile"
    done
    echo "✅ Profiles exported — review with git diff, then dots push"

# Git status
status-git:
    git status

# Open dotfiles in $EDITOR
edit:
    $EDITOR .

# List stow packages
list:
    @echo "{{PACKAGES}}"

# Remove all stow-managed symlinks from $HOME
nuke:
    @echo "🗑️  Removing stow-managed symlinks..."
    stow --delete --target=$HOME {{PACKAGES}}
    @echo "✅ Done. Run 'just sync' to re-link."

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
    file="{{file}}"
    rel="${file#$HOME/}"
    dest="$(pwd)/{{package}}/$rel"
    mkdir -p "$(dirname "$dest")"
    mv "$file" "$dest"
    stow --restow --target=$HOME {{package}}
    echo "✅ Adopted {{file}}"
