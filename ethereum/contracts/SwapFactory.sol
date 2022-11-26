// SPDX-License-Identifier: LGPLv3
pragma solidity ^0.8.5 .0;

import "./ERC2771Context.sol";
import "./IERC20.sol";
import "./Secp256k1.sol";

contract SwapFactory is ERC2771Context, Secp256k1 {
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
        // individual swap creator, Alice
        address payable owner;
        // address allowed to claim the ether in this contract
        address payable claimer;
        // the keccak256 hash of the expected public key derived from the secret `s_b`.
        // this public key is a point on the secp256k1 curve
        bytes32 pubKeyClaim;
        // the keccak256 hash of the expected public key derived from the secret `s_a`.
        // this public key is a point on the secp256k1 curve
        bytes32 pubKeyRefund;
        // timestamp (set at contract creation)
        // before which Alice can call either set_ready or refund
        uint256 timeout0;
        // timestamp after which Bob cannot claim, only Alice can refund.
        uint256 timeout1;
        // the asset being swapped: equal to address(0) for ETH, or an ERC-20 token address
        address asset;
        // the value of this swap.
        uint256 value;
        // choose random
        uint256 nonce;
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
    event Ready(bytes32 swapID);
    event Claimed(bytes32 swapID, bytes32 s);
    event Refunded(bytes32 swapID, bytes32 s);

    constructor(address trustedForwarder) ERC2771Context(trustedForwarder) {} // solhint-disable-line

    // newSwap creates a new Swap instance with the given parameters.
    // it returns the swap's ID.
    function newSwap(
        bytes32 _pubKeyClaim,
        bytes32 _pubKeyRefund,
        address payable _claimer,
        uint256 _timeoutDuration,
        address _asset,
        uint256 _value,
        uint256 _nonce
    ) public payable returns (bytes32) {
        Swap memory swap;
        swap.owner = payable(msg.sender);
        swap.pubKeyClaim = _pubKeyClaim;
        swap.pubKeyRefund = _pubKeyRefund;
        swap.claimer = _claimer;
        swap.timeout0 = block.timestamp + _timeoutDuration;
        swap.timeout1 = block.timestamp + (_timeoutDuration * 2);
        swap.asset = _asset;
        swap.value = _value;
        if (swap.asset == address(0)) {
            require(swap.value == msg.value, "value not same as ETH amount sent");
        } else {
            // transfer ERC-20 token into this contract
            // TODO: potentially check token balance before/after this step
            // and ensure the balance was increased by swap.value since fee-on-transfer
            // tokens are not supported
            IERC20(swap.asset).transferFrom(msg.sender, address(this), swap.value);
        }
        swap.nonce = _nonce;

        bytes32 swapID = keccak256(abi.encode(swap));

        // make sure this isn't overriding an existing swap
        require(swaps[swapID] == Stage.INVALID, "swap already exists");

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

    // Alice should call setReady() within t_0 once she verifies the XMR has been locked
    function setReady(Swap memory _swap) public {
        bytes32 swapID = keccak256(abi.encode(_swap));
        require(swaps[swapID] == Stage.PENDING, "swap is not in PENDING state");
        require(_swap.owner == msg.sender, "only the swap owner can call setReady");
        swaps[swapID] = Stage.READY;
        emit Ready(swapID);
    }

    // Bob can claim if:
    // - Alice has set the swap to `ready` or it's past t_0 but before t_1
    function claim(Swap memory _swap, bytes32 _s) public {
        _claim(_swap, _s);

        // send eth to caller (Bob)
        if (_swap.asset == address(0)) {
            _swap.claimer.transfer(_swap.value);
        } else {
            // TODO: this will FAIL for fee-on-transfer or rebasing tokens if the token
            // transfer reverts (i.e. if this contract does not contain _swap.value tokens),
            // exposing Bob's secret while giving him nothing

            // potential solution: wrap tokens into shares instead of absolute values
            // swap.value would then contain the share of the token
            IERC20(_swap.asset).transfer(_swap.claimer, _swap.value);
        }
    }

    // Bob can claim if:
    // - Alice has set the swap to `ready` or it's past t_0 but before t_1
    function claimRelayer(Swap memory _swap, bytes32 _s, uint256 fee) public {
        require(
            isTrustedForwarder(msg.sender),
            "claimRelayer can only be called by a trusted forwarder"
        );
        _claim(_swap, _s);

        // send ether to swap claimant, subtracting the relayer fee
        // which is sent to the originator of the transaction.
        // tx.origin is okay here, since it isn't for authentication purposes.
        if (_swap.asset == address(0)) {
            _swap.claimer.transfer(_swap.value - fee);
            payable(tx.origin).transfer(fee); // solhint-disable-line
        } else {
            // TODO: this will FAIL for fee-on-transfer or rebasing tokens if the token
            // transfer reverts (i.e. if this contract does not contain _swap.value tokens),
            // exposing Bob's secret while giving him nothing

            // potential solution: wrap tokens into shares instead of absolute values
            // swap.value would then contain the share of the token
            IERC20(_swap.asset).transfer(_swap.claimer, _swap.value - fee);
            IERC20(_swap.asset).transfer(tx.origin, fee); // solhint-disable-line
        }
    }

    function _claim(Swap memory _swap, bytes32 _s) internal {
        bytes32 swapID = keccak256(abi.encode(_swap));
        Stage swapStage = swaps[swapID];
        require(swapStage != Stage.INVALID, "invalid swap");
        require(swapStage != Stage.COMPLETED, "swap is already completed");
        require(_msgSender() == _swap.claimer, "only claimer can claim!");
        require(
            (block.timestamp >= _swap.timeout0 || swapStage == Stage.READY),
            "too early to claim!"
        );
        require(block.timestamp < _swap.timeout1, "too late to claim!");

        verifySecret(_s, _swap.pubKeyClaim);
        emit Claimed(swapID, _s);
        swaps[swapID] = Stage.COMPLETED;
    }

    // Alice can claim a refund:
    // - Until t_0 unless she calls set_ready
    // - After t_1
    function refund(Swap memory _swap, bytes32 _s) public {
        bytes32 swapID = keccak256(abi.encode(_swap));
        Stage swapStage = swaps[swapID];
        require(
            swapStage != Stage.COMPLETED && swapStage != Stage.INVALID,
            "swap is already completed"
        );
        require(msg.sender == _swap.owner, "refund must be called by the swap owner");
        require(
            block.timestamp >= _swap.timeout1 ||
                (block.timestamp < _swap.timeout0 && swapStage != Stage.READY),
            "it's the counterparty's turn, unable to refund, try again later"
        );

        verifySecret(_s, _swap.pubKeyRefund);
        emit Refunded(swapID, _s);

        // send asset back to owner==caller (Alice)
        swaps[swapID] = Stage.COMPLETED;
        if (_swap.asset == address(0)) {
            _swap.owner.transfer(_swap.value);
        } else {
            IERC20(_swap.asset).transfer(_swap.owner, _swap.value);
        }
    }

    function verifySecret(bytes32 _s, bytes32 pubKey) internal pure {
        require(
            mulVerify(uint256(_s), uint256(pubKey)),
            "provided secret does not match the expected public key"
        );
    }
}
