// SPDX-License-Identifier: MIT

pragma solidity ^0.8.5;

// import "./Ed25519.sol";
import "./Ed25519_alt.sol";

contract Swap {
    // Ed25519 library
    Ed25519 immutable ed25519;

    // contract creator, Alice
    address payable immutable owner;

    // the expected public key derived from the secret `s_b`.
    // this public key is a point on the ed25519 curve
    bytes32 public immutable pubKeyClaim;

    // the expected public key derived from the secret `s_a`.
    // this public key is a point on the ed25519 curve
    bytes32 public immutable pubKeyRefund;

    // timestamp (set at contract creation)
    // before which Alice can call either set_ready or refund
    uint256 public immutable timeout_0;

    // timestamp after which Bob cannot claim, only Alice can refund.
    uint256 public immutable timeout_1;

    // Alice sets ready to true when she sees the funds locked on the other chain.
    // this prevents Bob from withdrawing funds without locking funds on the other chain first
    bool isReady = false;

    event Constructed(bytes32 p);
    event IsReady(bool b);
    event Claimed(uint256 s);
    event Refunded(uint256 s);

    constructor(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund, uint256 _timeoutDuration) payable {
        owner = payable(msg.sender);
        pubKeyClaim = _pubKeyClaim;
        pubKeyRefund = _pubKeyRefund;
        timeout_0 = block.timestamp + _timeoutDuration;
        timeout_1 = block.timestamp + (_timeoutDuration * 2);
        ed25519 = new Ed25519();
        emit Constructed(_pubKeyRefund);
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
    function claim(uint256 _s) external {
        require(block.timestamp < timeout_1 && (block.timestamp >= timeout_0 || isReady), 
            "too late or early to claim!");

        verifySecret(_s, pubKeyClaim);
        emit Claimed(_s);

        // send eth to caller (Bob)
        selfdestruct(payable(msg.sender));
    }

    // Alice can claim a refund:
    // - Until t_0 unless she calls set_ready
    // - After t_1, if she called set_ready
    function refund(uint256 _s) external {
        require(
            block.timestamp >= timeout_1 || ( block.timestamp < timeout_0 && !isReady),
            "It's Bob's turn now, please wait!"
        );

        verifySecret(_s, pubKeyRefund);
        emit Refunded(_s);

        // send eth back to owner==caller (Alice)
        selfdestruct(owner);
    }

    function verifySecret(uint256 _s, bytes32 pubKey) internal view {
        // (uint256 px, uint256 py) = ed25519.derivePubKey(_s);
        (uint256 px, uint256 py) = ed25519.scalarMultBase(_s);
        uint256 canonical_p = py | ((px % 2) << 255);
        require(
            bytes32(canonical_p) == pubKey,
            "provided secret does not match the expected pubKey"
        );
    }
}
