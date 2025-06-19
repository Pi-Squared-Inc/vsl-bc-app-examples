// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.13;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";

contract Base is Script {
    bool isDest;
    uint256 transferAmount;
    uint256 inboundLimit;
    uint16 chainId;
    uint16 destChainId;
    uint256 privateKey;
    address owner;
    address sourceVslAddress;
    address sourceManagerAddress;
    address sourceTransceiverAddress;
    address sourceTokenAddress;
    address destManagerAddress;
    address destTransceiverAddress;
    address destTokenAddress;

    function setUp() public {
        isDest = vm.envBool("IS_DEST");
        transferAmount = vm.envUint("TRANSFER_AMOUNT") * 10 ** 18;
        inboundLimit = vm.envUint("INBOUND_LIMIT") * 10 ** 18;

        if (isDest) {
            privateKey = vm.envUint("DEST_PRIVATE_KEY");
            chainId = uint16(vm.envUint("DEST_CHAIN_ID"));
            sourceVslAddress = vm.envAddress("DEST_VSL_ADDRESS");
            sourceManagerAddress = vm.envAddress("DEST_MANAGER_ADDRESS");
            sourceTransceiverAddress = vm.envAddress(
                "DEST_TRANSCEIVER_ADDRESS"
            );
            sourceTokenAddress = vm.envAddress("DEST_TOKEN_ADDRESS");
            destChainId = uint16(vm.envUint("SRC_CHAIN_ID"));
            destManagerAddress = vm.envAddress("SRC_MANAGER_ADDRESS");
            destTransceiverAddress = vm.envAddress("SRC_TRANSCEIVER_ADDRESS");
            destTokenAddress = vm.envAddress("SRC_TOKEN_ADDRESS");
        } else {
            privateKey = vm.envUint("SRC_PRIVATE_KEY");
            chainId = uint16(vm.envUint("SRC_CHAIN_ID"));
            sourceVslAddress = vm.envAddress("SRC_VSL_ADDRESS");
            sourceManagerAddress = vm.envAddress("SRC_MANAGER_ADDRESS");
            sourceTransceiverAddress = vm.envAddress("SRC_TRANSCEIVER_ADDRESS");
            sourceTokenAddress = vm.envAddress("SRC_TOKEN_ADDRESS");
            destChainId = uint16(vm.envUint("DEST_CHAIN_ID"));
            destManagerAddress = vm.envAddress("DEST_MANAGER_ADDRESS");
            destTransceiverAddress = vm.envAddress("DEST_TRANSCEIVER_ADDRESS");
            destTokenAddress = vm.envAddress("DEST_TOKEN_ADDRESS");
        }
        owner = vm.addr(privateKey);
    }
}
