# Bitcoin Claim Generation in Rust

This Rust application generates claims and proofs based on Bitcoin blocks, serializing them as JSON files for further processing or verification.

## Project Structure

```
rs/
├── src/
│   ├── generation/
│   │   ├── block_processing.rs           # Claim generation logic
│   ├── generation.rs
│   ├── lib.rs
├── tests/
│   ├── block_processing_tests.rs         # Unit and integration tests
│   ├── test_claim.json                   # Example claim for testing
│   └── test_proof.json                   # Example proof for testing
├── Cargo.toml                            # Rust crate configuration
└── README.md                             # Project README
```

## Getting Started

### Prerequisites

- [Rust and Cargo](https://www.rust-lang.org/tools/install) (latest stable recommended)
- Internet connection (for fetching dependencies)

To build and run the application, ensure you have Rust and Cargo installed on your machine. You can follow the instructions on the [official Rust website](https://www.rust-lang.org/tools/install) to install Rust.

## Environment

Set the following environment variable (or add to a `.env` file, or even copy the `sample.env` file):

```
BITCOIN_RPC_URL=http://user:password@127.0.0.1:8332
```

### Building the Project

Navigate to the project directory and run the following command:

```
cargo build
```
