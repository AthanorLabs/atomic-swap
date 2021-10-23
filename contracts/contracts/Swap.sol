// SPDX-License-Identifier: MIT

pragma solidity 0.8.9;

import "./Ed25519.sol";

contract Swap {
    // Ed25519 library
    Ed25519 ed25519;

    // contract creator, Alice
    address payable owner;

    // the hash of the expected public key for which the secret `s_b` corresponds.
    // this public key is a point on the ed25519 curve, and is in 33-byte compressed format (?)
    bytes32 public pubKeyClaim;

    // the hash of the expected public key for which the secret `s_a` corresponds.
    // this public key is a point on the ed25519 curve, and is in 33-byte compressed format (?)
    bytes32 public pubKeyRefund;

    // time period from contract creation
    // during which Alice can call either set_ready or refund
    uint256 public timeout_0;

    // time period from the moment Alice calls `set_ready`
    // during which Bob can claim. After this, Alice can refund again
    uint256 public timeout_1;

    // ready is set to true when Alice sees the funds locked on the other chain.
    // this prevents Bob from withdrawing funds without locking funds on the other chain first
    bool isReady = false;

    event DerivedPubKeyClaim(uint256 s);
    event DerivedPubKeyRefund(uint256 s);

    constructor(
        bytes32 _pubKeyClaim,
        bytes32 _pubKeyRefund,
        Ed25519 _ed25519
    ) payable {
        owner = payable(msg.sender);
        pubKeyClaim = _pubKeyClaim;
        pubKeyRefund = _pubKeyRefund;
        timeout_0 = block.timestamp + 1 days;
        ed25519 = _ed25519;
    }

    // Alice must call set_ready() within t_0 once she verifies the XMR has been locked
    function set_ready() public {
        require(msg.sender == owner && block.timestamp < timeout_0);
        isReady = true;
        timeout_1 = block.timestamp + 1 days;
    }

    // Bob can claim if:
    // - Alice doesn't call set_ready or refund within t_0, or
    // - Alice calls ready within t_0, in which case Bob can call claim until t_1
    function claim(uint256 _s) external {
        if (isReady == true) {
            require(block.timestamp < timeout_1, "Too late to claim!");
        } else {
            require(
                block.timestamp >= timeout_0,
                "'isReady == false' cannot claim yet!"
            );
        }

        verifySecret(_s, pubKeyClaim);
        emit DerivedPubKeyClaim(_s);

        // send eth to caller (Bob)
        selfdestruct(payable(msg.sender));
    }

    // Alice can claim a refund:
    // - Until t_0 unless she calls set_ready
    // - After t_1, if she called set_ready
    function refund(uint256 _s) external {
        require(
            (!isReady && block.timestamp < timeout_0) ||
                (isReady && block.timestamp >= timeout_1)
        );

        verifySecret(_s, pubKeyRefund);
        emit DerivedPubKeyRefund(_s);

        // send eth back to owner==caller (Alice)
        selfdestruct(owner);
    }

    function verifySecret(uint256 _s, bytes32 pubKey) internal view {
        (uint256 px, uint256 py) = ed25519.derivePubKey(_s);
        bytes32 ph = keccak256(abi.encode(px, py));
        require(
            ph == pubKey,
            "provided secret does not match the expected pubKey"
        );
    }
}
