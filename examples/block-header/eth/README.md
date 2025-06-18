# ETH Block header settlement with VSL integration

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/) Version >= 2.19.0
- [Foundry](https://book.getfoundry.sh/getting-started/installation.html)

### Initialization

1. Ensure in [example/block-header/eth](./) folder, execute the following command to copy templates with environment variables:

   ```bash
   make prepare-envs
   ```

2. Go to [alchemy](https://www.alchemy.com) to get your API key for Ethereum (Or use any other RPC provider).

3. Fill [./claim-submitter/.env](./claim-submitter/.env) with your Alchemy API key.

   ```env
   SOURCE_RPC_ENDPOINT=https://eth-mainnet.g.alchemy.com/v2/api_key
   SOURCE_WEBSOCKET_ENDPOINT=wss://eth-mainnet.g.alchemy.com/v2/api_key
   ```

4. Execute the following command to generate the private key used for signing the block headers.

   ```bash
   make generate-header-signer-keys
   ```

5. Fill [./claim-submitter/.env](./claim-submitter/.env) values.

   ```env
   HEADER_SIGNER_PRIVATE_KEY=0x...
   HEADER_SIGNER_PUBLIC_KEY=0x...
   ```

6. Fill [./claim-verifier/.env](./claim-verifier/.env) values.

   ```env
   HEADER_SIGNER_PUBLIC_KEY=0x...
   ```

7. Execute the following command to generate the VSL key pairs.

   ```bash
   make generate-submitter-verifier-keys
   ```

8. Fill [./claim-submitter/.env](./claim-submitter/.env) values.

   ```env
   SUBMITTER_PRIVATE_KEY=...
   VERIFIER_ADDRESS=...
   ```

9. Fill [./claim-verifier/.env](./claim-verifier/.env) values.

   ```env
   VERIFIER_PRIVATE_KEY=...
   SUBMITTER_ADDRESS=...
   ```

10. Start ETH claim submitter and verifier services:

    ```bash
    make start
    ```

    - Or go back to the [example/block-header](../README.md) folder and follow the instruction to start all services
