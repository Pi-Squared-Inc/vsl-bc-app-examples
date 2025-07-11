# Wormhole Multichain Demo

PLease refer to the main [README](./README.md) for general information. This file explains how to run the demo with locally deployed Foundry chains.

## Prerequisites

- [Foundry](https://book.getfoundry.sh/getting-started/installation) (version stable v1.0 or higher)
- [Golang](https://go.dev/doc/install) (version 1.21.1 or higher)
- [anvil(Foundry)](https://book.getfoundry.sh/getting-started/installation): for starting the testnet
- [pnpm](https://pnpm.io/installation): Used for the UI
- [jq](https://jqlang.github.io/jq/download/)

## Preparation

1. Run a local VSL node or connect to a remote one. For local deployment you can find instructions in the [`VSL-CLI`](https://github.com/Pi-Squared-Inc/vsl-cli) and [`VSL-SDK`](https://github.com/Pi-Squared-Inc/vsl-sdk) repos. We need to provide these following environment variables that related to VSL:
   - VSL node RPC endpoint: `VSL_RPC`
   - Two VSL accounts with some funds for observer and verifier:
     - VSL observer account address: `VSL_CLIENT_ADDRESS`
     - VSL observer account private key: `VSL_CLIENT_PRIVATE_KEY`
     - VSL verifier account address: `VSL_VERIFIER_ADDRESS`
     - VSL verifier account private key: `VSL_VERIFIER_PRIVATE_KEY`

## Getting started

1. Prepare depends and environments:

   ```bash
   make install
   make prepare-env
   make prepare-deps
   ```

2. Start the source anvil node:

   ```bash
   make start-local-source-chain
   ```

3. Start the destination anvil node:

   ```bash
   make start-local-dest-chain
   ```

4. Modify the `.env` file, fill all environment variables with the example values

5. Deploy the PeerToken, Vsl, NttManager, and Transceiver contracts on the source and destination chains (if you want to deploy again, please stop and start fresh Anvil nodes to keep the contracts' addresses matched with the .env file):

   ```bash
   make deploy
   ```

6. Setup the contracts:

   ```bash
   make setup
   ```

7. Start the backend:

   ```bash
   make start-backend
   ```

8. Start the observer:

   ```bash
   make start-observer
   ```

9. Start the verifier:

   ```bash
   make start-verifier
   ```

10. Start the relayer:

    ```bash
    make start-relayer
    ```

11. Mint token on source chain and check the balance:

    ```bash
    make mint-source-token # Mint token and set the minter to the manager
    make check-source-balance # Check the balance
    ```

12. Transfer the token between source chain and dest chain:

    ```bash
    make transfer # Transfer the token between source chain and dest chain
    make check-source-balance # Check the balance on source chain
    ```

13. First, the observer will generate a claim and submit it to VSL. Then, the relayer will monitor VSL for new claim and automatically delivers to the destination chain. You can check the logs from the observer and relayer to see the process.

14. Once the relayer has successfully delivered the claim to the destination chain, verify the destination chain balance:

    ```bash
    make check-dest-balance # Check the balance again
    ```

## Start UI

1. Rename the `examples/wormhole-vsl/web/config/constant.local.tsx` file to replace the `examples/wormhole-vsl/web/config/constant.tsx` file.
2. Start the UI:

   ```bash
   make start-ui-dev
   ```

3. Open the browser and navigate to `http://localhost:3000`, you will see the UI.
