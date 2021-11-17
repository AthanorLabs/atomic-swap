// SPDX-License-Identifier: GPLv3

pragma solidity ^0.8.5;

import "./SwapOnChain.sol";
import "./TestUtils.sol";

contract SwapOnChainMock is SwapOnChain, TestUtils {
    constructor() SwapOnChain(0, 0) {}

    //
    // Do not apply view modifier
    // Gas can't be measured for method calls that don't return a receipt (i.e. pure/view)
    //
    function testVerifySecret(uint256 _s, bytes32 pubKey) external {
        // (uint256 px, uint256 py) = ed25519.derivePubKey(_s);
        (uint256 px, uint256 py) = ed25519.scalarMultBase(_s);
        uint256 canonical_p = py | ((px % 2) << 255);
        // console.log("py: %s", uint2hexstr(py));
        // console.log("px: %s", uint2hexstr(px));
        // console.log("derived:  %s", uint2hexstr(canonical_p));
        // TODO WTF - don't comment out the next line
        // All tests fail if nothing is printed
        console.log("provided: %s", uint2hexstr(uint256(pubKey)));

        require(bytes32(canonical_p) == pubKey, "wrong secret");
    }
}
