#!/usr/bin/env bash
set -e

install_extensions() {
  local profile="$1"
  shift
  while read -r ext; do
    code --profile "$profile" --install-extension "$ext" --force
  done < "$(dirname "$0")/extensions-common.txt"
  for ext in "$@"; do
    code --profile "$profile" --install-extension "$ext" --force
  done
}

install_extensions "ML" \
  ms-python.python ms-python.vscode-pylance ms-toolsai.jupyter

install_extensions "Java" \
  vscjava.vscode-java-pack vmware.vscode-spring-boot

install_extensions "WebDev" \
  dbaeumer.vscode-eslint esbenp.prettier-vscode bradlc.vscode-tailwindcss humao.rest-client ms-python.python ms-python.vscode-pylance

install_extensions "Rust" \
  rust-lang.rust-analyzer vadimcn.vscode-lldb
