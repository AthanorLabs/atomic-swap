// SPDX-License-Identifier: LGPLv3
pragma solidity ^0.8.19;

import {IERC20} from "./IERC20.sol";
import {Secp256k1} from "./Secp256k1.sol";

contract SwapCreator is Secp256k1 {
    // Swap state is PENDING when the swap is first created and funded
    // Alice sets Stage to READY when she sees the funds locked on the other chain.
    // this prevents Bob from withdrawing funds without locking funds on the other chain first
    // Stage is set to COMPLETED upon the swap value being claimed or refunded.
    enum Stage {
        INVALID,
        PENDING,
        READY,
        COMPLETED
    }

    struct Swap {
        // the swap initiator, Alice
        // address allowed to refund the ether for this swap
        address payable owner;
        // address allowed to claim the ether for this swap, Bob
        address payable claimer;
        // the keccak256 hash of the expected public key derived from the secret `s_b`.
        // this public key is a point on the secp256k1 curve
        bytes32 pubKeyClaim;
        // the keccak256 hash of the expected public key derived from the secret `s_a`.
        // this public key is a point on the secp256k1 curve
        bytes32 pubKeyRefund;
        // timestamp before which Alice can call either `setReady` or `refund`
        uint256 timeout0;
        // timestamp after which Bob cannot claim, only Alice can refund
        uint256 timeout1;
        // the asset being swapped: equal to address(0) for ETH, or an ERC-20 token address
        address asset;
        // the value of this swap
        uint256 value;
        // choose random
        uint256 nonce;
    }

    // RelaySwap contains additional information required for relayed transactions.
    // This entire structure is encoded and signed by the swap claimer, and the signature is
    // passed to the `claimRelayer` function.
    struct RelaySwap {
        // the swap the transaction is for
        Swap swap;
        // the fee, in wei, paid to the relayer
        uint256 fee;
        // hash of (relayer's payout address || 4-byte salt)
        bytes32 relayerHash;
        // address of the swap contract this transaction is meant for
        address swapCreator;
    }

    mapping(bytes32 => Stage) public swaps;

    event New(
        bytes32 swapID,
        bytes32 claimKey,
        bytes32 refundKey,
        uint256 timeout0,
        uint256 timeout1,
        address asset,
        uint256 value
    );
    event Ready(bytes32 indexed swapID);
    event Claimed(bytes32 indexed swapID, bytes32 indexed s);
    event Refunded(bytes32 indexed swapID, bytes32 indexed s);

    // returned when trying to initiate a swap with a zero value
    error ZeroValue();

    // returned when the pubKeyClaim or pubKeyRefund parameters for `newSwap` are zero
    error InvalidSwapKey();

    // returned when the claimer parameter for `newSwap` is the zero address
    error InvalidClaimer();

    // returned when the timeout0 or timeout1 parameters for `newSwap` are zero
    error InvalidTimeout();

    // returned when the ether sent with a `newSwap` transaction does not match the value parameter
    error InvalidValue();

    // returned when trying to initiate a swap with an ID that already exists
    error SwapAlreadyExists();

    // returned when trying to call `setReady` on a swap that is not in the PENDING stage
    error SwapNotPending();

    // returned when the caller of `setReady` or `refund` is not the swap owner
    error OnlySwapOwner();

    // returned when the signer of the relayed transaction is not the swap's claimer
    error OnlySwapClaimer();

    // returned when trying to call `claim` or `refund` on an invalid swap
    error InvalidSwap();

    // returned when trying to call `claim` or `refund` on a swap that's already completed
    error SwapCompleted();

    // returned when trying to call `claim` on a swap that's not set to ready or the first timeout has not been reached
    error TooEarlyToClaim();

    // returned when trying to call `claim` on a swap where the second timeout has been reached
    error TooLateToClaim();

    // returned when it's the counterparty's turn to claim and refunding is not allowed
    error NotTimeToRefund();

    // returned when the provided secret does not match the expected public key
    error InvalidSecret();

    // returned when the signature of a `RelaySwap` is invalid
    error InvalidSignature();

    // returned when the SwapCreator address is a `RelaySwap` is not the addres of this contract
    error InvalidContractAddress();

    // returned when the hash of the relayer address and salt passed to `claimRelayer`
    // does not match the relayer hash in `RelaySwap`
    error InvalidRelayerAddress();

    // newSwap creates a new Swap instance with the given parameters.
    // it returns the swap's ID.
    // _timeoutDuration0: duration between the current timestamp and timeout0
    // _timeoutDuration1: duration between timeout0 and timeout1
    function newSwap(
        bytes32 _pubKeyClaim,
        bytes32 _pubKeyRefund,
        address payable _claimer,
        uint256 _timeoutDuration0,
        uint256 _timeoutDuration1,
        address _asset,
        uint256 _value,
        uint256 _nonce
    ) public payable returns (bytes32) {
        if (_value == 0) revert ZeroValue();
        if (_asset == address(0)) {
            if (_value != msg.value) revert InvalidValue();
        } else {
            // transfer ERC-20 token into this contract
            // WARN: fee-on-transfer tokens are not supported
            IERC20(_asset).transferFrom(msg.sender, address(this), _value);
        }

        if (_pubKeyClaim == 0 || _pubKeyRefund == 0) revert InvalidSwapKey();
        if (_claimer == address(0)) revert InvalidClaimer();
        if (_timeoutDuration0 == 0 || _timeoutDuration1 == 0) revert InvalidTimeout();

        Swap memory swap = Swap({
            owner: payable(msg.sender),
            pubKeyClaim: _pubKeyClaim,
            pubKeyRefund: _pubKeyRefund,
            claimer: _claimer,
            timeout0: block.timestamp + _timeoutDuration0,
            timeout1: block.timestamp + _timeoutDuration0 + _timeoutDuration1,
            asset: _asset,
            value: _value,
            nonce: _nonce
        });

        bytes32 swapID = keccak256(abi.encode(swap));

        // make sure this isn't overriding an existing swap
        if (swaps[swapID] != Stage.INVALID) revert SwapAlreadyExists();

        emit New(
            swapID,
            _pubKeyClaim,
            _pubKeyRefund,
            swap.timeout0,
            swap.timeout1,
            swap.asset,
            swap.value
        );
        swaps[swapID] = Stage.PENDING;
        return swapID;
    }

    // Alice should call setReady() before timeout0 once she verifies the XMR has been locked
    function setReady(Swap memory _swap) public {
        bytes32 swapID = keccak256(abi.encode(_swap));
        if (swaps[swapID] != Stage.PENDING) revert SwapNotPending();
        if (_swap.owner != msg.sender) revert OnlySwapOwner();
        swaps[swapID] = Stage.READY;
        emit Ready(swapID);
    }

    // Bob can claim if:
    // - (Alice has set the swap to `ready` or it's past timeout0) and it's before timeout1
    function claim(Swap memory _swap, bytes32 _secret) public {
        if (msg.sender != _swap.claimer) revert OnlySwapClaimer();
        _claim(_swap, _secret);

        // send ether to swap claimer
        if (_swap.asset == address(0)) {
            _swap.claimer.transfer(_swap.value);
        } else {
            // WARN: this will FAIL for fee-on-transfer or rebasing tokens if the token
            // transfer reverts (i.e. if this contract does not contain _swap.value tokens),
            // exposing Bob's secret while giving him nothing.
            IERC20(_swap.asset).transfer(_swap.claimer, _swap.value);
        }
    }

    // Bob can claim if:
    // - (Alice has set the swap to `ready` or it's past timeout0) and it's before timeout1
    // It transfers the fee to the relayer address specified in `_relaySwap`.
    // Note: this function will revert if the swap value is less than the relayer fee;
    // in that case, `claim` must be called instead.
    function claimRelayer(
        RelaySwap memory _relaySwap,
        bytes32 _secret,
        address payable _relayer,
        uint32 _salt,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public {
        address signer = ecrecover(keccak256(abi.encode(_relaySwap)), v, r, s);
        if (signer != _relaySwap.swap.claimer) revert InvalidSignature();
        if (address(this) != _relaySwap.swapCreator) revert InvalidContractAddress();
        if (keccak256(abi.encodePacked(_relayer, _salt)) != _relaySwap.relayerHash)
            revert InvalidRelayerAddress();

        _claim(_relaySwap.swap, _secret);

        // send ether to swap claimer, subtracting the relayer fee
        if (_relaySwap.swap.asset == address(0)) {
            _relaySwap.swap.claimer.transfer(_relaySwap.swap.value - _relaySwap.fee);
            payable(_relayer).transfer(_relaySwap.fee);
        } else {
            // WARN: this will FAIL for fee-on-transfer or rebasing tokens if the token
            // transfer reverts (i.e. if this contract does not contain _swap.value tokens),
            // exposing Bob's secret while giving him nothing.
            IERC20(_relaySwap.swap.asset).transfer(
                _relaySwap.swap.claimer,
                _relaySwap.swap.value - _relaySwap.fee
            );
            IERC20(_relaySwap.swap.asset).transfer(_relayer, _relaySwap.fee);
        }
    }

    function _claim(Swap memory _swap, bytes32 _secret) internal {
        bytes32 swapID = keccak256(abi.encode(_swap));
        Stage swapStage = swaps[swapID];
        if (swapStage == Stage.INVALID) revert InvalidSwap();
        if (swapStage == Stage.COMPLETED) revert SwapCompleted();
        if (block.timestamp < _swap.timeout0 && swapStage != Stage.READY) revert TooEarlyToClaim();
        if (block.timestamp >= _swap.timeout1) revert TooLateToClaim();

        verifySecret(_secret, _swap.pubKeyClaim);
        emit Claimed(swapID, _secret);
        swaps[swapID] = Stage.COMPLETED;
    }

    // Alice can claim a refund:
    // - Until timeout0 unless she calls setReady
    // - After timeout1
    function refund(Swap memory _swap, bytes32 _secret) public {
        bytes32 swapID = keccak256(abi.encode(_swap));
        Stage swapStage = swaps[swapID];
        if (swapStage == Stage.INVALID) revert InvalidSwap();
        if (swapStage == Stage.COMPLETED) revert SwapCompleted();
        if (_swap.owner != msg.sender) revert OnlySwapOwner();
        if (
            block.timestamp < _swap.timeout1 &&
            (block.timestamp > _swap.timeout0 || swapStage == Stage.READY)
        ) revert NotTimeToRefund();

        verifySecret(_secret, _swap.pubKeyRefund);
        emit Refunded(swapID, _secret);

        // send asset back to swap owner
        swaps[swapID] = Stage.COMPLETED;
        if (_swap.asset == address(0)) {
            _swap.owner.transfer(_swap.value);
        } else {
            IERC20(_swap.asset).transfer(_swap.owner, _swap.value);
        }
    }

    function verifySecret(bytes32 _secret, bytes32 _hashedPubkey) internal pure {
        if (!mulVerify(uint256(_secret), uint256(_hashedPubkey))) revert InvalidSecret();
    }
}
