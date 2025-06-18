// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.20;

import "wormhole-solidity-sdk/Utils.sol";
import "ntt/interfaces/ITransceiver.sol";
import "ntt/interfaces/INttManager.sol";
import "ntt/Transceiver/Transceiver.sol";

// =============== Interfaces ===============
interface IVsl {
    function generateStateQueryClaim(
        uint16 destChainid,
        address senderVslTransciever,
        address receiverVslTransciever,
        bytes calldata transceiverMessage
    ) external;
}

contract VslTransceiver is Transceiver {
    // =============== Events ===============
    /// @notice Emitted when a peer transceiver is set.
    /// @dev Topic0
    ///      0xa559263ee060c7a2560843b3a064ff0376c9753ae3e2449b595a3b615d326466.
    /// @param chainId The chain ID of the peer.
    /// @param peerContract The address of the peer contract.
    event SetPeer(uint16 chainId, bytes32 peerContract);

    /// @notice Error the peer contract cannot be the zero address.
    /// @dev Selector: 0x26e0c7de.
    error InvalidWormholePeerZeroAddress();

    /// @notice The chain ID cannot be zero.
    /// @dev Selector: 0x3dd98b24.
    error InvalidWormholeChainIdZero();

    /// @notice Error if the peer has already been set.
    /// @dev Selector: 0xb55eeae9.
    /// @param chainId The chain ID of the peer.
    /// @param peerAddress The address of the peer.
    error PeerAlreadySet(uint16 chainId, bytes32 peerAddress);

    // =============== Constants ===============

    /// @dev Prefix for all TransceiverMessage payloads
    ///      This is 0x99'E''W''H'
    /// @notice Magic string (constant value set by messaging provider) that idenfies the payload as an transceiver-emitted payload.
    ///         Note that this is not a security critical field. It's meant to be used by messaging providers to identify which messages are Transceiver-related.
    bytes4 constant WH_TRANSCEIVER_PAYLOAD_PREFIX = 0x9945FF10;

    /// @dev Prefix for all Wormhole transceiver initialisation payloads
    ///      This is bytes4(keccak256("WormholeTransceiverInit"))
    bytes4 constant WH_TRANSCEIVER_INIT_PREFIX = 0x9c23bd3b;

    /// @dev Prefix for all Wormhole peer registration payloads
    ///      This is bytes4(keccak256("WormholePeerRegistration"))
    bytes4 constant WH_PEER_REGISTRATION_PREFIX = 0x18fc67c2;

    // =============== Storage ===============
    bytes32 private constant PEERS_SLOT =
        bytes32(uint256(keccak256("vslTransceiver.peers")) - 1);

    // =============== State ===============
    address public immutable vslAddress;

    // =============== Constructor ===============
    constructor(
        address _nttManager,
        address _vslAddress
    ) Transceiver(_nttManager) {
        vslAddress = _vslAddress;
    }

    // =============== Methods ===============
    function getTransceiverType()
        external
        pure
        override
        returns (string memory)
    {
        return "VslTransceiver";
    }

    function receiveMessages(
        bytes memory payload,
        bytes32 sourceAddress,
        uint16 sourceChain
    ) external {
        // parse the encoded Transceiver payload
        TransceiverStructs.TransceiverMessage memory parsedTransceiverMessage;
        TransceiverStructs.NttManagerMessage memory parsedNttManagerMessage;
        (parsedTransceiverMessage, parsedNttManagerMessage) = TransceiverStructs
            .parseTransceiverAndNttManagerMessage(
                WH_TRANSCEIVER_PAYLOAD_PREFIX,
                payload
            );

        _deliverToNttManager(
            sourceChain,
            parsedTransceiverMessage.sourceNttManagerAddress,
            parsedTransceiverMessage.recipientNttManagerAddress,
            parsedNttManagerMessage
        );
    }

    function setPeer(
        uint16 peerChainId,
        bytes32 peerContract
    ) external payable onlyOwner {
        if (peerChainId == 0) {
            revert InvalidWormholeChainIdZero();
        }
        if (peerContract == bytes32(0)) {
            revert InvalidWormholePeerZeroAddress();
        }

        bytes32 oldPeerContract = _getPeersStorage()[peerChainId];

        // We don't want to allow updating a peer since this adds complexity in the accountant
        // If the owner makes a mistake with peer registration they should deploy a new Wormhole
        // transceiver and register this new transceiver with the NttManager
        if (oldPeerContract != bytes32(0)) {
            revert PeerAlreadySet(peerChainId, oldPeerContract);
        }

        _getPeersStorage()[peerChainId] = peerContract;

        // Publish a message for this transceiver registration
        // TransceiverStructs.TransceiverRegistration
        //     memory registration = TransceiverStructs.TransceiverRegistration({
        //         transceiverIdentifier: WH_PEER_REGISTRATION_PREFIX,
        //         transceiverChainId: peerChainId,
        //         transceiverAddress: peerContract
        //     });
        // wormhole.publishMessage{value: msg.value}(
        //     0,
        //     TransceiverStructs.encodeTransceiverRegistration(registration),
        //     consistencyLevel
        // );

        emit SetPeer(peerChainId, peerContract);
    }

    function getPeer(uint16 chainId) public view returns (bytes32) {
        return _getPeersStorage()[chainId];
    }

    // =============== Internal ===============
    function _quoteDeliveryPrice(
        uint16 targetChain,
        TransceiverStructs.TransceiverInstruction memory instruction
    ) internal view override returns (uint256 nativePriceQuote) {
        return 0;
    }

    function _sendMessage(
        uint16 recipientChain,
        uint256 deliveryPayment,
        address caller,
        bytes32 recipientNttManagerAddress,
        bytes32 refundAddress,
        TransceiverStructs.TransceiverInstruction memory instruction,
        bytes memory nttManagerMessage
    ) internal override {
        (
            TransceiverStructs.TransceiverMessage memory transceiverMessage,
            bytes memory encodedTransceiverPayload
        ) = TransceiverStructs.buildAndEncodeTransceiverMessage(
                WH_TRANSCEIVER_PAYLOAD_PREFIX,
                toWormholeFormat(caller),
                recipientNttManagerAddress,
                nttManagerMessage,
                new bytes(0)
            );
        IVsl(vslAddress).generateStateQueryClaim(
            recipientChain,
            address(this),
            fromWormholeFormat(getPeer(recipientChain)),
            encodedTransceiverPayload
        );
    }

    function _getPeersStorage()
        internal
        pure
        returns (mapping(uint16 => bytes32) storage $)
    {
        uint256 slot = uint256(PEERS_SLOT);
        assembly ("memory-safe") {
            $.slot := slot
        }
    }
}
