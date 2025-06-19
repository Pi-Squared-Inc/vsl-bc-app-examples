use base::btc_rpc::{get_best_block_hash, get_block_by_hash, get_block_hash_at_height};
use bitcoin::{Block, Transaction, TxIn, TxOut};
use std::fs::File;
use std::io::BufReader;

use crate::types::{Claim, ClaimVerificationContext, GenerationError};

/// Fetches a Bitcoin block using the provided RPC URL, block hash, or block height.
///
/// # Arguments
/// * `rpc_url` - The URL of the Bitcoin node's RPC endpoint.
/// * `block_hash` - An optional block hash as a `String`. Ignored if `block_height` is provided.
/// * `block_height` - An optional block height as a `u64`. Takes precedence over `block_hash`.
///
/// # Behavior
/// - If both `block_hash` and `block_height` are provided, `block_height` takes precedence.
/// - If neither is provided, the latest block (tip) is fetched.
///
/// # Errors
/// Returns `GenerationError::FetchBlockError` if any RPC call fails.
///
/// # Returns
/// Returns the requested `bitcoin::Block` on success.
pub async fn fetch_block_data(
    rpc_url: &str,
    block_hash: Option<String>,
    block_height: Option<u64>,
) -> Result<Block, GenerationError> {
    // Fetch the block hash if neither hash nor height is provided
    let block_hash = if block_hash.is_none() && block_height.is_none() {
        let tip_hash = get_best_block_hash(rpc_url)
            .await
            .map_err(|e| GenerationError::FetchBlockError(e.to_string()))?;
        Some(tip_hash)
    } else if let Some(height) = block_height {
        // Fetch the block hash for the given height
        let hash_at_height = get_block_hash_at_height(rpc_url, height)
            .await
            .map_err(|e| GenerationError::FetchBlockError(e.to_string()))?;
        Some(hash_at_height)
    } else {
        block_hash
    };

    let block = get_block_by_hash(rpc_url, block_hash.as_ref().unwrap())
        .await
        .map_err(|e| GenerationError::FetchBlockError(e.to_string()))?;

    Ok(block)
}

/// Computes the prestate and poststate for a block's transactions.
///
/// # Arguments
/// * `transactions` - A vector of Bitcoin transactions from a block.
///
/// # Returns
/// A tuple containing:
/// - `prestate`: All inputs (`TxIn`) whose previous outputs were not created in this block (i.e., spent outputs from previous blocks).
/// - `poststate`: All outputs (`TxOut`) created in this block that are not spent by any input in this block (i.e., unspent outputs after this block).
///
/// # Details
/// - The function collects all outputs created in the block and all outpoints spent by inputs in the block.
/// - Prestate is formed by inputs that spend outputs not created in this block.
/// - Poststate is formed by outputs that are not spent by any input in this block.
pub fn generate_block_states(transactions: &Vec<Transaction>) -> (Vec<TxIn>, Vec<TxOut>) {
    use std::collections::HashSet;

    // Collect all outputs created in this block
    let block_outputs: HashSet<_> = transactions
        .iter()
        .flat_map(|tx| {
            let txid = tx.compute_txid();
            tx.output
                .iter()
                .enumerate()
                .map(move |(i, _)| (txid, i as u32))
        })
        .collect();

    // Collect all spent outpoints in this block
    let spent_outpoints: HashSet<_> = transactions
        .iter()
        .flat_map(|tx| {
            tx.input
                .iter()
                .map(|txin| (txin.previous_output.txid, txin.previous_output.vout))
        })
        .collect();

    // Prestate: inputs whose previous_output is NOT created in this block
    let prestate: Vec<TxIn> = transactions
        .iter()
        .flat_map(|tx| tx.input.iter().cloned())
        .filter(|txin| {
            !block_outputs.contains(&(txin.previous_output.txid, txin.previous_output.vout))
        })
        .collect();

    // Poststate: outputs that are NOT spent by any input in this block
    let spent_outpoints = &spent_outpoints;
    let poststate: Vec<TxOut> = transactions
        .iter()
        .flat_map(|tx| {
            let txid = tx.compute_txid();
            tx.output.iter().enumerate().filter_map(move |(vout, out)| {
                if !spent_outpoints.contains(&(txid, vout as u32)) {
                    Some(out.clone())
                } else {
                    None
                }
            })
        })
        .collect();

    (prestate, poststate)
}

