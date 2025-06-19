# This script runs the whole pipeline via the CLI commands:
#   1. It creates accoutns for the client and the verifier
#   2. It spins a mock attestation server that serves a valid attestation report for
#       i.   image classification on the goldfish sample image
#       ii.  LLM inference on the prompt "Who is Lebron James?"
#   3. It runs the client and verifier
#   4. If no errors occur, it checks that the client and verifier's balances
#      updated accordingly at the end of each workflow

set -euo pipefail

go build

# Create new accounts for client and verifier, update .env
echo "1. Creating accounts for client and verifier..."
./vsl-rpc-demo gen-address client
./vsl-rpc-demo gen-address verifier
echo "> Client and verifier accounts created."
echo "=========================================="
source .env

# Get initial balances of client and verifier
echo "2. Checking initial balances of client and verifier..."
balance_output=$(./vsl-rpc-demo check-balance $CLIENT_ADDR 2>&1)
client_initial_balance=$(echo "$balance_output" | grep "Balance (in attos):" | awk '{print $NF}')

balance_output=$(./vsl-rpc-demo check-balance $VERIFIER_ADDR 2>&1)
verifier_initial_balance=$(echo "$balance_output" | grep "Balance (in attos):" | awk '{print $NF}')
echo "> Client initial balance: $client_initial_balance"
echo "> Verifier initial balance: $verifier_initial_balance"
echo "=========================================="

# Function to run end-to-end test
run_e2e_test() {
  local test_name="$1"
  local mock_report="$2"
  local fee="$3"
  local prev_client_balance="$4"
  local prev_verifier_balance="$5"
  shift 5
  local validate_fee=1

  # Run verifier in the background (it will keep polling until it verifies exactly 1 claim)
  echo "i. Running verifier in the background..."
  ./vsl-rpc-demo verifier --num-claims 1 &
  echo "> Verifier started in the background."

  echo "ii. Starting mock attestation server on port 6000..."
  { echo -ne "HTTP/1.0 200 OK\r\nContent-Length: $(wc -c <"$mock_report")\r\n\r\n"; cat "$mock_report"; } | nc -l 6000 &
  echo "> Mock attestation server started."

  echo "iii. Running client for $test_name..."
  ./vsl-rpc-demo client "$@" --fee "$fee" --zero-nonce
  echo "> Client for $test_name executed."

  echo "iv. Checking updated balances of client and verifier..."
  balance_output=$(./vsl-rpc-demo check-balance $CLIENT_ADDR 2>&1)
  client_balance=$(echo "$balance_output" | grep "Balance (in attos):" | awk '{print $NF}')

  balance_output=$(./vsl-rpc-demo check-balance $VERIFIER_ADDR 2>&1)
  verifier_balance=$(echo "$balance_output" | grep "Balance (in attos):" | awk '{print $NF}')
  echo "> Client updated balance: $client_balance"
  echo "> Verifier updated balance: $verifier_balance"

  echo "v. Checking if the client's and verifier's balances are updated as expected..."
  expected_client_balance=$(echo "$prev_client_balance - $fee - $validate_fee" | bc)
  if [ "$client_balance" != "$expected_client_balance" ]; then
      echo "Error: Client balance assertion failed"
      echo "Expected: $expected_client_balance"
      echo "Actual: $client_balance"
      exit 1
  fi
  expected_verifier_balance=$(echo "$prev_verifier_balance + $fee - $validate_fee" | bc)
  if [ "$verifier_balance" != "$expected_verifier_balance" ]; then
      echo "Error: Verifier balance assertion failed"
      echo "Expected: $expected_verifier_balance"
      echo "Actual: $verifier_balance"
      exit 1
  fi
  echo "> Client and verifier balances updated as expected."
}

# Run end-to-end test for image classification
echo "4. Running end-to-end test for image classification..."
run_e2e_test "image classification" \
    "tests/goldfish_report.out" \
    100 \
    "$client_initial_balance" \
    "$verifier_initial_balance" \
    img_class --img ../common/attester/inference/src/sample/goldfish.jpeg
echo "=========================================="

# Get updated balances after the first test
balance_output=$(./vsl-rpc-demo check-balance $CLIENT_ADDR 2>&1)
client_updated_balance=$(echo "$balance_output" | grep "Balance (in attos):" | awk '{print $NF}')
balance_output=$(./vsl-rpc-demo check-balance $VERIFIER_ADDR 2>&1)
verifier_updated_balance=$(echo "$balance_output" | grep "Balance (in attos):" | awk '{print $NF}')

# Run end-to-end test for LLM inference
echo "5. Running end-to-end test for LLM inference..."
run_e2e_test "LLM inference" \
  "tests/prompt_report.out" \
  200 \
  "$client_updated_balance" \
  "$verifier_updated_balance" \
  llama --prompt "'Who is Lebron James?'"
echo "==========================================="
