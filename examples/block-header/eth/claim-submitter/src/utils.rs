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

pub fn to_hex(s: &str) -> Result<String, Box<dyn StdError + Send + Sync>> {
    if s.starts_with("0x") {
        if u64::from_str_radix(s.strip_prefix("0x").unwrap_or(s), 16).is_ok() {
            Ok(s.to_string())
        } else {
            Err(format!(
                "Invalid number format: {}, must be a hexadecimal or decimal integer",
                s
            )
            .into())
        }
    } else {
        if let Ok(num) = s.parse::<u64>() {
            Ok(format!("0x{:x}", num))
        } else {
            Err(format!(
                "Invalid number format: {}, must be a hexadecimal or decimal integer",
                s
            )
            .into())
        }
    }
}