/// Generates a `Claim` and its verification context for a given Bitcoin block.
///
/// # Arguments
/// * `rpc_url` - The URL of the Bitcoin node's RPC endpoint.
/// * `block_hash` - An optional block hash as a `String`. Ignored if `block_height` is provided.
/// * `block_height` - An optional block height as a `u64`. Takes precedence over `block_hash`.
///
/// # Behavior
/// - If both `block_hash` and `block_height` are provided, `block_height` takes precedence.
/// - If neither is provided, the latest block (tip) is used.
///
/// # Errors
/// Returns `GenerationError::FetchBlockError` if any RPC call fails.
///
/// # Returns
/// Returns a tuple containing:
/// - `Claim`: The generated claim, including prestate, block header, and poststate.
/// - `ClaimVerificationContext`: The context needed to verify the claim (e.g., all block transactions).
pub async fn generate(
    rpc_url: &str,
    block_hash: Option<String>,
    block_height: Option<u64>,
) -> Result<(Claim, ClaimVerificationContext), GenerationError> {
    let block = fetch_block_data(rpc_url, block_hash, block_height).await?;

    let (prestate, poststate) = generate_block_states(&block.txdata);

    Ok((
        Claim {
            prestate,
            transition: block.header,
            poststate,
        },
        ClaimVerificationContext {
            transactions: block.txdata,
        },
    ))
}

/// Generates a `Claim` and its verification context from a provided Bitcoin block.
///
/// # Arguments
/// * `block` - The `bitcoin::Block` to generate the claim and context from.
///
/// # Returns
/// Returns a tuple containing:
/// - `Claim`: The generated claim, including prestate, block header, and poststate.
/// - `ClaimVerificationContext`: The context needed to verify the claim (e.g., all block transactions).
pub fn generate_from_block(
    block: Block,
) -> Result<(Claim, ClaimVerificationContext), GenerationError> {
    let (prestate, poststate) = generate_block_states(&block.txdata);

    Ok((
        Claim {
            prestate,
            transition: block.header,
            poststate,
        },
        ClaimVerificationContext {
            transactions: block.txdata,
        },
    ))
}

/// Writes a `Claim` to a file in JSON format.
///
/// # Arguments
/// * `claim` - A reference to the `Claim` to be written.
/// * `path` - The file path where the claim should be saved.
///
/// # Errors
/// Returns `GenerationError::WriteClaimError` if the file cannot be created or the claim cannot be serialized.
///
/// # Returns
/// Returns `Ok(())` if the claim was successfully written to the file.
pub fn write_claim_to_file(claim: &Claim, path: &str) -> Result<(), GenerationError> {
    let file = File::create(path).map_err(|e| GenerationError::WriteClaimError(e.to_string()))?;
    serde_json::to_writer(file, claim).map_err(|e| GenerationError::WriteClaimError(e.to_string()))
}

/// Reads a `Claim` from a file in JSON format.
///
/// # Arguments
/// * `path` - The file path from which the claim should be read.
///
/// # Errors
/// Returns `GenerationError::WriteClaimError` if the file cannot be opened or the claim cannot be deserialized.
///
/// # Returns
/// Returns the deserialized `Claim` on success.
pub fn read_claim_from_file(path: &str) -> Result<Claim, GenerationError> {
    let file = File::open(path).map_err(|e| GenerationError::WriteClaimError(e.to_string()))?;
    let reader = BufReader::new(file);
    serde_json::from_reader(reader).map_err(|e| GenerationError::WriteClaimError(e.to_string()))
}

/// Writes a `ClaimVerificationContext` (proof) to a file in JSON format.
///
/// # Arguments
/// * `proof` - A reference to the `ClaimVerificationContext` to be written.
/// * `path` - The file path where the proof should be saved.
///
/// # Errors
/// Returns `GenerationError::WriteClaimError` if the file cannot be created or the proof cannot be serialized.
///
/// # Returns
/// Returns `Ok(())` if the proof was successfully written to the file.
pub fn write_proof_to_file(
    proof: &ClaimVerificationContext,
    path: &str,
) -> Result<(), GenerationError> {
    let file = File::create(path).map_err(|e| GenerationError::WriteClaimError(e.to_string()))?;
    serde_json::to_writer(file, proof).map_err(|e| GenerationError::WriteClaimError(e.to_string()))
}

/// Reads a `ClaimVerificationContext` (proof) from a file in JSON format.
///
/// # Arguments
/// * `path` - The file path from which the proof should be read.
///
/// # Errors
/// Returns `GenerationError::WriteClaimError` if the file cannot be opened or the proof cannot be deserialized.
///
/// # Returns
/// Returns the deserialized `ClaimVerificationContext` on success.
pub fn read_proof_from_file(path: &str) -> Result<ClaimVerificationContext, GenerationError> {
    let file = File::open(path).map_err(|e| GenerationError::WriteClaimError(e.to_string()))?;
    let reader = BufReader::new(file);
    serde_json::from_reader(reader).map_err(|e| GenerationError::WriteClaimError(e.to_string()))
}

#[cfg(test)]
mod tests {
    use super::*;
    use bitcoin::Block;
    use std::fs;

