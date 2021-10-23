pragma solidity 0.8.9;


contract Swap {

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

	event DerivedPubKeyClaim(uint256 x, uint256 y);
	event DerivedPubKeyRefund(uint256 x, uint256 y);

	constructor(
		bytes32 _pubKeyClaim,
		bytes32 _pubKeyRefund
	) payable {
      owner = payable(msg.sender);
		pubKeyClaim = _pubKeyClaim;
		pubKeyRefund = _pubKeyRefund;
		timeout_0 = block.timestamp + 1 days; // TODO: make configurable
	}

	function set_ready() public {
		require(msg.sender == owner);
		isReady = true;
    timeout_1 = block.timestamp + 1 days; // TODO: make configurable
	}

	function claim(
		uint256 _s
	) public {
		require(isReady == true, "contract is not ready!");
		// confirm that provided secret `_s` was used to derive pubKeyClaim
    /* Ed25519(_s) */

		/* emit DerivedPubKeyClaim(px, py); */
		/* bytes32 ph = keccak256(abi.encode(px, py)); */
		/* require(ph == pubKeyClaim, "provided public key does not match expected"); */

		// // send eth to caller
		payable(msg.sender).transfer(address(this).balance);
	}

  function refund_bob(
    uint256 _s
  ) public {
      // TODO required that secret provided corresponds to expectedPubKeyClaim

      /* emit DerivedPubKeyClaim(px, py); */

      require(block.timestamp < timeout_0);
  }

	function refund_alice(
		uint256 _s
	) public {
      require((block.timestamp <= timeout_0 && isReady == false) || block.timestamp <= timeout_1);

      // TODO require that the provided secret matches `pubKeyRefund`

      /* emit DerivedPubKeyRefund(px, py); */

      // send eth back to owner
      owner.transfer(address(this).balance);
	}
}
