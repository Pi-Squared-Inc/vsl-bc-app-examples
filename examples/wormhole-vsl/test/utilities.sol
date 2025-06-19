pragma solidity ^0.8.20;

contract Util{
    function chop4(bytes calldata sig) public pure returns (bytes memory){
        return sig[4:];
    }

    function chop64(bytes calldata sig) public pure returns (bytes memory){
        return sig[64:];
    }
}