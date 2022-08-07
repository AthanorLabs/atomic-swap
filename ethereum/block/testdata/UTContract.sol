// SPDX-License-Identifier: LGPLv3
pragma solidity ^0.8.5;

//
// Unit test contract. Our goal is a transaction that is easy to make fail, but only when it is
// mined, not when the transaction is sent. To do this, pass check_stamp a value that is equal
// to the timestamp of the last mined block. Gas estimation will pass, as it uses the timestamp
// from the last mined block, so we'll successfully send the transaction to the nework, but the
// transaction is guaranteed to fail in whatever block it is mined into, as the block's timestamp
// will be greater than what we passed in.
//
contract UTContract {
    uint256 private stamp;

    function check_stamp(uint256 _stamp) external {
        require(block.timestamp <= _stamp, "block.timestamp was not less than stamp");
        stamp = _stamp; // Prevent the function from being view-only
    }
}
