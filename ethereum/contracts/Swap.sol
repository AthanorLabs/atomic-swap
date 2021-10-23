// SPDX-License-Identifier: MIT

pragma solidity 0.8.9;

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

    // time period from contract creation
    // during which Alice can call either set_ready or refund
    uint256 public immutable timeout_0;

    // time period from the moment Alice calls `set_ready`
    // during which Bob can claim. After this, Alice can refund again
    uint256 public timeout_1;

    // Alice sets ready to true when she sees the funds locked on the other chain.
    // this prevents Bob from withdrawing funds without locking funds on the other chain first
    bool isReady = false;

    event Constructed(bytes32 p);
    event IsReady(bool b);
    event Claimed(uint256 s);
    event Refunded(uint256 s);

    constructor(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund) payable {
        owner = payable(msg.sender);
        pubKeyClaim = _pubKeyClaim;
        pubKeyRefund = _pubKeyRefund;
        timeout_0 = block.timestamp + 1 days;
        ed25519 = new Ed25519();
        emit Constructed(_pubKeyRefund);
    }

    // Alice must call set_ready() within t_0 once she verifies the XMR has been locked
    function set_ready() external {
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

        verifySecret(_s, pubKeyClaim);
        emit Claimed(_s);

        // send eth to caller (Bob)
        selfdestruct(payable(msg.sender));
    }

    // Alice can claim a refund:
    // - Until t_0 unless she calls set_ready
    // - After t_1, if she called set_ready
    function refund(uint256 _s) external {
        if (isReady == true) {
            require(
                block.timestamp >= timeout_1,
                "It's Bob's turn now, please wait!"
            );
        } else {
            require(block.timestamp < timeout_0, "Missed your chance!");
        }

        verifySecret(_s, pubKeyRefund);
        emit Refunded(_s);

        // send eth back to owner==caller (Alice)
        selfdestruct(owner);
    }

    function verifySecret(uint256 _s, bytes32 pubKey) public view {
        // (uint256 px, uint256 py) = ed25519.derivePubKey(_s);
        (uint256 px, uint256 py) = ed25519.scalarMultBase(_s);
        uint256 canonical_p = py | ((px >> (8 * 30 + 1)) & 0x80);
        require(
            bytes32(canonical_p) == pubKey,
            "provided secret does not match the expected pubKey"
        );
    }
}
