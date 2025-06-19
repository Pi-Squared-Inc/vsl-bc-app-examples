use std::error::Error as StdError;

use base::vsl_utils::vsl_rpc::VslRpcClient;

pub async fn get_nonce(
    vsl_rpc_client: &VslRpcClient,
    address: &str,
) -> Result<u64, Box<dyn StdError + Send + Sync>> {
    let nonce = vsl_rpc_client
        .get_account_nonce(&address.to_string())
        .await?;
    Ok(nonce)
}
