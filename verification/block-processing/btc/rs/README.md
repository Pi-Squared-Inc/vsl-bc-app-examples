# Claim Verification for Bitcoin

## Overview

This project is a Rust-based library for verifiying claims in Bitcoin block processing. It ensures the integrity of blockchain data by validating prestate, poststate, and transitions of Bitcoin blocks.

## Features

- **Prestate Validation**: Validates the prestate of transactions.
- **Poststate Validation**: Ensures the poststate matches the expected outputs.
- **Transition Validation**: Checks the integrity of block transitions, including Merkle root and witness commitment.
