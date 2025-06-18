#!/bin/bash

BACKEND_ENDPOINT="http://backend:3001"
MAX_RETRIES=15
RETRY_DELAY=5 # seconds

retry_count=0
while [ $retry_count -lt $MAX_RETRIES ]; do
  echo "Attempt $((retry_count + 1))/$MAX_RETRIES: Fetching block mirroring records..."
  response=$(curl -s -f --request GET --url "${BACKEND_ENDPOINT}/block_mirroring_records?page=0")
  curl_exit_status=$?

  if [ $curl_exit_status -ne 0 ]; then
    echo "Error: curl command failed with status $curl_exit_status."
  else
    block_mirroring_records_count=$(echo "$response" | jq '.records | length')
    jq_exit_status=$?

    if [ $jq_exit_status -ne 0 ] || ! [[ "$block_mirroring_records_count" =~ ^[0-9]+$ ]]; then
      echo "Error: Failed to parse JSON response or extract a valid claim count."
    elif [ "$block_mirroring_records_count" -gt 0 ]; then
      echo "Found $block_mirroring_records_count block mirroring record(s). Checking the first one..."
      first_block_mirroring_record=$(echo "$response" | jq '.records[0]')
      jq_extract_exit_status=$?

      if [ $jq_extract_exit_status -ne 0 ]; then
        echo "Error: Failed to extract the first block mirroring record using jq."
      else
        # Check if .claim_details.MirroringReth exists in the first claim
        has_reth=$(echo "$first_block_mirroring_record" | jq 'has("claim_details") and (.claim_details | has("MirroringReth"))')

        if [ "$has_reth" != "true" ]; then
          echo "First block mirroring record does not contain 'claim_details.MirroringReth'. Skipping this attempt."
        else
          # Check if .claim_details.MirroringReth.error is null
          reth_error=$(echo "$first_block_mirroring_record" | jq '.claim_details.MirroringReth.error')

          if [ "$reth_error" == "null" ]; then
            echo "Found the valid block mirroring record."
            exit 0
          else
            # Error condition: MirroringReth.error is not null
            echo "Error: Found a block mirroring record, but 'MirroringReth.error' is not null."
            echo "Problematic block mirroring record details:"
            echo "$first_block_mirroring_record"
            exit 1
          fi
        fi
      fi
    else
      # Block mirroring record count is 0
      echo "Block mirroring records count is 0."
    fi
  fi

  # Increment retry counter and wait if more retries are left
  retry_count=$((retry_count + 1))
  if [ $retry_count -lt $MAX_RETRIES ]; then
    echo "Retrying in $RETRY_DELAY seconds..."
    sleep $RETRY_DELAY
  else
    # Max retries reached without finding a *valid* claim
    echo "Error: No valid claim found after $MAX_RETRIES attempts."
    exit 1
  fi
done

exit 1
