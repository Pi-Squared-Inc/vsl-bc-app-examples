use bitcoin::Block;
use reqwest::Client;
use serde_json::json;
use std::error::Error;

/// Fetches the current block count (chain height) from a Bitcoin Core RPC server.
///
/// # Arguments
/// * `rpc_url` - The URL of the Bitcoin RPC server (e.g., "http://127.0.0.1:8332").
///
/// # Returns
/// * `Ok(u64)` - The current block count.
/// * `Err(Box<dyn Error>)` - If the request or parsing fails.
pub async fn get_block_count(rpc_url: &str) -> Result<u64, Box<dyn Error>> {
    let client = Client::new();
    let req_body = json!({
        "jsonrpc": "1.0",
        "id": "pi2_getblockcount",
        "method": "getblockcount",
        "params": []
    });

    let resp = client.post(rpc_url).json(&req_body).send().await?;

    let resp_json: serde_json::Value = resp.json().await?;
    if let Some(result) = resp_json.get("result") {
        if let Some(height) = result.as_u64() {
            Ok(height)
        } else {
            Err("Invalid result type for block count".into())
        }
    } else {
        Err("No result field in RPC response".into())
    }
}

/// Fetches the block hash at a given height from a Bitcoin Core RPC server.
///
/// # Arguments
/// * `rpc_url` - The URL of the Bitcoin RPC server (e.g., "http://127.0.0.1:8332").
/// * `height` - The block height to query.
///
/// # Returns
/// * `Ok(String)` - The block hash as a hex string.
/// * `Err(Box<dyn Error>)` - If the request or parsing fails.
pub async fn get_block_hash_at_height(
    rpc_url: &str,
    height: u64,
) -> Result<String, Box<dyn Error>> {
    let client = Client::new();
    let req_body = json!({
        "jsonrpc": "1.0",
        "id": "pi2_getblockhash",
        "method": "getblockhash",
        "params": [height]
    });

    let resp = client.post(rpc_url).json(&req_body).send().await?;

    let resp_json: serde_json::Value = resp.json().await?;
    if let Some(result) = resp_json.get("result") {
        if let Some(hash) = result.as_str() {
            Ok(hash.to_string())
        } else {
            Err("Invalid result type for block hash".into())
        }
    } else {
        Err("No result field in RPC response".into())
    }
}

/// Fetches the block information for a given block hash from a Bitcoin Core RPC server,
/// and returns a `bitcoin::Block`.
///
/// # Arguments
/// * `rpc_url` - The URL of the Bitcoin RPC server (e.g., "http://127.0.0.1:8332").
/// * `block_hash` - The block hash as a hex string.
///
/// # Returns
/// * `Ok(Block)` - The block as a `bitcoin::Block`.
/// * `Err(Box<dyn Error>)` - If the request or parsing fails.
pub async fn get_block_by_hash(rpc_url: &str, block_hash: &str) -> Result<Block, Box<dyn Error>> {
    let client = Client::new();
    let req_body = json!({
        "jsonrpc": "1.0",
        "id": "pi2_getblock",
        "method": "getblock",
        "params": [block_hash, 0]
    });

    let resp = client.post(rpc_url).json(&req_body).send().await?;
    let resp_json: serde_json::Value = resp.json().await?;
    let hex = resp_json
        .get("result")
        .and_then(|v| v.as_str())
        .ok_or("No result field or not a string in RPC response")?;

    let raw = hex::decode(hex)?;
    let block: Block = bitcoin::consensus::deserialize(&raw)?;
    Ok(block)
}

/// Fetches the best (tip) block hash from a Bitcoin Core RPC server.
///
/// # Arguments
/// * `rpc_url` - The URL of the Bitcoin RPC server (e.g., "http://127.0.0.1:8332").
///
/// # Returns
/// * `Ok(String)` - The best block hash as a hex string.
/// * `Err(Box<dyn Error>)` - If the request or parsing fails.
pub async fn get_best_block_hash(rpc_url: &str) -> Result<String, Box<dyn Error>> {
    let client = Client::new();
    let req_body = json!({
        "jsonrpc": "1.0",
        "id": "pi2_getbestblockhash",
        "method": "getbestblockhash",
        "params": []
    });

    let resp = client.post(rpc_url).json(&req_body).send().await?;
    let resp_json: serde_json::Value = resp.json().await?;
    if let Some(result) = resp_json.get("result") {
        if let Some(hash) = result.as_str() {
            Ok(hash.to_string())
        } else {
            Err("Invalid result type for best block hash".into())
        }
    } else {
        Err("No result field in RPC response".into())
    }
}
