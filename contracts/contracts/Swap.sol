// SPDX-License-Identifier: MIT

pragma solidity 0.8.9;

import "./Ed25519_mock.sol";

contract Swap {
    Ed25519 ed25519;
    // the hash of the expected public key for which the secret `s_b` corresponds.
    // this public key is a point on the ed25519 curve, and is in 33-byte compressed format (?)
    bytes32 public pubKeyClaim;

    // the hash of the expected public key for which the secret `s_a` corresponds.
    // this public key is a point on the ed25519 curve, and is in 33-byte compressed format (?)
    bytes32 public pubKeyRefund;

    // time period since contract creation,
    // when a refund is allowed for Alice before she calls `ready`
    // (i.e. before Bob locks monero)
    uint256 timeout_0;

    // time period since calling `ready`,
    // when a refund is allowed for Alice if Bob doesn't claim
    uint256 timeout_1;

    // contract creator, Alice
    address payable owner;

    // ready is set to true when Alice sees the funds locked on the other chain.
    // this prevents Bob from withdrawing funds without locking funds on the other chain first
    bool isReady = false;

    event DerivedPubKeyClaim(uint256 s);
    event DerivedPubKeyRefund(uint256 s);

    constructor(bytes32 _pubKeyClaim, bytes32 _pubKeyRefund) payable {
        owner = payable(msg.sender);
        pubKeyClaim = _pubKeyClaim;
        pubKeyRefund = _pubKeyRefund;
        timeout_0 = block.timestamp + 1 days; // TODO: make configurable
        ed25519 = new Ed25519();
    }

    function set_ready() public {
        require(msg.sender == owner);
        isReady = true;
        timeout_1 = block.timestamp + 1 days; // TODO: make configurable
    }

    function claim(uint256 _s) public {
        // Bob can claim either when:
        // Alice never calls ready, and timeout_0 has passed, or
        // Alice called ready and Bob is within timeframe timeout_1
        if (isReady == true) {
            require(block.timestamp < timeout_1, "Too late to claim!");
        } else {
            require(
                block.timestamp >= timeout_0,
                "'isReady == false' cannot claim yet!"
            );
        }
        // confirm that provided secret `_s` was used to derive pubKeyClaim
        (uint256 px, uint256 py) = ed25519.scalarMultBase(_s);
        bytes32 ph = keccak256(abi.encode(px, py));
        require(
            ph == pubKeyClaim,
            "provided secret does not match the expected pubKey"
        );
        emit DerivedPubKeyClaim(_s);

        // // send eth to caller
        payable(msg.sender).transfer(address(this).balance);
    }

    function refund(uint256 _s) public {
        require(
            (block.timestamp <= timeout_0 && isReady == false) ||
                block.timestamp <= timeout_1
        );

        (uint256 px, uint256 py) = ed25519.scalarMultBase(_s);
        bytes32 ph = keccak256(abi.encode(px, py));
        require(
            ph == pubKeyRefund,
            "provided secret does not match the expected pubKey"
        );
        emit DerivedPubKeyRefund(_s);

        // send eth back to owner
        owner.transfer(address(this).balance);
    }
}
