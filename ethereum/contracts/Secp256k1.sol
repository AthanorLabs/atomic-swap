// SPDX-License-Identifier: GPLv3
// modified from https://github.com/1Address/ecsol/blob/master/contracts/EC.sol

pragma solidity ^0.8.5;

contract Secp256k1 {
    uint256 constant gx =
        0x79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798;
    uint256 constant gy =
        0x483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8;
    uint256 constant n =
        0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F;
    uint256 constant a = 0;
    uint256 constant b = 7;

    uint256 constant m =
        0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141;

    //
    // Based on the original idea of Vitalik Buterin:
    // https://ethresear.ch/t/you-can-kinda-abuse-ecrecover-to-do-ecmul-in-secp256k1-today/2384/9
    //
    // Verifies that `Q = [s] G` on the secp256k1 curve
    // qKeccak is defined as uint256(keccak256(abi.encodePacked(qx, qy))
    //
    function mulVerify(uint256 s, uint256 qKeccak)
        external
        pure
        returns (bool)
    {
        address signer = ecrecover(
            0,
            gy % 2 != 0 ? 28 : 27,
            bytes32(gx),
            bytes32(mulmod(s, gx, m))
        );
        return address(uint160(qKeccak)) == signer;
    }
}
