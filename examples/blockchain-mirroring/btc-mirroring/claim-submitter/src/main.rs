use alloy::{consensus::Signed, primitives::Address};
use base::btc_rpc::get_block_count;
use base::vsl_utils::timestamp::Timestamp;
use base::vsl_utils::vsl_rpc::VslRpcClient;
use base::vsl_utils::vsl_types::{
    IntoSigned, SettledVerifiedClaim, SubmittedClaim, Timestamped, private_key_to_signer,
};
use dotenvy::dotenv;
use generation_block_processing_btc::generation::block_processing::{
    fetch_block_data, generate_from_block,
};
use generation_block_processing_btc::types::{Claim, ClaimVerificationContext};
use log::{error, info};
use reqwest::Client;
use serde_json::json;
use std::env;
use std::sync::Arc;
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use tokio::time::sleep;

const BLOCK_SEARCH_COOLDOWN: Duration = Duration::from_secs(60);

#[tokio::main]
async fn main() {
    info!("Starting claim submitter...");
    info!("Initializing environment variables...");
    dotenv().ok();
    env_logger::Builder::from_env("RUST_LOG").init();
    let submitter_private_key =
        env::var("SUBMITTER_PRIVATE_KEY").expect("SUBMITTER_PRIVATE_KEY not set");
    let _ = env::var("VERIFIER_ADDRESS").expect("VERIFIER_ADDRESS not set");
    let http_url = env::var("SOURCE_RPC_ENDPOINT").expect("SOURCE_RPC_ENDPOINT not set");
    info!("Environment variables initialized successfully");
    let backend_url = env::var("BACKEND_API_ENDPOINT")
        .expect("BACKEND_API_ENDPOINT environment variable is not set");

    info!("Initializing Signer with private key");
    let submitter_signer = private_key_to_signer(&submitter_private_key);
    let submitter_address = submitter_signer.address();
    info!("Signer initialized successfully");

    info!("Initializing VSL RPC Client");
    let vsl_rpc_url = env::var("VSL_RPC_URL").expect("VSL_RPC_URL environment variable is not set");
    let vsl_rpc_client = Arc::new(
        VslRpcClient::new(&vsl_rpc_url)
            .await
            .expect("Failed to create VslRpcClient"),
    );
    info!("VSL RPC Client initialized successfully");

    info!("Initializing backend client");
    let backend_client = reqwest::Client::builder()
        .build()
        .expect("Failed to build backend client");
    info!("Backend client initialized successfully");

    // Initialize last processed block height (None means process from the current tip)
    let mut last_block_height: Option<u64> = None;
    loop {
        let latest_height = match get_block_count(&http_url).await {
            Ok(height) => height,
            Err(e) => {
                error!("Failed to fetch block height: {}", e);
                sleep(BLOCK_SEARCH_COOLDOWN).await;
                continue;
            }
        };

        let start_height = last_block_height.map(|h| h + 1).unwrap_or(latest_height);

        for height in start_height..=latest_height {
            let block_data = match fetch_block_data(&http_url, None, Some(height)).await {
                Ok(block) => block,
                Err(e) => {
                    error!("Failed to fetch block at height {}: {}", height, e);
                    break;
                }
            };

            info!("Processing block at height: {}", height);

            match generate_from_block(block_data) {
                Ok((claim, verification_context)) => {
                    info!("Claim generated successfully for block {}", height);
                    let vsl_rpc_client = Arc::clone(&vsl_rpc_client);
                    let claim_clone = claim.clone();
                    let backend_url = backend_url.clone();
                    let backend_client = backend_client.clone();
                    tokio::spawn({
                        async move {
                            handle_new_heads_claim_submit(
                                &vsl_rpc_client,
                                &backend_client,
                                &backend_url,
                                height,
                                claim_clone,
                                verification_context.clone(),
                                &submitter_address,
                            )
                            .await;
                        }
                    });
                }
                Err(e) => {
                    error!("Error generating claim for block {}: {}", height, e);
                }
            }
        }
        last_block_height = Some(latest_height);
        info!(
            "Waiting for {} seconds before checking for new blocks...",
            BLOCK_SEARCH_COOLDOWN.as_secs()
        );
        sleep(BLOCK_SEARCH_COOLDOWN).await;
    }
}

