# Block header settlement demo

## Overview

This demo shows the complete process of block header settlement for Bitcoin and Ethereum on VSL.

The demo allows users to submit block headers from Bitcoin and Ethereum chains and verify them using VSL's verifiable settlement infrastructure. When a block header is submitted, the system:

1. Generates a VSL claim representing the block header data
2. Creates a proof of the block header's validity
3. Verifies the proof using VSL's verification system
4. Settles the verified block header data on VSL

This demo showcases how VSL enables trustless and decentralized cross-chain verification of blockchain data. In its current form, a claim submitter acts as a trusted observer, monitoring Bitcoin and Ethereum nodes and generating claims by signing each new block header with its private key. Verification involves checking that the headers were indeed signed by the claim submitter, ensuring their validity and allowing applications to securely consume data across chains. A potential enhancement to this setup would involve extending the observer’s role to also verify that each block adheres to the consensus rules of its respective chain.

It includes the following components:

- [btc](./btc/README.md): Service that generate and verify Bitcoin block headers.
- [eth](./eth/README.md): Service that generate and verify Ethereum block headers.
- [backend](./backend/README.md): The backend service provides the HTTP APIs and functionalities to store and serve the verification results
- [frontend](./frontend/README.md): The demo frontend application

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/) Version >= 2.19.0

### Initialization

1. Ensure in [examples/block-header](./) folder, execute the following commands:

```bash
make prepare-backend &&
make prepare-frontend
```

2. Go to [btc](./btc/README.md) folder and follow the instructions to set up the Bitcoin block header settlement service.
3. Go to [eth](./eth/README.md) folder and follow the instructions to set up the Ethereum block header settlement service.

4. Run the demo:

```bash
make start-demo
```

3. Stop the demo:

```bash
make stop-demo
```

### VSL Deployment

The demo uses the devnet RPC VSL endpoints. If you would like to run the demo with a local VSL deployment, change endpoints in `.env` files used in the demo. You can find more details about running VSL in the [`VSL-CLI`](https://github.com/Pi-Squared-Inc/vsl-cli) and [`VSL-SDK`](https://github.com/Pi-Squared-Inc/vsl-sdk) repos. 
