use alloy::primitives::Address;
use anyhow::Result;
use base::vsl_utils::timestamp::Timestamp;
use reqwest;
use serde_json::{json, Value};

use crate::types::Claim;

pub async fn list_settled_claims(
    rpc: &str,
    timestamp: Timestamp,
    address: Address,
) -> Result<Vec<Claim>> {
    let client = reqwest::Client::new();

    println!("Since: {:?}", timestamp);
    let payload = json!({
        "jsonrpc": "2.0",
        "method": "vsl_listSettledClaimsForReceiver",
        "params": {
            "since": {
                "seconds": timestamp.seconds(),
                "nanos": timestamp.nanos()
            },
            "address": address.to_string()
        },
        "id": 1
    });

    let response = client
        .post(rpc)
        .header("Content-Type", "application/json")
        .json(&payload)
        .send()
        .await?;

    let response_value = response.json::<Value>().await?;

    let result = response_value["result"]
        .as_array()
        .ok_or("Invalid response format")
        .map_err(|_| anyhow::anyhow!("Invalid response format"))?;

    let claims = result
        .iter()
        .filter_map(|item| {
            serde_json::from_value::<Claim>(item.clone())
                .ok()
                .filter(|claim| !claim.data.verifiers.is_empty())
        })
        .collect::<Vec<Claim>>();

    Ok(claims)
}
