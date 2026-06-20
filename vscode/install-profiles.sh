#!/usr/bin/env bash
set -e

PROFILES=(ML Java WebDev Rust)

declare -A PROFILE_EXTENSIONS
PROFILE_EXTENSIONS[ML]="ms-python.python ms-python.vscode-pylance ms-toolsai.jupyter"
PROFILE_EXTENSIONS[Java]="vscjava.vscode-java-pack vmware.vscode-spring-boot"
PROFILE_EXTENSIONS[WebDev]="dbaeumer.vscode-eslint esbenp.prettier-vscode bradlc.vscode-tailwindcss humao.rest-client ms-python.python ms-python.vscode-pylance"
PROFILE_EXTENSIONS[Rust]="rust-lang.rust-analyzer vadimcn.vscode-lldb"

for profile in "${PROFILES[@]}"; do
  # Install common extensions
  while read -r ext; do
    code --profile "$profile" --install-extension "$ext" --force
  done < "$(dirname "$0")/extensions-common.txt"

  # Install profile-specific extensions
  for ext in ${PROFILE_EXTENSIONS[$profile]}; do
    code --profile "$profile" --install-extension "$ext" --force
  done
done
