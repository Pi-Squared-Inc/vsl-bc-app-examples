pragma solidity ^0.8.0;

library Hex {
    /// @notice Optimized hex string to bytes conversion using assembly
    /// @param hexString The hex string to convert (must have 0x prefix)
    /// @return result The decoded bytes
    function hexStringToBytesAssembly(
        string memory hexString
    ) internal pure returns (bytes memory result) {
        bytes memory hexBytes = bytes(hexString);
        require(hexBytes.length >= 2, "String too short");
        require(
            hexBytes[0] == "0" && (hexBytes[1] == "x" || hexBytes[1] == "X"),
            "Missing 0x prefix"
        );

        uint256 dataLength = (hexBytes.length - 2) / 2;
        require((hexBytes.length - 2) % 2 == 0, "Invalid hex string length");

        result = new bytes(dataLength);

        assembly {
            let hexData := add(hexBytes, 0x22) // Skip length prefix + "0x"
            let resultData := add(result, 0x20) // Skip length prefix

            for {
                let i := 0
            } lt(i, dataLength) {
                i := add(i, 1)
            } {
                let byteOffset := mul(i, 2)
                let high := byte(0, mload(add(hexData, byteOffset)))
                let low := byte(0, mload(add(hexData, add(byteOffset, 1))))

                // Convert ASCII to hex value
                high := sub(high, mul(lt(high, 58), 48))
                high := sub(high, mul(and(lt(high, 71), gt(high, 64)), 55))
                high := sub(high, mul(and(lt(high, 103), gt(high, 96)), 87))

                low := sub(low, mul(lt(low, 58), 48))
                low := sub(low, mul(and(lt(low, 71), gt(low, 64)), 55))
                low := sub(low, mul(and(lt(low, 103), gt(low, 96)), 87))

                mstore8(add(resultData, i), or(shl(4, high), low))
            }
        }
    }
}
