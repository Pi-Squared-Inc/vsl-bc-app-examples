pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/Vsl/Vsl.sol";
import "./utilities.sol";
import "../src/EVMViewFnVerifier/libs/PodRegistry.sol";
import "../src/EVMViewFnVerifier/libs/PodECDSA.sol";

contract ContractUSLTest is Test {
    Vsl sourceVsl;
    Vsl destVsl;
    Util utils;

    struct RelayMessage {
        uint16 srcChainid;
        uint16 destChainid;
        address senderTransceiver;
        address receiverTransceiver;
        bytes transceiverMessage;
    }

    event genStateQueryClaim(
        uint16 srcChainid,
        uint16 destChainid,
        uint blockNumber,
        address contractAddress,
        bytes viewFunctionEncoding
    );

    function setUp() public {
        // TODO: Remove Pod Registry
        address[] memory committee = new address[](4);
        committee[0] = address(0xD64C0A2A1BAe8390F4B79076ceaE7b377B5761a3);
        committee[1] = address(0x8646d958225301A00A6CB7b6609Fa23bab87DA7C);
        committee[2] = address(0x7D5761b7b49fC7BFdD499E3AE908a4aCFe0807E6);
        committee[3] = address(0x06aD294f74dc98bE290E03797e745CF0D9c03dA2);

        PodRegistry podRegistry = new PodRegistry(committee);
        PodECDSA.PodConfig memory podConfig = PodECDSA.PodConfig({
            quorum: 2,
            registry: podRegistry
        });

        sourceVsl = new Vsl(1, podConfig);
        destVsl = new Vsl(2, podConfig);
        utils = new Util();
    }

    function test_GenerateClaim_Nonce() public {
        uint16 cid = sourceVsl.chainid();
        assertEq(sourceVsl.messageNonce(cid), 0);

        sourceVsl.generateStateQueryClaim(
            cid,
            address(this),
            address(0x0),
            abi.encode(0x0)
        );
        assertEq(sourceVsl.messageNonce(cid), 1);
    }

    function test_GenerateClaim_Event() public {
        uint16 sourceChainid = sourceVsl.chainid();
        uint16 destChainid = destVsl.chainid();

        //vm package is an example of a cheatcodes: https://book.getfoundry.sh/forge/cheatcodes
        vm.expectEmit(true, true, false, false);
        emit genStateQueryClaim(
            sourceChainid,
            destChainid,
            1,
            address(sourceVsl),
            abi.encode(0x0)
        );
        sourceVsl.generateStateQueryClaim(
            destChainid,
            address(sourceVsl),
            address(0x0),
            abi.encode(0x0)
        );
    }

    function test_GenerateClaim_Claim() public {
        string memory test = vm.readFile("./test/claim2.json");

        uint16 srcChainid = sourceVsl.chainid();
        uint16 destChainid = destVsl.chainid();
        uint32 blknumber = uint32(vm.parseJsonUint(test, ".event.blockNumber"));

        address sender = vm.parseJsonAddress(
            test,
            ".event.senderUslTransciever"
        );
        address receiver = vm.parseJsonAddress(
            test,
            ".event.receiverUslTransciever"
        );
        bytes memory nttmsg = vm.parseJsonBytes(
            test,
            ".event.transceiverMessage"
        );

        vm.chainId(srcChainid);
        vm.roll(blknumber);
        sourceVsl.generateStateQueryClaim(
            destChainid,
            sender,
            receiver,
            nttmsg
        );

        assertEq(
            sourceVsl.relays(destChainid, 0),
            abi.decode(vm.parseJsonBytes(test, ".raw.output"), (bytes)),
            "encoded relay message does not match"
        );
    }

    function test_parseStateQueryClaim_1() public view {
        string memory test = vm.readFile("./test/claim1.json");
        bytes memory claim = abi.decode(vm.parseJson(test, ".hex"), (bytes));

        (
            uint256 chainid,
            uint256 blknumber,
            address from,
            bytes memory viewFunctionEncoding,
            bytes memory relayMessage
        ) = destVsl.parseStateQueryClaim(claim);

        assertEq(
            chainid,
            vm.parseJsonUint(test, ".raw.chainId"),
            "Chain ID does not match"
        );
        assertEq(
            blknumber,
            vm.parseJsonUint(test, ".raw.blockNumber"),
            "Blknumber does not match"
        );
        assertEq(
            from,
            vm.parseJsonAddress(test, ".raw.to"),
            "Contract address does not match"
        );
        assertEq(
            viewFunctionEncoding,
            vm.parseJsonBytes(test, ".raw.input"),
            "The decoded function encoding does not macth"
        );
        assertEq(
            relayMessage,
            vm.parseJsonBytes(test, ".raw.output"),
            "Message to be relayed does not match"
        );
    }

    function test_parseStateQueryClaim_2() public view {
        string memory test = vm.readFile("./test/claim2.json");
        bytes memory claim = vm.parseJsonBytes(test, ".hex");

        (
            uint256 chainid,
            uint256 blknumber,
            address from,
            bytes memory viewFunctionEncoding,
            bytes memory relayMessage
        ) = destVsl.parseStateQueryClaim(claim);

        assertEq(
            chainid,
            vm.parseJsonUint(test, ".raw.chainId"),
            "Chain ID does not match"
        );
        assertEq(
            blknumber,
            vm.parseJsonUint(test, ".raw.blockNumber"),
            "Blknumber does not match"
        );
        assertEq(
            from,
            vm.parseJsonAddress(test, ".raw.to"),
            "Contract address does not match"
        );
        assertEq(
            viewFunctionEncoding,
            vm.parseJsonBytes(test, ".raw.input"),
            "The decoded function encoding does not macth"
        );
        assertEq(
            relayMessage,
            vm.parseJsonBytes(test, ".raw.output"),
            "Message to be relayed does not match"
        );
    }
}
