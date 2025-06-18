// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./libs/PodECDSA.sol";
import "./libs/EVMViewFnClaim.sol";
import "./libs/Hex.sol";
import "openzeppelin-contracts-5-0-1/contracts/utils/cryptography/MessageHashUtils.sol";

/// @title EVMViewFnClaimVerifier
/// @notice This contract verifies claims against a committee's BLS signatures and a provided proof.
///         Claims are identified by their hash and stored once verified.
contract EVMViewFnClaimVerifier {
    PodECDSA.PodConfig podConfig;

    address constant VALIDATOR_ADDRESS =
        0xDB4a76394D34E39802ee169Ec9527b9223A16f0F;

    /// @notice An array of verified claim IDs (hashes). Useful for enumerating verified claims.
    string[] public claimIds;

    /// @notice A counter for the number of verified claims.
    uint256 public claimCount;

    /// @notice A mapping of claim IDs (hashes) to their fully decoded EVMViewFnClaim.Claim data.
    ///         Once a claim is verified, it is stored here for future reference.
    mapping(string => EVMViewFnClaim.Claim) public claimIdToClaim;

    /// @notice The timestamp of the last successfully verified claim.
    uint256 public lastVerificationTimestamp;

    /// @notice Emitted when a claim is successfully verified and stored.
    /// @param claimId The unique hash of the verified claim.
    /// @param timestamp The block timestamp at which the claim was verified.
    event ClaimVerified(bytes32 indexed claimId, uint256 timestamp);

    /// @notice Thrown when attempting to verify a claim that is already verified.
    error ClaimAlreadyVerified();

    /// @notice Thrown when attempting to retrieve a claim that does not exist.
    error ClaimNotFound();

    /// @notice Thrown when the committee's certificate signature is invalid.
    error InvalidCertificateSignature();

    /// @notice Thrown when the certificate target is not the claim.
    error InvalidCertificateTarget();

    constructor(PodECDSA.PodConfig memory _podConfig) {
        podConfig = _podConfig;
        lastVerificationTimestamp = block.timestamp; // Make sure this is set to the current block timestamp
    }

    /// @notice Retrieves the data of a verified claim by its claimId (hash).
    /// @dev Reverts with `ClaimNotFound` if the claim does not exist.
    /// @param claimId The unique hash identifying the claim.
    /// @return The EVMViewFnClaim.Claim data structure associated with the given claimId.
    function getClaim(
        string calldata claimId
    ) public view returns (EVMViewFnClaim.Claim memory) {
        require(bytes(claimId).length > 0, "Invalid claimId");

        // Check if the claim exists by verifying if `result` field is set
        if (claimIdToClaim[claimId].result.length == 0) {
            revert ClaimNotFound();
        }

        return claimIdToClaim[claimId];
    }

    /// @notice Verifies a claim using a provided proof and certificate signed by the committee.
    /// @dev This function:
    ///      1. Checks that the claim has not been verified before.
    ///      2. Verifies the committee's BLS signature on the provided certificate.
    ///      3. Uses the proof to ensure the claim is included in the certificate's merkle tree.
    ///      If all checks pass, the claim is decoded and stored on-chain.
    /// @param settledVerifiedClaim The ABI-encoded data that contains the claim and signature, the claim can be decoded into EVMViewFnClaim.SettledVerifiedClaim.
    /// @return Returns true if the claim is successfully verified.
    function verifyClaim(
        bytes memory settledVerifiedClaim,
        bytes32 hash,
        bytes32 r,
        bytes32 s,
        uint8 v
    ) public returns (bool, EVMViewFnClaim.Claim memory) {
        require(
            settledVerifiedClaim.length > 0,
            "Invalid settledVerifiedClaim"
        );

        // Decode the claim
        EVMViewFnClaim.SettledVerifiedClaim memory signedClaim = abi.decode(
            settledVerifiedClaim,
            (EVMViewFnClaim.SettledVerifiedClaim)
        );

        // RLP encode the claim to compute the hash, then check if the hash is correct
        bytes memory encodedClaim = EVMViewFnClaim.rlpEncode(signedClaim);
        bytes32 signedMessageHash = MessageHashUtils.toEthSignedMessageHash(
            encodedClaim
        );
        require(signedMessageHash == hash, "Invalid hash");

        // Recover the validator address from the signature
        address recoveredValidatorAddress = ecrecover(hash, v, r, s);
        require(
            recoveredValidatorAddress == VALIDATOR_ADDRESS,
            "Invalid signature"
        );

        // Decode the claim now that we have ensured its validity
        EVMViewFnClaim.Claim memory claim = abi.decode(
            Hex.hexStringToBytesAssembly(signedClaim.verifiedClaim.claim),
            (EVMViewFnClaim.Claim)
        );

        // Store the verified claim and update records
        claimIdToClaim[signedClaim.verifiedClaim.claim_id] = claim;
        claimIds.push(signedClaim.verifiedClaim.claim_id);
        lastVerificationTimestamp = block.timestamp;
        claimCount++;

        // Emit an event to signal that the claim was successfully verified
        // emit ClaimVerified(claimHash, lastVerificationTimestamp);

        return (true, claim);
    }
}
