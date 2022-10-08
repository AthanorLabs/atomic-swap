# JSON-RPC API

The `swapd` program automatically starts a JSON-RPC server that can be used to interact with the swap network and make/take swap offers.

## `net` namespace

### `net_addresses`

Get the libp2p listening addresses of the node.

Parameters:
- none

Returns:
- `addresses`: list of libp2p multiaddresses the swap daemon is currently listening on.

Example:

```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_addresses","params":{}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"addresses":["/ip4/192.168.0.101/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2","/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2","/ip4/38.88.101.233/tcp/14815/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2"]},"id":"0"}
```

### `net_discover`

Discover peers on the network via DHT that have active swap offers.

Parameters:
- `provides` (optional): one of `ETH` or `XMR`, depending on which offer you are searching for. **Note**: Currently only `XMR` offers are supported. Default is `XMR`.
- `searchTime` (optional): duration in seconds for which to perform the search. Default is 12s.

Returns:
- `peers`: list of lists of peers's multiaddresses. A peer may have multiple multiaddresses, so the nested list pertains to a single peer.

Example:

```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_discover","params":{"searchTime":3}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"peers":[["/ip4/127.0.0.1/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7","/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7"]]},"id":"0"}
```

### `net_queryAll`

Discover peers on the network via DHT that have active swap offers and gets all their swap offers.

Parameters:
- `provides` (optional): one of `ETH` or `XMR`, depending on which offer you are searching for. **Note**: Currently only `XMR` offers are supported. Default is `XMR`.
- `searchTime` (optional): duration in seconds for which to perform the search. Default is 12s.

Returns:
- `peersWithOffers`: list of peers's multiaddresses and their current offers.

Example:

```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_queryAll","params":{"searchTime":3}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"PeersWithOffers":[{"peer":["/ip4/206.189.47.220/tcp/9900/p2p/12D3KooWGVzz2d2LSceVFFdqTYqmQXTqc5eWziw7PLRahCWGJhKB"],"offers":[{"ID":"a41b00034daee28df414ba337b3ddf942893a117f9a9fcf62bd5a664738710db","Provides":"XMR","MinimumAmount":0.1,"MaximumAmount":1,"ExchangeRate":0.5}]},{"peer":["/ip4/161.35.110.210/tcp/9900/p2p/12D3KooWS8iKxqsGTiL3Yc1VaAfg99U5km1AE7bWYQiuavXj3Yz6"],"offers":[{"ID":"25188edd7573f43fca5760f0aacdc1a358171a8fc6bdf11876fa937f77fc583c","Provides":"XMR","MinimumAmount":0.1,"MaximumAmount":1,"ExchangeRate":0.5}]}]},"id":"0"}
```

### `net_queryPeer`

Query a specific peer for their current active offers.

Parameters:
- `multiaddr`: multiaddress of the peer to query. Found via `net_discover`.

Returns:
- `offers`: list of the peer's current active offers.

Example:

```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_queryPeer","params":{"multiaddr":"/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7"}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"offers":[{"ID":[207,75,240,26,7,117,160,209,63,164,27,20,81,110,75,137,3,67,0,112,122,23,84,224,217,155,101,246,203,111,255,185],"Provides":"XMR","MinimumAmount":0.1,"MaximumAmount":1,"ExchangeRate":0.05}]},"id":"0"}
```

### `net_makeOffer`

Make a new swap offer and advertise it on the network. **Note:** Currently only XMR offers can be made.

Parameters:
- `minimumAmount`: minimum amount to swap, in XMR.
- `maximumAmount`: maximum amount to swap, in XMR.
- `exchangeRate`: exchange rate of ETH-XMR for the swap, expressed in a fraction of XMR/ETH. For example, if you wish to trade 10 XMR for 1 ETH, the exchange rate would be 0.1.
- `ethAsset`: (optional) Ethereum asset to trade, either an ERC-20 token address or the zero address for regular ETH. default: regular ETH

Returns:
- `offerID`: ID of the swap offer.

Example:
```bash
curl -X POST http://127.0.0.1:5002 -d '{"jsonrpc":"2.0","id":"0","method":"net_makeOffer","params":{"minimumAmount":1, "maximumAmount":10, "exchangeRate": 0.1}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"offerID":"12b9d56a4c568c772a4e099aaed03a457256d6680562be2a518753f75d75b7ad"},"id":"0"}
```


### `net_takeOffer`

Take an advertised swap offer. This call will initiate and execute an atomic swap. **Note:** You must be the ETH holder to take a swap.

