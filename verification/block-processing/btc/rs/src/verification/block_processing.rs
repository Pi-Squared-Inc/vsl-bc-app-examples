use bitcoin::{Block, Transaction, TxIn, TxOut, block::Header};
use generation_block_processing_btc::{
    generation::block_processing::generate_block_states,
    types::{Claim, ClaimVerificationContext},
};
use serde::{Deserialize, Serialize};

/// Represents errors that can occur during the verification of Bitcoin block claims.
///
/// This enum categorizes the possible failure modes when verifying block states and transitions,
/// providing detailed error messages for each type of verification failure.
///
/// # Variants
/// - `PrestateVerificationError(String)`: The prestate does not match the expected prestate.
/// - `PoststateVerificationError(String)`: The poststate does not match the expected poststate.
/// - `TransitionVerificationError(String)`: The block transition is invalid.
/// - `MerkleRootVerificationError(String)`: The block's Merkle root does not match the expected value.
/// - `WitnessCommitmentVerificationError(String)`: The block's witness commitment is invalid.
///
/// Each variant contains a `String` describing the specific error encountered.
#[derive(Debug, thiserror::Error, Serialize, Deserialize)]
pub enum VerificationError {
    #[error("Prestate verification failed: {0}")]
    PrestateVerificationError(String),
    #[error("Poststate verification failed: {0}")]
    PoststateVerificationError(String),
    #[error("Transition verification failed: {0}")]
    TransitionVerificationError(String),
    #[error("Merkle root verification failed: {0}")]
    MerkleRootVerificationError(String),
    #[error("Witness commitment verification failed: {0}")]
    WitnessCommitmentVerificationError(String),
}

/// Verifies that the provided prestate and poststate match the expected states derived from the given transactions.
///
/// This function checks:
/// - That the `prestate` matches the expected prestate generated from the transactions in the `proof`.
/// - That the `poststate` matches the expected poststate generated from the transactions in the `proof`.
///
/// # Arguments
/// * `prestate` - The list of transaction inputs representing the state before block processing.
/// * `poststate` - The list of transaction outputs representing the state after block processing.
/// * `proof` - The `ClaimVerificationContext` containing the transactions used to generate the expected states.
///
/// # Returns
/// * `Ok(())` if both the prestate and poststate are valid.
/// * `Err(VerificationError)` if either the prestate or poststate does not match the expected values.
///
/// # Errors
/// Returns a `VerificationError` if:
/// - The prestate does not match the expected prestate.
/// - The poststate does not match the expected poststate.
pub fn verify_block_states(
    prestate: &Vec<TxIn>,
    poststate: &Vec<TxOut>,
    proof: &ClaimVerificationContext,
) -> Result<(), VerificationError> {
    let (expected_prestate, expected_poststate) = generate_block_states(&proof.transactions);
    if !prestate.eq(&expected_prestate) {
        return Err(VerificationError::PrestateVerificationError(
            "Prestate does not match the expected prestate".to_string(),
        ));
    }
    if !poststate.eq(&expected_poststate) {
        return Err(VerificationError::PoststateVerificationError(
            "Poststate does not match the expected poststate".to_string(),
        ));
    }
    Ok(())
}

/// Verifies the integrity of a Bitcoin block transition by checking Merkle root and witness commitment.
///
/// This function constructs a `Block` from the provided header and transactions,
/// then performs the following checks:
/// - Verifies that the block's Merkle root matches the expected value in the header.
/// - Verifies that the block's witness commitment is correct.
///
/// # Arguments
/// * `transition` - The block header representing the transition to verify.
/// * `context` - The list of transactions included in the block.
///
/// # Returns
/// * `Ok(())` if both the Merkle root and witness commitment are valid.
/// * `Err(VerificationError)` if either check fails, with details about the failure.
///
/// # Errors
/// Returns a `VerificationError` if:
/// - The Merkle root does not match the expected value.
/// - The witness commitment is invalid.
pub fn verify_transition(
    transition: &Header,
    context: &Vec<Transaction>,
) -> Result<(), VerificationError> {
    let block = Block {
        header: transition.clone(),
        txdata: context.clone(),
    };
    if block.check_merkle_root() == false {
        return Err(VerificationError::MerkleRootVerificationError(
            "Merkle root does not match".to_string(),
        ));
    }

    if block.check_witness_commitment() == false {
        return Err(VerificationError::WitnessCommitmentVerificationError(
            "Witness root does not match".to_string(),
        ));
    }

    Ok(())
}

/// Verifies the validity of a Bitcoin block claim against the provided verification context.
///
/// This function performs the following checks:
/// - Verifies that the prestate and poststate of the claim match the expected states derived from the transactions.
/// - Ensures the block transition (header and transactions) has a valid Merkle root and witness commitment.
///
/// # Arguments
/// * `claim` - The `Claim` object containing prestate, poststate, and transition data to be verified.
/// * `context` - The `ClaimVerificationContext` providing the set of transactions and auxiliary data needed for verification.
///
/// # Returns
/// * `Ok(())` if the claim is valid.
/// * `Err(VerificationError)` if any verification step fails, with details about the failure.
///
/// # Errors
/// Returns a `VerificationError` if:
/// - The prestate or poststate does not match the expected values.
/// - The Merkle root or witness commitment is invalid.
pub fn verify(claim: &Claim, context: &ClaimVerificationContext) -> Result<(), VerificationError> {
    verify_block_states(&claim.prestate, &claim.poststate, &context)?;
    verify_transition(&claim.transition, &context.transactions)?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use bitcoin::Block;
    use generation_block_processing_btc::generation::block_processing::{
        generate_from_block, read_claim_from_file, read_proof_from_file,
    };

    fn get_test_block() -> Block {
        let test_claim = read_claim_from_file("src/verification/test_claim.json").unwrap();
        let test_proof = read_proof_from_file("src/verification/test_proof.json").unwrap();
        let test_block = Block {
            header: test_claim.transition,
            txdata: test_proof.transactions,
        };
        test_block
    }

    /// Tests that a valid claim and context pass verification successfully.
    ///
    /// This test loads a test block, generates a claim and verification context from it, and asserts that the `verify` function returns `Ok(())`, indicating successful validation.
    #[tokio::test]
    async fn test_verify_success() {
        let block = get_test_block();
        let (claim, context) = generate_from_block(block).unwrap();
        let validation_result = verify(&claim, &context).unwrap();
        assert_eq!(validation_result, ());
    }

    /// Tests that an invalid context causes claim verification to fail.
    ///
    /// This test loads a test block, generates a claim and context, then modifies the context by removing the first transaction. It asserts that the `verify` function returns an error, indicating that the claim is no longer valid with the modified context.
    #[tokio::test]
    async fn test_verify_failure() {
        let block = get_test_block();
        let (claim, context) = generate_from_block(block).unwrap();

        // remove  the first transaction from the context to simulate a failure
        let mut modified_transactions = context.transactions.clone();
        modified_transactions.remove(0);
        let modified_context = ClaimVerificationContext {
            transactions: modified_transactions,
        };

        let validation_result = verify(&claim, &modified_context);
        assert!(validation_result.is_err());
    }
}
