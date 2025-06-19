// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.13;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";
import {Vsl} from "../src/Vsl/Vsl.sol";
import {PeerToken} from "../src/NttToken/PeerToken.sol";
import {NttManager} from "ntt/NttManager/NttManager.sol";
import {IManagerBase} from "ntt/interfaces/IManagerBase.sol";
import {VslTransceiver} from "../src/Vsl/VslTransceiver.sol";
import {EVMViewFnClaimVerifier} from "../src/EVMViewFnVerifier/EVMViewFnClaimVerifier.sol";
import {ERC1967Proxy} from "openzeppelin-contracts/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {PodRegistry} from "../src/EVMViewFnVerifier/libs/PodRegistry.sol";
import {PodECDSA} from "../src/EVMViewFnVerifier/libs/PodECDSA.sol";

contract Deploy is Script {
    bool isDest;
    uint16 chainId;
    uint256 privateKey;
    address owner;
    address podRegistryAddress;

    function setUp() public {
        isDest = vm.envBool("IS_DEST");

        if (isDest) {
            privateKey = vm.envUint("DEST_PRIVATE_KEY");
            chainId = uint16(vm.envUint("DEST_CHAIN_ID"));
        } else {
            privateKey = vm.envUint("SRC_PRIVATE_KEY");
            chainId = uint16(vm.envUint("SRC_CHAIN_ID"));
        }
        owner = vm.addr(privateKey);

        podRegistryAddress = vm.envOr("POD_REGISTRY_ADDRESS", address(0));
    }

    function run() public {
        vm.startBroadcast(privateKey);

        // TODO: Remove Pod Registry
        PodRegistry podRegistry;
        // Pod registry will be deployed by the pod team. But for the private networks we just deploy it from another key.
        if (podRegistryAddress == address(0)) {
            podRegistry = deployPodRegistry();
        } else {
            podRegistry = PodRegistry(podRegistryAddress);
        }

        address tokenAddress = deployPeerToken();
        console.log("Token address:", tokenAddress);

        address vslAddress = deployVsl(podRegistry);
        console.log("VSL address:", vslAddress);

        address managerAddress = deployNttManager(tokenAddress);
        console.log("NttManager address:", managerAddress);

        address transceiverAddress = deployTransceiver(
            managerAddress,
            vslAddress
        );
        console.log("Transceiver address:", transceiverAddress);

        vm.stopBroadcast();
    }

    function deployPeerToken() internal returns (address) {
        PeerToken peerToken = new PeerToken(
            "PeerToken",
            "PT",
            owner,
            msg.sender
        );
        return address(peerToken);
    }

    function deployVsl(PodRegistry podRegistry) internal returns (address) {
        PodECDSA.PodConfig memory podConfig = PodECDSA.PodConfig({
            quorum: 2,
            registry: podRegistry
        });

        Vsl vsl = new Vsl(chainId, podConfig);
        return address(vsl);
    }

    function deployPodRegistry() internal returns (PodRegistry) {
        address[] memory committee = new address[](4);
        committee[0] = address(0x5A29ADC28eF780461e909AeD0E9eC31CB79cfA32);
        committee[1] = address(0x1CF6A9667F236CA9c3bEA0a5cbD9925CAbDE3309);
        committee[2] = address(0xFEDE83A4dFefd84A71a171Fb1D8a83E88549CBb2);
        committee[3] = address(0xE39B8DC304AE7521575Fb8F497aC37d560329239);

        PodRegistry podRegistry = new PodRegistry(committee);

        return podRegistry;
    }

    function deployNttManager(address tokenAddress) internal returns (address) {
        // Implementation
        NttManager implementation = new NttManager(
            tokenAddress,
            IManagerBase.Mode.BURNING,
            chainId,
            86400,
            false
        );

        // Proxy
        NttManager proxy = NttManager(
            address(new ERC1967Proxy(address(implementation), ""))
        );
        proxy.initialize();
        return address(proxy);
    }

    function deployTransceiver(
        address managerAddress,
        address vslAddress
    ) internal returns (address) {
        VslTransceiver implementation = new VslTransceiver(
            managerAddress,
            vslAddress
        );
        VslTransceiver proxy = VslTransceiver(
            address(new ERC1967Proxy(address(implementation), ""))
        );
        proxy.initialize();
        return address(proxy);
    }
}
