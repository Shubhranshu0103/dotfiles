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
    if ! ./bin/dotsync check brew; then
      brew bundle --file=Brewfile
      ./bin/dotsync record brew
    fi
    for pkg in {{PACKAGES}}; do
      if ! ./bin/dotsync check "stow_${pkg}"; then
        stow --restow --target=$HOME "$pkg"
        ./bin/dotsync record "stow_${pkg}"
      fi
    done
    if ! ./bin/dotsync check tpm; then
      if [ -f "$HOME/.tmux/plugins/tpm/scripts/install_plugins.sh" ]; then
        TMUX_PLUGIN_MANAGER_PATH="$HOME/.tmux/plugins/" \
          ~/.tmux/plugins/tpm/scripts/install_plugins.sh
        ./bin/dotsync record tpm
      else
        echo "⚠️  TPM not installed — skipping plugin sync (run install.sh first)"
      fi
    fi
    echo "✅ Sync done"

# Pull remote main, rebase staging, then sync
update:
    #!/usr/bin/env bash
    set -e
    echo "⬇️  Pulling latest..."
    git checkout main
    git pull origin main
    if git show-ref --verify --quiet refs/heads/staging; then
      git checkout staging
      git rebase main
      git checkout main
    fi
    just sync
    echo "✅ Updated"

# Export VS Code profiles, commit everything to staging, push for PR review: dots push "my message"
push message:
    #!/usr/bin/env bash
    set -e
    just vscode-export
    git checkout staging 2>/dev/null || git checkout -b staging
    git add -A
    git commit -m "{{message}}"
    git push origin staging
    git checkout main
    echo "✅ Pushed to staging — approve the PR on GitHub to merge to main"

# Apply macOS system preferences (run after editing macos/defaults.sh)
macos:
    #!/usr/bin/env bash
    set -e
    if ! ./bin/dotsync check macos; then
      bash macos/defaults.sh
      ./bin/dotsync record macos
    else
      echo "macos defaults already up-to-date"
    fi

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
    set -e
    VSCODE_USER="$HOME/Library/Application Support/Code/User"
    STORAGE="$VSCODE_USER/globalStorage/storage.json"
    for profile in ML Java WebDev Rust; do
      location=$(jq -r --arg n "$profile" '.userDataProfiles[] | select(.name == $n) | .location' "$STORAGE")
      if [ -z "$location" ] || [ "$location" = "null" ]; then
        echo "⚠️  $profile: not found in VS Code storage, skipping"
        continue
      fi
      ext_file="$VSCODE_USER/profiles/$location/extensions.json"
      if [ ! -f "$ext_file" ]; then
        echo "⚠️  $profile: extensions file missing, skipping"
        continue
      fi
      extensions=$(jq -r '.[].identifier.id' "$ext_file" | tr '[:upper:]' '[:lower:]' | sort | tr '\n' ' ' | sed 's/ $//')
      lower=$(echo "$profile" | tr '[:upper:]' '[:lower:]')
      printf '{"name":"%s","extensions":"%s"}\n' "$profile" "$extensions" > "vscode/profiles/${lower}.code-profile"
      echo "✓ $profile"
    done
    default_ext="$HOME/.vscode/extensions/extensions.json"
    if [ -f "$default_ext" ]; then
      extensions=$(jq -r '.[].identifier.id' "$default_ext" | tr '[:upper:]' '[:lower:]' | sort | tr '\n' ' ' | sed 's/ $//')
      printf '{"name":"Default","extensions":"%s"}\n' "$extensions" > "vscode/profiles/default.code-profile"
      echo "✓ Default"
    fi
    ./bin/dotsync record vscode_profiles
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
