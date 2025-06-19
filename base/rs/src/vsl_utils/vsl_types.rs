use std::str::FromStr;

use crate::vsl_utils::timestamp::Timestamp;
use alloy::consensus::Signed;
use alloy::consensus::transaction::{RlpEcdsaDecodableTx, RlpEcdsaEncodableTx};
use alloy::eips::Typed2718;
use alloy::hex::FromHex as _;
use alloy::primitives::{Address, Keccak256, SignatureError, eip191_hash_message};
use alloy::signers::k256::SecretKey;
use alloy::signers::local::PrivateKeySigner;
use alloy::signers::{Error, Signature, SignerSync};
use alloy_rlp::{Decodable, Encodable, RlpDecodable, RlpEncodable};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema)]
pub struct InitAccount {
    pub account: String,
    pub initial_balance: String,
}

// TODO: Consider refactoring SubmittedClaim and VerifiedClaim to merge common fields
#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema)]
pub struct Timestamped<T> {
    pub id: String,
    pub data: T,
    pub timestamp: Timestamp,
}

impl<T> Timestamped<T> {
    pub fn new(id: String, timestamp: Timestamp, data: T) -> Self {
        Self {
            id,
            data,
            timestamp,
        }
    }
}

#[derive(Serialize, Deserialize, Debug)]
pub struct AccountSpec {
    pub id: String,
    pub balance: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct TokenSpec {
    pub ticker_symbol: String,
    pub creator_id: String,
    pub creator_balance: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct GenesisSettings {
    pub accounts: Vec<AccountSpec>,
    pub tokens: Vec<TokenSpec>,
}

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema, RlpEncodable, RlpDecodable)]
pub struct SubmittedClaim {
    /// the claim to be verified
    pub claim: String,
    /// the claim type
    pub claim_type: String,
    /// the proof of the claim
    pub proof: String,
    /// the client nonce
    pub nonce: String,
    /// the list of verifiers to receive this claim (currently a singleton list)
    pub to: Vec<String>,
    //the minimum quorum of signatures (currently 1)
    pub quorum: u16,
    // the client account requesting verification
    pub from: String,
    // the time after which the claim is dropped if not enough verifications are received
    pub expires: Timestamp,
    /// the total fee for verification and claim validation
    pub fee: String,
}

impl IdentifiableClaim for SubmittedClaim {
    fn claim_str(&self) -> &str {
        &self.claim
    }

    fn claim_nonce_str(&self) -> &str {
        &self.nonce
    }

    fn claim_owner_str(&self) -> &str {
        &self.from
    }
}

impl HasReceivers for SubmittedClaim {
    fn receivers(&self) -> Vec<Address> {
        self.to
            .iter()
            .map(|a| Address::from_str(&a).unwrap())
            .collect()
    }
}

pub trait HasSender {
    /// The sender of this transaction
    /// Commonly the sender is recovered from a signature on the transaction instead.
    fn sender(&self) -> Option<Address> {
        None
    }
}

impl<T> HasSender for Signed<T>
where
    T: Encodable,
{
    fn sender(&self) -> Option<Address> {
        let mut msg: Vec<u8> = Vec::new();
        self.tx().encode(&mut msg);
        self.signature().recover_address_from_msg(msg).ok()
    }
}

impl<T> HasSender for Timestamped<T>
where
    T: HasSender,
{
    fn sender(&self) -> Option<Address> {
        self.data.sender()
    }
}

pub trait HasReceivers {
    /// The intended receivers of this transaction
    fn receivers(&self) -> Vec<Address> {
        vec![]
    }
}

impl<T> HasReceivers for Signed<T>
where
    T: HasReceivers,
{
    fn receivers(&self) -> Vec<Address> {
        self.tx().receivers()
    }
}

impl<T> HasReceivers for Timestamped<T>
where
    T: HasReceivers,
{
    fn receivers(&self) -> Vec<Address> {
        self.data.receivers()
    }
}

/// Trait to turn an [Encodable] object into a [Signed] one.
///
/// The object is encoded intro a stream of bytes, then prefixed according to
/// the [EIP-191](https://eips.ethereum.org/EIPS/eip-191) standard
/// (to allow using the [Signed] methods for extracting the address)
pub trait IntoSigned: Sized + Encodable + HasSender {
    fn into_signed(self, signer: &PrivateKeySigner) -> Result<Signed<Self>, Error> {
        let mut buf: Vec<u8> = Vec::new();
        self.encode(&mut buf);
        let hash = eip191_hash_message(buf);
        let sig = signer.sign_hash_sync(&hash)?;
        Ok(Signed::new_unchecked(self, sig, hash))
    }

    fn recover_address(&self, sig: &Signature) -> Result<Address, SignatureError> {
        let mut msg: Vec<u8> = Vec::new();
        self.encode(&mut msg);
        sig.recover_address_from_msg(&msg)
    }

    fn check(signed: &Signed<Self>) -> bool {
        let data = signed.tx();
        let signer = match data.recover_address(signed.signature()) {
            Ok(address) => address,
            Err(_) => return false,
        };
        let Some(address) = data.sender() else {
            return false;
        };
        return signer == address;
    }

