use alloy::sol_types::SolValue;
use base::signer::verify_message;
use block_header_common::types::{Claim, ClaimVerificationContext, EncodableClaim};
use generation_block_header_eth::generation::block_header::EncodableEthereumHeader;
use sha2::{Digest, Sha256};

#[derive(Debug, thiserror::Error)]
pub enum VerificationError {
    #[error("Failed to serialize block header: {0}")]
    SerializeBlockHeaderError(String),
    #[error("Failed to validate block header: {0}")]
    VerifyBlockError(String),
    #[error("Decoding claim failed: {0}")]
    DecodeClaimError(String),
}

/// Verifies a claim against a given verification context.
///
/// # Arguments
///
/// * `hex_encoded_claim` - A reference to the hex-encoded claim string to be validated.
/// * `public_key` - A reference to a `&str` public key that was used to sign the claim.
/// * `context` - A reference to the `ClaimVerificationContext` instance that provides the
/// necessary context for verification.
///
/// # Returns
///
/// * `Ok(())` if the claim is valid.
/// * `Err(VerificationError)` if the claim is invalid or if an error occurs during verification.
pub fn verify(
    hex_encoded_claim: &String,
    public_key: &str,
    context: &ClaimVerificationContext,
) -> Result<(), VerificationError> {
    // Decode hex to bytes
    let claim_bytes = hex::decode(hex_encoded_claim)
        .map_err(|e| VerificationError::SerializeBlockHeaderError(e.to_string()))?;

    // Decode claim using ABI decoding (same as Bitcoin)
    let decoded_encodable_claim = <EncodableClaim as SolValue>::abi_decode(&claim_bytes)
        .map_err(|e| VerificationError::DecodeClaimError(e.to_string()))?;

    // Convert to typed claim
    let decoded_claim: Claim<EncodableEthereumHeader> = decoded_encodable_claim
        .try_into()
        .map_err(|e: &'static str| VerificationError::SerializeBlockHeaderError(e.to_string()))?;

    // Encode JUST the header part using ABI encoding (same as Bitcoin)
    let header_bytes = decoded_claim.header.abi_encode();

    // Hash the encoded header
    let message_bytes: [u8; 32] = Sha256::digest(&header_bytes).into();

    // Verify the signature
    verify_message(public_key, message_bytes, &context.signature)
        .map_err(|e| VerificationError::VerifyBlockError(e.to_string()))?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::verify;
    use block_header_common::types::ClaimVerificationContext;
    use std::fs::read_to_string;

    #[test]
    fn test_verify() {
        let current_dir = std::env::current_dir().unwrap();
        let claim_path = current_dir.join("src/verification/block_header_test_mock_claim.json");
        let context_path =
            current_dir.join("src/verification/block_header_test_mock_claim_context.json");

        let claim: String = serde_json::from_str(&read_to_string(&claim_path).unwrap())
            .expect("Failed to parse claim");
        let claim_context: ClaimVerificationContext =
            serde_json::from_str(&read_to_string(&context_path).unwrap())
                .expect("Failed to parse context");

        let public_key: &str = "033ae6e24aa5d203b6d468d8519fb94930739df84cd3aeab475571424875298cab";

        let result = verify(&claim, &public_key, &claim_context);
        assert!(
            result.is_ok(),
            "Claim verification failed: {:?}",
            result.err()
        );
    }
}
