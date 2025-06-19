#!/bin/bash

set -e

CONTEXT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# Run the wormhole tests
$CONTEXT_ROOT/wormhole.sh

# Add more tests here