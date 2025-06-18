# Blockchain Mirroring for Reth

This is the service that mirrors the Reth node, generates and validates the block processing claim with the Reth execution client.

## Prerequisites

- [Rust](https://www.rust-lang.org/tools/install): Version >= 1.86.0

## Getting Started

1. Start the Reth node. Please check the [README](../../../generation/block-processing/evm/rs/README.md) for how to start the Reth node.
2. Start the backend service. Please check the [README](../backend/README.md) for how to start the backend service.
3. Prepare the environment variables

   ```bash
   cp sample.env .env
   ```

   Update the `SOURCE_RPC_ENDPOINT` and `SOURCE_WEBSOCKET_ENDPOINT` environment variables with the Reth node RPC and WebSocket endpoints.

4. Run the Reth mirroring service

   ```bash
   RUST_LOG=info cargo run
   ```