async fn handle_new_heads_claim_submit(
    vsl_rpc_client: &VslRpcClient,
    backend_client: &Client,
    backend_url: &str,
    block_number: u64,
    claim: Claim,
    verification_context: ClaimVerificationContext,
    submitter_address: &Address,
) {
    let verifier_address = env::var("VERIFIER_ADDRESS").expect("VERIFIER_ADDRESS not set");
    let current_time = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs();
    let expiration_time = current_time + 10 * 60; // 10 minutes from now
    let nonce = vsl_rpc_client
        .get_account_nonce(&submitter_address.to_string())
        .await;

    let nonce = match nonce {
        Ok(n) => n,
        Err(err) => {
            error!("Error fetching nonce: {:?}", err);
            return;
        }
    };
    let submitted_claim = SubmittedClaim {
        claim: serde_json::to_string(&claim).unwrap(),
        claim_type: "BitcoinBlock".to_string(),
        proof: serde_json::to_string(&verification_context).unwrap(),
        nonce: nonce.to_string(),
        to: vec![verifier_address.to_string()],
        quorum: 1,
        from: submitter_address.to_string(),
        expires: Timestamp::from_seconds(expiration_time),
        fee: "0x1".to_string(),
    };

    let submitter_private_key =
        env::var("SUBMITTER_PRIVATE_KEY").expect("SUBMITTER_PRIVATE_KEY not set");
    let signed_submitted_claim = submitted_claim
        .into_signed(&private_key_to_signer(&submitter_private_key))
        .unwrap();

    let claim_id = vsl_rpc_client.submit_claim(signed_submitted_claim).await;
    match claim_id {
        Ok(claim_id) => {
            info!("Claim submitted successfully with ID: {}", claim_id);
            info!("Submitting claim to backend...");
            let response = submit_claim_to_backend(
                &backend_client,
                backend_url,
                block_number,
                claim_id.clone(),
                "BitcoinBlock".to_string(),
            )
            .await;
            if let Err(err) = response {
                error!("Error submitting claim to backend: {:?}", err);
                return;
            }
            info!("Claim submitted to backend successfully");
            match poll_for_settlement(&vsl_rpc_client, claim_id.clone(), 10000).await {
                Ok(Some(_)) => {
                    info!("Claim: {:?} settled successfully", claim_id);
                }
                Ok(None) => {
                    info!("Claim: {:?} not settled within max attempts.", claim_id);
                }
                Err(err) => {
                    error!("Error polling for settlement: {:?}", err);
                }
            }
        }
        Err(err) => {
            error!("Error submitting claim: {:?}", err);
            return;
        }
    }
}

async fn poll_for_settlement(
    vsl_rpc_client: &VslRpcClient,
    claim_id: String,
    polling_interval_ms: u64,
) -> Result<
    Option<Timestamped<Signed<SettledVerifiedClaim>>>,
    Box<dyn std::error::Error + Send + Sync>,
> {
    use std::time::Duration;
    use tokio::time::sleep;

    loop {
        sleep(Duration::from_millis(polling_interval_ms)).await;
        match vsl_rpc_client.get_settled_claim_by_id(&claim_id).await {
            Ok(settled_claim) => {
                return Ok(Some(settled_claim));
            }
            Err(e) => {
                if e.to_string().contains("Claim not found") {
                    info!(
                        "Claim {:?} not settled yet, retrying in {} ms...",
                        claim_id, polling_interval_ms
                    );
                } else {
                    error!("Error fetching claim: {:?}", e);
                }
                info!(
                    "Claim {:?} not settled yet, retrying in {} ms...",
                    claim_id, polling_interval_ms
                );
            }
        }
    }
}

async fn submit_claim_to_backend(
    backend_client: &Client,
    backend_url: &str,
    block_number: u64,
    claim_id: String,
    client: String,
) -> Result<bool, Box<dyn std::error::Error + Send + Sync>> {
    let payload = json!({
        "block_number": block_number,
        "execution_client": client,
        "claim_id": claim_id,
    })
    .to_string();

    let response = backend_client
        .post(&format!("{}/block_mirroring_btc_record", backend_url))
        .body(payload)
        .send()
        .await?;

    Ok(response.status().is_success())
}
