// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.13;

import {Base} from "./Base.sol";
import {console} from "forge-std/console.sol";
import {Vsl} from "../src/Vsl/Vsl.sol";
import {PeerToken} from "../src/NttToken/PeerToken.sol";
import {NttManager} from "ntt/NttManager/NttManager.sol";
import {IManagerBase} from "ntt/interfaces/IManagerBase.sol";
import {VslTransceiver} from "../src/Vsl/VslTransceiver.sol";
import {EVMViewFnClaimVerifier} from "../src/EVMViewFnVerifier/EVMViewFnClaimVerifier.sol";

contract Setup is Base {
    function run() public {
        vm.startBroadcast(privateKey);

        _setNttPeer(destChainId, sourceManagerAddress, destManagerAddress);
        _setTransceiverPeer(
            destChainId,
            sourceTransceiverAddress,
            destTransceiverAddress
        );
        _setManagerTransceiver(sourceManagerAddress, sourceTransceiverAddress);
        _setOutboundLimit(sourceManagerAddress);
        _setThreshold(sourceManagerAddress);
        _setInboundLimit(sourceManagerAddress, inboundLimit, destChainId);
        PeerToken(sourceTokenAddress).setMinter(sourceManagerAddress);

        vm.stopBroadcast();
    }

    function _setNttPeer(
        uint16 chainId,
        address managerAddress,
        address peerAddress
    ) internal {
        NttManager(managerAddress).setPeer(
            chainId,
            bytes32(uint256(uint160(peerAddress))),
            18,
            0
        );
    }

    function _setTransceiverPeer(
        uint16 chainId,
        address transceiverAddress,
        address peerAddress
    ) internal {
        VslTransceiver(transceiverAddress).setPeer(
            chainId,
            bytes32(uint256(uint160(peerAddress)))
        );
    }

    function _setManagerTransceiver(
        address managerAddress,
        address transceiverAddress
    ) internal {
        IManagerBase(managerAddress).setTransceiver(transceiverAddress);
    }

    function _setOutboundLimit(address managerAddress) internal {
        NttManager(managerAddress).setOutboundLimit(
            184467440737095516150000000000
        );
    }

    function _setThreshold(address managerAddress) internal {
        IManagerBase(managerAddress).setThreshold(1);
    }

    function _setInboundLimit(
        address managerAddress,
        uint256 inboundLimit,
        uint16 chainId
    ) internal {
        NttManager(managerAddress).setInboundLimit(inboundLimit, chainId);
    }
}
