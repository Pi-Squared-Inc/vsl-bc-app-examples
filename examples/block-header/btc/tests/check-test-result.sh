#!/bin/bash
set -e

echo "=== EXTRACTING LOGS FROM CONTAINERS ==="
SUBMITTER_CONTAINER=$(docker ps -a --format "{{.Names}}" | grep claim-submitter | head -1)
VERIFIER_CONTAINER=$(docker ps -a --format "{{.Names}}" | grep claim-verifier | head -1)

if [[ -z "$SUBMITTER_CONTAINER" ]]; then
    echo "Error: claim-submitter container not found"
    echo "Available containers:"
    docker ps -a --format "table {{.Names}}\t{{.Status}}"
    exit 1
fi

if [[ -z "$VERIFIER_CONTAINER" ]]; then
    echo "Error: claim-verifier container not found"
    echo "Available containers:"
    docker ps -a --format "table {{.Names}}\t{{.Status}}"
    exit 1
fi

echo "Found containers:"
echo "  - Submitter: $SUBMITTER_CONTAINER"
echo "  - Verifier:  $VERIFIER_CONTAINER"

SUBMITTER_LOG=$(docker logs "$SUBMITTER_CONTAINER" 2>&1)
VERIFIER_LOG=$(docker logs "$VERIFIER_CONTAINER" 2>&1)

echo "=== CLAIM-SUBMITTER LOGS ==="
echo "$SUBMITTER_LOG"
echo ""

echo "=== CLAIM-VERIFIER LOGS ==="
echo "$VERIFIER_LOG"
echo ""

echo "=== ANALYZING LOGS ==="
submitter_claim_id=$(echo "$SUBMITTER_LOG" | grep 'Claim submitted successfully with ID:' | sed -E 's/.*Claim submitted successfully with ID: ([^ ]+).*/\1/' | head -1)

if [[ -z "$submitter_claim_id" ]]; then
    echo "Error: No claim ID found in logs"
    exit 1
fi

echo "Extracted IDs:"
echo "  - Submitter claim ID: '$submitter_claim_id'"

echo "=== READING ADDRESSES FROM ENV FILES ==="

RPC_URL="http://localhost:44444"

SUBMITTER_ADDRESS=$(grep "^SUBMITTER_ADDRESS=" "$PWD/claim-verifier/.env" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'" | head -1)
VERIFIER_ADDRESS=$(grep "^VERIFIER_ADDRESS=" "$PWD/claim-submitter/.env" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'" | head -1)

echo "Addresses:"
echo "  - Submitter: $SUBMITTER_ADDRESS"
echo "  - Verifier:  $VERIFIER_ADDRESS"

SEC=0
NANOS=0

TIMESTAMP_JSON="{\"seconds\":$SEC,\"nanos\":$NANOS}"

echo "=== RPC: listSubmittedClaimsForReceiver (Verifier) ==="
SUBMITTED_CLAIMS=$(curl -s -X POST -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"vsl_listSubmittedClaimsForReceiver\",\"params\":{\"address\":\"$VERIFIER_ADDRESS\",\"since\":$TIMESTAMP_JSON}}" \
  "$RPC_URL")
  
if ! echo "$SUBMITTED_CLAIMS" | grep -q "$submitter_claim_id"; then
    echo "FAILURE: Claim ID not in submitted claims for verifier"
    exit 1
fi
echo "Claim found in verifier's submitted claims"

echo "=== RPC: listSettledClaimsForReceiver (Submitter) ==="
SETTLED_CLAIMS=$(curl -s -X POST -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"vsl_listSettledClaimsForReceiver\",\"params\":{\"address\":\"$SUBMITTER_ADDRESS\",\"since\":$TIMESTAMP_JSON}}" \
  "$RPC_URL")

if ! echo "$SETTLED_CLAIMS" | grep -q "$submitter_claim_id"; then
    echo "FAILURE: Claim ID not in settled claims for submitter"
    exit 1
fi

echo "Claim found in submitter's settled claims"
echo "=== FINAL VERIFICATION COMPLETE ==="
echo "All checks passed successfully!"
