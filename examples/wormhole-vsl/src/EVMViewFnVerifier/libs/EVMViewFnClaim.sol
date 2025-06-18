// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../../RLP/RLPEncode.sol";

library EVMViewFnClaim {
    struct EVMCall {
        address from;
        address to;
        bytes input;
    }

    struct EVMMetadata {
        uint256 chainId;
    }

    struct VerifiedClaim {
        string claim;
        string claim_id;
        string claim_type;
        string claim_owner;
    }

    struct SettledVerifiedClaim {
        VerifiedClaim verifiedClaim;
        string[] verifiers;
    }

    struct Claim {
        string claimType;
        string trustBaseSpec;
        Header assumptions;
        EVMCall action;
        bytes result;
        EVMMetadata metadata;
    }

    struct Header {
        bytes32 parentHash;
        bytes32 uncleHash;
        address coinbase;
        bytes32 root;
        bytes32 txHash;
        bytes32 receiptHash;
        bytes bloom;
        uint256 difficulty;
        uint256 number;
        uint256 gasLimit;
        uint256 gasUsed;
        uint256 time;
        bytes extra;
        bytes32 mixDigest;
        bytes8 nonce;
    }

    // Bitmask for the first callData in the proof.
    // This is used to determine the path of the first callData in the proof.
    // For example, 2 = 0b010, which means that the inclusion proof for the first calldata is [right, left, right]
    // TODO: Maybe use a proper global index for this?
    uint256 constant FIRST_CALLDATA_BITMASK = 2;

    /**
     * @dev Returns true if a `leaf` can be proved to be a part of a Merkle tree
     * defined by `root`. For this, a `proof` must be provided, containing
     * sibling hashes on the branch from the leaf to the root of the tree.
     */
    function verify(
        bytes32[] memory proof,
        bytes32 root,
        bytes32 leaf
    ) internal pure returns (bool) {
        return processProof(proof, leaf) == root;
    }

    function processProof(
        bytes32[] memory proof,
        bytes32 leaf
    ) public pure returns (bytes32) {
        // Use a uint256 as a bitmask where each bit represents left/right position
        // For this example: 0b010 = 2 represents [false, true, false]
        uint256 positionBitmask = FIRST_CALLDATA_BITMASK;

        for (uint256 i = 0; i < proof.length; i++) {
            if ((positionBitmask >> i) & 1 == 1) {
                // Check if the bit at position i is 1 (left). In this case, the proof[i] is on the left
                leaf = keccak256(abi.encodePacked(proof[i], leaf));
            } else {
                // Check if the bit at position i is 0 (right). In this case, the proof[i] is on the right
                leaf = keccak256(abi.encodePacked(leaf, proof[i]));
            }
        }

        return leaf;
    }

    function rlpEncode(
        SettledVerifiedClaim memory claim
    ) internal pure returns (bytes memory) {
        bytes[] memory verifiedClaimFields = new bytes[](4);
        verifiedClaimFields[0] = RLPEncode.encodeString(
            claim.verifiedClaim.claim
        );
        verifiedClaimFields[1] = RLPEncode.encodeString(
            claim.verifiedClaim.claim_id
        );
        verifiedClaimFields[2] = RLPEncode.encodeString(
            claim.verifiedClaim.claim_type
        );
        verifiedClaimFields[3] = RLPEncode.encodeString(
            claim.verifiedClaim.claim_owner
        );
        bytes memory encodedVerifiedClaim = RLPEncode.encodeList(
            verifiedClaimFields
        );

        bytes[] memory verifiers = new bytes[](claim.verifiers.length);
        for (uint256 i = 0; i < claim.verifiers.length; i++) {
            verifiers[i] = RLPEncode.encodeString(claim.verifiers[i]);
        }
        bytes memory encodedVerifierSignatures = RLPEncode.encodeList(
            verifiers
        );

        bytes[] memory settledClaimFields = new bytes[](2);
        settledClaimFields[0] = encodedVerifiedClaim;
        settledClaimFields[1] = encodedVerifierSignatures;
        bytes memory encodedSettledClaim = RLPEncode.encodeList(
            settledClaimFields
        );
        return encodedSettledClaim;
    }
}
