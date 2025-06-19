// use alloy_genesis::Genesis;
use alloy_rlp::Decodable;
use generation_block_processing_evm::types::{Claim, ClaimVerificationContext};
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
        // TODO: make it configurable
        MAINNET.clone(),
    )
    .map_err(|e| VerificationError::VerifyBlockError(e.to_string()))?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use generation_block_processing_evm::types::{Claim, ClaimVerificationContext};
    use serde_json;
    use std::env::current_dir;
    use std::fs::read_to_string;

    use crate::verification::block_processing::verify;

    /// This test validates the `verify_claim` function by using mock data.
    ///
    /// Required files for the test, you can generate them using the `test_generate_claim` function in `generation/block-processing/evm/rs/src/generation/block_processing.rs`:
    /// - `src/verification/block_processing_test_mock_claim.json`: Contains a mock `Claim` object in JSON format.
    /// - `src/verification/block_processing_test_mock_verification_context.json`: Contains a mock `ClaimVerificationContext` object in JSON format.
    ///
    /// The test reads these files, deserializes their contents into the appropriate types,
    /// and then calls the `verify_claim` function to ensure it behaves as expected.
    #[test]
    fn test_verify() {
        let current_dir = current_dir().unwrap();
        let claim_path = current_dir.join("src/verification/block_processing_test_mock_claim.json");
        let context_path = current_dir
            .join("src/verification/block_processing_test_mock_verification_context.json");

        let claim: Claim =
            serde_json::from_str(&read_to_string(&claim_path).expect("Failed to read claim file"))
                .expect("Failed to parse claim");
        let context: ClaimVerificationContext = serde_json::from_str(
            &read_to_string(&context_path).expect("Failed to read context file"),
        )
        .expect("Failed to parse context");

        let result = verify(&claim, &context);
        assert!(
            result.is_ok(),
            "Claim verification failed: {:?}",
            result.err()
        );
    }
}
