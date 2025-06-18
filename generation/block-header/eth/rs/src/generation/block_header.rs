use alloy::consensus::Header;
use alloy::primitives::{B64, B256, BlockHash, Bloom, Bytes, U256};
use alloy::rpc::types::Header as EthereumHeader;
use alloy::{eips::BlockNumberOrTag, providers::Provider, sol, sol_types::SolValue};
use base::signer::Signer;
use block_header_common::types::{
    Claim, ClaimType, ClaimVerificationContext, EncodableClaim, Metadata, Network,
};
use sha2::{Digest, Sha256};

sol! {
    struct EncodableEthereumHeader {
        bytes32 hash;
        bytes32 parent_hash;
        bytes32 ommers_hash;
        address beneficiary;
        bytes32 state_root;
        bytes32 transactions_root;
        bytes32 receipts_root;
        bytes logs_bloom;
        bytes32 difficulty;
        uint64 number;
        uint64 gas_limit;
        uint64 gas_used;
        uint64 timestamp;
        bytes extra_data;
        bytes32 mix_hash;
        uint64 nonce;
        uint64 base_fee_per_gas;
        bytes32 withdrawals_root;
        uint64 blob_gas_used;
        uint64 excess_blob_gas;
        bytes32 parent_beacon_block_root;
        bytes32 requests_hash;
        bytes32 total_difficulty;
        bytes32 size;
    }
}

impl From<EthereumHeader> for EncodableEthereumHeader {
    fn from(header: EthereumHeader) -> Self {
        EncodableEthereumHeader {
            hash: header.hash.0.into(),
            parent_hash: header.inner.parent_hash.into(),
            ommers_hash: header.inner.ommers_hash.into(),
            beneficiary: header.inner.beneficiary,
            state_root: header.inner.state_root.into(),
            transactions_root: header.inner.transactions_root.into(),
            receipts_root: header.inner.receipts_root.into(),
            logs_bloom: header.inner.logs_bloom.0.to_vec().into(),
            difficulty: B256::from_slice(&header.inner.difficulty.to_be_bytes::<32>()),
            number: header.inner.number,
            gas_limit: header.inner.gas_limit,
            gas_used: header.inner.gas_used,
            timestamp: header.inner.timestamp,
            extra_data: header.inner.extra_data.to_vec().into(),
            mix_hash: header.inner.mix_hash.into(),
            nonce: u64::from_be_bytes(header.inner.nonce.0),
            base_fee_per_gas: header.inner.base_fee_per_gas.unwrap_or_default(),
            withdrawals_root: header.inner.withdrawals_root.map_or(B256::ZERO, |h| h),
            blob_gas_used: header.inner.blob_gas_used.unwrap_or_default(),
            excess_blob_gas: header.inner.excess_blob_gas.unwrap_or_default(),
            parent_beacon_block_root: header
                .inner
                .parent_beacon_block_root
                .map_or(B256::ZERO, |h| h),
            requests_hash: header.inner.requests_hash.map_or(B256::ZERO, |h| h),
            total_difficulty: header
                .total_difficulty
                .map_or(B256::ZERO, |td| B256::from_slice(&td.to_be_bytes::<32>())),
            size: header
                .size
                .map_or(B256::ZERO, |s| B256::from_slice(&s.to_be_bytes::<32>())),
        }
    }
}

impl From<EncodableEthereumHeader> for EthereumHeader {
    fn from(encodable: EncodableEthereumHeader) -> Self {
        let mut logs_bloom_array = [0u8; 256];
        let copy_len = encodable.logs_bloom.len().min(256);
        logs_bloom_array[..copy_len].copy_from_slice(&encodable.logs_bloom[..copy_len]);

        let consensus_header = Header {
            parent_hash: encodable.parent_hash,
            ommers_hash: encodable.ommers_hash,
            beneficiary: encodable.beneficiary,
            state_root: encodable.state_root,
            transactions_root: encodable.transactions_root,
            receipts_root: encodable.receipts_root,
            logs_bloom: Bloom::from(logs_bloom_array),
            difficulty: U256::from_be_bytes(encodable.difficulty.0),
            number: encodable.number,
            gas_limit: encodable.gas_limit,
            gas_used: encodable.gas_used,
            timestamp: encodable.timestamp,
            extra_data: Bytes::from(encodable.extra_data.to_vec()),
            mix_hash: encodable.mix_hash,
            nonce: B64::new(encodable.nonce.to_be_bytes()),
            base_fee_per_gas: if encodable.base_fee_per_gas == 0 {
                None
            } else {
                Some(encodable.base_fee_per_gas)
            },
            withdrawals_root: if encodable.withdrawals_root == B256::ZERO {
                None
            } else {
                Some(encodable.withdrawals_root)
            },
            blob_gas_used: if encodable.blob_gas_used == 0 {
                None
            } else {
                Some(encodable.blob_gas_used)
            },
            excess_blob_gas: if encodable.excess_blob_gas == 0 {
                None
            } else {
                Some(encodable.excess_blob_gas)
            },
            parent_beacon_block_root: if encodable.parent_beacon_block_root == B256::ZERO {
                None
            } else {
                Some(encodable.parent_beacon_block_root)
            },
            requests_hash: if encodable.requests_hash == B256::ZERO {
                None
            } else {
                Some(encodable.requests_hash)
            },
        };

        EthereumHeader {
            hash: BlockHash::from(encodable.hash.0),
            inner: consensus_header,
            total_difficulty: if encodable.total_difficulty == B256::ZERO {
                None
            } else {
                Some(U256::from_be_bytes(encodable.total_difficulty.0))
            },
            size: if encodable.size == B256::ZERO {
                None
            } else {
                Some(U256::from_be_bytes(encodable.size.0))
            },
        }
    }
}

