#!/bin/bash

set -e

# Check if DEST_VSL_ADDRESS is provided
if [ -z "$1" ]; then
  echo "Error: DEST_VSL_ADDRESS is not provided."
  exit 1
fi

CONTEXT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

$CONTEXT_ROOT/stop_relayer.sh

pushd $CONTEXT_ROOT/../../../ >/dev/null

cd $CONTEXT_ROOT/../../../relayer

# Build relayer Docker image
docker build -q --build-context root=../../../ -t relayer:latest . >/dev/null

# Run relayer Docker container
docker run -d --net vsl-network --name relayer --rm \
  -e DEST_PK=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
  -e DEST_RPC=http://dest-chain:8545 \
  -e DEST_VSL_ADDRESS=$1 \
  -e VSL_RPC=http://vsl-core:44444 \
  -e VSL_CLIENT_PRIVATE_KEY=a3a0fd7cf0ef00256689d7e77262ea5ebac5349f2c724b15489ea7725d440338 \
  -e WORMHOLE_BACKEND_API_ENDPOINT=http://backend:3001 \
  relayer:latest >/dev/null

popd >/dev/null
