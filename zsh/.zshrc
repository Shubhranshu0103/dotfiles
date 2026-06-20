# Plugins (sourced from Homebrew)
source $(brew --prefix)/share/zsh-autosuggestions/zsh-autosuggestions.zsh
source $(brew --prefix)/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh

# Tools
eval "$(starship init zsh)"
eval "$(zoxide init zsh)"
eval "$(mise activate zsh)"
[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh

# PATH
export PATH="$HOME/.local/bin:$HOME/bin:$PATH"

# Dotfiles task runner — run from anywhere
alias dots="just --justfile ~/dotfiles/Justfile --working-directory ~/dotfiles"

# World shortcuts
export WORLD="$HOME/Desktop/world"
export CODE="$WORLD/code"
export PROJECTS="$CODE/projects"
alias world="cd $WORLD"
alias proj="cd $PROJECTS"

# Git aliases
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

# Shell reload
alias reload="source ~/.zshrc"

# Tree — t [depth] [flags] in any order
t() {
  local depth=2
  local args=()
  for arg in "$@"; do
    if [[ "$arg" =~ ^[0-9]+$ ]]; then
      depth="$arg"
    else
      args+=("$arg")
    fi
  done
  tree -L "$depth" "${args[@]}"
}

# Better defaults
alias ls="eza --icons"
alias ll="eza -la --icons"
alias cat="bat"
alias grep="rg"
