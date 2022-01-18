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

```
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_addresses","params":{}}' -H 'Content-Type: application/json'
```

```
{"jsonrpc":"2.0","result":{"addresses":["/ip4/192.168.0.101/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2","/ip4/127.0.0.1/tcp/9933/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2","/ip4/38.88.101.233/tcp/14815/p2p/12D3KooWAYn1T8Lu122Pav4zAogjpeU61usLTNZpLRNh9gCqY6X2"]},"id":"0"}
```

### `net_discover`

Discover peers on the network via DHT that have active swap offers.

Parameters:
- `provides` (optional): one of `ETH` or `XMR`, depending on which offer you are searching for. **Note**: Currently only `XMR` offers are supported. Default is `XMR`.
- `searchTime` (optional): duration in seconds for which to perform the search. Default is 12s.

Returns:
- `peers`: list of lists of peers's multiaddresses. A peer may have multiple multiaddresses, so the nested list pertains to a single peer.

Example:

```
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_discover","params":{"searchTime":3}}' -H 'Content-Type: application/json'
```

```
{"jsonrpc":"2.0","result":{"peers":[["/ip4/127.0.0.1/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7","/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7"]]},"id":"0"}
```

### `net_queryPeer`

Query a specific peer for their current active offers.

Parameters:
- `multiaddr`: multiaddress of the peer to query. Found via `net_discover`.

Returns:
- `offers`: list of the peer's current active offers.

Example:

```
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_queryPeer","params":{"multiaddr":"/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7"}}' -H 'Content-Type: application/json'
```

```
{"jsonrpc":"2.0","result":{"offers":[{"ID":[207,75,240,26,7,117,160,209,63,164,27,20,81,110,75,137,3,67,0,112,122,23,84,224,217,155,101,246,203,111,255,185],"Provides":"XMR","MinimumAmount":0.1,"MaximumAmount":1,"ExchangeRate":0.05}]},"id":"0"}
```

### `net_makeOffer`

Make a new swap offer and advertise it on the network. **Note:** Currently only XMR offers can be made.

Parameters:
- `minimumAmount`: minimum amount to swap, in XMR.
- `maximumAmount`: maximum amount to swap, in XMR.
- `exchangeRate`: exchange rate of ETH-XMR for the swap, expressed in a fraction of XMR/ETH. For example, if you wish to trade 10 XMR for 1 ETH, the exchange rate would be 0.1.

Returns:
- `offerID`: ID of the swap offer.

Example:
```
curl -X POST http://127.0.0.1:5002 -d '{"jsonrpc":"2.0","id":"0","method":"net_makeOffer","params":{"minimumAmount":1, "maximumAmount":10, "exchangeRate": 0.1}}' -H 'Content-Type: application/json'
```

```
{"jsonrpc":"2.0","result":{"offerID":"12b9d56a4c568c772a4e099aaed03a457256d6680562be2a518753f75d75b7ad"},"id":"0"}
```


### `net_takeOffer`

Take an advertised swap offer. This call will initiate and execute an atomic swap. **Note:** You must be the ETH holder to take a swap.

Parameters:
- `multiaddr`: multiaddress of the peer to swap with.
- `offerID`: ID of the swap offer.
- `providesAmount`: amount of ETH you will be providing. Must be between the offer's `minimumAmount * exchangeRate` and `maximumAmount * exchangeRate`. For example, if the offer has a minimum of 1 XMR and a maximum of 5 XMR and an exchange rate of 0.1, you must provide between 0.1 ETH and 0.5 ETH.

Example:
```
curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_takeOffer","params":{"multiaddr":"/ip4/192.168.0.101/tcp/9934/p2p/12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7", "offerID":"12b9d56a4c568c772a4e099aaed03a457256d6680562be2a518753f75d75b7ad", "providesAmount": 0.3}}' -H 'Content-Type: application/json'
```

```
{"jsonrpc":"2.0","result":{"success":true,"receivedAmount":2.999999999999},"id":"0"}
```


## `personal` namespace

### `personal_setMoneroWalletFile`

Sets the node's monero wallet file. The wallet file must be in the directory specified by `--wallet-dir` when starting the `monero-wallet-rpc` server.

Parameters:
- `walletFile`: name of the wallet file.
- `walletPassword`: password to the wallet.

Returns:
- none

Example:
```
curl -X POST http://127.0.0.1:5002 -d '{"jsonrpc":"2.0","id":"0","method":"personal_setMoneroWalletFile","params":{"walletFile":"test-wallet", "walletPassword": ""}}' -H 'Content-Type: application/json'
```
```
{"jsonrpc":"2.0","result":null,"id":"0"}
```