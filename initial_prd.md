# Dotfiles Repository — PRD

**Author:** Shubhranshu  
**Status:** Draft  
**Target Machine:** MacBook Pro M5 Pro, 24GB RAM, macOS  
**Repo:** `github.com/you/dotfiles` → cloned to `~/Desktop/world/code/dotfiles/` → symlinked at `~/dotfiles`

---

## 1. Goal

A single Git repository that fully describes a personal development environment. Cloning the repo and running one script (`./install.sh`) on a fresh macOS machine should reproduce the complete setup — terminal, shell, editor, toolchains, aliases, fonts, themes, and system preferences — with no manual steps.

Changes are always made inside the repo. A `just`-based command surface (via `Justfile`) manages syncing those changes to the system and pushing them to remote.

---

## 2. Design Principles

- **Repo is the source of truth.** Never edit configs via their respective applications. Edit in `~/dotfiles`, then sync.
- **Stow for symlinks.** GNU Stow manages all symlinking. One package per tool.
- **Homebrew for packages.** A curated `Brewfile` installs all CLI tools and GUI apps. Nothing installed ad hoc.
- **Mise for runtimes.** One tool manages all language version pinning (Node, Python, Go, Rust).
- **Catppuccin Mocha everywhere.** Consistent theme across terminal, tmux, Neovim, VS Code, and Starship.
- **No plugin managers where avoidable.** Zsh plugins sourced directly via Homebrew. Keeps the shell fast and transparent.
- **`just` for task running.** A `Justfile` at the repo root replaces any custom CLI scripts. `just --list` is self-documenting; no help text to maintain.

---

## 3. Repository Structure

```
dotfiles/
├── Brewfile                  # All Homebrew packages and casks
├── Justfile                  # Task runner — sync, update, push, etc.
├── install.sh                # Bootstrap script for a fresh machine
│
├── macos/                    # macOS system defaults (not stowed)
│   └── defaults.sh
│
├── zsh/                      # Zsh shell config
│   ├── .zshrc
│   └── .zprofile
│
├── tmux/                     # Tmux config
│   └── .tmux.conf
│
├── nvim/                     # Neovim (hand-rolled Lua)
│   └── .config/
│       └── nvim/
│           ├── init.lua
│           └── lua/
│               ├── options.lua
│               ├── keymaps.lua
│               ├── plugins.lua
│               └── lsp.lua
│
├── git/                      # Git config and global ignore
│   ├── .gitconfig
│   └── .gitignore_global
│
├── ghostty/                  # Ghostty terminal config
│   └── .config/
│       └── ghostty/
│           └── config
│
├── starship/                 # Starship prompt
│   └── .config/
│       └── starship.toml
│
├── mise/                     # Mise runtime version manager
│   └── .config/
│       └── mise/
│           └── config.toml
│
└── vscode/                   # VS Code profiles and extensions
    ├── profiles/
    │   ├── ml.code-profile
    │   ├── java.code-profile
    │   ├── webdev.code-profile
    │   └── rust.code-profile
    ├── extensions-common.txt
    └── install-profiles.sh
```

---

## 4. Stow Packages

Each folder is a Stow package. `stow <package>` symlinks its contents into `$HOME`.

| Package | What it manages |
|---|---|
| `zsh` | `.zshrc`, `.zprofile` |
| `tmux` | `.tmux.conf` |
| `nvim` | `.config/nvim/` |
| `git` | `.gitconfig`, `.gitignore_global` |
| `ghostty` | `.config/ghostty/config` |
| `starship` | `.config/starship.toml` |
| `mise` | `.config/mise/config.toml` |

VS Code is handled separately via a symlink script (see §8) because its config lives in `~/Library/`.

The `Justfile` and `macos/defaults.sh` live at the repo root and are not stowed — they are invoked directly.

---

## 5. Brewfile

### CLI Tools
```ruby
brew "zsh" unless File.exist?("/bin/zsh")
brew "zsh-autosuggestions"
brew "zsh-syntax-highlighting"
brew "starship"
brew "zoxide"
brew "fzf"
brew "ripgrep"
brew "fd"
brew "bat"
brew "eza"
brew "jq"
brew "htop"
brew "stow"
brew "just"
brew "mise"
```

