use alloy::consensus::Signed;
use alloy::dyn_abi::SolType;
use alloy::primitives::Address;
use alloy::providers::Provider;
use alloy::rpc::types::Header;
use alloy::signers::local::PrivateKeySigner;
use base::backend_client::{BackendClient, CreateBlockHeaderRequest};
use base::signer::Signer;
use base::vsl_utils::timestamp::Timestamp;
use base::vsl_utils::vsl_rpc::VslRpcClient;
use base::vsl_utils::vsl_types::{
    IntoSigned as _, SettledVerifiedClaim, SubmittedClaim, Timestamped, private_key_to_signer,
};
use block_header_common::types::{Claim, EncodableClaim};
use dotenvy::dotenv;
use generation_block_header_eth::generation::block_header::{EncodableEthereumHeader, generate};
use log::{error, info};
use reqwest::Url;
use std::env;
use std::str::FromStr;
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};
use utils::to_hex;
use web3::futures::StreamExt;
use web3::types::BlockHeader;
pub mod utils;

#[tokio::main]
async fn main() {
    dotenv().ok();
    env_logger::Builder::from_env("RUST_LOG").init();

    let submitter_private_key = env::var("SUBMITTER_PRIVATE_KEY")
        .expect("SUBMITTER_PRIVATE_KEY environment variable is not set");

    let submitter_signer = Arc::new(private_key_to_signer(&submitter_private_key));
    let submitter_address = submitter_signer.address();

    let verifier_address_str =
        env::var("VERIFIER_ADDRESS").expect("VERIFIER_ADDRESS environment variable is not set");

    let verifier_address = Arc::new(
        Address::from_str(&verifier_address_str).expect("Failed to parse VERIFIER_ADDRESS"),
    );

    let backend_url = env::var("BACKEND_URL").expect("BACKEND_URL environment variable is not set");

    let backend_client = Arc::new(BackendClient::new(&backend_url));

    let vsl_url = env::var("VSL_URL").expect("VSL_URL environment variable is not set");
    let vsl_rpc_client = Arc::new(
        VslRpcClient::new(&vsl_url)
            .await
            .expect("Failed to create VslRpcClient"),
    );

    let http_url = env::var("SOURCE_RPC_ENDPOINT")
        .expect("SOURCE_RPC_ENDPOINT environment variable is not set");
    let ws_url = env::var("SOURCE_WEBSOCKET_ENDPOINT")
        .expect("SOURCE_WEBSOCKET_ENDPOINT environment variable is not set");

    let provider = Arc::new(
        alloy::providers::ProviderBuilder::new()
            .connect_http(Url::parse(&http_url).expect("Invalid HTTP URL")),
    );

    let header_signer_private_key = env::var("HEADER_SIGNER_PRIVATE_KEY")
        .expect("HEADER_SIGNER_PRIVATE_KEY environment variable is not set");

    info!("Initializing Signer with private key");
    let header_signer = Arc::new(
        Signer::from_hex_key(&header_signer_private_key)
            .expect("Failed to create signer from private key"),
    );
    info!("Signer initialized successfully");

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

    info!("Check if account already exists");

    let balance_response = vsl_rpc_client
        .get_balance(&submitter_address.to_string())
        .await;

    info!("Balance: {:?}", balance_response);

    while let Some(new_block) = sub.next().await {
        match new_block {
            Ok(block) => {
                let provider = Arc::clone(&provider);
                let header_signer = Arc::clone(&header_signer);
                let vsl_rpc_client = Arc::clone(&vsl_rpc_client);
                let backend_client = Arc::clone(&backend_client);
                let submitter_signer = Arc::clone(&submitter_signer);
                let verifier_address = Arc::clone(&verifier_address);

                tokio::spawn(async move {
                    handle_new_heads_subscription(
                        &vsl_rpc_client,
                        &backend_client,
                        provider,
                        &header_signer,
                        block,
                        &submitter_signer,
                        &verifier_address,
                    )
                    .await;
                });
            }
            Err(err) => {
                error!("Error receiving new block: {:?}", err);
            }
        }
    }
}

async fn handle_new_heads_subscription(
    vsl_rpc_client: &VslRpcClient,
    backend_client: &BackendClient,
    provider: Arc<dyn Provider + Send + Sync>,
    header_signer: &Signer,
    block: BlockHeader,
    submitter_signer: &PrivateKeySigner,
    verifier_address: &Address,
) {
    let block_number = match block.number {
        Some(number) => number.as_u64(),
        None => {
            error!("Block number is missing");
            return;
        }
    };

    let claim_result = generate(&*provider, header_signer, block_number).await;

    match claim_result {
        Ok((hex_encoded_claim, verification_context)) => {
            let current_time = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_secs();

            let nonce = utils::get_nonce(vsl_rpc_client, &submitter_signer.address().to_string())
                .await
                .expect("Failed to get nonce");

            let expiration_time = current_time + 10 * 60;

            let submitted_claim = SubmittedClaim {
                claim: hex_encoded_claim.clone(),
                claim_type: "BlockHeader".to_string(),
                proof: serde_json::to_string(&verification_context).unwrap(),
                nonce: nonce.to_string(),
                to: vec![verifier_address.to_string()],
                quorum: 1,
                from: submitter_signer.address().to_string(),
                expires: Timestamp::from_seconds(expiration_time),
                fee: to_hex("1")
                    .map_err(|e| format!("Failed to convert fee to hex: {}", e))
                    .unwrap_or("0x1".into()),
            };

            let signed_submitted_claim = submitted_claim.into_signed(&submitter_signer).unwrap();

            let claim_id = vsl_rpc_client.submit_claim(signed_submitted_claim).await;

            match claim_id {
                Ok(claim_id) => {
                    info!("Claim submitted successfully with ID: {}", claim_id);

                    let claim_bytes = hex::decode(&hex_encoded_claim)
                        .map_err(|e| format!("Failed to decode claim: {}", e))
                        .unwrap();

                    let decoded_encodable_claim = EncodableClaim::abi_decode(&claim_bytes)
                        .map_err(|e| format!("Failed to decode claim: {}", e))
                        .unwrap();

                    let decoded_claim: Claim<EncodableEthereumHeader> = decoded_encodable_claim
                        .try_into()
                        .map_err(|e: &'static str| format!("Failed to convert claim: {}", e))
                        .unwrap();

                    let claim: Claim<Header> = decoded_claim
                        .map_header(|encodable_header| Header::from(encodable_header).into());

                    let create_header_request = CreateBlockHeaderRequest {
                        claim_id: claim_id.clone(),
                        chain: "Ethereum".to_string(),
                        block_number,
                        claim: Some(serde_json::to_string(&claim).unwrap()),
                        error: None,
                        verification_time: None,
                    };

                    let backend_response = backend_client
                        .add_block_header_record(create_header_request)
                        .await;

                    match backend_response {
                        Ok(_) => {
                            info!("Backend updated successfully for claim ID: {}", claim_id);
                        }
                        Err(err) => {
                            error!("Error adding block header record to backend: {:?}", err);
                        }
                    }

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
