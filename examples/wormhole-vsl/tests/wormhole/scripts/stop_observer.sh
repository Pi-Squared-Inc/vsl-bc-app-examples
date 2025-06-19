#!/bin/bash

CONTEXT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# Stop observer if it is running
if docker ps -q -f name=observer | grep -q .; then
  docker stop observer >/dev/null
fi