### Terminal & Multiplexer
```ruby
brew "tmux"
brew "tpm"                    # Tmux Plugin Manager
```

### Editor
```ruby
brew "neovim"
```

### Git
```ruby
brew "git"
brew "gh"
brew "lazygit"
```

### Fonts
```ruby
tap "homebrew/cask-fonts"
cask "font-jetbrains-mono-nerd-font"
```

### GUI Apps (Casks)
```ruby
cask "ghostty"
cask "visual-studio-code"
cask "raycast"
```

---

## 6. Tool Configurations

### 6.1 Ghostty

Native Metal-rendered terminal. Replaces WezTerm for simplicity — same Catppuccin theme, full ligature support, zero Lua required.

```ini
font-family = JetBrains Mono
font-size = 14
font-feature = calt
font-feature = liga
theme = catppuccin-mocha
window-padding-x = 8
window-padding-y = 8
background-opacity = 0.95
cursor-style = block
cursor-style-blink = false
mouse-hide-while-typing = true
shell-integration = detect
```

### 6.2 Zsh

Hand-rolled config. No plugin manager — plugins installed via Homebrew and sourced directly.

Key contents of `.zshrc`:

```zsh
# Plugins (sourced from Homebrew)
source $(brew --prefix)/share/zsh-autosuggestions/zsh-autosuggestions.zsh
source $(brew --prefix)/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh

# Tools
eval "$(starship init zsh)"
eval "$(zoxide init zsh)"
eval "$(mise activate zsh)"
[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh

# PATH
export PATH="$HOME/bin:$PATH"

# Dotfiles task runner — run from anywhere
alias dots="just --justfile ~/dotfiles/Justfile --working-directory ~/dotfiles"

# World shortcuts
export WORLD="$HOME/Desktop/world"
export CODE="$WORLD/code"
export PROJECTS="$CODE/projects"
alias world="cd $WORLD"
alias proj="cd $PROJECTS"

# Git aliases (curated from GitAlias)
alias gs="git status"
alias ga="git add"
alias gc="git commit"
alias gp="git push"
alias gl="git pull"
alias glo="git log --oneline --graph --decorate"
alias gco="git checkout"
alias gcb="git checkout -b"
alias gbd="git branch -d"
alias grb="git rebase"
alias gst="git stash"
alias gstp="git stash pop"
alias gd="git diff"
alias gds="git diff --staged"

# Tmux aliases
alias ta="tmux attach -t"
alias tn="tmux new -s"
alias tl="tmux list-sessions"
alias tk="tmux kill-session -t"

# Better defaults
alias ls="eza --icons"
alias ll="eza -la --icons"
alias cat="bat"
alias grep="rg"
```

### 6.3 Tmux

```conf
# Prefix
unbind C-b
set -g prefix `
bind ` send-prefix

# Vim navigation
setw -g mode-keys vi
bind h select-pane -L
bind j select-pane -D
bind k select-pane -U
bind l select-pane -R

# Theme
set -g @plugin 'catppuccin/tmux'
set -g @catppuccin_flavour 'mocha'

# Session persistence
set -g @plugin 'tmux-plugins/tmux-resurrect'
set -g @plugin 'tmux-plugins/tmux-continuum'
set -g @continuum-restore 'on'

# TPM
set -g @plugin 'tmux-plugins/tpm'
run '~/.tmux/plugins/tpm/tpm'
```

### 6.4 Starship

Minimal two-line prompt. Language versions shown contextually — only when inside a relevant project.

```toml
format = """
$directory$git_branch$git_status$cmd_duration
$character"""

[directory]
truncation_length = 3
truncate_to_repo = true

[git_branch]
symbol = " "
format = "[$symbol$branch]($style) "
style = "bold mauve"

[git_status]
ahead = "⇡${count}"
behind = "⇣${count}"
modified = "!${count}"
untracked = "?${count}"
staged = "+${count}"

[cmd_duration]
min_time = 2_000
format = "[ $duration]($style) "

[character]
success_symbol = "[❯](bold green)"
error_symbol = "[❯](bold red)"

[python]
detect_files = ["requirements.txt", "pyproject.toml", ".python-version"]

[rust]
detect_files = ["Cargo.toml"]

[nodejs]
detect_files = ["package.json"]

[golang]
detect_files = ["go.mod"]

[package]
disabled = true

[aws]
disabled = true
```

