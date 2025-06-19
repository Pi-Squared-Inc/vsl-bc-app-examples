// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.13;

import {console} from "forge-std/console.sol";
import {Vsl} from "../src/Vsl/Vsl.sol";
import {PeerToken} from "../src/NttToken/PeerToken.sol";
import {NttManager} from "ntt/NttManager/NttManager.sol";
import {Base} from "./Base.sol";

contract Operations is Base {
    function mintToken() public {
        vm.startBroadcast(privateKey);

        PeerToken(sourceTokenAddress).setMinter(owner);
        PeerToken(sourceTokenAddress).mint(owner, transferAmount);
        PeerToken(sourceTokenAddress).setMinter(sourceManagerAddress);

        vm.stopBroadcast();
    }

    function checkBalance() public {
        vm.startBroadcast(privateKey);

        uint256 balance = PeerToken(sourceTokenAddress).balanceOf(owner);
        console.log("Balance:", balance);

        vm.stopBroadcast();
    }

    function transfer() public {
        vm.startBroadcast(privateKey);

        PeerToken(sourceTokenAddress).approve(
            sourceManagerAddress,
            transferAmount
        );
        NttManager(sourceManagerAddress).transfer(
            transferAmount,
            destChainId,
            bytes32(uint256(uint160(owner)))
        );

        vm.stopBroadcast();
    }
}