    fn check_and_strip_signature(signed: Signed<Self>) -> Option<Self> {
        if !Self::check(&signed) {
            return None;
        };
        Some(signed.strip_signature())
    }
}

/// Implement the (undocumented) [RlpEcdsaDecodableTx] and [RlpEcdsaEncodableTx] traits
/// needed for some uses of [Signed], by making a dummy implementation of [Typed2718] with
/// a code of 0u8, and otherwise forwarding to [alloy_rlp::Decodable] and [alloy_rlp::Encodable].
macro_rules! impl_rlp_ecdsa_glue {
    ($type:ty) => {
        impl Typed2718 for $type {
            fn ty(&self) -> u8 {
                0
            }
        }
        impl RlpEcdsaEncodableTx for $type {
            fn rlp_encoded_fields_length(&self) -> usize {
                self.length()
            }

            fn rlp_encode_fields(&self, out: &mut dyn alloy_rlp::BufMut) {
                self.encode(out);
            }
        }

        impl RlpEcdsaDecodableTx for $type {
            const DEFAULT_TX_TYPE: u8 = 0;

            fn rlp_decode_fields(buf: &mut &[u8]) -> alloy_rlp::Result<Self> {
                Decodable::decode(buf)
            }
        }

        impl IntoSigned for $type {}
    };
}

impl HasSender for SubmittedClaim {
    fn sender(&self) -> Option<Address> {
        Address::from_str(&self.from).ok()
    }
}

impl_rlp_ecdsa_glue!(SubmittedClaim);

/// A settled (verified) claim
#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema, RlpEncodable, RlpDecodable)]
pub struct SettledVerifiedClaim {
    /// the claim which was verified
    pub verified_claim: VerifiedClaim,
    /// the addresses of the verifiers which have signed the `verified_claim` object
    pub verifiers: Vec<String>,
}

impl HasSender for SettledVerifiedClaim {
    fn sender(&self) -> Option<Address> {
        self.verified_claim.sender()
    }
}

impl HasReceivers for SettledVerifiedClaim {
    fn receivers(&self) -> Vec<Address> {
        self.verified_claim.receivers()
    }
}

impl_rlp_ecdsa_glue!(SettledVerifiedClaim);

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema, RlpEncodable, RlpDecodable)]
pub struct VerifiedClaim {
    /// the original claim which was verified and now settled
    pub claim: String,
    /// the id (hex-encoded 256 bit hash) of the claim (useful for retrieving the full data of the claim)
    pub claim_id: String,
    /// the claim type
    pub claim_type: String,
    /// the (Ethereum-style) address of the client which produced this claim
    pub claim_owner: String,
}

impl IdentifiableClaim for VerifiedClaim {
    fn claim_str(&self) -> &str {
        &self.claim
    }

    fn claim_nonce_str(&self) -> &str {
        unimplemented!()
    }

    fn claim_owner_str(&self) -> &str {
        &self.claim_owner
    }

    fn claim_id(&self) -> String {
        self.claim_id.clone()
    }
}

impl HasSender for VerifiedClaim {}

impl HasReceivers for VerifiedClaim {
    fn receivers(&self) -> Vec<Address> {
        std::vec![Address::from_str(&self.claim_owner).unwrap()]
    }
}

impl_rlp_ecdsa_glue!(VerifiedClaim);

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema, RlpDecodable, RlpEncodable)]
pub struct SettleClaimMessage {
    /// The address of the verifier requesting claim settlement
    pub from: String,
    /// The nonce of the verifier requesting claim settlement
    pub nonce: String,
    /// The id of the claim for which claim settlement is requested
    pub target_claim_id: String,
}

impl HasSender for SettleClaimMessage {
    fn sender(&self) -> Option<Address> {
        let Ok(addr) = self.from.parse() else {
            return None;
        };
        Some(addr)
    }
}

impl_rlp_ecdsa_glue!(SettleClaimMessage);

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema, RlpDecodable, RlpEncodable)]
pub struct PayMessage {
    pub from: String,
    pub to: String,
    pub amount: String,
    pub nonce: String,
}

impl IdentifiableClaim for PayMessage {
    fn claim_str(&self) -> &str {
        todo!()
    }

    fn claim_nonce_str(&self) -> &str {
        &self.nonce
    }

    fn claim_owner_str(&self) -> &str {
        &self.from
    }

    fn claim_id(&self) -> String {
        Self::claim_id_hash(
            self.claim_owner_str(),
            self.claim_nonce_str(),
            &serde_json::to_string(&ValidatorVerifiedClaim::from(self)).unwrap(),
        )
    }
}

impl HasSender for PayMessage {
    fn sender(&self) -> Option<Address> {
        Address::from_str(&self.from).ok()
    }
}

impl_rlp_ecdsa_glue!(PayMessage);

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema, RlpDecodable, RlpEncodable)]
pub struct CreateAssetMessage {
    /// The address of the account creating the asset
    pub account_id: String,
    /// The nonce of the account creating the asset
    pub nonce: String,
    pub ticker_symbol: String,
    pub total_supply: String,
}

