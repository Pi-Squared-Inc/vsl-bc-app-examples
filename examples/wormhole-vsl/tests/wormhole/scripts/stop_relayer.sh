#!/bin/bash

CONTEXT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# Stop claim generator if it is running
if docker ps -q -f name=relayer | grep -q .; then
  docker stop relayer >/dev/null
fi
