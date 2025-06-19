use std::{
    env,
    str::FromStr,
    time::{SystemTime, UNIX_EPOCH},
};

pub mod utils;
use alloy::{
    consensus::Signed, dyn_abi::SolType, primitives::Address, signers::local::PrivateKeySigner,
};
use base::{
    backend_client::{BackendClient, CreateBlockHeaderRequest},
    btc_rpc::get_block_count,
    signer::Signer,
    vsl_utils::{
        timestamp::Timestamp,
        vsl_rpc::VslRpcClient,
        vsl_types::{
            IntoSigned, SettledVerifiedClaim, SubmittedClaim, Timestamped, private_key_to_signer,
        },
    },
};
use bitcoin::block::Header;
use block_header_common::types::{Claim, EncodableClaim};
use dotenvy::dotenv;
use generation_block_header_btc::generation::block_header::{EncodableBitcoinHeader, generate};
use log::{error, info};
use utils::to_hex;

#[tokio::main]
async fn main() {
    dotenv().ok();
    env_logger::Builder::from_env("RUST_LOG").init();

    let submitter_private_key = env::var("SUBMITTER_PRIVATE_KEY")
        .expect("SUBMITTER_PRIVATE_KEY environment variable is not set");

    let verifier_address_str =
        env::var("VERIFIER_ADDRESS").expect("VERIFIER_ADDRESS environment variable is not set");

    let vsl_url = env::var("VSL_URL").expect("VSL_URL environment variable is not set");

    let backend_url = env::var("BACKEND_URL").expect("BACKEND_URL environment variable is not set");

    let backend_client = BackendClient::new(&backend_url);

    let pool_interval = env::var("POOL_INTERVAL")
        .unwrap_or_else(|_| "10".to_string())
        .parse::<u64>()
        .expect("Failed to parse POOL_INTERVAL");

    let btc_rpc_url =
        env::var("BITCOIN_RPC_URL").expect("BITCOIN_RPC_URL environment variable is not set");

    let header_signer_private_key = env::var("HEADER_SIGNER_PRIVATE_KEY")
        .expect("HEADER_SIGNER_PRIVATE_KEY environment variable is not set");

    let header_signer = Signer::from_hex_key(&header_signer_private_key)
        .expect("Failed to create signer from private key");

    let submitter_signer = private_key_to_signer(&submitter_private_key);
    let submitter_address = submitter_signer.address();

    let verifier_address =
        Address::from_str(&verifier_address_str).expect("Failed to parse VERIFIER_ADDRESS");

    let vsl_rpc_client = VslRpcClient::new(&vsl_url)
        .await
        .expect("Failed to create VslRpcClient");

    let balance_response = vsl_rpc_client
        .get_balance(&submitter_address.to_string())
        .await;

    info!("Balance: {:?}", balance_response);

    info!(
        "Starting listener with pool interval: {} seconds",
        pool_interval
    );

    let _ = start_listener(
        &vsl_rpc_client,
        &backend_client,
        &btc_rpc_url,
        &header_signer,
        &submitter_signer,
        &verifier_address,
        pool_interval,
    )
    .await;
}

pub async fn start_listener(
    vsl_rpc_client: &VslRpcClient,
    backend_client: &BackendClient,
    btc_rpc_url: &str,
    header_signer: &Signer,
    submitter_signer: &PrivateKeySigner,
    verifier_address: &Address,
    pool_interval: u64,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut current_block = get_block_count(btc_rpc_url).await?;

    info!("Starting listener from block height: {}", current_block);

    loop {
        info!(
            "Started new block check, last processed block height: {}",
            current_block
        );

        let last_chain_block_height = get_block_count(btc_rpc_url).await?;

        if last_chain_block_height >= current_block {
            for block_number in current_block..=last_chain_block_height {
                info!("Processing block number: {}", block_number);

                match process_single_block(
                    btc_rpc_url,
                    backend_client,
                    vsl_rpc_client,
                    header_signer,
                    submitter_signer,
                    verifier_address,
                    block_number,
                )
                .await
                {
                    Ok(_) => {
                        current_block = block_number;
                        info!("Successfully processed block {}", block_number);
                    }
                    Err(e) => {
                        error!("Error processing block {}: {:?}", block_number, e);
                    }
                }
            }
        } else {
            info!(
                "No new blocks to process, last chain block height: {}",
                last_chain_block_height
            );
        }

        tokio::time::sleep(tokio::time::Duration::from_secs(pool_interval)).await;
    }
}

async fn process_single_block(
    btc_rpc_url: &str,
    backend_client: &BackendClient,
    vsl_rpc_client: &VslRpcClient,
    header_signer: &Signer,
    submitter_signer: &PrivateKeySigner,
    verifier_address: &Address,
    block_number: u64,
) -> Result<(), Box<dyn std::error::Error>> {
    let claim_result = generate(btc_rpc_url, header_signer, block_number).await?;
    let (hex_encoded_claim, verification_context) = claim_result;

    let submitter_nonce = utils::get_nonce(vsl_rpc_client, &submitter_signer.address().to_string())
        .await
        .expect("Failed to get nonce");

    let current_time = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs();
    let expiration_time = current_time + 10 * 60;

    let submitted_claim = SubmittedClaim {
        claim: hex_encoded_claim.clone(),
        claim_type: "BlockHeader".to_string(),
        proof: serde_json::to_string(&verification_context).unwrap(),
        nonce: submitter_nonce.to_string(),
        to: vec![verifier_address.to_string()],
        quorum: 1,
        from: submitter_signer.address().to_string(),
        expires: Timestamp::from_seconds(expiration_time),
        fee: to_hex("1").map_err(|e| format!("Failed to convert fee to hex: {}", e))?,
    };

    println!("Submitting claim: {:?}", submitted_claim);

    let signed_submitted_claim = submitted_claim.into_signed(&submitter_signer)?;

    match vsl_rpc_client.submit_claim(signed_submitted_claim).await {
        Ok(claim_id) => {
            info!("Claim submitted successfully with ID: {}", claim_id);

            let claim_bytes = hex::decode(&hex_encoded_claim)
                .map_err(|e| format!("Failed to decode claim: {}", e))?;

            let decoded_encodable_claim = EncodableClaim::abi_decode(&claim_bytes)
                .map_err(|e| format!("Failed to decode claim: {}", e))?;

            let decoded_claim: Claim<EncodableBitcoinHeader> =
                decoded_encodable_claim
                    .try_into()
                    .map_err(|e: &'static str| format!("Failed to convert claim: {}", e))?;

            let claim: Claim<Header> =
                decoded_claim.map_header(|encodable_header| Header::from(encodable_header).into());

            let create_header_request = CreateBlockHeaderRequest {
                claim_id: claim_id.clone(),
                chain: "Bitcoin".to_string(),
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
        Err(e) => {
            error!("Error submitting claim: {:?}", e);
        }
    }

    Ok(())
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