### 6.5 Git

```ini
[user]
  name = Shubhranshu
  email = shubhranshusingh.work@gmail.com

[core]
  editor = nvim
  excludesfile = ~/.gitignore_global
  autocrlf = input

[pull]
  rebase = true

[push]
  default = current

[alias]
  # Log
  lg = log --oneline --graph --decorate --all
  lp = log --pretty=format:'%C(yellow)%h %C(reset)%s %C(dim)%an, %ar'

  # Workflow
  undo = reset HEAD~1 --mixed
  wip = !git add -A && git commit -m 'wip'
  unwip = !git log -n 1 | grep -q 'wip' && git reset HEAD~1

  # Branch
  br = branch --format='%(HEAD) %(color:yellow)%(refname:short)%(color:reset) - %(contents:subject) %(color:green)(%(committerdate:relative))%(color:reset)'
  cleanup = !git branch --merged | grep -v '\\*\\|main\\|master\\|develop' | xargs -n 1 git branch -d
```

`.gitignore_global`:
```
.DS_Store
.env
.env.local
*.log
node_modules/
.idea/
.vscode/
*.swp
*.swo
__pycache__/
.pytest_cache/
.mypy_cache/
dist/
build/
.Trash/
```

### 6.6 Mise

```toml
[tools]
node = "22"
python = "3.12"
go = "1.23"

# Rust is managed by rustup for richer toolchain control
# mise just ensures rustup is present via Brewfile
```

---

## 7. VS Code Profiles

Four profiles sharing a common extension base. Managed via exported `.code-profile` files and an install script.

### Profile Matrix

| Extension | ML | Java | WebDev | Rust |
|---|:---:|:---:|:---:|:---:|
| **Vim** | ✓ | ✓ | ✓ | ✓ |
| **Catppuccin theme** | ✓ | ✓ | ✓ | ✓ |
| **GitLens** | ✓ | ✓ | ✓ | ✓ |
| **Error Lens** | ✓ | ✓ | ✓ | ✓ |
| **Even Better TOML** | ✓ | ✓ | ✓ | ✓ |
| Python + Pylance | ✓ | | ✓ | |
| Jupyter | ✓ | | | |
| Extension Pack for Java | | ✓ | | |
| Spring Boot Tools | | ✓ | | |
| ESLint + Prettier | | | ✓ | |
| Tailwind CSS IntelliSense | | | ✓ | |
| REST Client | | | ✓ | |
| rust-analyzer | | | | ✓ |
| CodeLLDB | | | | ✓ |

### `extensions-common.txt`
```
vscodevim.vim
catppuccin.catppuccin-vsc
eamodio.gitlens
usernamehw.errorlens
tamasfe.even-better-toml
```

### `install-profiles.sh`
```bash
#!/usr/bin/env bash
PROFILES=(ML Java WebDev Rust)

for profile in "${PROFILES[@]}"; do
  code --profile "$profile" \
    --import "$(dirname "$0")/profiles/${profile,,}.code-profile"
done

while read -r ext; do
  for profile in "${PROFILES[@]}"; do
    code --profile "$profile" --install-extension "$ext"
  done
done < "$(dirname "$0")/extensions-common.txt"
```

---

## 8. macOS System Defaults

Run once during bootstrap via `install.sh`. These are not stowed — they write directly to macOS preference files.

