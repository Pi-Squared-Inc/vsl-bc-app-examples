#!/bin/bash

set -e

CONTEXT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

$CONTEXT_ROOT/stop_verifier.sh

pushd $CONTEXT_ROOT/../../../ >/dev/null

# Build verifier Docker image
cd $CONTEXT_ROOT/../../../verifier
docker build -q --build-context root=../../../ -t verifier:latest . >/dev/null

# Run verifier Docker container
docker run -d --net vsl-network --name verifier --rm \
  -e VSL_RPC=http://vsl-core:44444 \
  -e VSL_VERIFIER_ADDRESS=0xB078F143F926fa85Bcf455AF78846321b2c5F1A6 \
  -e VSL_VERIFIER_PRIVATE_KEY=0a06f5103d2b4584f3d057e32d5540025cda8181b371469ae69b5e2212f4722d \
  verifier:latest >/dev/null

popd >/dev/null
