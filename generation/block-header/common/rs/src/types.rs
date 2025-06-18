use alloy::{dyn_abi::SolType, sol, sol_types::SolValue};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub enum ClaimType {
    BlockHeader,
}

#[derive(Serialize, Deserialize, Debug)]
pub enum Network {
    Ethereum,
    Bitcoin,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Metadata {
    pub chain_id: u64,
    pub network: Network,
}

#[derive(Serialize, Deserialize, Debug)]
/// Represents a claim with associated metadata, and assumptions.
///
/// # Fields
/// - `claim_type`: The type of the claim, represented by the `ClaimType` enum.
/// - `metadata`: Additional metadata associated with the claim.
/// - `assumptions`: Assumptions or headers related to the claim.
pub struct Claim<H> {
    pub claim_type: ClaimType,
    pub metadata: Metadata,
    pub header: H,
}

#[derive(Serialize, Deserialize, Debug)]
/// Represents the context required for verifying a claim, which includes
/// the necessary verification data.
pub struct ClaimVerificationContext {
    pub signature: String,
}

sol! {
    enum EncodableClaimType {
        BlockHeader
    }

    enum EncodableNetwork {
        Ethereum,
        Bitcoin
    }

    struct EncodableMetadata {
        uint64 chainId;
        EncodableNetwork network;
    }

 struct EncodableClaim {
        EncodableClaimType claimType;
        EncodableMetadata metadata;
        bytes header;
    }

}

impl From<ClaimType> for EncodableClaimType {
    fn from(claim_type: ClaimType) -> Self {
        match claim_type {
            ClaimType::BlockHeader => EncodableClaimType::BlockHeader,
        }
    }
}

impl TryFrom<EncodableClaimType> for ClaimType {
    type Error = &'static str;

    fn try_from(encodable: EncodableClaimType) -> Result<Self, Self::Error> {
        match encodable {
            EncodableClaimType::BlockHeader => Ok(ClaimType::BlockHeader),
            EncodableClaimType::__Invalid => Err("Invalid claim type"),
        }
    }
}

impl From<Network> for EncodableNetwork {
    fn from(network: Network) -> Self {
        match network {
            Network::Ethereum => EncodableNetwork::Ethereum,
            Network::Bitcoin => EncodableNetwork::Bitcoin,
        }
    }
}

impl TryFrom<EncodableNetwork> for Network {
    type Error = &'static str;

    fn try_from(encodable: EncodableNetwork) -> Result<Self, Self::Error> {
        match encodable {
            EncodableNetwork::Ethereum => Ok(Network::Ethereum),
            EncodableNetwork::Bitcoin => Ok(Network::Bitcoin),
            EncodableNetwork::__Invalid => Err("Invalid network"),
        }
    }
}

impl From<Metadata> for EncodableMetadata {
    fn from(metadata: Metadata) -> Self {
        EncodableMetadata {
            chainId: metadata.chain_id,
            network: metadata.network.into(),
        }
    }
}

impl TryFrom<EncodableMetadata> for Metadata {
    type Error = &'static str;

    fn try_from(encodable: EncodableMetadata) -> Result<Self, Self::Error> {
        Ok(Metadata {
            chain_id: encodable.chainId,
            network: encodable.network.try_into()?,
        })
    }
}

impl<H> From<Claim<H>> for EncodableClaim
where
    H: SolValue,
{
    fn from(claim: Claim<H>) -> Self {
        EncodableClaim {
            claimType: claim.claim_type.into(),
            metadata: claim.metadata.into(),
            header: claim.header.abi_encode().into(),
        }
    }
}

impl<H> TryFrom<EncodableClaim> for Claim<H>
where
    H: SolType<RustType = H>,
{
    type Error = &'static str;

    fn try_from(encodable: EncodableClaim) -> Result<Self, Self::Error> {
        Ok(Claim {
            claim_type: encodable.claimType.try_into()?,
            metadata: encodable.metadata.try_into()?,
            header: <H as SolType>::abi_decode(&encodable.header)
                .map_err(|_| "Failed to decode header")?,
        })
    }
}

impl<H> Claim<H> {
    pub fn map_header<T>(self, f: impl FnOnce(H) -> T) -> Claim<T> {
        Claim {
            claim_type: self.claim_type,
            metadata: self.metadata,
            header: f(self.header),
        }
    }
}
