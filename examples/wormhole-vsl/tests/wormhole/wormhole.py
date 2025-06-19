import os
import subprocess
import time
import unittest

import requests

VSL_RPC = "http://127.0.0.1:44444"
OBSERVER_ENDPOINT = "http://127.0.0.1:10001"
BACKEND_ENDPOINT = "http://127.0.0.1:3001"
ADDRESS = "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))


def run_command(command, cwd):
    try:
        result = subprocess.run(
            command,
            cwd=cwd,
            shell=True,
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(e.stderr)
        return None


def mint_token():
    command = "IS_DEST=false forge script script/Operations.s.sol --rpc-url $SRC_RPC_URL --broadcast --slow --sig 'mintToken()'"
    return run_command(command, f"{SCRIPT_DIR}/../../") is not None


def get_account_balance(chain="source"):
    if chain == "source":
        command = "IS_DEST=false forge script script/Operations.s.sol --rpc-url $SRC_RPC_URL --broadcast --slow --sig 'checkBalance()' | awk '/Balance: / {print $2}'"
    else:
        command = "IS_DEST=true forge script script/Operations.s.sol --rpc-url $DEST_RPC_URL --broadcast --slow --sig 'checkBalance()' | awk '/Balance: / {print $2}'"

    try:
        result = subprocess.run(
            command,
            cwd=f"{SCRIPT_DIR}/../../",
            shell=True,
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        output = result.stdout.strip()
        return int(output)
    except subprocess.CalledProcessError as e:
        print(f"Failed to check balance on {chain} chain. {e.stderr}")
        return None


def transfer_token():
    command = "IS_DEST=false forge script script/Operations.s.sol --rpc-url $SRC_RPC_URL --broadcast --slow --sig 'transfer()'"
    if run_command(command, f"{SCRIPT_DIR}/../../") is None:
        return None

    command = "cat broadcast/Operations.s.sol/31337/transfer-latest.json | jq -r '.transactions[1].hash'"
    return run_command(command, f"{SCRIPT_DIR}/../../")


def fetch_claim_from_database():
    try:
        response = requests.get(f"{BACKEND_ENDPOINT}/claims?address={ADDRESS}&limit=1")
        if response.status_code == 200:
            claims = response.json()
            if claims and len(claims) > 0:
                claim_id = claims[0].get("claim_id")
                return claim_id
        return None
    except Exception as e:
        print(f"Request failed with error: {e}")
        return None


def fetch_claim_from_vsl(claim_id):
    try:
        response = requests.post(
            f"{VSL_RPC}",
            json={
                "jsonrpc": "2.0",
                "method": "vsl_getSettledClaimById",
                "params": {
                    "claim_id": claim_id,
                },
                "id": 1,
            },
        )
        if response.status_code == 200:
            result = response.json()
            # Check for JSON-RPC errors
            if "error" in result:
                print(f"VSL error: {result['error']}")
                return False
            # Check if we have a valid JSON-RPC response with result
            if (
                result.get("jsonrpc") == "2.0"
                and "result" in result
                and result["result"] is not None
            ):
                return True
        return False
    except Exception as e:
        print(f"VSL request failed with error: {e}")
        return False


def generate_claim(tx_hash):
    response = requests.post(
        f"{OBSERVER_ENDPOINT}/generate_claim", json={"transaction_hash": tx_hash}
    )
    return response.status_code == 200


class TestWormhole(unittest.TestCase):
    def test_transfer(self):
        max_attempts = 10

        self.assertTrue(mint_token(), "Failed to mint token on source chain.")

        balance_on_source1 = get_account_balance(chain="source")
        self.assertEqual(
            balance_on_source1,
            100000000000000000000,
            "Balance on source chain is not correct.",
        )

        tx_hash = transfer_token()
        self.assertIsNotNone(tx_hash, "Failed to get transaction hash.")

        balance_on_source2 = get_account_balance(chain="source")
        self.assertEqual(
            balance_on_source2, 0, "Balance on source chain is not correct."
        )

        self.assertTrue(generate_claim(tx_hash), "Failed to generate claim.")

        # Check if the claim was created in the database
        print("Fetching claim from database...")
        claim_id = None
        for i in range(max_attempts):
            print(f"Attempt {i + 1}/{max_attempts}")
            claim_id = fetch_claim_from_database()
            if claim_id:
                break
            time.sleep(5)
        else:
            self.fail("Failed to fetch claim from database after 10 attempts.")
        print(f"Successfully retrieved claim from database: {claim_id}")

        # Check if the claim was created in the VSL
        print(f"Fetching claim {claim_id} from VSL...")
        claim_found_in_vsl = False
        for i in range(max_attempts):
            print(f"Attempt {i + 1}/{max_attempts}")
            claim_found_in_vsl = fetch_claim_from_vsl(claim_id)
            if claim_found_in_vsl:
                break
            time.sleep(5)
        else:
            self.fail("Failed to fetch claim from VSL after 10 attempts.")

        print(f"Successfully retrieved claim from VSL: {claim_id}")

        # Poll destination chain balance every 5 seconds for up to 10 attempts
        # Return True if balance becomes non-zero, False if it remains zero after all attempts
        print("Checking balance on destination chain...")
        max_attempts = 20
        for i in range(max_attempts):
            print(f"Attempt {i + 1}/{max_attempts}")
            balance_on_dest = get_account_balance(chain="dest")
            if balance_on_dest != 0:
                break
            time.sleep(5)
        else:
            self.fail("Balance on destination chain is not correct.")


if __name__ == "__main__":
    unittest.main()
