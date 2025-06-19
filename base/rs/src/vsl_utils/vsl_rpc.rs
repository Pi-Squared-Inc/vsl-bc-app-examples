use super::timestamp::Timestamp;
use super::vsl_types::{SettleClaimMessage, SettledVerifiedClaim, SubmittedClaim, Timestamped};
use crate::vsl_utils::vsl_types::PayMessage;
use alloy::consensus::Signed;
use jsonrpsee::core::client::{ClientT, Subscription, SubscriptionClientT};
use jsonrpsee::core::params::ObjectParams;
use jsonrpsee::http_client::{HttpClient, HttpClientBuilder};
use jsonrpsee::ws_client::{WsClient, WsClientBuilder};
use std::error::Error as StdError;

#[derive(Debug)]
pub struct VslRpcClient {
    http_client: HttpClient,
    ws_client: Option<WsClient>,
}

impl VslRpcClient {
    /// Create a new HTTP-only client connected to the given URL
    pub async fn new(http_url: &str) -> Result<Self, Box<dyn StdError + Send + Sync>> {
        let http_client = HttpClientBuilder::default()
            .max_request_size(u32::MAX)
            .max_response_size(u32::MAX)
            .request_timeout(std::time::Duration::from_secs(120))
            .build(http_url)?;
        Ok(Self {
            http_client,
            ws_client: None,
        })
    }

    /// Create a new client with both HTTP and WebSocket support for subscriptions
    pub async fn new_with_subscription_support(
        http_url: &str,
        ws_url: &str,
    ) -> Result<Self, Box<dyn StdError + Send + Sync>> {
        let http_client = HttpClientBuilder::default()
            .max_response_size(u32::MAX)
            .build(http_url)?;

        let ws_client = WsClientBuilder::default()
            .max_response_size(u32::MAX)
            .build(ws_url)
            .await?;

        Ok(Self {
            http_client,
            ws_client: Some(ws_client),
        })
    }

    /// Submits a request-for-verification claim
    /// Returns the submitted claim hash (claim ID) as a String
    pub async fn submit_claim(
        &self,
        claim: Signed<SubmittedClaim>,
    ) -> Result<String, Box<dyn StdError + Send + Sync>> {
        let mut params = ObjectParams::new();
        params.insert("claim", claim)?;
        Ok(self.http_client.request("vsl_submitClaim", params).await?)
    }

    /// Submits a verified claim for settlement
    /// Returns the settled claim hash (claim ID) as a String
    pub async fn settle_claim(
        &self,
        settled_claim: Signed<SettleClaimMessage>,
    ) -> Result<String, Box<dyn StdError + Send + Sync>> {
        let mut params = ObjectParams::new();
        params.insert("settled_claim", settled_claim)?;
        Ok(self.http_client.request("vsl_settleClaim", params).await?)
    }

    /// Yields recent settled claims for a receiver
    /// Use None for address to get all claims
    pub async fn list_settled_claims_for_receiver(
        &self,
        address: Option<&String>,
        since: Timestamp,
    ) -> Result<Vec<Timestamped<Signed<SettledVerifiedClaim>>>, Box<dyn StdError + Send + Sync>>
    {
        let mut params = ObjectParams::new();
        params.insert("address", address)?;
        params.insert("since", since)?;
        Ok(self
            .http_client
            .request("vsl_listSettledClaimsForReceiver", params)
            .await?)
    }

    /// Yields recent claim verification requests for a receiver
    pub async fn list_submitted_claims_for_receiver(
        &self,
        address: &String,
        since: Timestamp,
    ) -> Result<Vec<Timestamped<Signed<SubmittedClaim>>>, Box<dyn StdError + Send + Sync>> {
        let mut params = ObjectParams::new();
        params.insert("address", address)?;
        params.insert("since", since)?;
        Ok(self
            .http_client
            .request("vsl_listSubmittedClaimsForReceiver", params)
            .await?)
    }

    /// Yields recent settled claims from a sender address
    pub async fn list_settled_claims_for_sender(
        &self,
        address: &String,
        since: Timestamp,
    ) -> Result<Vec<Timestamped<Signed<SettledVerifiedClaim>>>, Box<dyn StdError + Send + Sync>>
    {
        let mut params = ObjectParams::new();
        params.insert("address", address)?;
        params.insert("since", since)?;
        Ok(self
            .http_client
            .request("vsl_listSettledClaimsForSender", params)
            .await?)
    }