```bash
# Key repeat (essential for vim users)
defaults write NSGlobalDomain KeyRepeat -int 2
defaults write NSGlobalDomain InitialKeyRepeat -int 15

# Trackpad
# defaults write com.apple.driver.AppleBluetoothMultitouch.trackpad Clicking -bool true
# defaults write NSGlobalDomain com.apple.swipescrolldirection -bool false

# Finder
defaults write com.apple.finder AppleShowAllFiles -bool true
defaults write com.apple.finder ShowPathbar -bool true
defaults write NSGlobalDomain AppleShowAllExtensions -bool true

defaults write com.apple.finder FXPreferredViewStyle -string "Nlsv"
defaults write com.apple.finder _FXShowPosixPathInTitle -bool true

# Dock
defaults write com.apple.dock autohide -bool true
defaults write com.apple.dock autohide-delay -float 0
defaults write com.apple.dock show-recents -bool false
defaults write com.apple.dock tilesize -int 48

# Screenshots
defaults write com.apple.screencapture location "$HOME/Desktop/world/docs/screenshots"
defaults write com.apple.screencapture type -string "png"
defaults write com.apple.screencapture disable-shadow -bool true

# No .DS_Store on external drives
defaults write com.apple.desktopservices DSDontWriteNetworkStores -bool true
defaults write com.apple.desktopservices DSDontWriteUSBStores -bool true

# Mission Control
defaults write com.apple.dock mru-spaces -bool false

killall Dock Finder
```

---

## 9. The `Justfile`

Lives at `~/dotfiles/Justfile`. Invoked via the `dots` alias (defined in `.zshrc`) which points `just` at the dotfiles repo regardless of current directory.

```zsh
alias dots="just --justfile ~/dotfiles/Justfile --working-directory ~/dotfiles"
```

Running `dots` with no arguments prints all available recipes via `just --list`.

### `Justfile`

```just
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

# Adopt a loose file into a package: dots add ~/.somerc zsh
add file package:
    #!/usr/bin/env bash
    rel="${{file}#$HOME/}"
    dest="$(pwd)/{{package}}/$rel"
    mkdir -p "$(dirname "$dest")"
    mv "{{file}}" "$dest"
    stow --restow --target=$HOME {{package}}
    echo "✅ Adopted {{file}}"
```

### Command surface

| Command | What it does |
|---|---|
| `dots sync` | Restow all packages — apply repo changes to system |
| `dots update` | `git pull` + sync + `brew bundle` — full refresh from remote |
| `dots push "message"` | Stage all, commit with message, push to remote |
| `dots status` | `git status` inside the dotfiles repo |
| `dots edit` | Open dotfiles dir in `$EDITOR` |
| `dots list` | List all stow packages |
| `dots add <file> <pkg>` | Move a loose file into a package and stow it |

### Day-to-day workflow

```
Edit file in ~/dotfiles/
        ↓
dots sync          ← apply to system (for new files / restructures)
        ↓
(test the change)
        ↓
dots push "feat: update tmux prefix"   ← commit + push when satisfied
```

On another machine or after a fresh clone:

```
./install.sh       ← first time only
dots update        ← all subsequent refreshes
```

---

## 10. Bootstrap — `install.sh`

Runs top to bottom on a fresh macOS machine.

```bash
#!/usr/bin/env bash
set -e

DOTFILES="$HOME/dotfiles"

echo "🚀 Bootstrapping machine..."

# 1. Homebrew
if ! command -v brew &>/dev/null; then
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
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
```

---

## 11. Theme Consistency

Catppuccin Mocha applied uniformly across the stack:

| Layer | How |
|---|---|
| Ghostty | `theme = catppuccin-mocha` (built-in) |
| Tmux | `catppuccin/tmux` plugin, `flavour = mocha` |
| Neovim | `catppuccin/nvim` plugin |
| VS Code | Catppuccin VSC extension (all profiles) |
| Starship | `style = "bold mauve"` (Catppuccin palette) |

---

## 12. Out of Scope (for now)

- **Java toolchain** — deferred; add to Brewfile and Mise when needed
- **SSH config** — managed manually; never commit private keys
- **Secrets/tokens** — never in this repo; use macOS Keychain or a secrets manager
- **Nix/home-manager** — revisit if Linux dev machines enter the picture
- **Neovim LSP deep config** — incremental; built out as Lipi research progresses

---

## 13. Open Questions

- [ ] Which Git aliases to carry over from work Mac? (audit against GitAlias repo)
- [ ] Neovim plugin manager — lazy.nvim is the standard choice for hand-rolled configs
- [ ] TPM install path — confirm `~/.tmux/plugins/tpm` or manage via Brewfile
- [ ] Rustup integration with Mise — decide if Mise handles Rust version or defers fully to rustup
