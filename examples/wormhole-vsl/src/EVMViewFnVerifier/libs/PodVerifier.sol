// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {BLS} from "./bls/BLS.sol";

library PodVerifier {
    bytes constant DOMAIN_SEPARATOR = "WARLOCK-CHAOS-V01-CS01-SHA-256";

    struct MerkleProof {
        bytes32[] path;
        uint256 generalizedIndex;
    }

    struct Committee {
        uint256[4][] publicKeys;
    }

    struct CertifiedReceipt {
        bytes32 target;
        bytes32 receiptRoot;
        uint256 bitmask;
        uint256[2] aggregateSignature;
        MerkleProof proof;
    }

    function verifyMerkleProof(
        bytes32 root,
        bytes32 target,
        MerkleProof memory proof
    )
        public
        pure
        returns (bool)
    {
        if (proof.generalizedIndex == 0 && proof.path.length == 0) {
            return target == root;
        }

        bytes32 currentHash = target;
        uint256 currentIndex = proof.generalizedIndex;
        for (uint256 i = 0; i < proof.path.length; i++) {
            if (currentIndex == 0) {
                return false;
            }

            bytes32 sibling = proof.path[i];
            if (currentIndex % 2 == 0) {
                currentHash = keccak256(abi.encodePacked(sibling, currentHash));
                currentIndex = (currentIndex - 2) / 2;
            } else {
                currentHash = keccak256(abi.encodePacked(currentHash, sibling));
                currentIndex = (currentIndex - 1) / 2;
            }
        }

        return currentIndex == 0 && currentHash == root;
    }

    function verifyCertifiedReceipt(
        Committee memory committee,
        CertifiedReceipt memory certifiedReceipt
    )
        public
        view
        returns (bool)
    {
        uint256[4][] memory publicKeys = new uint256[4][](committee.publicKeys.length);
        uint256 publicKeyCount = 0;
        for (uint256 i = 0; i < committee.publicKeys.length; i++) {
            if ((certifiedReceipt.bitmask & (1 << i)) != 0) {
                publicKeys[publicKeyCount++] = committee.publicKeys[i];
            }
        }

        uint256[4][] memory publicKeysTrimmed = new uint256[4][](publicKeyCount);
        for (uint256 i = 0; i < publicKeyCount; i++) {
            publicKeysTrimmed[i] = publicKeys[i];
        }

        uint256[2] memory message =
            BLS.hashToPoint(abi.encodePacked(DOMAIN_SEPARATOR), abi.encode(certifiedReceipt.receiptRoot));

        (bool pairingSuccess, bool callSuccess) =
            BLS.verifyMultiple(certifiedReceipt.aggregateSignature, publicKeysTrimmed, message);

        if (!pairingSuccess || !callSuccess) {
            return false;
        }

        return verifyMerkleProof(
            certifiedReceipt.receiptRoot,
            certifiedReceipt.target,
            certifiedReceipt.proof
        );
    }
}
