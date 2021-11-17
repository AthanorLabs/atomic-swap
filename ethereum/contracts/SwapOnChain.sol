// SPDX-License-Identifier: GPLv3

pragma solidity ^0.8.5;

// import "./Ed25519.sol";
import "./Swap.sol";
import "./Ed25519_alt.sol";

contract SwapOnChain is Swap {
    // Ed25519 library
    Ed25519 immutable ed25519;

    // pubKeyClaim and pubKeyRefund are the Monero public keys
    // derived from the secrets `s_b` and `s_a`, respectively.
    // This are points on the ed25519 curve.
    constructor(bytes32 pubKeyClaim, bytes32 pubKeyRefund)
        payable
        Swap(pubKeyClaim, pubKeyRefund)
    {
        ed25519 = new Ed25519();
    }

    function verifySecret(uint256 _s, bytes32 ctment) internal view override {
        // (uint256 px, uint256 py) = ed25519.derivePubKey(_s);
        (uint256 px, uint256 py) = ed25519.scalarMultBase(_s);
        uint256 canonical_p = py | ((px % 2) << 255);

        require(bytes32(canonical_p) == ctment, "wrong secret");
    }
}
