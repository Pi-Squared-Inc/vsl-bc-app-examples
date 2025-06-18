mod contract;
mod types;
mod vsl;

use std::{env, str::FromStr};

use alloy::primitives::Address;
use alloy::signers::local::PrivateKeySigner;
use anyhow::Result;
use base::vsl_utils::timestamp::Timestamp;
use contract::{claim_exists, deliver_claim, get_vsl_contract};
use dotenv::dotenv;
use reqwest::Url;
use serde_json::json;
use tokio::time::{sleep, Duration};
use types::TheVSL;
use vsl::list_settled_claims;

const LOOP_INTERVAL: u64 = 5; // seconds

#[tokio::main]
async fn main() {
    dotenv().ok();
    let rpc: &str = &env::var("DEST_RPC").unwrap();
    let rpc_url = Url::parse(rpc).unwrap();
    let pk: &str = &env::var("DEST_PK").unwrap();
    let vsl_rpc: &str = &env::var("VSL_RPC").unwrap();

    // VSL observer
    let vsl_observer_address_env: &str = &env::var("VSL_OBSERVER_ADDRESS").unwrap();
    let vsl_observer_address = Address::from_str(vsl_observer_address_env).unwrap();

    // Destination VSL contract
    let dest_vsl_contract_address: &str = &env::var("DEST_VSL_ADDRESS").unwrap();
    let dest_vsl_contract = get_vsl_contract(dest_vsl_contract_address, rpc_url.clone(), pk)
        .await
        .unwrap();

    // Wormhole backend
    let wormhole_backend_api_endpoint: &str = &env::var("WORMHOLE_BACKEND_API_ENDPOINT").unwrap();
    let wormhole_backend_api_client = reqwest::Client::new();

    let mut since: Timestamp = Timestamp::now();

    loop {
        match relay_round(
            &dest_vsl_contract,
            since.clone(),
            vsl_rpc,
            vsl_observer_address,
            wormhole_backend_api_endpoint,
            &wormhole_backend_api_client,
        )
        .await
        {
            Ok(ts) => since = ts,
            Err(e) => {
                eprintln!("Error during relay_round: {:?}", e);
            }
        }

        sleep(Duration::from_secs(LOOP_INTERVAL)).await;
    }
}

async fn relay_round(
    vsl_contract: &TheVSL,
    since: Timestamp,
    vsl_rpc: &str,
    address: Address,
    wormhole_backend_api_endpoint: &str,
    api_client: &reqwest::Client,
) -> Result<Timestamp> {
    let claims = list_settled_claims(vsl_rpc, since, address).await?;
    if claims.is_empty() {
        println!("No new claims to relay.");
        return Ok(since);
    }

    let mut since = since;

    println!("Relaying {} claims", claims.len());

    for claim in claims {
        let claim_id = claim.clone().data.verified_claim.claim_id;
        if claim_exists(&vsl_contract, claim_id.clone()).await? {
            println!("Claim already processed: {:?}", claim_id);
            continue;
        }

        println!("Found unprocessed claim: {:?}", claim_id);

        match deliver_claim(&vsl_contract, claim.clone()).await {
            Ok(tx_hash) => {
                println!("New claim verified: {:?}", claim_id);

                since = since.max(claim.timestamp.clone().tick());

                // Update claim destination transaction hash in Wormhole backend
                let upsert_claim_api_url =
                    format!("{}/claim/{}", wormhole_backend_api_endpoint, claim_id);
                let response = api_client
                    .put(upsert_claim_api_url)
                    .json(&json!({ "destination_transaction_hash": tx_hash }))
                    .send()
                    .await?;
                if response.status() != 200 {
                    eprintln!(
                        "Error updating claim destination transaction hash in Wormhole backend: {:?}",
                        response.status()
                    );
                } else {
                    println!(
                        "Claim destination transaction hash updated in Wormhole backend: {:?}",
                        tx_hash
                    );
                }
            }
            Err(e) => {
                eprintln!("Error verifying claim: {:?}", e);
            }
        }
    }

    Ok(since.clone())
}
