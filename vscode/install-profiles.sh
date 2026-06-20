#!/usr/bin/env bash
set -e

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
