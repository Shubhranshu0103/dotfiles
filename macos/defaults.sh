#!/usr/bin/env bash
set -e

echo "🍎 Applying macOS defaults..."

# Key repeat (essential for vim users)
defaults write NSGlobalDomain KeyRepeat -int 2
defaults write NSGlobalDomain InitialKeyRepeat -int 15

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
mkdir -p "$HOME/Desktop/world/docs/screenshots"
defaults write com.apple.screencapture location "$HOME/Desktop/world/docs/screenshots"
defaults write com.apple.screencapture type -string "png"
defaults write com.apple.screencapture disable-shadow -bool true

# No .DS_Store on external drives
defaults write com.apple.desktopservices DSDontWriteNetworkStores -bool true
defaults write com.apple.desktopservices DSDontWriteUSBStores -bool true

# Mission Control — don't rearrange spaces by recent use
defaults write com.apple.dock mru-spaces -bool false

killall Dock Finder

echo "✅ macOS defaults applied."
