# VSL Blockchain Applications

The Verifiable Settlement Layer (VSL) is an infrastructure-level decentralized network that offers 
scalable, affordable, and customizable verifiability to all web3 protocols, applications, and users. 
It encodes payments, transactions, computations, assets, and data as VSL claims — the fundamental building 
blocks of anything verifiable or provable across platforms, ecosystems, and blockchains.

This repository contains several example applications build on top of VSL. You can find details in corresponding subdirectories:

## Block Header Settlement Demo
The [block-header](./examples/block-header) example demonstrates how to use VSL for block header settlement of Bitcoin and Ethereum. It includes a complete implementation with services for generating and verifying block headers, a backend API, and a frontend interface. This example showcases how VSL can be used to verify blockchain data across different networks.

## Blockchain Mirroring Demo
The [blockchain-mirroring](./examples/blockchain-mirroring) example demonstrates how to use VSL for mirroring and validating blockchain data using different execution clients. It uses implementations for Geth and Reth nodes and can also use verifiers with support for KEVM execution client (not open-sourced yet). This example shows how VSL can be used to verify the correctness of blockchain state transitions across different execution clients.

## Wormhole Multichain Demo
The [wormhole-vsl](./examples/wormhole-vsl) example showcases the integration of VSL with the Wormhole protocol for cross-chain token transfers. It demonstrates how to use VSL to verify and settle cross-chain transactions between different networks (e.g., Ethereum Sepolia and Arbitrum Sepolia). This example includes smart contracts, a relayer service, and a web interface for managing cross-chain transfers.


