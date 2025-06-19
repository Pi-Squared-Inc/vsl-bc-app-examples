#!/usr/bin/env bash
# filepath: /home/jin/vsl-ai-clients/example/set-up-env-file.sh

set -e

# List of directories containing sample.env
DIRS=(
  "vsl-rpc-demo"
  "vsl-rpc-demo/cmd/backend-server"
  "frontend"
  "common/attester"
)

for dir in "${DIRS[@]}"; do
  if [ -f "$dir/sample.env" ]; then
    cp "$dir/sample.env" "$dir/.env"
    echo "Copied $dir/sample.env to $dir/.env"
  else
    echo "Warning: $dir/sample.env not found, skipping."
  fi
done