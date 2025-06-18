use alloy::consensus::Signed;
use alloy::primitives::Address;
use alloy::signers::local::PrivateKeySigner;
use base::vsl_utils::timestamp::Timestamp;
use base::vsl_utils::vsl_rpc::VslRpcClient;
use base::vsl_utils::vsl_types::SettledVerifiedClaim;
use base::vsl_utils::vsl_types::{IntoSigned, SubmittedClaim, Timestamped, private_key_to_signer};
use dotenv::dotenv;
use env_logger::Env;
use generation_block_processing_evm::generation::block_processing::generate;
use jsonrpsee::http_client::HttpClient;
use log::{error, info};
use reqwest::Client;
use serde_json::json;
use std::env;
use std::str::FromStr;
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};
use web3::futures::StreamExt;
use web3::types::U64;

#[tokio::main]
async fn main() {
    dotenv().ok();
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();
    let submitter_private_key = env::var("SUBMITTER_PRIVATE_KEY")
        .expect("SUBMITTER_PRIVATE_KEY environment variable is not set");

    let submitter_signer = private_key_to_signer(&submitter_private_key);

    let verifier_address_str =
        env::var("VERIFIER_ADDRESS").expect("VERIFIER_ADDRESS environment variable is not set");

    let verifier_address =
        Address::from_str(&verifier_address_str).expect("Failed to parse VERIFIER_ADDRESS");

    let vsl_url = env::var("VSL_URL").expect("VSL_URL environment variable is not set");
    let vsl_rpc_client = Arc::new(
        VslRpcClient::new(&vsl_url)
            .await
            .expect("Failed to create VslRpcClient"),
    );
    let http_url = env::var("SOURCE_RPC_ENDPOINT")
        .expect("SOURCE_RPC_ENDPOINT environment variable is not set");
    let backend_url = env::var("BACKEND_API_ENDPOINT")
        .expect("BACKEND_API_ENDPOINT environment variable is not set");
    let ws_url = env::var("SOURCE_WEBSOCKET_ENDPOINT")
        .expect("SOURCE_WEBSOCKET_ENDPOINT environment variable is not set");

    info!("Initializing HTTP client with URL: {}", http_url);
    let http_client = HttpClient::builder()
        .max_response_size(40 * 1024 * 1024) // 40 MB
        .request_timeout(std::time::Duration::from_secs(1000))
        .build(http_url)
        .expect("Failed to build HTTP client");
    info!("HTTP client initialized successfully");

    info!("Initializing backend client");
    let backend_client = reqwest::Client::builder()
        .build()
        .expect("Failed to build backend client");
    info!("Backend client initialized successfully");

    info!("Connecting to WebSocket at URL: {}", ws_url);
    let ws = web3::transports::WebSocket::new(ws_url.as_str())
        .await
        .expect(&format!("Failed to connect to WebSocket at {}", ws_url));
    info!("WebSocket connection established successfully");

    let web3 = web3::Web3::new(ws.clone());
    info!("Subscribing to new heads");
    let mut sub = web3
        .eth_subscribe()
        .subscribe_new_heads()
        .await
        .expect("Failed to subscribe");
    info!("Subscription to new heads successfully");

    while let Some(new_block) = sub.next().await {
        match new_block {
            Ok(block) => {
                let block_number = match block.number {
                    Some(num) => num,
                    None => {
                        error!("Received block without a number, skipping");
                        continue;
                    }
                };
                tokio::spawn({
                    let http_client = http_client.clone();
                    let backend_url = backend_url.clone();
                    let backend_client = backend_client.clone();
                    let vsl_rpc_client = vsl_rpc_client.clone();
                    let submitter_signer = submitter_signer.clone();

                    async move {
                        handle_new_heads_subscription(
                            &vsl_rpc_client,
                            &backend_client,
                            &backend_url,
                            http_client,
                            block_number,
                            &submitter_signer,
                            &verifier_address,
                        )
                        .await;
                    }
                });
            }
            Err(err) => {
                error!("Error receiving new block: {:?}", err);
            }
        }
    }

    sub.unsubscribe().await.expect("Failed to unsubscribe");
}

async fn handle_new_heads_subscription(
    vsl_rpc_client: &VslRpcClient,
    backend_client: &Client,
    backend_url: &str,
    client: HttpClient,
    block_number: U64,
    submitter_signer: &PrivateKeySigner,
    verifier_address: &Address,
) {
    let claim_result = generate(&client, block_number.as_u64()).await;
    match claim_result {
        Ok((claim, verification_context)) => {
            let nonce = vsl_rpc_client.get_account_nonce(
                &submitter_signer.address().to_string(),
            ).await;

            let nonce = match nonce {
                Ok(n) => n,
                Err(err) => {
                    error!("Error fetching nonce: {:?}", err);
                    return;
                }
            };

            let current_time = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_secs();

            let expiration_time = current_time + 10 * 60; // 10 minutes from now

            let submitted_claim = SubmittedClaim {
                claim: serde_json::to_string(&claim).unwrap(),
                claim_type: "MirroringReth".to_string(),
                proof: serde_json::to_string(&verification_context).unwrap(),
                nonce: nonce.to_string(),
                to: vec![verifier_address.to_string()],
                quorum: 1,
                from: submitter_signer.address().to_string(),
                expires: Timestamp::from_seconds(expiration_time),
                fee: "0x1".to_string(), // TODO: Update when VSL RPC is ready
            };

            let signed_submitted_claim = submitted_claim.into_signed(&submitter_signer).unwrap();

            let claim_id = vsl_rpc_client.submit_claim(signed_submitted_claim).await;

            match claim_id {
                Ok(claim_id) => {
                    info!("Claim submitted successfully with ID: {}", claim_id);
                    info!("Submitting claim to backend...");
                    let response = submit_claim_to_backend(
                        &backend_client,
                        backend_url,
                        block_number.as_u64(),
                        claim_id.clone(),
                        "MirroringReth".to_string(),
                    )
                    .await;
                    if let Err(err) = response {
                        error!("Error submitting claim to backend: {:?}", err);
                        return;
                    }
                    info!("Claim submitted to backend successfully");
                    match poll_for_settlement(vsl_rpc_client, claim_id.clone(), 10000).await {
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
        Err(err) => {
            error!("Error generating claim: {:?}", err);
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
        .post(&format!("{}/block_mirroring_record", backend_url))
        .body(payload)
        .send()
        .await?;

    Ok(response.status().is_success())
}