Parameters:
- `multiaddr`: multiaddress of the peer to swap with.
- `offerID`: ID of the swap offer.
- `providesAmount`: amount of ETH you will be providing. Must be between the offer's `minimumAmount * exchangeRate` and `maximumAmount * exchangeRate`. For example, if the offer has a minimum of 1 XMR and a maximum of 5 XMR and an exchange rate of 0.1, you must provide between 0.1 ETH and 0.5 ETH.

Returns:
- null

Example:
```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_takeOffer","params":{"multiaddr":"/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7", "offerID":"12b9d56a4c568c772a4e099aaed03a457256d6680562be2a518753f75d75b7ad", "providesAmount": 0.3}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":null,"id":"0"}
```

### `net_takeOfferSync`

Take an advertised swap offer. This call will initiate and execute an atomic swap. It will not return until the swap has completed, after which it will return whether the swap was successful or not. **Note:** You must be the ETH holder to take a swap.

Parameters:
- `multiaddr`: multiaddress of the peer to swap with.
- `offerID`: ID of the swap offer.
- `providesAmount`: amount of ETH you will be providing. Must be between the offer's `minimumAmount * exchangeRate` and `maximumAmount * exchangeRate`. For example, if the offer has a minimum of 1 XMR and a maximum of 5 XMR and an exchange rate of 0.1, you must provide between 0.1 ETH and 0.5 ETH.

Returns:
- `status`: the swap's status, one of `success`, `refunded`, or `aborted`.

Example:
```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_takeOffer","params":{"multiaddr":"/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7", "offerID":"12b9d56a4c568c772a4e099aaed03a457256d6680562be2a518753f75d75b7ad", "providesAmount": 0.3}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{status":"success"},"id":"0"}
```


## `personal` namespace

### `personal_balances`

Returns combined information of both the Monero and Ethereum account addresses  and balances.

Parameters:
- none

Returns:
- `monero_address`: primary monero address of the swapd wallet
- `piconero_balance`: balance the swapd wallet in piconero
- `piconero_unlocked_balance`: balance the swapd wallet in piconero that is spendable immediately
- `blocks_to_unlock`: number of blocks until the full piconero_balance will be unlocked
- `eth_address`: address of the swapd ethereum wallet
- `wei_balance`: balance of the ethereum wallet in wei

Example:
```bash
curl -X POST http://127.0.0.1:5002 -d '{"jsonrpc":"2.0","id":"0","method":"personal_balances","params":{}}' -H 'Content-Type: application/json'
#{"jsonrpc":"2.0","result":{"monero_address":"47RP5qtFwN2fEsRtiXQ5Pe4BDB5UxLxFbbRbvQy4sCLzN8xZxaJTBw25JE7Saz4fCngcY5ZbCk1XN3squfGQzs2pVjgG6tb","piconero_balance":2250425843583586,"piconero_unlocked_balance":175824411726902,"blocks_to_unlock":59,"eth_address":"0xFFcf8FDEE72ac11b5c542428B35EEF5769C409f0","wei_balance":999987682387589565906},"id":"0"}
```

## `swap` namespace

### `swap_cancel`

Attempts to cancel an ongoing swap.

Parameters:
- `id`: id of the swap to refund

Returns:
- `status`: exit status of the swap.

Example:
```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"swap_cancel","params":{"id": "17c01ad48a1f75c1456932b12cb51d430953bb14ffe097195b1f8cace7776e70"}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"status":"Success"},"id":"0"}
```

### `swap_getOngoing`

Gets information about the ongoing swap, if there is one.

Parameters:
- none

Returns:
- `id`: the swap's ID.
- `provided`: the coin provided during the swap.
- `providedAmount`: the amount of coin provided during the swap.
- `receivedAmount`: the amount of coin expected to be received during the swap.
- `exchangeRate`: the exchange rate of the swap, expressed in a ratio of XMR/ETH.
- `status`: the swap's status; should always be "ongoing".

Example:
```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"swap_getOngoing","params":{"id":"17c01ad48a1f75c1456932b12cb51d430953bb14ffe097195b1f8cace7776e70"}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"id":3,"provided":"ETH","providedAmount":0.05,"receivedAmount":0,"exchangeRate":0,"status":"ongoing"},"id":"0"}
```

### `swap_getPastIDs`

Gets all past swap IDs.

Parameters:
- none

Returns:
- `ids`: a list of all past swap IDs.

Example:
```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"swap_getPastIDs","params":{}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"ids":["7492ceb4d0f5f45ecd5d06923b35cae406d1406cd685ce1ba184f2a40c683ac2","17c01ad48a1f75c1456932b12cb51d430953bb14ffe097195b1f8cace7776e70"]},"id":"0"}
```

### `swap_getPast`

Gets a past swap information for the given swap ID.

