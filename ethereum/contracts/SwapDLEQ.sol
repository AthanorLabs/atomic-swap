// SPDX-License-Identifier: GPLv3

pragma solidity ^0.8.5;

import "./Swap.sol";
import "./Secp256k1.sol";

contract SwapDLEQ is Swap {
    Secp256k1 immutable secp256k1;

    // TODO modify in go
    constructor(bytes32 claimCtment, bytes32 refundCtment)
        payable
        Swap(claimCtment, refundCtment)
    {
        secp256k1 = new Secp256k1();
    }

    function verifySecret(uint256 _s, bytes32 ctment) internal view override {
        require(secp256k1.mulVerify(_s, uint256(ctment)), "wrong secret");
    }
}
