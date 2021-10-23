// SPDX-License-Identifier: MIT

pragma solidity 0.8.9;

import "./Ed25519.sol";

contract Swap {
    // Ed25519 library
    Ed25519 ed25519;

    // contract creator, Alice
    address payable owner;

    // the expected public key for which the secret `s_b` corresponds.
    // this public key is a point on the ed25519 curve, and is in 33-byte compressed format (?)
    bytes32 public x_bob;
    bytes32 public y_bob;

    // the expected public key for which the secret `s_a` corresponds.
    // this public key is a point on the ed25519 curve, and is in 33-byte compressed format (?)
    bytes32 public x_alice;
    bytes32 public y_alice;

    // time period from contract creation
    // during which Alice can call either set_ready or refund
    uint256 public timeout_0;

    // time period from the moment Alice calls `set_ready`
    // during which Bob can claim. After this, Alice can refund again
    uint256 public timeout_1;

    // ready is set to true when Alice sees the funds locked on the other chain.
    // this prevents Bob from withdrawing funds without locking funds on the other chain first
    bool isReady = false;

    event Constructed(bytes32 x, bytes32 y);
    event IsReady(bool b);
    event Claimed(uint256 s);
    event Refunded(uint256 s);

    constructor(
        bytes32 _x_alice,
        bytes32 _y_alice,
        bytes32 _x_bob,
        bytes32 _y_bob
    ) payable {
        owner = payable(msg.sender);
        x_alice = _x_alice;
        y_alice = _y_alice;
        x_bob = _x_bob;
        y_bob = _y_bob;
        timeout_0 = block.timestamp + 1 days;
        ed25519 = new Ed25519();
        emit Constructed(x_alice, y_alice);
    }

    // Alice must call set_ready() within t_0 once she verifies the XMR has been locked
    function set_ready() public {
        require(msg.sender == owner && block.timestamp < timeout_0);
        isReady = true;
        timeout_1 = block.timestamp + 1 days;
        emit IsReady(true);
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

        verifySecret(_s, x_bob, y_bob);
        emit Claimed(_s);

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
        emit Refunded(_s);

        // send eth back to owner==caller (Alice)
        selfdestruct(owner);
    }

    function verifySecret(
        uint256 _s,
        bytes32 x,
        bytes32 y
    ) internal view {
        (uint256 px, uint256 py) = ed25519.derivePubKey(_s);
        require(
            bytes32(px) == x && bytes32(py) == y,
            "provided secret does not match the expected pubKey"
        );
    }
}