#[derive(Debug, thiserror::Error)]
pub enum GenerationError {
    #[error("Failed to fetch block: {0}")]
    FetchBlockError(String),
    #[error("Failed to fetch chain config: {0}")]
    FetchChainConfigError(String),
    #[error("Failed to serialize block header: {0}")]
    SerializeBlockHeader(String),
    #[error("Block not found: {0}")]
    BlockNotFound(String),
    #[error("Failed to sign message: {0}")]
    SignMessageError(String),
}

/// Asynchronously generates a claim for a given block number by calling the JSON-RPC API.
///
/// # Arguments
///
/// * `provider` - A reference to an `Provider` instance used to communicate with the JSON-RPC API.
/// * `signer` - A reference to a `Signer` instance used to sign the claim.
/// * `block_number` - The block number for which the claim is generated.
///
/// # Returns
///
/// * `Ok((String, ClaimVerificationContext))` if the claim is generated successfully.
/// * `Err(GenerationError)` if an error occurs during the process.
pub async fn generate(
    provider: &dyn Provider,
    signer: &Signer,
    block_number: u64,
) -> Result<(String, ClaimVerificationContext), GenerationError> {
    let block = provider
        .get_block_by_number(BlockNumberOrTag::Number(block_number))
        .await
        .map_err(|e| GenerationError::FetchBlockError(e.to_string()))?
        .ok_or_else(|| {
            GenerationError::BlockNotFound(format!("Block not found at height: {}", block_number))
        })?;

    let chain_id = provider
        .get_chain_id()
        .await
        .map_err(|e| GenerationError::FetchChainConfigError(e.to_string()))?;

    let encodable_header = EncodableEthereumHeader::from(block.header.clone());

    let header_bytes = encodable_header.abi_encode();

    let message_bytes: [u8; 32] = Sha256::digest(&header_bytes).into();

    let signature = signer
        .sign_message(message_bytes)
        .map_err(|e| GenerationError::SignMessageError(e.to_string()))?;

    let network = Network::Ethereum;

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
    use super::*;
    use alloy::{eips::BlockNumberOrTag, providers::Provider, transports::http::reqwest::Url};
    use base::signer::Signer;
    use block_header_common::types::{Claim, EncodableClaim};
    use std::{env::current_dir, io::Write};

    #[tokio::test]
    async fn test_generate() {
        let current_dir = current_dir().unwrap();

        let provider = alloy::providers::ProviderBuilder::new()
            .on_http(Url::parse("http://162.55.127.90:8545").unwrap());

        let signer_hex = "0x4c0883a69102937d6231471b5dbb62f0e8b11d6f4c2e7f8a9c2e7f8a9c2e7f8a";
        let signer = Signer::from_hex_key(signer_hex).unwrap();

        let block_number = 22323458;

        let raw_block = provider
            .get_block_by_number(BlockNumberOrTag::Number(block_number))
            .await
            .unwrap()
            .unwrap();

        let (generated_claim_hex, generated_claim_verification_context) =
            generate(&provider, &signer, block_number)
                .await
                .expect("Failed to generate claim");

        let claim_bytes = hex::decode(&generated_claim_hex).expect("Failed to decode claim hex");

        let decoded_encodable_claim = <EncodableClaim as SolValue>::abi_decode(&claim_bytes)
            .expect("Failed to decode encodable claim");

        let decoded_claim: Claim<EncodableEthereumHeader> = decoded_encodable_claim
            .try_into()
            .expect("Failed to convert encodable claim back to claim");

        let decoded_ethereum_header: EthereumHeader = decoded_claim.header.into();

        assert_eq!(decoded_ethereum_header.hash, raw_block.header.hash);
        assert_eq!(decoded_ethereum_header.inner, raw_block.header.inner);
        assert_eq!(
            decoded_ethereum_header.total_difficulty,
            raw_block.header.total_difficulty
        );
        assert_eq!(decoded_ethereum_header.size, raw_block.header.size);

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
