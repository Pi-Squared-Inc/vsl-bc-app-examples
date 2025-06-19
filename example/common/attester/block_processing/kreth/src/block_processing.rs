// use alloy_genesis::Genesis;
use alloy_rlp::Decodable;
use crate::types::{Claim, ClaimVerificationContext};
use reth_chainspec::MAINNET;
use reth_primitives_traits::SealedBlock;
use reth_stateless::validation::stateless_validation;

#[derive(Debug, thiserror::Error)]
pub enum VerificationError {
    #[error("Failed to decode block: {0}")]
    DecodeBlockError(String),
    #[error("Failed to verify block: {0}")]
    VerifyBlockError(String),
}

/// Verifies a claim against a given verification context.
///
/// # Arguments
///
/// * `claim` - A reference to the `Claim` object that needs to be verified.
/// * `context` - A reference to the `ClaimVerificationContext` that provides
///   the necessary context for verification.
///
/// # Returns
///
/// * `Ok(())` if the claim is successfully verified.
/// * `Err(VerifyClaimError)` if the verification fails, with detailed error
///   information.
pub fn verify(claim: &Claim, context: &ClaimVerificationContext) -> Result<(), VerificationError> {
    let mut raw_block_slice: &[u8] = &claim.result;
    let decoded: SealedBlock<alloy_consensus::Block<reth_primitives::TransactionSigned>> =
        SealedBlock::<reth_primitives::Block>::decode(&mut raw_block_slice)
            .map_err(|e| VerificationError::DecodeBlockError(e.to_string()))?;

    let recovered_block = decoded.clone().try_recover().unwrap();

    // let chain_spec: ChainSpec = Genesis {
    //     config: claim.metadata.chain_config.clone(),
    //     ..Default::default()
    // }
    // .into();

    stateless_validation(
        recovered_block,
        context.witness.clone(),
        MAINNET.clone(),
    )
    .map_err(|e| VerificationError::VerifyBlockError(e.to_string()))?;
    Ok(())
}
