/// Generates a claim for a specific block number by fetching and processing the block data.
///
/// # Arguments
///
/// * `client` - A reference to an `HttpClient` used to interact with the JSON-RPC API.
/// * `block_number` - The block number for which the claim is to be generated.
///
/// # Returns
///
/// Returns a `Result` containing a tuple of:
/// - `Claim`: The generated claim containing block processing details.
/// - `ClaimVerificationContext`: The context required for verifying the claim.
///
/// On failure, returns a `GenerationError` indicating the type of error encountered.
///
/// # Errors
///
/// This function can return the following errors:
/// - `GenerationError::FetchBlockError`: If fetching the block data fails.
/// - `GenerationError::FetchChainConfigError`: If fetching the chain configuration fails.
/// - `GenerationError::DecodeBlockError`: If decoding the block data fails.
/// - `GenerationError::FetchBlockHashError`: If fetching the block hash fails.
/// - `GenerationError::FetchExecutionWitnessError`: If fetching the execution witness fails.
///
use alloy_rlp::Decodable;
use crate::types::{Claim, ClaimType, ClaimVerificationContext, Metadata};
use jsonrpsee::http_client::HttpClient;
use reth_primitives_traits::SealedBlock;
use reth_rpc_api::DebugApiClient;
use reth_stateless::ExecutionWitness;

#[derive(Debug, thiserror::Error)]
pub enum GenerationError {
    #[error("Failed to fetch block: {0}")]
    FetchBlockError(String),
    #[error("Failed to fetch chain config: {0}")]
    FetchChainConfigError(String),
    #[error("Failed to decode block: {0}")]
    DecodeBlockError(String),
    #[error("Failed to fetch block hash: {0}")]
    FetchBlockHashError(String),
    #[error("Failed to fetch execution witness: {0}")]
    FetchExecutionWitnessError(String),
}

/// Asynchronously generates a claim for a given block number by interacting with the JSON-RPC API.
///
/// # Arguments
///
/// * `client` - A reference to an `HttpClient` used to communicate with the JSON-RPC API.
/// * `block_number` - The block number for which the claim is to be generated.
pub async fn generate(
    client: &HttpClient,
    block_number: u64,
) -> Result<(Claim, ClaimVerificationContext), GenerationError> {
    let raw_block = DebugApiClient::raw_block(client, block_number.into())
        .await
        .map_err(|e| GenerationError::FetchBlockError(e.to_string()))?;

    let mut raw_block_slice: &[u8] = &raw_block;
    let decoded: SealedBlock<alloy_consensus::Block<reth_primitives::TransactionSigned>> =
        SealedBlock::<reth_primitives::Block>::decode(&mut raw_block_slice)
            .map_err(|e| GenerationError::DecodeBlockError(e.to_string()))?;

    let recovered_block = decoded.clone().try_recover().unwrap();

    let chain_config = DebugApiClient::debug_chain_config(client)
        .await
        .map_err(|e| GenerationError::FetchChainConfigError(e.to_string()))?;

    let witness =
        DebugApiClient::debug_execution_witness_by_block_hash(client, recovered_block.hash())
            .await
            .map_err(|e| GenerationError::FetchExecutionWitnessError(e.to_string()))?;

    Ok((
        Claim {
            claim_type: ClaimType::BlockProcessing,
            assumptions: recovered_block.header().clone(),
            metadata: Metadata {
                chain_id: chain_config.chain_id,
            },
            result: raw_block,
        },
        ClaimVerificationContext {
            // TODO: No conversion is needed anymore when ExecutionWitness has a unified type. see details in https://github.com/paradigmxyz/reth/commit/3e5c230f4df14ebc0af2c5f2a71c45edd6c6ea00#diff-077fe63dfd04c1eef775929690afe80b1554904d5bb4260900f68de550783dbaR2.
            witness: ExecutionWitness {
                state: witness.state,
                codes: witness.codes.clone(),
                keys: witness.keys.clone(),
                headers: witness.headers.clone(),
            },
        },
    ))
}

#[cfg(test)]
mod tests {
    use super::*;
    use jsonrpsee::http_client::HttpClient;
    use serde_json;
    use std::env::current_dir;
    use std::io::Write;

