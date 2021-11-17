// SPDX-License-Identifier: GPLv3

pragma solidity ^0.8.5;

abstract contract Swap {
    // contract creator, Alice
    address payable immutable owner;

    // will allow Bob to claim the funds if he reveals the secret
    bytes32 public immutable claimCtment;

    // will allow Alice to claim a refund if she reveals the secret
    bytes32 public immutable refundCtment;

    // time period from contract creation
    // during which Alice can call either set_ready or refund
    uint256 public immutable timeout_0;

    // time period from the moment Alice calls `set_ready`
    // during which Bob can claim. After this, Alice can refund again
    uint256 public timeout_1;

    // Alice sets ready to true when she sees the funds locked on the other chain.
    // This prevents Bob from withdrawing funds without locking funds on the other chain first
    bool public isReady = false;

    event Constructed(bytes32 p);
    event IsReady();
    event Claimed(uint256 s);
    event Refunded(uint256 s);

    constructor(bytes32 _claimCtment, bytes32 _refundCtment) payable {
        owner = payable(msg.sender);
        claimCtment = _claimCtment;
        refundCtment = _refundCtment;
        timeout_0 = block.timestamp + 1 days;

        // TODO if this is not the public key it doesn't make sense to emit it
        // Bob needs to verify both anyway
        // Either the event contains both (though we won't be able to listen for it,
        // so I don't think there's any use), or we
        // - Emit nothing
        // - Notify bob of deployment
        // - Bob reads both commitments through getters and
        // - Verifies they match the information exchanged with Alice
        emit Constructed(_refundCtment);
    }

    // Alice must call set_ready() within t_0 once she verifies the XMR has been locked
    function set_ready() external {
        require(!isReady && msg.sender == owner && block.timestamp < timeout_0);
        isReady = true;
        timeout_1 = block.timestamp + 1 days;
        emit IsReady();
    }

    // Bob can claim if:
    // - Alice doesn't call set_ready or refund within t_0, or
    // - Alice calls ready within t_0, in which case Bob can call claim until t_1
    function claim(uint256 _s) external {
        if (isReady) {
            require(
                block.timestamp < timeout_1,
                "Too late to claim! Pray that Alice claims a refund."
            );
        } else {
            require(
                block.timestamp >= timeout_0,
                "Please wait until Alice has called set_ready or the first timeout is reached."
            );
        }

        verifySecret(_s, claimCtment);
        emit Claimed(_s);

        // send eth to caller (Bob)
        selfdestruct(payable(msg.sender));
    }

    // Alice can claim a refund:
    // - Until t_0 unless she calls set_ready
    // - After t_1, if she called set_ready
    function refund(uint256 _s) external {
        if (isReady) {
            require(
                block.timestamp >= timeout_1,
                "Bob can now claim the funds until the second timeout, please wait!"
            );
        } else {
            require(
                block.timestamp < timeout_0,
                "Too late for a refund! Pray that Bob claims his ETH."
            );
        }

        verifySecret(_s, refundCtment);
        emit Refunded(_s);

        // send eth back to owner==caller (Alice)
        selfdestruct(owner);
    }

    function verifySecret(uint256 _s, bytes32 ctment) internal view virtual;
}