    /// Yields recent claim verification requests from a sender address
    pub async fn list_submitted_claims_for_sender(
        &self,
        address: &String,
        since: Timestamp,
    ) -> Result<Vec<Timestamped<Signed<SubmittedClaim>>>, Box<dyn StdError + Send + Sync>> {
        let mut params = ObjectParams::new();
        params.insert("address", address)?;
        params.insert("since", since)?;
        Ok(self
            .http_client
            .request("vsl_listSubmittedClaimsForSender", params)
            .await?)
    }

    /// Retrieves a settled claim by its unique claim ID
    pub async fn get_settled_claim_by_id(
        &self,
        claim_id: &String,
    ) -> Result<Timestamped<Signed<SettledVerifiedClaim>>, Box<dyn StdError + Send + Sync>> {
        let mut params = ObjectParams::new();
        params.insert("claim_id", claim_id)?;
        Ok(self
            .http_client
            .request("vsl_getSettledClaimById", params)
            .await?)
    }

    /// Retrieves the native token balance of a given account
    /// Returns the balance value as a string-encoded float
    pub async fn get_balance(
        &self,
        account_id: &String,
    ) -> Result<String, Box<dyn StdError + Send + Sync>> {
        let mut params = ObjectParams::new();
        params.insert("account_id", account_id)?;
        Ok(self.http_client.request("vsl_getBalance", params).await?)
    }

    /// Checks if the server is up and ready
    /// Returns "ok" if the server is healthy
    pub async fn get_health(&self) -> Result<String, Box<dyn StdError + Send + Sync>> {
        let params = ObjectParams::new();
        Ok(self.http_client.request("vsl_getHealth", params).await?)
    }

    /// Subscribe to real-time settled claims for a receiver
    /// Use None for address to receive all settled claims
    pub async fn subscribe_to_settled_claims_for_receiver(
        &self,
        address: Option<&String>,
    ) -> Result<
        Subscription<Timestamped<Signed<SettledVerifiedClaim>>>,
        Box<dyn StdError + Send + Sync>,
    > {
        let ws_client = self
            .ws_client
            .as_ref()
            .ok_or("WebSocket client not initialized. Use new_with_subscription_support()")?;

        let mut params = ObjectParams::new();
        params.insert("address", address)?;

        let subscription = ws_client
            .subscribe(
                "vsl_subscribeToSettledClaimsForReceiver",
                params,
                "vsl_unsubscribeToSettledClaimsForReceiver",
            )
            .await?;

        Ok(subscription)
    }

    /// Subscribe to real-time submitted claims for a receiver
    pub async fn subscribe_to_submitted_claims_for_receiver(
        &self,
        address: &String,
    ) -> Result<Subscription<Timestamped<Signed<SubmittedClaim>>>, Box<dyn StdError + Send + Sync>>
    {
        let ws_client = self
            .ws_client
            .as_ref()
            .ok_or("WebSocket client not initialized. Use new_with_subscription_support()")?;

        let mut params = ObjectParams::new();
        params.insert("address", address)?;

        let subscription = ws_client
            .subscribe(
                "vsl_subscribeToSubmittedClaimsForReceiver",
                params,
                "vsl_unsubscribeToSubmittedClaimsForReceiver",
            )
            .await?;

        Ok(subscription)
    }

    /// Pays a payment to a receiver
    pub async fn pay(
        &self,
        payment: Signed<PayMessage>,
    ) -> Result<String, Box<dyn StdError + Send + Sync>> {
        let mut params = ObjectParams::new();
        params.insert("payment", payment)?;
        Ok(self.http_client.request("vsl_pay", params).await?)
    }

    /// Retrieves the nonce for a given account
    pub async fn get_account_nonce(
        &self,
        account_id: &String,
    ) -> Result<u64, Box<dyn StdError + Send + Sync>> {
        let mut params = ObjectParams::new();
        params.insert("account_id", account_id)?;
        Ok(self
            .http_client
            .request("vsl_getAccountNonce", params)
            .await?)
    }
}
