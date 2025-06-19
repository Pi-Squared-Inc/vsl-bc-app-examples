use alloy::{sol, sol_types::SolValue};
use base::{
    btc_rpc::{get_block_by_hash, get_block_hash_at_height},
    signer::Signer,
};
use bitcoin::{block::Header, hashes::Hash};
use block_header_common::types::{
    Claim, ClaimType, ClaimVerificationContext, EncodableClaim, Metadata, Network,
};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

sol! {
    #[derive(Debug,Serialize, Deserialize)]
    struct EncodableBitcoinHeader {
        int32 version;
        bytes32 prev_blockhash;
        bytes32 merkle_root;
        uint32 time;
        uint32 bits;
        uint32 nonce;
    }
}

impl From<Header> for EncodableBitcoinHeader {
    fn from(value: Header) -> Self {
        EncodableBitcoinHeader {
            version: value.version.to_consensus(),
            prev_blockhash: value.prev_blockhash.to_raw_hash().to_byte_array().into(),
            merkle_root: value.merkle_root.as_byte_array().into(),
            time: value.time,
            bits: value.bits.to_consensus(),
            nonce: value.nonce,
        }
    }
}

impl From<EncodableBitcoinHeader> for Header {
    fn from(value: EncodableBitcoinHeader) -> Self {
        Header {
            version: bitcoin::blockdata::block::Version::from_consensus(value.version),
            prev_blockhash: bitcoin::BlockHash::from_byte_array(value.prev_blockhash.into()),
            merkle_root: bitcoin::TxMerkleNode::from_byte_array(value.merkle_root.into()),
            time: value.time,
            bits: bitcoin::CompactTarget::from_consensus(value.bits),
            nonce: value.nonce,
        }
    }
}

#[derive(Debug, thiserror::Error)]
pub enum GenerateError {
    #[error("Failed to fetch block hash: {0}")]
    FetchBlockHashError(String),
    #[error("Failed to fetch block header: {0}")]
    FetchBlockHeaderError(String),
    #[error("Failed to serialize block header: {0}")]
    SerializeBlockHeaderError(String),
    #[error("Failed to get chain info: {0}")]
    GetChainInfoError(String),
    #[error("Failed to sign message: {0}")]
    SignMessageError(String),
    #[error("Failed to decode claim: {0}")]
    DecodeClaimError(String),
}

/// Asynchronously generates a claim for a given block number by calling the JSON-RPC API.
///
/// # Arguments
///
/// * `client` - A reference to a `Client` object for making JSON-RPC calls.
/// * `signer` - A reference to a `Signer` instance used to sign the claim.
/// * `block_number` - The block number for which the claim is generated.
///
/// # Returns
///
///  * `Ok((String, ClaimVerificationContext))` - A tuple containing the claim and the
///  verification context.
///
/// # Errors
///
/// * `Err(GenerateError)` - An error if any of the operations fail
///
/// * `Header` type - https://github.com/bitcoin/bitcoin/blob/345457b542b6a980ccfbc868af0970a6f91d1b82/src/primitives/block.h#L20
pub async fn generate(
    rpc_url: &str,
    signer: &Signer,
    block_number: u64,
) -> Result<(String, ClaimVerificationContext), GenerateError> {
    let block_hash = get_block_hash_at_height(rpc_url, block_number)
        .await
        .map_err(|e| GenerateError::FetchBlockHashError(e.to_string()))?;
    let block = get_block_by_hash(rpc_url, &block_hash)
        .await
        .map_err(|e| GenerateError::FetchBlockHeaderError(e.to_string()))?;

    let block_header = block.header;
    let chain_id = 1;

    let network = Network::Bitcoin;

    let encodable_header = EncodableBitcoinHeader::from(block_header.clone());

    let header_bytes = encodable_header.abi_encode();

    let message_bytes: [u8; 32] = Sha256::digest(&header_bytes).into();

    let signature = signer
        .sign_message(message_bytes)
        .map_err(|e| GenerateError::SignMessageError(e.to_string()))?;

    let encodable_header = EncodableBitcoinHeader::from(block_header.clone());

    let claim = Claim {
        claim_type: ClaimType::BlockHeader,
        metadata: Metadata { chain_id, network },
        header: encodable_header,
    };

    let encodable_claim: EncodableClaim = claim.into();

    let encoded_claim = encodable_claim.abi_encode();

    let hex_encoded_claim = hex::encode(encoded_claim);

    let claim_verification_context = ClaimVerificationContext { signature };

    Ok((hex_encoded_claim, claim_verification_context))
}

#[cfg(test)]
mod tests {
    use super::{EncodableBitcoinHeader, generate};
    use alloy::dyn_abi::SolType;
    // Fixed: Use EncodableBitcoinHeader, not EncodableHeader
    use base::{
        btc_rpc::{get_block_by_hash, get_block_hash_at_height},
        signer::Signer,
    };
    use bitcoin::block::Header;
    use block_header_common::types::{Claim, EncodableClaim}; // Added EncodableClaim import
    use hex;
    use std::{env::current_dir, io::Write};

    #[tokio::test]
    async fn test_generate() {
        let current_dir = current_dir().unwrap();

        let btc_rpc = "https://bitcoin-mainnet.g.alchemy.com/v2/Yjxf3Pt_A6aWcjdpyuXG9GME74JOzcgR
";
        let private_key_hex = "0x4c0883a69102937d6231471b5dbb62f0e8b11d6f4c2e7f8a9c2e7f8a9c2e7f8a";

        let signer = Signer::from_hex_key(private_key_hex).unwrap();

        let block_number = 896519;

        let block_hash = get_block_hash_at_height(btc_rpc, block_number)
            .await
            .unwrap();

        let block_header = get_block_by_hash(btc_rpc, &block_hash)
            .await
            .unwrap()
            .header;

        let (generated_claim_hex, generated_claim_verification_context) =
            generate(&btc_rpc, &signer, block_number).await.unwrap();

        let claim_bytes = hex::decode(&generated_claim_hex).expect("Failed to decode claim hex");

        let decoded_encodable_claim =
            EncodableClaim::abi_decode(&claim_bytes).expect("Failed to decode encodable claim");

        let decoded_claim: Claim<EncodableBitcoinHeader> = decoded_encodable_claim
            .try_into()
            .expect("Failed to convert encodable claim back to claim");

        let claim_header = Header::from(decoded_claim.header);

        assert_eq!(claim_header, block_header, "Block header mismatch");

        let mut file = std::fs::File::create(
            current_dir.join("src/generation/block_header_test_mock_claim.json"),
        )
        .unwrap();
        let claim = serde_json::to_string(&generated_claim_hex).unwrap();
        file.write_all(claim.as_bytes()).unwrap();
        file.flush().unwrap();

        let mut file = std::fs::File::create(
            current_dir.join("src/generation/block_header_test_mock_claim_context.json"),
        )
        .unwrap();
        let json_claim_context =
            serde_json::to_string(&generated_claim_verification_context).unwrap();
        file.write_all(json_claim_context.as_bytes()).unwrap();
        file.flush().unwrap();
    }
}
