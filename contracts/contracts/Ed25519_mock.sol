pragma solidity ^0.8.9;

contract Ed25519 {
    function scalarMultBase(uint256 s) public view returns (uint256, uint256) {
        return (s, s);
    }
}