Paramters:
- `id`: the swap ID.

Returns:
- `provided`: the coin provided during the swap.
- `providedAmount`: the amount of coin provided during the swap.
- `receivedAmount`: the amount of coin received during the swap.
- `exchangeRate`: the exchange rate of the swap, expressed in a ratio of XMR/ETH.
- `status`: the swap's status, one of `success`, `refunded`, or `aborted`.

Example:
```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"swap_getPast","params":{"id": "17c01ad48a1f75c1456932b12cb51d430953bb14ffe097195b1f8cace7776e70"}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"provided":"ETH","providedAmount":0.05,"receivedAmount":1,"exchangeRate":20,"status":"success"},"id":"0"}
```

### `swap_getStage`

Gets the stage of an ongoing swap.

Parameters:
- `id`: id of the swap to get the stage of

Returns:
- `stage`: stage of the swap
- `info`: description of the swap's stage

Example:
```bash
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"swap_getStage","params":{"id": "17c01ad48a1f75c1456932b12cb51d430953bb14ffe097195b1f8cace7776e70"}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"stage":"KeysExchanged", "info":"keys have been exchanged, but no value has been locked"},"id":"0"}
```

## websocket subscriptions

The daemon also runs a websockets server that can be used to subscribe to push notifications for updates. You can use the command-line tool `wscat` to easily connect to a websockets server.

### `swap_subscribeStatus`

Subscribe to updates of status of a swap. Pushes a notification each time the stage updates, and a final push when the swap completes, containing its completion status.

Paramters:
- `id`: the swap ID.

Returns:
- `status`: the swap's status.

Example:
```bash
wscat -c ws://localhost:5001/ws
# Connected (press CTRL+C to quit)
# > {"jsonrpc":"2.0", "method":"swap_subscribeStatus", "params": {"id": "7492ceb4d0f5f45ecd5d06923b35cae406d1406cd685ce1ba184f2a40c683ac2"}, "id": 0}
# < {"jsonrpc":"2.0","result":{"stage":"ETHLocked"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"refunded"},"error":null,"id":null}
```

### `net_makeOfferAndSubscribe`

Make a swap offer and subscribe to updates on it. A notification will be pushed with the swap ID when the offer is taken, as well as status updates after that, until the swap has completed.

Parameters:
- `minimumAmount`: minimum amount to swap, in XMR.
- `maximumAmount`: maximum amount to swap, in XMR.
- `exchangeRate`: exchange rate of ETH-XMR for the swap, expressed in a fraction of XMR/ETH. For example, if you wish to trade 10 XMR for 1 ETH, the exchange rate would be 0.1.
- `ethAsset`: (optional) Ethereum asset to trade, either an ERC-20 token address or the zero address for regular ETH. default: regular ETH

Returns:
- `offerID`: ID of the swap offer.
- `id`: ID of the swap, when the offer is taken and a swap is initiated.
- `status`: the swap's status.

Example (including notifications when swap is taken):
```bash
wscat -c ws://localhost:5002/ws
# Connected (press CTRL+C to quit)
# > {"jsonrpc":"2.0", "method":"net_makeOfferAndSubscribe", "params": {"minimumAmount": 0.1, "maximumAmount": 1, "exchangeRate": 0.05}, "id": 0}
# < {"jsonrpc":"2.0","result":{"offerID":"cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"id":0},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"ExpectingKeys"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"KeysExchanged"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"XMRLocked"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"Success"},"error":null,"id":null}
```

### `net_takeOfferAndSubscribe`

Take an advertised swap offer and subscribe to updates on it. This call will initiate and execute an atomic swap. 

Parameters:
- `multiaddr`: multiaddress of the peer to swap with.
- `offerID`: ID of the swap offer.
- `providesAmount`: amount of ETH you will be providing. Must be between the offer's `minimumAmount * exchangeRate` and `maximumAmount * exchangeRate`. For example, if the offer has a minimum of 1 XMR and a maximum of 5 XMR and an exchange rate of 0.1, you must provide between 0.1 ETH and 0.5 ETH.

Returns:
- `id`: ID of the initiated swap.
- `status`: the swap's status.

Example:
```bash
wscat -c ws://localhost:5001/ws
# Connected (press CTRL+C to quit)
# > {"jsonrpc":"2.0", "method":"net_takeOfferAndSubscribe", "params": {"multiaddr": "/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7", "offerID": "cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9", "providesAmount": 0.05}, "id": 0}
# < {"jsonrpc":"2.0","result":{"id":0},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"ExpectingKeys"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"ETHLocked"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"ContractReady"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"stage":"Success"},"error":null,"id":null}
```
