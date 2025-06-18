use alloy::{
    network::EthereumWallet,
    providers::{
        fillers::{
            BlobGasFiller, ChainIdFiller, FillProvider, GasFiller, JoinFill, NonceFiller,
            WalletFiller,
        },
        Identity, RootProvider,
    },
    sol,
};
use base::vsl_utils::timestamp::Timestamp;
use serde::Deserialize;

use crate::contract::VSL::VSLInstance;

pub type TheVSL = VSLInstance<
    FillProvider<
        JoinFill<
            JoinFill<
                Identity,
                JoinFill<GasFiller, JoinFill<BlobGasFiller, JoinFill<NonceFiller, ChainIdFiller>>>,
            >,
            WalletFiller<EthereumWallet>,
        >,
        RootProvider,
    >,
>;

#[derive(Clone, Debug, Deserialize)]
pub struct Claim {
    pub data: ClaimData,
    pub timestamp: Timestamp,
}

#[derive(Clone, Debug, Deserialize)]
pub struct ClaimData {
    pub verified_claim: VerifiedClaim,
    pub verifiers: Vec<String>,
    pub r: String,
    pub s: String,
    pub v: String,
    pub hash: String,
}

sol! {
    #[derive(Debug, Deserialize)]
    struct VerifiedClaim {
        string claim;
        string claim_id;
        string claim_type;
        string claim_owner;
    }

    #[derive(Debug, Deserialize)]
    struct SettledVerifiedClaim {
        VerifiedClaim verifiedClaim;
        string[] verifiers;
    }
}
