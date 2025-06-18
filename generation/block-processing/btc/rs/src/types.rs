use bitcoin::{Transaction, TxIn, TxOut, block::Header};
use serde::{Deserialize, Serialize};

/// Represents a claim about a Bitcoin block state transition.
///
/// A `Claim` contains:
/// - `prestate`: The set of transaction inputs (UTXOs) before the block is applied.
/// - `transition`: The block header representing the state transition.
/// - `poststate`: The set of transaction outputs (UTXOs) after the block is applied.
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claim {
    pub prestate: Vec<TxIn>,
    pub transition: Header,
    pub poststate: Vec<TxOut>,
}

/// Context for verifying a claim, containing the transactions relevant to the claim.
///
/// A `ClaimVerificationContext` provides the set of transactions that are used to verify the correctness of a `Claim`.
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ClaimVerificationContext {
    pub transactions: Vec<Transaction>,
}

/// Errors that can occur during claim generation.
///
/// `GenerationError` enumerates possible failures when generating a claim, such as:
/// - Fetching a block from the Bitcoin node or RPC endpoint (`FetchBlockError`)
/// - Writing a claim or proof to a file (`WriteClaimError`)
///
/// Each variant contains a `String` describing the underlying error.
#[derive(Debug, thiserror::Error, Clone)]
pub enum GenerationError {
    #[error("Failed to fetch block: {0}")]
    FetchBlockError(String),
    #[error("Failed to write claim to file: {0}")]
    WriteClaimError(String),
}
