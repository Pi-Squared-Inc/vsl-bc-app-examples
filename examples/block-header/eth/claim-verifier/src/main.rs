use alloy::primitives::Address;
use base::backend_client::BackendClient;
use base::vsl_utils::vsl_types::{IdentifiableClaim, SettleClaimMessage};
use base::vsl_utils::{
    timestamp::Timestamp,
    vsl_rpc::VslRpcClient,
    vsl_types::{IntoSigned as _, private_key_to_signer},
};
use block_header_common::types::ClaimVerificationContext;
use dotenv::dotenv;
use log::{error, info};
use std::{collections::HashSet, env, str::FromStr};
use verification_block_header_eth::verification::block_header::verify;
pub mod utils;

#[tokio::main]
async fn main() {
    dotenv().ok();
    env_logger::Builder::from_env("RUST_LOG").init();

    let mut settled_claims_ids = HashSet::<String>::new();

    let verifier_private_key = env::var("VERIFIER_PRIVATE_KEY")
        .expect("VERIFIER_PRIVATE_KEY environment variable is not set");

    let verifier_signer = private_key_to_signer(&verifier_private_key);
    let verifier_address = verifier_signer.address();

    let submitter_address_str =
        env::var("SUBMITTER_ADDRESS").expect("SUBMITTER_ADDRESS environment variable is not set");
    let submitter_address =
        Address::from_str(&submitter_address_str).expect("Failed to parse SUBMITTER_ADDRESS");

    let backend_url = env::var("BACKEND_URL").expect("BACKEND_URL environment variable is not set");

    let backend_client = BackendClient::new(&backend_url);

    let vsl_url = env::var("VSL_URL").expect("VSL_URL environment variable is not set");
    let vsl_rpc_client = VslRpcClient::new(&vsl_url)
        .await
        .expect("Failed to create VslRpcClient");

    let mut header_signer_public_key = env::var("HEADER_SIGNER_PUBLIC_KEY")
        .expect("HEADER_SIGNER_PUBLIC_KEY environment variable is not set");
    header_signer_public_key = header_signer_public_key
        .trim_start_matches("0x")
        .to_string();

    let mut since = Timestamp::from_seconds(0);

    let balance_response = vsl_rpc_client
        .get_balance(&verifier_address.to_string())
        .await;

    info!("Balance: {:?}", balance_response);

    info!("Update the settled claims list");

    let settled_claims_response = vsl_rpc_client
        .list_settled_claims_for_receiver(Some(&submitter_address.to_string()), since)
        .await;

    match settled_claims_response {
        Ok(settled_claim) => {
            info!("Fetched {} settled claims", settled_claim.len());

            for claim in settled_claim {
                let claim_id = &claim.data.tx().verified_claim.claim_id;

                since = claim.timestamp;
                settled_claims_ids.insert(claim_id.clone());
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

                    let claim = &submitted_claim.data.tx().claim;

                    let claim_id = submitted_claim.data.tx().claim_id();

                    if settled_claims_ids.contains(&claim_id) {
                        info!("Claim ID: {:?} already processed", claim_id);
                        continue;
                    }

                    info!("Claim ID: {:?} will be verified", claim_id);

                    let parsed_context: ClaimVerificationContext =
                        serde_json::from_str(&submitted_claim.data.tx().proof)
                            .expect("Failed to parse claim verification context");

                    let start_time = std::time::Instant::now();
                    match verify(&claim, &header_signer_public_key, &parsed_context) {
                        Ok(_) => {
                            info!("Claim ID: {} is valid", claim_id);
                            let elapsed_time = start_time.elapsed();

                            let elapsed_time_micros = u64::try_from(elapsed_time.as_micros())
                                .expect("Failed to convert elapsed time to milliseconds");

                            info!(
                                "Claim ID: {} is valid, validated in {} ms",
                                claim_id, elapsed_time_micros
                            );

                            let verifier_nonce =
                                utils::get_nonce(&vsl_rpc_client, &verifier_address.to_string())
                                    .await
                                    .expect("Failed to get nonce");

                            let settle_message = SettleClaimMessage {
                                from: verifier_address.to_string(),
                                nonce: verifier_nonce.to_string(),
                                target_claim_id: claim_id.clone(),
                            };

                            let signed_settle_message = settle_message
                                .into_signed(&verifier_signer)
                                .expect("Failed to sign settle message");

                            match vsl_rpc_client.settle_claim(signed_settle_message).await {
                                Ok(settled_claim_id) => {
                                    info!("Claim ID: {} settled successfully", &settled_claim_id);

                                    backend_client
                                        .update_block_header_record_verification_time(
                                            &settled_claim_id,
                                            elapsed_time_micros,
                                        )
                                        .await
                                        .expect("Failed to update verification time in backend");

                                    // Insert after using it for the test
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

        info!("Sleeping for 5 seconds before checking for new claims...");
        tokio::time::sleep(tokio::time::Duration::from_secs(5)).await;
    }
}
