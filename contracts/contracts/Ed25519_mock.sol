pragma solidity ^0.8.9;

contract Ed25519 {
    function scalarMultBase(uint s) public view returns (uint, uint) {
        return (s, s);
    }

}