    /// This test generates mock data for block processing and stores it in files.
    /// The files include:
    /// - `block_processing_test_mock_raw_block.txt`: Contains the raw block data, used for Mocking.
    /// - `block_processing_test_mock_witness.json`: Contains the execution witness data in JSON format, used for Mocking.
    /// - `block_processing_test_mock_chain_config.json`: Contains the chain configuration in JSON format, used for Mocking.
    /// - `block_processing_test_mock_claim.json`: Contains the generated claim in JSON format, used for claim validator unit testing.
    /// - `block_processing_test_mock_verification_context.json`: Contains the claim verification context in JSON format, used for claim validator unit testing.
    ///
    /// These files are used for testing and debugging purposes.
    #[tokio::test]
    async fn test_generate() {
        let current_dir = current_dir().unwrap();
        let http_client = HttpClient::builder()
            .max_response_size(40 * 1024 * 1024) // 40 MB
            .request_timeout(std::time::Duration::from_secs(1000))
            // RPC API URL for getting the block info & witness
            .build("http://162.55.127.90:8545")
            .expect("Failed to build HTTP client");
        // Block number to fetch
        let block_number: u64 = 22515065;

        let raw_block = DebugApiClient::raw_block(&http_client, block_number.into())
            .await
            .expect("Failed to fetch block");

        let mut raw_block_slice: &[u8] = &raw_block;
        let decoded: SealedBlock<alloy_consensus::Block<reth_primitives::TransactionSigned>> =
            SealedBlock::<reth_primitives::Block>::decode(&mut raw_block_slice)
                .expect("Failed to decode block");

        let recovered_block = decoded.clone().try_recover().unwrap();

        let chain_config = DebugApiClient::debug_chain_config(&http_client)
            .await
            .expect("Failed to fetch chain config");

        let witness = DebugApiClient::debug_execution_witness_by_block_hash(
            &http_client,
            recovered_block.hash(),
        )
        .await
        .expect("Failed to fetch execution witness");

        // store recovered block in block_processing_test_mock_raw_block.txt
        let mut file = std::fs::File::create(
            current_dir.join("src/generation/block_processing_test_mock_raw_block.txt"),
        )
        .expect("Failed to create file");

        file.write_all(&raw_block.0)
            .expect("Failed to write raw block to file");
        file.flush().expect("Failed to flush file");

        // store witness in block_processing_test_mock_witness.json
        let mut file = std::fs::File::create(
            current_dir.join("src/generation/block_processing_test_mock_witness.json"),
        )
        .expect("Failed to create file");
        let witness_json =
            serde_json::to_string(&witness).expect("Failed to serialize witness to JSON");
        file.write_all(witness_json.as_bytes())
            .expect("Failed to write witness to file");
        file.flush().expect("Failed to flush file");

        // store chain config in block_processing_test_mock_chain_config.json
        let mut file = std::fs::File::create(
            current_dir.join("src/generation/block_processing_test_mock_chain_config.json"),
        )
        .expect("Failed to create file");
        let chain_config_json =
            serde_json::to_string(&chain_config).expect("Failed to serialize chain config to JSON");
        file.write_all(chain_config_json.as_bytes())
            .expect("Failed to write chain config to file");
        file.flush().expect("Failed to flush file");

        let claim = Claim {
            claim_type: ClaimType::BlockProcessing,
            assumptions: recovered_block.header().clone(),
            metadata: Metadata {
                chain_id: chain_config.chain_id,
            },
            result: raw_block,
        };
        let context = ClaimVerificationContext {
            witness: ExecutionWitness {
                state: witness.state,
                codes: witness.codes.clone(),
                keys: witness.keys.clone(),
                headers: witness.headers.clone(),
            },
        };

        let claim_json = serde_json::to_string(&claim).expect("Failed to serialize claim to JSON");
        let context_json =
            serde_json::to_string(&context).expect("Failed to serialize context to JSON");
        // store claim in block_processing_test_mock_claim.json
        let mut file = std::fs::File::create(
            current_dir.join("src/generation/block_processing_test_mock_claim.json"),
        )
        .expect("Failed to create file");
        file.write_all(claim_json.as_bytes())
            .expect("Failed to write claim to file");
        file.flush().expect("Failed to flush file");
        let mut file = std::fs::File::create(
            current_dir.join("src/generation/block_processing_test_mock_verification_context.json"),
        )
        .expect("Failed to create file");
        file.write_all(context_json.as_bytes())
            .expect("Failed to write context to file");
        file.flush().expect("Failed to flush file");
        // Check if the files exist
        assert!(
            std::path::Path::new(
                &current_dir.join("src/generation/block_processing_test_mock_claim.json")
            )
            .exists()
        );
        assert!(
            std::path::Path::new(
                &current_dir
                    .join("src/generation/block_processing_test_mock_verification_context.json")
            )
            .exists()
        );
        assert!(
            std::path::Path::new(
                &current_dir.join("src/generation/block_processing_test_mock_witness.json")
            )
            .exists()
        );
        assert!(
            std::path::Path::new(
                &current_dir.join("src/generation/block_processing_test_mock_chain_config.json")
            )
            .exists()
        );
        assert!(
            std::path::Path::new(
                &current_dir.join("src/generation/block_processing_test_mock_raw_block.txt")
            )
            .exists()
        );
    }
}
