# Protocol

## Current version

See [this issue describing the update](https://github.com/noot/atomic-swap/issues/36).

```
gas used now to deploy Swap.sol: 640005
gas used previously to deploy Swap.sol: 1855645
improvement: ~2.9x

gas used now for the Claim() or Refund() call: 14729
gas used previously for the Claim() or Refund() call: 938818
improvement: ~64x
```

## Initial version

Alice has ETH and wants XMR, Bob has XMR and wants ETH. They come to an agreement to do the swap and the amounts they will swap.

#### Initial (offchain) phase
- Alice and Bob each generate Monero secret keys (which consist of secret spend and view keys): (`s_a`, `v_a`) and (`s_b`, `v_b`), which are used to construct valid points on the ed25519 curve (ie. public keys): `P_a` and `P_b` accordingly. Alice sends Bob her public key and Bob sends Alice his public spend key and private view key. Note: The XMR will be locked in the account with address corresponding to the public key `P_a + P_b`. Bob needs to send his private view key so Alice can check that Bob actually locked the amount of XMR he claims he will.

#### Step 1.
Alice deploys a smart contract on Ethereum and locks her ETH in it. The contract has the following properties:
- it is non-destructible

- it contains two timestamps, `t_0` and `t_1`, before and after which different actions are authorized.

- it is constructed containing `P_a` and`P_b`, so that if Alice or Bob reveals their secret by calling the contract, the contract will verify that the secret corresponds to the expected public key that it was initalized with.

- it has a `Ready()` function which can only be called by Alice. Once `Ready()` is invoked, Bob can proceed with redeeming his ether. Alice has until the `t_0` timestamp to call `Ready()` - once `t_0` passes, then the contract automatically allows Bob to claim his ether, up until some second timestamp `t_1`.

- it has a `Claim()` function which can only be called by Bob after `Ready()` is called or `t_0` passes, up until the timestamp `t_1`. After `t_1`, Bob can no longer claim the ETH.

- `Claim()` takes one parameter from Bob: `s_b`. Once `Claim()` is called, the ETH is transferred to Bob, and simultaneously Bob reveals his secret and thus Alice can claim her XMR by combining her and Bob's secrets.

- it has a `Refund()` function that can only be called by Alice and only before `Ready()` is called *or* `t_0` is reached. Once `Ready()` is invoked, Alice can no longer call `Refund()` until the next timestamp `t_1`.  If Bob doesn't claim his ether by `t_1`, then `Refund()` can be called by Alice once again.

- `Refund()` takes one parameter from Alice: `s_a`. This allows Alice to get her ETH back in case Bob goes offline, but it simulteneously reveals her secret, allowing Bob to regain access to the XMR he locked.

#### Step 2. 
Bob sees the smart contract has been deployed with the correct parameters. He sends his XMR to an account address constructed from `P_a + P_b`. Thus, the funds can only be accessed by an entity having both `s_a` and `s_b`, as the secret spend key to that account is `s_a + s_b`. The funds are viewable by someone having `v_a + v_b`.

Note: `Refund()` and `Claim()` cannot be called at the same time. This is to prevent the case of front-running where, for example, Bob tries to claim, so his secret `s_b` is in the mempool, and then Alice tries to call `Refund()` with a higher priority while also transferring the XMR in the account controlled by `s_a + s_b`. If her call goes through before Bob's and Bob doesn't notice this happening in time, then Alice will now have *both* the ETH and the XMR. Due to this case, Alice and Bob should not call `Refund()` or `Claim()` when they are approaching `t_0` or `t_1` respectively, as their transaction may not go through in time.

#### Step 3.
Alice sees that the XMR has been locked, and the amount is correct (as she knows `v_a` and Bob send her `v_b` in the first key exchange step). She calls `Ready()` on the smart contract if the XMR has been locked. If the amount of XMR locked is incorrect, Alice calls `Refund()` to abort the swap and reclaim her ETH.

From this point on, Bob can redeem his ether by calling `Claim(s_b)`, which transfers the ETH to him.

By redeeming, Bob reveals his secret. Now Alice is the only one that has both `s_a` and `s_b` and she can access the monero in the account created from `P_a + P_b`.

#### What could go wrong

- **Alice locked her ETH, but Bob doesn't lock his XMR**. Alice has until time `t_0` to call `Refund()` to reclaim her ETH, which she should do if `t_0` is soon.

- **Alice called `Ready()`, but Bob never redeems.** Deadlocks are prevented thanks to a second timelock `t_1`, which re-enables Alice to call refund after it, while disabling Bob's ability to claim.

- **Alice never calls `ready` within `t_0`**. Bob can still claim his ETH by waiting until after `t_0` has passed, as the contract automatically allows him to call `Claim()`.