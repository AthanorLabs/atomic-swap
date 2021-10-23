// SPDX-License-Identifier: MIT

pragma solidity 0.8.9;

import "./EllipticCurve.sol";

/**
 * @title Ed25519 Elliptic Curve
 * @notice Particularization of Elliptic Curve for ed25519 curve
 */
contract Ed25519 {
    uint256 public constant GX =
        15112221349535400772501151409588531511454012693041857206046113283949847762202;
    uint256 public constant GY =
        46316835694926478169428394003475163141307993866256225615783033603165251855960;
    uint256 public constant AA =
        37095705934669439343138083508754565189542113879843219016388785533085940283555;
    uint256 public constant PP =
        0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFED;

    /// @notice Public Key derivation from private key
    /// @param privKey The private key
    /// @return (qx, qy) The Public Key
    function derivePubKey(uint256 privKey)
        external
        pure
        returns (uint256, uint256)
    {
        return EllipticCurve.ecMul(privKey, GX, GY, AA, PP);
    }
}
