#!/bin/bash

CONTEXT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# Stop verifier if it is running
if docker ps -q -f name=verifier | grep -q .; then
  docker stop verifier >/dev/null
fi
