pragma solidity 0.6.12;

import "./elliptic-curve-solidity/examples/Secp256k1.sol";

contract Swap {
	Secp256k1 secp256k1;

	// the hash where the pre-image must be disclosed to redeem the eth
	// in this contract.
	// it is the keccak256 hash of Bob's secret (s_b), which when disclosed
	// allows Alice, the contract creator to redeem funds on another chain.
	bytes32 public hashRedeem;

	// the hash of the expected public key for which the secret s_b corresponds.
	// this public key lies on the secp256k1 curve, and is in 33-byte compressed format.
	bytes32 public expectedPublicKey;

	// the hash where the pre-image must be disclosed by the contract owner, Alice, after time
	// `t` to refund the eth in this contract.
	// it is the keccak256 hash of Alice's secret, which when disclosed 
	// allows Bob to refund their coins on the other chain.
	bytes32 public hashRefund;

	// time after which a refund is allowed
	uint256 timeout;

	// contract creator, Alice
	address payable owner;

	// ready is set to true when Alice sees the funds locked on the other chain.
	// this prevents Bob from withdrawing funds without locking funds on the other chain first
	bool isReady = false;

	event DerivedPubKeyRedeem(uint256 x, uint256 y);

	constructor(
		bytes32 _hashRedeem,
		bytes32 _expectedPublicKey, 
		bytes32 _hashRefund
	) public payable {
		owner = msg.sender;
		hashRedeem = _hashRedeem;
		expectedPublicKey = _expectedPublicKey;
		hashRefund = _hashRefund;
		timeout = now + 1 days; // TODO: make configurable
		secp256k1 = new Secp256k1();
	}

	function ready() public {
		require(msg.sender == owner);
		isReady = true; 
	}

	function redeem(
		uint256 _s
	) public {
		require(isReady == true, "contract is not ready!");

		// // confirm that provided secret `_s` is pre-image of `hashRedeem`
		// bytes32 h0 = keccak256(abi.encode(_s));
		// require(h0 == hashRedeem, "pre-image for redeem was incorrect");

		// confirm that secret corresponds to provided public key
		(uint256 px, uint256 py) = secp256k1.derivePubKey(_s);
		emit DerivedPubKeyRedeem(px, py);
		bytes32 ph = keccak256(abi.encode(px, py));
		require(ph == expectedPublicKey, "provided public key does not match expected");

		// // send eth to caller
		// msg.sender.transfer(address(this).balance);
	}

	function refund(
		uint256 _s
	) public {
		// confirm that provided secret is pre-image of `hashRefund`
		bytes32 h = keccak256(abi.encode(_s));
		require(h == hashRefund);

		// send eth back to owner
		owner.transfer(address(this).balance);
	}
}