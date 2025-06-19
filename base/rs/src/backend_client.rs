use reqwest::Client;
use serde::Serialize;

#[derive(Debug, Clone)]
pub struct BackendClient {
    client: Client,
    base_url: String,
}

#[derive(Debug, Serialize)]
pub struct CreateBlockHeaderRequest {
    pub claim_id: String,
    pub chain: String,
    pub block_number: u64,
    pub claim: Option<String>,
    pub error: Option<String>,
    pub verification_time: Option<u64>,
}

#[derive(Debug, Serialize)]
pub struct UpdateVerificationTimeRequest {
    pub verification_time: u64,
}

impl BackendClient {
    pub fn new(base_url: impl Into<String>) -> Self {
        Self {
            client: Client::new(),
            base_url: base_url.into(),
        }
    }

    pub async fn add_block_header_record(
        &self,
        request: CreateBlockHeaderRequest,
    ) -> Result<serde_json::Value, reqwest::Error> {
        let url = format!("{}/block_header_record", self.base_url);
        self.client
            .post(&url)
            .json(&request)
            .send()
            .await?
            .json()
            .await
    }

    pub async fn update_block_header_record_verification_time(
        &self,
        claim_id: &str,
        verification_time: u64,
    ) -> Result<serde_json::Value, reqwest::Error> {
        let url = format!(
            "{}/block_header_records/{}/verification_time",
            self.base_url, claim_id
        );
        let request = UpdateVerificationTimeRequest { verification_time };
        self.client
            .put(&url)
            .json(&request)
            .send()
            .await?
            .json()
            .await
    }
}