impl IdentifiableClaim for CreateAssetMessage {
    fn claim_str(&self) -> &str {
        todo!()
    }

    fn claim_nonce_str(&self) -> &str {
        &self.nonce
    }

    fn claim_owner_str(&self) -> &str {
        &self.account_id
    }

    fn claim_id(&self) -> String {
        Self::claim_id_hash(
            self.claim_owner_str(),
            self.claim_nonce_str(),
            &serde_json::to_string(&ValidatorVerifiedClaim::from(self)).unwrap(),
        )
    }
}

impl HasSender for CreateAssetMessage {
    fn sender(&self) -> Option<Address> {
        Address::from_str(&self.account_id).ok()
    }
}

impl_rlp_ecdsa_glue!(CreateAssetMessage);

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema, RlpEncodable, RlpDecodable)]
pub struct TransferAssetMessage {
    /// The address of the account transfering the asset
    pub from: String,
    /// The nonce of the account transfering the asset
    pub nonce: String,
    /// The id of the asset (returned when asset was created)
    pub asset_id: String,
    /// The address of the account receiving the asset
    pub to: String,
    /// The amount (of asset) to be transfered
    pub amount: String,
}

impl IdentifiableClaim for TransferAssetMessage {
    fn claim_str(&self) -> &str {
        todo!()
    }

    fn claim_nonce_str(&self) -> &str {
        &self.nonce
    }

    fn claim_owner_str(&self) -> &str {
        &self.from
    }

    fn claim_id(&self) -> String {
        Self::claim_id_hash(
            self.claim_owner_str(),
            self.claim_nonce_str(),
            &serde_json::to_string(&ValidatorVerifiedClaim::from(self)).unwrap(),
        )
    }
}

impl HasSender for TransferAssetMessage {
    fn sender(&self) -> Option<Address> {
        Address::from_str(&self.from).ok()
    }
}

impl_rlp_ecdsa_glue!(TransferAssetMessage);

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema, RlpEncodable, RlpDecodable)]
pub struct SetStateMessage {
    /// The address of the account requesting its state to be changed
    pub from: String,
    /// The nonce of the account requesting its state to be changed
    pub nonce: String,
    /// The new state
    pub state: String,
}

impl HasSender for SetStateMessage {
    fn sender(&self) -> Option<Address> {
        Address::from_str(&self.from).ok()
    }
}

impl_rlp_ecdsa_glue!(SetStateMessage);

#[derive(Debug, Clone, Serialize, Deserialize, JsonSchema)]
pub enum ValidatorVerifiedClaim {
    Payment(PayMessage),
    AssetCreation(CreateAssetMessage),
    AssetTransfer(TransferAssetMessage),
}

impl From<PayMessage> for ValidatorVerifiedClaim {
    fn from(value: PayMessage) -> Self {
        ValidatorVerifiedClaim::Payment(value)
    }
}

impl From<&PayMessage> for ValidatorVerifiedClaim {
    fn from(value: &PayMessage) -> Self {
        ValidatorVerifiedClaim::Payment(value.clone())
    }
}

impl From<CreateAssetMessage> for ValidatorVerifiedClaim {
    fn from(value: CreateAssetMessage) -> Self {
        ValidatorVerifiedClaim::AssetCreation(value)
    }
}

impl From<&CreateAssetMessage> for ValidatorVerifiedClaim {
    fn from(value: &CreateAssetMessage) -> Self {
        ValidatorVerifiedClaim::AssetCreation(value.clone())
    }
}

impl From<TransferAssetMessage> for ValidatorVerifiedClaim {
    fn from(value: TransferAssetMessage) -> Self {
        ValidatorVerifiedClaim::AssetTransfer(value)
    }
}

impl From<&TransferAssetMessage> for ValidatorVerifiedClaim {
    fn from(value: &TransferAssetMessage) -> Self {
        ValidatorVerifiedClaim::AssetTransfer(value.clone())
    }
}

pub fn private_key_to_signer(private_key: &str) -> PrivateKeySigner {
    let bytes = <[u8; 32]>::from_hex(private_key).expect("Could not extract private key");
    let secret_key = SecretKey::from_bytes(&bytes.into()).expect("could not parse private key");
    PrivateKeySigner::from(secret_key)
}

pub fn private_key_to_public(private_key: &str) -> String {
    let bytes = <[u8; 32]>::from_hex(private_key).expect("Could not extract private key");
    let secret_key = SecretKey::from_bytes(&bytes.into()).expect("could not parse private key");
    secret_key.public_key().to_string()
}

pub trait IdentifiableClaim {
    fn claim_str(&self) -> &str;
    fn claim_nonce_str(&self) -> &str;
    fn claim_owner_str(&self) -> &str;
    fn claim_id(&self) -> String {
        Self::claim_id_hash(
            self.claim_owner_str(),
            self.claim_nonce_str(),
            self.claim_str(),
        )
    }

    fn claim_id_hash(owner: &str, nonce: &str, claim: &str) -> String {
        let mut hasher = Keccak256::new();
        hasher.update(owner);
        hasher.update(nonce);
        hasher.update(claim);
        hasher.finalize().to_string()
    }
}
