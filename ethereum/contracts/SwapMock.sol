// SPDX-License-Identifier: GPLv3

pragma solidity ^0.8.5;

import "./Swap.sol";
import "./TestUtils.sol";

contract SwapMock is Swap, TestUtils {
    uint256 constant gx =
        0x79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798;
    uint256 constant gy =
        0x483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8;
    uint256 constant n =
        0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F;
    uint256 constant a = 0;
    uint256 constant b = 7;

    uint256 constant m =
        0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141;

    address payable mockAddress = payable(0);

    constructor() Swap(0, 0, mockAddress, 60) {}

    function testVerifySecret(uint256 s, bytes32 ctment) view external {
        address signer = ecrecover(
            0,
            gy % 2 != 0 ? 28 : 27,
            bytes32(gx),
            bytes32(mulmod(s, gx, m))
        );

        // console.log("ctment: %s", uint2hexstr(uint256(ctment)));
        // console.log("addr: %s", address(uint160(uint256(ctment))));
        // console.log("s: %s", s);
        console.log("derived: %s", signer);

        require(address(uint160(uint256(ctment))) == signer);
    }
}
