# Blockchain Submitter for Geth

This is the service that monitors the Geth fullnode, generates block processing claims, and submits them to a remote RPC endpoint for verification.

## Prerequisites

- [Golang](https://go.dev/doc/install): Version >= 1.24.1

## Getting Started

1. Start the Geth fullnode. Please check the [README](../../../generation/block-processing/evm/go/README.md) for how to start the Geth fullnode.
2. Ensure the verifier service is running and accessible to receive claims.
3. Prepare the environment variables

   ```bash
   cp sample.env .env
   ```

   Update the environment variables:
   - `SOURCE_RPC_ENDPOINT` and `SOURCE_WEBSOCKET_ENDPOINT` with the Geth node RPC and WebSocket endpoints
   - `REMOTE_RPC_ENDPOINT` with the verifier service endpoint

4. Install the dependencies

   ```bash
   go mod download
   ```

5. Start the Geth submitter service

   ```bash
   go run ./cmd/main.go
   ```

## Architecture

The submitter service:

1. **Monitors** a Geth fullnode for new blocks
2. **Generates** block processing claims using the generation logic
3. **Submits** claims to a remote RPC endpoint (verifier service)

The verification and backend submission logic has been moved to the verifier service for better separation of concerns.
