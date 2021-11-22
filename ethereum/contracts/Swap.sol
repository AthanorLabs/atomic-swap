// SPDX-License-Identifier: MIT

pragma solidity ^0.8.5;

contract Swap {
    // contract creator, Alice
    address payable immutable owner;

    // address allowed to claim the ether in this contract
    address payable immutable claimer;

    // the expected hash of the secret `s_b`.
    bytes32 public immutable claimHash;

    // the expected hash of the secret `s_a`.
    bytes32 public immutable refundHash;

    // timestamp (set at contract creation)
    // before which Alice can call either set_ready or refund
    uint256 public immutable timeout_0;

    // timestamp after which Bob cannot claim, only Alice can refund.
    uint256 public immutable timeout_1;

    // Alice sets ready to true when she sees the funds locked on the other chain.
    // this prevents Bob from withdrawing funds without locking funds on the other chain first
    bool isReady = false;

    event Constructed(bytes32 claimHash, bytes32 refundHash);
    event IsReady(bool b);
    event Claimed(bytes32 s);
    event Refunded(bytes32 s);

    constructor(bytes32 _claimHash, bytes32 _refundHash, address payable _claimer, uint256 _timeoutDuration) payable {
        owner = payable(msg.sender);
        claimHash = _claimHash;
        refundHash = _refundHash;
        claimer = _claimer;
        timeout_0 = block.timestamp + _timeoutDuration;
        timeout_1 = block.timestamp + (_timeoutDuration * 2);
        emit Constructed(claimHash, refundHash);
    }

    // Alice must call set_ready() within t_0 once she verifies the XMR has been locked
    function set_ready() external {
        require(!isReady && msg.sender == owner);
        isReady = true;
        emit IsReady(true);
    }

    // Bob can claim if:
    // - Alice doesn't call set_ready or refund within t_0, or
    // - Alice calls ready within t_0, in which case Bob can call claim until t_1
    function claim(bytes32 _s) external {
        require(msg.sender == claimer, "only claimer can claim!");
        require(block.timestamp < timeout_1 && (block.timestamp >= timeout_0 || isReady), 
            "too late or early to claim!");
        require(keccak256(abi.encode(_s)) == claimHash, "secret is not preimage to claimHash");
        emit Claimed(_s);

        // send eth to caller (Bob)
        selfdestruct(payable(msg.sender));
    }

    // Alice can claim a refund:
    // - Until t_0 unless she calls set_ready
    // - After t_1, if she called set_ready
    function refund(bytes32 _s) external {
        require(msg.sender == owner);
        require(
            block.timestamp >= timeout_1 || ( block.timestamp < timeout_0 && !isReady),
            "It's Bob's turn now, please wait!"
        );
        require(keccak256(abi.encode(_s)) == refundHash, "secret is not preimage to refundHash");
        emit Refunded(_s);

        // send eth back to owner==caller (Alice)
        selfdestruct(owner);
    }
}