    // Use these tests to save the files as json in the tests directory in order to reuse them
    const TEST_BLOCK_HASH: &str =
        "0000000000000000000258705d7087258920fcba1af024f911503a2379ff393f";
    const TEST_BLOCK_HEIGHT: u64 = 896462;

    fn get_test_block() -> Block {
        let test_claim = read_claim_from_file("src/generation/test_claim.json").unwrap();
        let test_proof = read_proof_from_file("src/generation/test_proof.json").unwrap();
        let test_block = Block {
            header: test_claim.transition,
            txdata: test_proof.transactions,
        };
        test_block
    }

    /// Tests fetching a Bitcoin block by hash, by height, and by tip using the RPC endpoint.
    /// Also checks that the fetched block matches the test block from JSON and that fetching by hash and height yields the same block.
    #[tokio::test]
    async fn test_fetch_block_data() {
        let rpc_url = "https://bitcoin-rpc.publicnode.com";

        // By hash
        let block_hash = Some(TEST_BLOCK_HASH.to_string());
        let hash_result = fetch_block_data(&rpc_url, block_hash, None).await;
        assert!(hash_result.is_ok());

        // By height
        let block_height = Some(TEST_BLOCK_HEIGHT);
        let height_result = fetch_block_data(&rpc_url, None, block_height).await;
        assert!(height_result.is_ok());

        // Assert that hash and height results are the same
        let hash_block = hash_result.unwrap();
        let height_block = height_result.unwrap();
        assert!(hash_block.eq(&height_block));

        // Assert that the block is the same as the block in the test jsons
        let test_block = get_test_block();
        assert!(hash_block.eq(&test_block));

        // By none
        let none_result = fetch_block_data(&rpc_url, None, None).await;
        assert!(none_result.is_ok());
    }

    /// Tests the generation of prestate and poststate from a block's transactions.
    /// Asserts that both prestate and poststate are non-empty for the test block.
    #[tokio::test]
    async fn test_generate_block_states() {
        let block = get_test_block();
        let (prestate, poststate) = generate_block_states(&block.txdata);
        assert!(prestate.len() > 0);
        assert!(poststate.len() > 0);
    }

    /// Tests generating a claim and verification context from a block fetched via RPC.
    /// Asserts that the generation succeeds for a known block hash.
    #[tokio::test]
    async fn test_generate() {
        let rpc_url = "https://bitcoin-rpc.publicnode.com";

        let block_hash = Some(TEST_BLOCK_HASH.to_string());
        let result = generate(&rpc_url, block_hash, None).await;
        assert!(result.is_ok());
    }

    /// Tests generating a claim and verification context directly from a provided block.
    /// Asserts that the generation succeeds for the test block loaded from JSON.
    #[tokio::test]
    async fn test_generate_from_block() {
        let block = get_test_block();
        let result = generate_from_block(block);
        assert!(result.is_ok());
    }

    /// Tests writing a claim to a file and reading it back.
    /// Asserts that the claim read from the file matches the original claim.
    /// Cleans up the test file after the test.
    #[tokio::test]
    async fn test_write_and_read_claim_to_file() {
        let rpc_url = "https://bitcoin-rpc.publicnode.com";

        let block_hash = Some(TEST_BLOCK_HASH.to_string());
        let (claim, _) = generate(&rpc_url, block_hash, None).await.unwrap();

        let path = "src/generation/serialize_test_claim.json";
        // Write claim to file
        let write_result = write_claim_to_file(&claim, path);
        assert!(write_result.is_ok());

        // Read claim from file
        let read_result = read_claim_from_file(path);
        assert!(read_result.is_ok());
        let read_claim = read_result.unwrap();

        // The claim read from file should be equal to the original
        assert_eq!(claim.prestate, read_claim.prestate);
        assert_eq!(claim.transition, read_claim.transition);
        assert_eq!(claim.poststate, read_claim.poststate);

        // Clean up
        let _ = fs::remove_file(path);
    }

    /// Tests writing a proof (ClaimVerificationContext) to a file and reading it back.
    /// Asserts that the proof read from the file matches the original proof.
    /// Cleans up the test file after the test.
    #[tokio::test]
    async fn test_write_and_read_proof_to_file() {
        let rpc_url = "https://bitcoin-rpc.publicnode.com";

        let block_hash = Some(TEST_BLOCK_HASH.to_string());
        let (_, proof) = generate(&rpc_url, block_hash, None).await.unwrap();

        let path = "src/generation/serialize_test_proof.json";
        // Write proof to file
        let write_result = write_proof_to_file(&proof, path);
        assert!(write_result.is_ok());

        // Read proof from file
        let read_result = read_proof_from_file(path);
        assert!(read_result.is_ok());
        let read_proof = read_result.unwrap();

        // The proof read from file should be equal to the original
        assert_eq!(proof.transactions, read_proof.transactions);

        // Clean up
        let _ = fs::remove_file(path);
    }
}
