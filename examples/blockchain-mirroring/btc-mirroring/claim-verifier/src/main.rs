use base::vsl_utils::{
    timestamp::Timestamp,
    vsl_rpc::VslRpcClient,
    vsl_types::{IdentifiableClaim,IntoSigned as _, SettleClaimMessage, private_key_to_signer},
};
use dotenvy::dotenv;
use generation_block_processing_btc::types::{Claim, ClaimVerificationContext};
use log::{error, info};
use std::{collections::HashSet, env};
use verification_block_processing_btc::verification::block_processing::verify;
use reqwest::Client;
use serde_json::json;

const CLAIM_SEARCH_COOLDOWN: u64 = 10; // seconds

#[tokio::main]
async fn main() {
    info!("Starting claim verifier...");
    info!("Initializing environment variables...");
    dotenv().ok();
    env_logger::Builder::from_env("RUST_LOG").init();

    let verifier_private_key = env::var("VERIFIER_PRIVATE_KEY")
        .expect("VERIFIER_PRIVATE_KEY environment variable is not set");

    let submitter_address_str =
        env::var("SUBMITTER_ADDRESS").expect("SUBMITTER_ADDRESS environment variable is not set");

    let vsl_rpc_url = env::var("VSL_RPC_URL").expect("VSL_RPC_URL environment variable is not set");
    info!("Environment variables initialized successfully");
    let backend_url = env::var("BACKEND_API_ENDPOINT")
        .expect("BACKEND_API_ENDPOINT environment variable is not set");

    let mut settled_claims_ids = HashSet::<String>::new();

    info!("Initializing Signer with private key");
    let verifier_signer = private_key_to_signer(&verifier_private_key);
    let verifier_address = verifier_signer.address();
    info!("Signer initialized successfully");

    info!("Initializing VSL RPC Client");
    let vsl_rpc_client = VslRpcClient::new(&vsl_rpc_url)
        .await
        .expect("Failed to create VslRpcClient");
    info!("VSL RPC Client initialized successfully");

    info!("Update the settled claims list");
    let mut since = Timestamp::from_seconds(0);
    let settled_claims_response = vsl_rpc_client
        .list_settled_claims_for_receiver(Some(&submitter_address_str), since)
        .await;

    // order settled_claims_response by timestamp
    let settled_claims_response = settled_claims_response.map(|mut claims| {
        claims.sort_by_key(|claim| claim.timestamp);
        claims
    });

    match settled_claims_response {
        Ok(settled_claim) => {
            info!("Fetched {} settled claims", settled_claim.len());

            for claim in settled_claim {
                let claim_id = claim.data.tx().verified_claim.claim_id();
                info!("Processing settled claim ID: {:?}", claim_id);
                since = claim.timestamp;
                settled_claims_ids.insert(claim_id);
            }
        }

        Err(e) => {
            error!("Error fetching settled claims: {}", e);
            return;
        }
    }

    info!("Settled claims list updated");

    loop {
        let result = vsl_rpc_client
            .list_submitted_claims_for_receiver(&verifier_address.to_string(), since)
            .await;
        info!("Fetching claims for address: {}", verifier_address);

        match result {
            Ok(submitted_claims) => {
                info!("Fetched {} claims", submitted_claims.len());

                for submitted_claim in submitted_claims {
                    since = since.max(submitted_claim.timestamp.tick());

                    let inside_claim = &submitted_claim.data.tx().claim;
                    let nonce = vsl_rpc_client
                        .get_account_nonce(&verifier_address.to_string())
                        .await;

                    let nonce = match nonce {
                        Ok(n) => n,
                        Err(err) => {
                            error!("Error fetching nonce: {:?}", err);
                            continue;
                        }
                    };

                    let claim_id = submitted_claim.data.tx().claim_id();

                    if settled_claims_ids.contains(&claim_id) {
                        info!("Claim ID: {:?} already processed", claim_id);
                        continue;
                    }

                    info!("Claim ID: {:?} will be verified", claim_id);

                    let parsed_claim: Claim =
                        serde_json::from_str(&inside_claim).expect("Failed to parse claim");

                    let parsed_context: ClaimVerificationContext =
                        serde_json::from_str(&submitted_claim.data.tx().proof)
                            .expect("Failed to parse claim verification context");

                    let start_time = std::time::Instant::now();
                    match verify(&parsed_claim, &parsed_context) {
                        Ok(_) => {
                            let elapsed_time = start_time.elapsed();
                            info!("Claim ID: {} is valid", claim_id);

                            let verified_claim = SettleClaimMessage {
                                from: verifier_address.to_string(),
                                nonce: nonce.to_string(),
                                target_claim_id: claim_id.clone(),
                            };

                            let signed_verified_claim = verified_claim
                                .into_signed(&private_key_to_signer(&verifier_private_key))
                                .unwrap();

                            match vsl_rpc_client.settle_claim(signed_verified_claim).await {
                                Ok(settled_claim_id) => {
                                    info!("Claim ID: {} settled successfully", &settled_claim_id);
                                    let response = set_verification_time_for_claim(
                                        &Client::new(),
                                        &backend_url,
                                        settled_claim_id.clone(),
                                        "BitcoinBlock".to_string(),
                                        u64::try_from(elapsed_time.as_micros()).unwrap(),
                                    )
                                    .await;

                                    if response.is_ok() {
                                        info!(
                                            "Verification time set for claim ID: {}",
                                            settled_claim_id
                                        );
                                    } else {
                                        error!(
                                            "Failed to set verification time for claim ID: {}",
                                            settled_claim_id
                                        );
                                    }

                                    settled_claims_ids.insert(settled_claim_id);
                                }
                                Err(error) => {
                                    error!("Failed to settle claim: {}", error);
                                }
                            }
                        }
                        Err(e) => {
                            error!("Claim ID: {} is invalid: {:?}", claim_id, e);
                        }
                    }
                }
            }
            Err(e) => {
                error!("Error fetching claims: {}", e);
                break;
            }
        }

        info!(
            "Sleeping for {} seconds before next claim search",
            CLAIM_SEARCH_COOLDOWN
        );
        tokio::time::sleep(tokio::time::Duration::from_secs(CLAIM_SEARCH_COOLDOWN)).await;
    }
}

async fn set_verification_time_for_claim(
    backend_client: &Client,
    backend_url: &str,
    claim_id: String,
    client: String,
    verification_time: u64,
) -> Result<bool, Box<dyn std::error::Error + Send + Sync>> {
    let payload = json!({
        "execution_client": client,
        "claim_id": claim_id,
        "verification_time": verification_time,
    })
    .to_string();

    let response = backend_client
        .post(&format!("{}/block_mirroring_btc_record", backend_url))
        .body(payload)
        .send()
        .await?;

    Ok(response.status().is_success())
}