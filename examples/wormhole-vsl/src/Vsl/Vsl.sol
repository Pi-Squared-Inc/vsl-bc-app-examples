// SPDX-License-Identifier: Apache 2
// This VSL endpoint contract is for the use of Wormhole integration
pragma solidity ^0.8.20;

import {PodECDSA} from "../EVMViewFnVerifier/libs/PodECDSA.sol";
import {EVMViewFnClaimVerifier} from "../EVMViewFnVerifier/EVMViewFnClaimVerifier.sol";
import {EVMViewFnClaim} from "../EVMViewFnVerifier/libs/EVMViewFnClaim.sol";
contract Vsl is EVMViewFnClaimVerifier {
    // =============== Structs ===============
    struct RelayMessage {
        uint16 srcChainid;
        uint16 destChainid;
        address senderTransceiver;
        address receiverTransceiver;
        bytes transceiverMessage;
    }

    // =============== Storage ===============
    // map: destChainid -> (nonce -> RelayMessage)
    mapping(uint16 => mapping(uint => bytes)) public relays;
    mapping(uint16 => uint) public messageNonce;

    // =============== State ===============
    uint16 public immutable chainid;
    address public verifierAddress;

    // =============== Events ===============
    event genStateQueryClaim(
        uint16 srcChainid,
        uint16 destChainid,
        uint blockNumber,
        address contractAddress,
        bytes viewFunctionEncoding
    );

    event DeliveredMessage(
        uint16 destChainid,
        address to,
        bytes transceiverMessage
    );

    // =============== Constructor ===============
    constructor(
        uint16 _chainid,
        PodECDSA.PodConfig memory _podConfig
    ) EVMViewFnClaimVerifier(_podConfig) {
        chainid = _chainid;
    }

    // =============== Functions ===============

    function generateStateQueryClaim(
        uint16 destChainid,
        address senderVslTransciever,
        address receiverVslTransciever,
        bytes memory transceiverMessage
    ) public {
        uint srcChainid = chainid; //There might need a conversion to Wormhole Chainid.
        RelayMessage memory message = RelayMessage(
            uint16(srcChainid),
            destChainid,
            senderVslTransciever,
            receiverVslTransciever,
            transceiverMessage
        );
        bytes memory relayMessage = abi.encode(message);

        uint nonce = messageNonce[destChainid];
        relays[destChainid][nonce] = relayMessage;

        _nextMessageNonce(destChainid);

        uint blockNumber = block.number;

        emit genStateQueryClaim(
            uint16(srcChainid),
            destChainid,
            blockNumber,
            address(this), // For now, keep this to accomadate a common pattern. Need to be reviewed in future.
            abi.encodeWithSignature(
                "relays(uint16,uint256)",
                destChainid,
                nonce
            )
        );
    }

    function deliverClaim(
        bytes calldata settledVerifiedClaim,
        bytes32 hash,
        bytes32 r,
        bytes32 s,
        uint8 v
    ) external {
        (bool result, EVMViewFnClaim.Claim memory claim) = super.verifyClaim(
            settledVerifiedClaim,
            hash,
            r,
            s,
            v
        );
        require(result, "Claim validation failed.");

        RelayMessage memory message = abi.decode(
            abi.decode(claim.result, (bytes)),
            (RelayMessage)
        );

        require(
            chainid == message.destChainid,
            "Recipeint chain id does not match."
        );

        _deliverMessage(message);
    }

    function _deliverMessage(RelayMessage memory message) internal {
        address recipient = message.receiverTransceiver;
        bytes memory tmsg = message.transceiverMessage;

        (bool success, ) = recipient.call(
            abi.encodeWithSignature(
                "receiveMessages(bytes,bytes32,uint16)",
                tmsg,
                bytes32(uint256(uint160(recipient))),
                message.srcChainid
            )
        );

        require(
            success,
            "Failed to deliver message to receiverTransceiver contract"
        );

        emit DeliveredMessage(message.destChainid, recipient, tmsg);
    }

    function _nextMessageNonce(uint16 destChainid) private {
        messageNonce[destChainid] += 1;
    }

    /**
     * TODO: This is a duplicate function. We should remove it.
     */
    function parseStateQueryClaim(
        bytes calldata encodedClaim //It has to be calldata as the data location, since memory data location does not support bytes array access.
    )
        public
        pure
        returns (uint256, uint256, address, bytes memory, bytes memory)
    {
        EVMViewFnClaim.Claim memory claim = abi.decode(
            encodedClaim,
            (EVMViewFnClaim.Claim)
        );
        return (
            claim.metadata.chainId,
            claim.assumptions.number,
            claim.action.to,
            claim.action.input,
            claim.result
        );
    }
}
