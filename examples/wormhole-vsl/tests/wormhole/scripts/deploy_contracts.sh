#!/bin/bash

set -e

deploy_contracts() {
  local CONTEXT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
  local WORMHOLE_DIR=$CONTEXT_ROOT/../../../

  pushd $WORMHOLE_DIR >/dev/null

  # Set up the environment
  export INBOUND_LIMIT=1000
  export TRANSFER_AMOUNT=100
  export SRC_CHAIN_ID=1
  export SRC_RPC_URL="http://localhost:8545"
  export SRC_PRIVATE_KEY="0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
  export DEST_CHAIN_ID=2
  export DEST_RPC_URL="http://localhost:8546"
  export DEST_PRIVATE_KEY="0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

  deploy_and_extract() {
    local IS_DEST=$1
    local RPC_URL=$2
    local output=$(forge clean && IS_DEST=$IS_DEST forge script --via-ir $WORMHOLE_DIR/script/Deploy.s.sol --rpc-url $RPC_URL --broadcast --slow)
    echo "$output" | awk -v prefix=$3 '
      /Token address:/ {print prefix "_TOKEN_ADDRESS=" $NF}
      /VSL address:/ {print prefix "_VSL_ADDRESS=" $NF}
      /NttManager address:/ {print prefix "_MANAGER_ADDRESS=" $NF}
      /Transceiver address:/ {print prefix "_TRANSCEIVER_ADDRESS=" $NF}
    '
  }

  eval $(deploy_and_extract false $SRC_RPC_URL "SRC")
  eval $(deploy_and_extract true $DEST_RPC_URL "DEST")

  # Export environment variables
  export SRC_TOKEN_ADDRESS SRC_VSL_ADDRESS SRC_MANAGER_ADDRESS SRC_TRANSCEIVER_ADDRESS
  export DEST_TOKEN_ADDRESS DEST_VSL_ADDRESS DEST_MANAGER_ADDRESS DEST_TRANSCEIVER_ADDRESS

  # Setup the contracts
  forge clean && IS_DEST=false forge script $WORMHOLE_DIR/script/Setup.s.sol --rpc-url $SRC_RPC_URL --broadcast --slow >/dev/null
  forge clean && IS_DEST=true forge script $WORMHOLE_DIR/script/Setup.s.sol --rpc-url $DEST_RPC_URL --broadcast --slow >/dev/null

  echo "Deployed contracts successfully"
  echo "SRC_TOKEN_ADDRESS: $SRC_TOKEN_ADDRESS"
  echo "SRC_VSL_ADDRESS: $SRC_VSL_ADDRESS"
  echo "SRC_MANAGER_ADDRESS: $SRC_MANAGER_ADDRESS"
  echo "SRC_TRANSCEIVER_ADDRESS: $SRC_TRANSCEIVER_ADDRESS"
  echo "DEST_TOKEN_ADDRESS: $DEST_TOKEN_ADDRESS"
  echo "DEST_VSL_ADDRESS: $DEST_VSL_ADDRESS"
  echo "DEST_MANAGER_ADDRESS: $DEST_MANAGER_ADDRESS"
  echo "DEST_TRANSCEIVER_ADDRESS: $DEST_TRANSCEIVER_ADDRESS"

  popd >/dev/null
}

deploy_contracts
