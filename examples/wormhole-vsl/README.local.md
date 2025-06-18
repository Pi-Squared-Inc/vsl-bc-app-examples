# Wormhole Multichain Demo

PLease refer to the main [README](./README.md) for general information. This file explains how to run the demo with locally deployed Foundry chains.

## Prerequisites

- [Foundry](https://book.getfoundry.sh/getting-started/installation) (version stable v1.0 or higher)
- [Golang](https://go.dev/doc/install) (version 1.21.1 or higher)
- [anvil(Foundry)](https://book.getfoundry.sh/getting-started/installation): for starting the testnet
- [pnpm](https://pnpm.io/installation): Used for the UI
- [jq](https://jqlang.github.io/jq/download/)

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

4. Deploy the PeerToken, Vsl, NttManager, and Transceiver contracts on the source and destination chains (if you want to deploy again, please stop and start fresh Anvil nodes to keep the contracts' addresses matched with the .env file):

   ```bash
   make deploy
   ```

5. Setup the contracts:

   ```bash
   make setup
   ```

6. Start the backend:

   ```bash
   make start-backend
   ```

7. Start the observer:

   ```bash
   make start-observer
   ```

8. Start the verifier:

   ```bash
   make start-verifier
   ```

9. Start the relayer:

   > **Note:** The relayer image was hosted on Pi Squared's GitHub Container Registry, you will need to login your Docker environment to the GitHub Container Registry first. Please refer to [Authenticating to the Container registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry) for more details.

   ```bash
   make start-relayer-dev
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

13. First, the observer will generate a claim and submit it to POD. Then, the relayer will monitor POD for new claim and automatically delivers to the destination chain. You can check the logs from the observer and relayer to see the process.

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
