#!/bin/bash

set -e

CONTEXT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
SRC_RPC_URL="http://127.0.0.1:8545"
DEST_RPC_URL="http://127.0.0.1:8546"

start_chains() {
  docker compose up -d --build >/dev/null
}

check_block_number() {
  local endpoint=$1
  while true; do
    BLOCK_NUMBER_HEX=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}' $endpoint | jq -r '.result')
    BLOCK_NUMBER=$((BLOCK_NUMBER_HEX))
    if [ "$BLOCK_NUMBER" -gt 0 ]; then
      break
    else
      echo "Block number is $BLOCK_NUMBER, waiting..."
      sleep 2
    fi
  done
}

run_tests() {
  pipenv install >/dev/null
  pipenv run python -m unittest -v wormhole.py
}

tear_down() {
  ./scripts/stop_observer.sh
  ./scripts/stop_verifier.sh
  ./scripts/stop_relayer.sh
  docker compose down
  docker volume prune -f >/dev/null
  docker network prune -f >/dev/null
  docker image prune -f >/dev/null
}

main() {
  pushd $CONTEXT_ROOT/wormhole >/dev/null
  start_chains
  check_block_number $SRC_RPC_URL
  check_block_number $DEST_RPC_URL
  source scripts/deploy_contracts.sh
  scripts/start_observer.sh $SRC_VSL_ADDRESS
  scripts/start_verifier.sh
  scripts/start_relayer.sh $DEST_VSL_ADDRESS
  run_tests
  tear_down
  popd >/dev/null
}

main
