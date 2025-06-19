// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/EVMViewFnVerifier/libs/EVMViewFnClaim.sol";

contract ViewFnClaimTest is Test {
    function setUp() public {}

    function test_verifyInclusionProof() public pure {
        bytes32 claimHash = 0xb3dcd8eef90e46f5827338ab3f093a4a7c4367cb64a5ae792d168d0600f2761e;
        bytes32 root = 0x90b95add2edb5e1255e15127d174e2feb803911ab5d12fb2253527ebfda7ea6d;

        bytes32[] memory proof = new bytes32[](3);
        proof[
            0
        ] = 0xfb4e3aed4bc84dbb7a3656200eec8ef3bfdb72a9b777fc47edf039df8d9e8e06;
        proof[
            1
        ] = 0x5d54d4e6f573cbe8b90870f2292d6d6a20c86b561ca1909b6832329297c40721;
        proof[
            2
        ] = 0x773b43f60cdc2f559c37490c5e3ea16a2b2ff6feacd6f8cd80119c64013517e7;

        bytes32 result = EVMViewFnClaim.processProof(proof, claimHash);
        assertEq(result, root);
    }

    function testEncodeDecodeClaim() public view {
        // Create test data
        EVMViewFnClaim.Header memory header = EVMViewFnClaim.Header({
            parentHash: bytes32(uint256(1)),
            uncleHash: bytes32(uint256(2)),
            coinbase: address(0x123),
            root: bytes32(uint256(3)),
            txHash: bytes32(uint256(4)),
            receiptHash: bytes32(uint256(5)),
            bloom: new bytes(256),
            difficulty: 100,
            number: 12345,
            gasLimit: 30000000,
            gasUsed: 21000,
            time: block.timestamp,
            extra: new bytes(0),
            mixDigest: bytes32(uint256(6)),
            nonce: bytes8(uint64(123456))
        });

        EVMViewFnClaim.EVMCall memory evmCall = EVMViewFnClaim.EVMCall({
            from: address(0x456),
            to: address(0x789),
            input: abi.encodeWithSignature("balanceOf(address)", address(0x123))
        });

        EVMViewFnClaim.EVMMetadata memory metadata = EVMViewFnClaim
            .EVMMetadata({chainId: 1});

        EVMViewFnClaim.Claim memory originalClaim = EVMViewFnClaim.Claim({
            claimType: "EVM_VIEW_FN",
            trustBaseSpec: "1.0.0",
            assumptions: header,
            action: evmCall,
            result: abi.encodePacked(bytes32(uint256(1000))),
            metadata: metadata
        });

        // Test encoding and decoding
        bytes memory encoded = abi.encode(originalClaim);
        EVMViewFnClaim.Claim memory decodedClaim = abi.decode(
            encoded,
            (EVMViewFnClaim.Claim)
        );

        // Verify all fields match
        assertEq(
            decodedClaim.claimType,
            originalClaim.claimType,
            "ClaimType mismatch"
        );
        assertEq(
            decodedClaim.trustBaseSpec,
            originalClaim.trustBaseSpec,
            "TrustBaseSpec mismatch"
        );
        assertEq(decodedClaim.result, originalClaim.result, "Result mismatch");
        assertEq(
            decodedClaim.metadata.chainId,
            originalClaim.metadata.chainId,
            "ChainId mismatch"
        );

        // Verify header fields
        assertEq(
            decodedClaim.assumptions.parentHash,
            originalClaim.assumptions.parentHash,
            "ParentHash mismatch"
        );
        assertEq(
            decodedClaim.assumptions.uncleHash,
            originalClaim.assumptions.uncleHash,
            "UncleHash mismatch"
        );
        assertEq(
            decodedClaim.assumptions.coinbase,
            originalClaim.assumptions.coinbase,
            "Coinbase mismatch"
        );
        assertEq(
            decodedClaim.assumptions.number,
            originalClaim.assumptions.number,
            "Block number mismatch"
        );

        // Verify EVM call fields
        assertEq(
            decodedClaim.action.from,
            originalClaim.action.from,
            "From address mismatch"
        );
        assertEq(
            decodedClaim.action.to,
            originalClaim.action.to,
            "To address mismatch"
        );
        assertEq(
            decodedClaim.action.input,
            originalClaim.action.input,
            "Input data mismatch"
        );
    }

    function testEmptyClaimEncodeDecode() public pure {
        // Test with minimal data
        EVMViewFnClaim.Header memory header = EVMViewFnClaim.Header({
            parentHash: bytes32(0),
            uncleHash: bytes32(0),
            coinbase: address(0),
            root: bytes32(0),
            txHash: bytes32(0),
            receiptHash: bytes32(0),
            bloom: new bytes(256),
            difficulty: 0,
            number: 0,
            gasLimit: 0,
            gasUsed: 0,
            time: 0,
            extra: new bytes(0),
            mixDigest: bytes32(0),
            nonce: bytes8(0)
        });

        EVMViewFnClaim.EVMCall memory evmCall = EVMViewFnClaim.EVMCall({
            from: address(0),
            to: address(0),
            input: new bytes(0)
        });

        EVMViewFnClaim.EVMMetadata memory metadata = EVMViewFnClaim
            .EVMMetadata({chainId: 0});

        EVMViewFnClaim.Claim memory emptyClaim = EVMViewFnClaim.Claim({
            claimType: "",
            trustBaseSpec: "",
            assumptions: header,
            action: evmCall,
            result: abi.encodePacked(bytes32(uint256(0))),
            metadata: metadata
        });

        bytes memory encoded = abi.encode(emptyClaim);
        EVMViewFnClaim.Claim memory decoded = abi.decode(
            encoded,
            (EVMViewFnClaim.Claim)
        );

        // Verify empty claim encoding/decoding
        assertEq(
            decoded.claimType,
            emptyClaim.claimType,
            "Empty claimType mismatch"
        );
        assertEq(
            decoded.trustBaseSpec,
            emptyClaim.trustBaseSpec,
            "Empty trustBaseSpec mismatch"
        );
        assertEq(decoded.result, emptyClaim.result, "Empty result mismatch");
        assertEq(
            decoded.metadata.chainId,
            emptyClaim.metadata.chainId,
            "Empty chainId mismatch"
        );
    }

    function testFuzzEncodeDecodeClaim(
        string memory claimType,
        string memory trustBaseSpec,
        bytes memory result,
        uint256 chainId
    ) public pure {
        EVMViewFnClaim.Header memory header = EVMViewFnClaim.Header({
            parentHash: bytes32(uint256(1)),
            uncleHash: bytes32(uint256(2)),
            coinbase: address(0x123),
            root: bytes32(uint256(3)),
            txHash: bytes32(uint256(4)),
            receiptHash: bytes32(uint256(5)),
            bloom: new bytes(256),
            difficulty: 100,
            number: 12345,
            gasLimit: 30000000,
            gasUsed: 21000,
            time: 0,
            extra: new bytes(0),
            mixDigest: bytes32(uint256(6)),
            nonce: bytes8(uint64(123456))
        });

        EVMViewFnClaim.EVMCall memory evmCall = EVMViewFnClaim.EVMCall({
            from: address(0x456),
            to: address(0x789),
            input: abi.encodeWithSignature("balanceOf(address)", address(0x123))
        });

        EVMViewFnClaim.EVMMetadata memory metadata = EVMViewFnClaim
            .EVMMetadata({chainId: chainId});

        EVMViewFnClaim.Claim memory claim = EVMViewFnClaim.Claim({
            claimType: claimType,
            trustBaseSpec: trustBaseSpec,
            assumptions: header,
            action: evmCall,
            result: result,
            metadata: metadata
        });

        bytes memory encoded = abi.encode(claim);
        EVMViewFnClaim.Claim memory decoded = abi.decode(
            encoded,
            (EVMViewFnClaim.Claim)
        );

        assertEq(decoded.claimType, claim.claimType, "Fuzz claimType mismatch");
        assertEq(
            decoded.trustBaseSpec,
            claim.trustBaseSpec,
            "Fuzz trustBaseSpec mismatch"
        );
        assertEq(decoded.result, claim.result, "Fuzz result mismatch");
        assertEq(
            decoded.metadata.chainId,
            claim.metadata.chainId,
            "Fuzz chainId mismatch"
        );
    }

    function test_decodeClaimFromTestdata() public pure {
        bytes
            memory claim = hex"000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000044000000000000000000000000000000000000000000000000000000000000005200000000000000000000000000000000000000000000000000000000000007a69000000000000000000000000000000000000000000000000000000000000000945564d56696577466e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010feeb0b5382ab7fbd3b341a9694c89731e9b58bd58c7e2578f039f3710945b91dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d493470000000000000000000000000000000000000000000000000000000000000000399759893ebb865a56f5cbd37f4fcf2ceedcaed3a5e970cf01d5885597a95ee2d23fec354794389dd3ea4a833bfa9c945075ea0c582e0cdb434ac38959bd6c6ecbf4885748d04361fc44ede78e7e8b7a8189ef70bc72a8c47c1264cb490f824c00000000000000000000000000000000000000000000000000000000000001e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a80000000000000000000000000000000000000000000000000000000001c9c38000000000000000000000000000000000000000000000000000000000000867dd00000000000000000000000000000000000000000000000000000000679303e40000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000100000400000000000000000000000000000000000000000000800000000000000000004000000000000010010000000000000000000080000000000000000000000000000000000000000000000000200000000000001000008000000000020000000000000100000000400000000080000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000001000000000000000000000000060000002000000000000000000000000020200000040000000000200000000000000000000000000002000000000000010000000001000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266000000000000000000000000dc64a140aa3e981100a9beca4e685f962f0cf6c9000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000440bee938600000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001c00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000008a791620dd6260079bf849dc5567adc3f2fdc3180000000000000000000000008a791620dd6260079bf849dc5567adc3f2fdc31800000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000d99945ff100000000000000000000000000165878a594ca255338adfa4d48449f69242eb8f0000000000000000000000000165878a594ca255338adfa4d48449f69242eb8f00910000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266004f994e54540800000002540be400000000000000000000000000cf7ed3acca5a467e9e704c703e8d87f634fb0fc9000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb922660002000000000000000000";
        EVMViewFnClaim.Claim memory decoded = abi.decode(
            claim,
            (EVMViewFnClaim.Claim)
        );
        console.log("decoded.assumptions.number", decoded.assumptions.number);
    }
}
