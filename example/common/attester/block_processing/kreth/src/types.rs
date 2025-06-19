// use alloy_genesis::ChainConfig;
use alloy_primitives::Bytes;
use reth_primitives_traits::Header;
use reth_stateless::ExecutionWitness;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub enum ClaimType {
    BlockProcessing,
    ViewFN,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Metadata {
    pub chain_id: u64,
}

#[derive(Serialize, Deserialize, Debug)]
/// Represents a claim with associated metadata, result, and assumptions.
///
/// # Fields
/// - `claim_type`: The type of the claim, represented by the `ClaimType` enum.
/// - `metadata`: Additional metadata associated with the claim.
/// - `result`: The post-processed RLP encoded block.
/// - `assumptions`: Assumptions or headers related to the claim.
pub struct Claim {
    pub claim_type: ClaimType,
    pub metadata: Metadata,
    pub result: Bytes,
    pub assumptions: Header,
}

#[derive(Serialize, Deserialize, Debug)]
/// Represents the context required for verifying a claim, which includes
/// the necessary execution witness data.
pub struct ClaimVerificationContext {
    pub witness: ExecutionWitness,
}
