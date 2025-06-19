#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

CONTEXT_DIR=$(pwd)
TEST_CHAIN_ID=1
HTTP_RPC_ENDPOINT=http://mock-rpc:8545
WS_RPC_ENDPOINT=ws://mock-rpc:8546
MAX_RETRIES=15
RETRY_DELAY=2

wait_for_rpc_endpoint() {
  echo "Waiting for HTTP RPC endpoint at $HTTP_RPC_ENDPOINT..."
  for ((i = 1; i <= MAX_RETRIES; i++)); do
    response=$(curl -s -X POST --connect-timeout 2 --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' -H "Content-Type: application/json" $HTTP_RPC_ENDPOINT)
    exit_code=$?
    if [ $exit_code -eq 0 ] && ! echo "$response" | grep -q '"error":'; then
      echo "HTTP RPC endpoint $HTTP_RPC_ENDPOINT is up! Response: $response"
      return 0
    fi
    echo "Attempt $i/$MAX_RETRIES failed for HTTP RPC ($HTTP_RPC_ENDPOINT). Retrying in $RETRY_DELAY second(s)..."
    sleep $RETRY_DELAY
  done
  echo "Error: HTTP RPC endpoint $HTTP_RPC_ENDPOINT did not become available."
  exit 1
}

# --- Run Integration Tests ---
echo "Running integration tests..."

echo "Sending test notification to HTTP RPC endpoint..."
curl -X POST $HTTP_RPC_ENDPOINT/notify
echo "Test notification sent."

# Run the check-test-result-geth.sh for Geth
./check-test-result-geth.sh
# Run the check-test-result-Reth.sh for Reth
./check-test-result-reth.sh
./check-test-result-btc.sh
# Add other check-test-result-<service>.sh here

echo "Integration tests finished."
