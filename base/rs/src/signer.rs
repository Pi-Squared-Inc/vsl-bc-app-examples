use std::str::FromStr;

use hex;
use secp256k1::{Message, PublicKey, Secp256k1, SecretKey, ecdsa::Signature};

#[derive(Debug, thiserror::Error)]
pub enum SignerError {
    #[error("Failed to decode hex: {0}")]
    HexDecodeError(String),
    #[error("Failed to create secret key: {0}")]
    SecretKeyError(String),
    #[error("Failed to parse signature: {0}")]
    SignatureParseError(String),
    #[error("Failed to create message: {0}")]
    MessageCreationError(String),
    #[error("Failed to verify signature: {0}")]
    VerificationError(String),
}
/// A struct representing a signer that can sign and verify messages using the secp256k1 elliptic curve
///
/// This implementation provides methods for creating signatures, verifying signatures,
/// and retrieving the public key associated with the signer's private key.
///
/// # Examples
///
/// ```
/// use base::signer::Signer;
///
/// // Create a new signer from a private key
/// let hex_key = "0x4c0883a69102937d6231471b5dbb62f0e8b11d6f4c2e7f8a9c2e7f8a9c2e7f8a";
/// let signer = Signer::from_hex_key(hex_key).unwrap();
///
/// // Sign a message
/// let message = [0u8; 32];
/// let signature = signer.sign_message(message).unwrap();
/// ```
#[derive(Clone)]
pub struct Signer {
    secret_key: SecretKey,
    secp: Secp256k1<secp256k1::All>,
}

impl Signer {
    /// Create a new Signer from a hex-encoded private key
    ///
    /// # Arguments
    ///
    /// * `hex_key` - A string slice representing the hex-encoded private key.
    ///
    /// # Returns
    ///
    /// * A Result containing the Signer object or an error if the key is invalid.
    pub fn from_hex_key(hex_key: &str) -> Result<Self, SignerError> {
        let clean_key = hex_key.trim_start_matches("0x");

        let key_bytes =
            hex::decode(clean_key).map_err(|e| SignerError::HexDecodeError(e.to_string()))?;

        let secp = Secp256k1::new();

        let secret_key = SecretKey::from_slice(&key_bytes)
            .map_err(|e| SignerError::SecretKeyError(e.to_string()))?;

        Ok(Self { secret_key, secp })
    }

    /// Sign a message using the private key
    ///
    /// # Arguments
    ///
    /// * `message_bytes` - A 32-byte array representing the message to be signed.
    ///
    /// # Returns
    ///
    /// * A Result containing the hex-encoded signature or an error if signing fails.
    pub fn sign_message(&self, message_bytes: [u8; 32]) -> Result<String, SignerError> {
        let message = Message::from_digest(message_bytes);
        let signature = self.secp.sign_ecdsa(message, &self.secret_key);

        Ok(hex::encode(signature.serialize_compact()))
    }
}

/// Verifies a message signature using the public key
///
/// # Arguments
///
/// * `public_key` - A string slice representing the hex-encoded public key.
/// * `message_bytes` - A 32-byte array representing the message to be verified.
/// * `signature_hex` - A string slice representing the hex-encoded signature.
///
/// # Returns
///
/// * A Result containing a boolean indicating whether the signature is valid or an error if
/// verification fails.
pub fn verify_message(
    public_key: &str,
    message_bytes: [u8; 32],
    signature_hex: &str,
) -> Result<bool, SignerError> {
    let secp = Secp256k1::verification_only();

    let message = Message::from_digest(message_bytes);

    let signature = Signature::from_compact(
        &hex::decode(signature_hex).map_err(|e| SignerError::SignatureParseError(e.to_string()))?,
    )
    .map_err(|e| SignerError::SignatureParseError(e.to_string()))?;

    let public_key =
        PublicKey::from_str(public_key).map_err(|e| SignerError::HexDecodeError(e.to_string()))?;

    secp.verify_ecdsa(message, &signature, &public_key)
        .map(|_| true)
        .map_err(|e| SignerError::VerificationError(e.to_string()))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sign_message() {
        let hex_key = "0x4c0883a69102937d6231471b5dbb62f0e8b11d6f4c2e7f8a9c2e7f8a9c2e7f8a";
        let signer = Signer::from_hex_key(hex_key).unwrap();

        let message_bytes: [u8; 32] = [0; 32];
        let signature = signer.sign_message(message_bytes).unwrap();

        let valid_signature = "258fdc6df377fc508e94666ff58ec7ee356a9ee3b031f364f0eaf4cfe5c311a20c410006c21c73de06794d0a7da44d872657da6b8b1123574d85ac036d48b041";

        assert_eq!(signature, valid_signature);
        assert_eq!(signature.len(), 128);
    }

    #[test]
    fn test_verify_message() {
        let public_key: &str = "033ae6e24aa5d203b6d468d8519fb94930739df84cd3aeab475571424875298cab";

        let valid_signature = "258fdc6df377fc508e94666ff58ec7ee356a9ee3b031f364f0eaf4cfe5c311a20c410006c21c73de06794d0a7da44d872657da6b8b1123574d85ac036d48b041";

        let invalid_message_bytes = [1; 32];
        let message_bytes: [u8; 32] = [0; 32];

        let result = verify_message(public_key, invalid_message_bytes, valid_signature);
        assert!(
            result.is_err(),
            "Should return an error for an invalid message"
        );

        let invalid_signature = "258fdc6df377fc508e94666ff58ec7ee356a9ee3b031f364f0eaf4cfe5c311a20c410006c21c73de06794d0a7da44d872657da6b8b1123574d85ac036d48b040";
        let result = verify_message(public_key, message_bytes, invalid_signature);
        assert!(
            result.is_err(),
            "Should return an error for an invalid signature"
        );

        let is_valid = verify_message(public_key, message_bytes, valid_signature).unwrap();
        assert_eq!(is_valid, true);
    }
}
