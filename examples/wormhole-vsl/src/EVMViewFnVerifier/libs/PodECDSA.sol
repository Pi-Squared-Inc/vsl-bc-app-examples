// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ECDSA} from "./ECDSA.sol";
import {MerkleTree} from "./MerkleTree.sol";
import {IPodRegistry} from "./PodRegistry.sol";

library PodECDSA {
    struct PodConfig {
        uint256 quorum;
        IPodRegistry registry;
    }

    struct CertifiedReceipt {
        bytes32 receiptRoot;
        bytes aggregateSignature;
    }

    struct Certificate {
        CertifiedReceipt certifiedReceipt;
        bytes32 leaf;
        MerkleTree.Proof proof;
    }

    struct MultiCertificate {
        CertifiedReceipt certifiedReceipt;
        bytes32[] leaves;
        MerkleTree.MultiProof proof;
    }

    function verifyCertifiedReceipt(PodConfig memory podConfig, CertifiedReceipt memory certifiedReceipt)
        public
        view
        returns (bool)
    {
        address[] memory validators =
            ECDSA.recoverSigners(certifiedReceipt.receiptRoot, certifiedReceipt.aggregateSignature);
        return podConfig.registry.computeWeight(validators) >= podConfig.quorum;
    }

    function verifyCertificate(PodConfig memory podConfig, Certificate memory certificate) public view returns (bool) {
        return verifyCertifiedReceipt(podConfig, certificate.certifiedReceipt)
            && MerkleTree.verify(certificate.certifiedReceipt.receiptRoot, certificate.leaf, certificate.proof);
    }

    function verifyMultiCertificate(PodConfig memory podConfig, MultiCertificate memory certificate)
        public
        view
        returns (bool)
    {
        return verifyCertifiedReceipt(podConfig, certificate.certifiedReceipt)
            && MerkleTree.verifyMulti(certificate.certifiedReceipt.receiptRoot, certificate.leaves, certificate.proof);
    }
}