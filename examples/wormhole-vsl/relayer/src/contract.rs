use std::str::FromStr;

use crate::types::{Claim, SettledVerifiedClaim, TheVSL};
use alloy::{
    contract::Error,
    network::EthereumWallet,
    primitives::{ruint::aliases::U256, Address, Bytes, FixedBytes, TxHash},
    providers::ProviderBuilder,
    signers::local::PrivateKeySigner,
    sol,
    sol_types::{SolError, SolValue},
    transports::RpcError,
};
use anyhow::{anyhow, Result};
use reqwest::Url;

sol!(
    #![sol(rpc)]
    #[derive(Debug, PartialEq, Eq)]
    VSL,
    "./src/abi.json"
);

pub async fn get_vsl_contract(address: &str, url: Url, private_key: &str) -> Result<TheVSL> {
    let address: Address = address.parse()?;
    let private_key = private_key.to_string();

    let signer: PrivateKeySigner = private_key
        .parse()
        .map_err(|_| anyhow!("Invalid private key"))?;
    let wallet = EthereumWallet::from(signer);

    let provider = ProviderBuilder::new().wallet(wallet).connect_http(url);

    let contract = VSL::new(address, provider);
    Ok(contract)
}

pub async fn deliver_claim(contract: &TheVSL, claim: Claim) -> Result<TxHash> {
    let hash = FixedBytes::from_str(&claim.data.hash).unwrap();
    let r_u256 = U256::from_str(&claim.data.r).unwrap();
    let r = FixedBytes::from(r_u256);
    let s_u256 = U256::from_str(&claim.data.s).unwrap();
    let s = FixedBytes::from(s_u256);
    let v = U256::from_str(&claim.data.v).unwrap();
    let adjusted_v = (v + U256::from(27)).to::<u8>();
    let signed_claim = SettledVerifiedClaim {
        verifiedClaim: claim.data.verified_claim,
        verifiers: claim.data.verifiers,
    };
    let signed_claim_bytes = Bytes::from(signed_claim.abi_encode());
    let tx_hash = match contract
        .deliverClaim(signed_claim_bytes, hash, r, s, adjusted_v)
        .send()
        .await
    {
        Ok(receipt) => Ok(receipt.tx_hash().clone()),
        Err(e) => {
            if let Error::TransportError(RpcError::ErrorResp(error_payload)) = &e {
                let data = error_payload
                    .data
                    .as_deref()
                    .map(|data| data.get())
                    .ok_or_else(|| anyhow!("No additional data in the error payload"))?;

                return if !data.starts_with("\"0x") {
                    Err(anyhow!("{}", error_payload))
                } else {
                    println!("data: {:?}", data);
                    println!(
                        "data.trim_start_matches(\"0x\"): {:?}",
                        data.trim_start_matches("\"0x\"")
                    );
                    let error_bytes: [u8; 4] =
                        hex::decode(data.trim_start_matches("\"0x").trim_end_matches("\""))
                            .map_err(|e| anyhow!("Failed to decode hex: {}", e))?
                            .try_into()
                            .map_err(|_| anyhow!("Failed to convert decoded bytes to [u8; 4]"))?;

                    Err(match error_bytes {
                        _ => anyhow!("RPC Error Response: selector: {:?}", error_bytes),
                    })
                };
            }
            Err(anyhow!(e))
        }
    };

    match tx_hash {
        Ok(tx_hash) => Ok(tx_hash),
        Err(e) => Err(e),
    }
}

pub async fn claim_exists(contract: &TheVSL, id: String) -> Result<bool> {
    let res = contract.getClaim(id).call().await;

    if let Ok(_) = res {
        return Ok(true);
    }

    if let Err(Error::TransportError(RpcError::ErrorResp(error_payload))) = &res {
        let clean_hex = error_payload
            .data
            .as_deref()
            .map(|data| data.get())
            .ok_or_else(|| anyhow!("No additional data in the error payload"))?
            .trim_start_matches("\"0x")
            .trim_end_matches("\"");

        let error_bytes: [u8; 4] = hex::decode(clean_hex)
            .map_err(|e| anyhow!("Failed to decode hex: {}", e))?
            .try_into()
            .map_err(|_| anyhow!("Failed to convert decoded bytes to [u8; 4]"))?;

        if error_bytes == VSL::ClaimNotFound::SELECTOR {
            return Ok(false);
        }
    }

    Err(anyhow!(res.err().unwrap()))
}
