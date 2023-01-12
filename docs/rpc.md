# JSON-RPC API

The `swapd` program automatically starts a JSON-RPC server that can be used to interact
with the swap network and make/take swap offers.

## `net` namespace

### `net_addresses`

Get the local libp2p listening addresses of the node. Unless you have a public IP
directly attached to your host, these are not the addresses that remote hosts will
directly connect to.

Parameters:
- none

Returns:
- `addresses`: list of libp2p multiaddresses the swap daemon is currently listening on.

Example:

```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"net_addresses","params":{}}' \
| jq .
```
```
{
  "jsonrpc": "2.0",
  "result": {
    "addresses": [
      "/ip4/172.31.32.254/tcp/9900/p2p/12D3KooWQQWDJ7KA1Fwdf2ejWz9VXHKvY8cC5PB7Sf34fbEGbsgV",
      "/ip4/127.0.0.1/tcp/9900/p2p/12D3KooWQQWDJ7KA1Fwdf2ejWz9VXHKvY8cC5PB7Sf34fbEGbsgV",
      "/ip4/172.31.32.254/udp/9900/quic-v1/p2p/12D3KooWQQWDJ7KA1Fwdf2ejWz9VXHKvY8cC5PB7Sf34fbEGbsgV",
      "/ip4/127.0.0.1/udp/9900/quic-v1/p2p/12D3KooWQQWDJ7KA1Fwdf2ejWz9VXHKvY8cC5PB7Sf34fbEGbsgV"
    ]
  },
  "id": "0"
}
```

### `net_discover`

Discover peers on the network via DHT that have active swap offers.

Parameters:
- `provides` (optional): one of `ETH` or `XMR`, depending on which offer you are searching
  for. **Note**: Currently only `XMR` offers are supported. Default is `XMR`.
- `searchTime` (optional): time in seconds to perform the search. Default is 12s.

Returns:
- `peers`: list of lists of peers's multiaddresses. A peer may have multiple multiaddresses, so the nested list pertains to a single peer.

Example:

```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"net_discover","params":{"searchTime":3}}' \
| jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "peerIDs": [
      [
        "12D3KooWHLUrLnJtUbaGzTSi6azZavKhNgUZTtSiUZ9Uy12v1eZ7",
        "12D3KooWGBw6ScWiL6k3pKNT2LR9o6MVh5CtYj1X8E1rdKueYLjv"
      ]
    ]
  },
  "id": "0"
}
```

### `net_queryAll`

Discover peers on the network via DHT that have active swap offers and gets all their swap offers.

Parameters:
- `provides` (optional): one of `ETH` or `XMR`, depending on which offer you are searching
  for. **Note**: Currently only `XMR` offers are supported. Default is `XMR`.
- `searchTime` (optional): duration in seconds for which to perform the search. Default is 12s.

Returns:
- `peersWithOffers`: list of peers's multiaddresses and their current offers.

Example:

```bash
curl -s -X POST http://127.0.0.1:5001 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"net_queryAll","params":{"searchTime":3}}' \
| jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "peersWithOffers": [
      {
        "peerID": "12D3KooWGVzz2d2LSceVFFdqTYqmQXTqc5eWziw7PLRahCWGJhKB",
        "offers": [
          {
            "offerID": "0xa7429fdb7ce0c0b19bd2450cb6f8274aa9d86b3e5f9386279e95671c24fd8381",
            "provides": "XMR",
            "minAmount": "0.1",
            "maxAmount": "1",
            "exchangeRate": "0.5",
            "ethAsset": "ETH"
          }
        ]
      },
      {
        "peerID": "12D3KooWS8iKxqsGTiL3Yc1VaAfg99U5km1AE7bWYQiuavXj3Yz6",
        "offers": [
          {
            "offerID": "0x25188edd7573f43fca5760f0aacdc1a358171a8fc6bdf11876fa937f77fc583c",
            "minAmount": "0.1",
            "maxAmount": "1",
            "provides": "XMR",
            "exchangeRate": "0.49",
            "ethAsset": "ETH"
          }
        ]
      }
    ]
  },
  "id": "0"
}
```

### `net_queryPeer`

Query a specific peer for their current active offers.

Parameters:
- `multiaddr`: multiaddress of the peer to query. Found via `net_discover`.

Returns:
- `offers`: list of the peer's current active offers.

Example:

```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"net_queryPeer","params":
{"peerID":"12D3KooWGBw6ScWiL6k3pKNT2LR9o6MVh5CtYj1X8E1rdKueYLjv"}}' \
| jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "offers": [
      {
        "version": "0.1.0",
        "offerID": "0xa7429fdb7ce0c0b19bd2450cb6f8274aa9d86b3e5f9386279e95671c24fd8381",
        "provides": "XMR",
        "minAmount": "0.5",
        "maxAmount": "1",
        "exchangeRate": "0.1",
        "ethAsset": "ETH"
      }
    ]
  },
  "id": "0"
}
```

### `net_makeOffer`

Make a new swap offer and advertise it on the network. **Note:** Currently only XMR offers can be made.

Parameters:
- `minAmount`: minimum amount to swap, in XMR.
- `maxAmount`: maximum amount to swap, in XMR.
- `exchangeRate`: exchange rate of ETH-XMR for the swap, expressed in a fraction of
  XMR/ETH. For example, if you wish to trade 10 XMR for 1 ETH, the exchange rate would be
  0.1.
- `ethAsset`: (optional) Ethereum asset to trade, either an ERC-20 token address or the
  zero address for regular ETH. default: regular ETH
- `relayerEndpoint`: (optional) RPC endpoint of the relayer to use for submitting claim
  transactions.
- `relayerCommission`: (optional) Commission in percentage that the relayer receives for
  submitting the claim transaction.

Returns:
- `offerID`: ID of the swap offer.

Example:
```bash
curl -s -X POST http://127.0.0.1:5001 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"net_makeOffer",
"params":{"minAmount":"1", "maxAmount":"10", "exchangeRate": "0.1"}}' \
| jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "peerID": "12D3KooWGBw6ScWiL6k3pKNT2LR9o6MVh5CtYj1X8E1rdKueYLjv",
    "offerID": "0x9549685d15cd9a136111db755e5440b4c95e266ba39dc0c84834714d185dc6f0"
  },
  "id": "0"
}
```

### `net_takeOffer`

Take an advertised swap offer. This call will initiate and execute an atomic swap.
**Note:** You must be the ETH holder to take a swap.

Parameters:
- `peerID`: ID of the peer to swap with.
- `offerID`: ID of the swap offer.
- `providesAmount`: amount of ETH you will be providing. Must be between the offer's
  `minAmount * exchangeRate` and `maxAmount * exchangeRate`. For example, if the offer has
  a minimum of 1 XMR and a maximum of 5 XMR and an exchange rate of 0.1, you must provide
  between 0.1 ETH and 0.5 ETH.

Returns:
- null

Example:
```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"net_takeOffer",
  "params":{
    "peerID":"12D3KooWGBw6ScWiL6k3pKNT2LR9o6MVh5CtYj1X8E1rdKueYLjv",
    "offerID":"0x9549685d15cd9a136111db755e5440b4c95e266ba39dc0c84834714d185dc6f0",
    "providesAmount": "0.3"
  }
}'
```
```json
{"jsonrpc":"2.0","result":null,"id":"0"}
```

### `net_takeOfferSync`

Take an advertised swap offer. This call will initiate and execute an atomic swap. It will
not return until the swap has completed, after which it will return whether the swap was
successful or not. **Note:** You must be the ETH holder to take a swap.

Parameters:
- `peerID`: ID of the peer to swap with.
- `offerID`: ID of the swap offer.
- `providesAmount`: amount of ETH you will be providing. Must be between the offer's
  `minimumAmount * exchangeRate` and `maximumAmount * exchangeRate`. For example, if the
  offer has a minimum of 1 XMR and a maximum of 5 XMR and an exchange rate of 0.1, you
  must provide between 0.1 ETH and 0.5 ETH.

Returns:
- `status`: the swap's status, one of `Success`, `Refunded`, or `Aborted`.

Example:
```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"net_takeOfferSync","params":{
  "peerID": "12D3KooWGBw6ScWiL6k3pKNT2LR9o6MVh5CtYj1X8E1rdKueYLjv",
  "offerID":"0xa7429fdb7ce0c0b19bd2450cb6f8274aa9d86b3e5f9386279e95671c24fd8381",
  "providesAmount": "0.03"
  }
}'
```
```json
{"jsonrpc":"2.0","result":{"status":"Success"},"id":"0"}
```


## `personal` namespace

### `personal_balances`

Returns combined information of both the Monero and Ethereum account addresses  and balances.

Parameters:
- none

Returns:
- `moneroAddress`: primary monero address of the swapd wallet
- `piconeroBalance`: balance the swapd wallet in piconero
- `piconeroUnlockedBalance`: balance the swapd wallet in piconero that is spendable immediately
- `blocksToUnlock`: number of blocks until the full piconero_balance will be unlocked
- `ethAddress`: address of the swapd ethereum wallet
- `weiBalance`: balance of the ethereum wallet in wei

Example:
```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"personal_balances","params":{}}' | jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "moneroAddress": "5BVXdWxKp5aMWRfkAiWYb38dPuDFDwTYwCL5ymSoe9CPcLN3c8BanUsiBG8KaGtmQ8W6X2yzCCsvsGjuSYvn8LSZUUV7QB3",
    "piconeroBalance": 149935630269820,
    "piconeroUnlockedBalance": 138815986625976,
    "blocksToUnlock": 37,
    "ethAddress": "0x297d1DdeA7224252fD629442989C569f23Ffc7FD",
    "weiBalance": 429169302264321300
  },
  "id": "0"
}
```

### `personal_setSwapTimeout`

Configures the `_timeoutDuration` used when the ethereum newSwap transaction is created.
This method is only for testing. In non-dev networks, the swaps are configured to fail if
you don't use the defaults.

Parameters:
- `timeout`: duration value in seconds 

Returns:
- null

```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"personal_setSwapTimeout","params":{"timeout":7200}}'
```
```json
{"jsonrpc":"2.0","result":null,"id":"0"}
```

### `personal_setSwapTimeout`

Sets the duration between swap initiation and t0 and t0 and t1, in seconds.

Parameters:
- `duration`: duration of timeout, in seconds

Returns:
- null

Example:
```bash
curl -X POST http://127.0.0.1:5002 -d '{"jsonrpc":"2.0","id":"0","method":"personal_setSwapTimeout","params":{"duration": 120}}' -H 'Content-Type: application/json'
#{"jsonrpc":"2.0","result":null,"id":"0"}
```

### `personal_getSwapTimeout`

Returns the duration between swap initiation and t0 and t0 and t1, in seconds

Parameters:
- none

Returns:
- `timeout`: timeout in seconds

Example:
```bash
curl -X POST http://127.0.0.1:5002 -d '{"jsonrpc":"2.0","id":"0","method":"personal_getSwapTimeout","params":{}}' -H 'Content-Type: application/json'
#{"jsonrpc":"2.0","result":{"timeout":120},"id":"0"}
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
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"swap_cancel",
"params":{"offerID": "0x17c01ad48a1f75c1456932b12cb51d430953bb14ffe097195b1f8cace7776e70"}}'
```
```json
{"jsonrpc":"2.0","result":{"status":"Success"},"id":"0"}
```

### `swap_getOngoing`

Gets information for the specified ongoing swap.

Parameters:
- `offerID`: the swap's ID.

Returns:
- `provided`: the coin provided during the swap.
- `providedAmount`: the amount of coin provided during the swap.
- `receivedAmount`: the amount of coin expected to be received during the swap.
- `exchangeRate`: the exchange rate of the swap, expressed in a ratio of XMR/ETH.
- `status`: the swap's status; should always be "ongoing".

Example:
```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"swap_getOngoing",
"params":{"offerID":"0xa7429fdb7ce0c0b19bd2450cb6f8274aa9d86b3e5f9386279e95671c24fd8381"}}' | jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "provided": "ETH",
    "providedAmount": "0.01",
    "receivedAmount": "1",
    "exchangeRate": "0.01",
    "status": "ETHLocked"
  },
  "id": "0"
}
```

### `swap_getPastIDs`

Gets all past swap IDs.

Parameters:
- none

Returns:
- `ids`: a list of all past swap IDs.

Example:
```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"swap_getPastIDs","params":{}}' | jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "ids": [
      "0x6ca19b496426bfbb97ccab16473db4cf50faf77074809701c8478afe7d8370d9",
      "0x9549685d15cd9a136111db755e5440b4c95e266ba39dc0c84834714d185dc6f0",
      "0xa55ba276c4c6bc77713776cd50fa2d20d31b8b26ed67be458e0c1a5794721587",
      "0x2ed05e12ecd45d992b523ee52a516f51d15480d4b7805a29733f9f000efc17d3",
      "0xa7429fdb7ce0c0b19bd2450cb6f8274aa9d86b3e5f9386279e95671c24fd8381"
    ]
  },
  "id": "0"
}
```

### `swap_getPast`

Gets a past swap information for the given swap ID.

Paramters:
- `offerID`: the swap ID.

Returns:
- `provided`: the coin provided during the swap.
- `providedAmount`: the amount of coin provided during the swap.
- `receivedAmount`: the amount of coin received during the swap.
- `exchangeRate`: the exchange rate of the swap, expressed in a ratio of XMR/ETH.
- `status`: the swap's status, one of `success`, `refunded`, or `aborted`.

Example:
```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"swap_getPast",
"params":{"offerID": "0xa7429fdb7ce0c0b19bd2450cb6f8274aa9d86b3e5f9386279e95671c24fd8381"}}' \
| jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "provided": "ETH",
    "providedAmount": "0.01",
    "receivedAmount": "1",
    "exchangeRate": "0.01",
    "status": "Success"
  },
  "id": "0"
}
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
curl -s -X POST http://127.0.0.1:5001 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"swap_getStage",
"params":{"offerID": "0xbe6cb622906510e69339fa5d8e7d60c90bad762deb8d06985466dd9144809040"}}' \
| jq
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "stage": "KeysExchanged",
    "info": "keys have been exchanged, but no value has been locked"
  },
  "id": "0"
}
```

### `swap_suggestedExchangeRate`

Returns the current mainnet exchange rate expressed as the XMR/ETH price ratio.

Parameters:
- none

Returns:
- `ethPrice`: the current ETH/USD price multiplied by 10^8.
- `xmrPrice`: the current XMR/USD price multiplied by 10^8.
- `exchangeRate`: the exchange rate expressed as the XMR/ETH price ratio.

- `ethUpdatedAt`: time when the ETH price was last updated (RFC 3339 formatted)
- `ethPrice`: current ETH/USD price (8 decimal points or less)
- `xmrUpdatedAt`: time when the XMR price was last updated (RFC 3339 formatted)
- `xmrPrice`: the current XMR/USD price (8 decimal points or less)
- `exchangeRate`: "0.119571"

Example:
```bash
curl -s -X POST http://127.0.0.1:5000 -H 'Content-Type: application/json' -d \
'{"jsonrpc":"2.0","id":"0","method":"swap_suggestedExchangeRate","params":{}}' \
| jq .
```
```json
{
  "jsonrpc": "2.0",
  "result": {
    "ethUpdatedAt": "2023-01-12T14:26:11-06:00",
    "ethPrice": "1430.10000000",
    "xmrUpdatedAt": "2023-01-12T14:22:23-06:00",
    "xmrPrice": "170.99780000",
    "exchangeRate": "0.119571"
  },
  "id": "0"
}


curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"swap_suggestedExchangeRate","params":{}}' -H 'Content-Type: application/json'
# {"jsonrpc":"2.0","result":{"ethPrice":118530759250,"xmrPrice":14453000000,"exchangeRate":0.12193459395224451},"id":"0"}
```

## websocket subscriptions

The daemon also runs a websockets server that can be used to subscribe to push
notifications for updates. You can use the command-line tool `wscat` to easily connect to
a websockets server.

### `swap_subscribeStatus`

Subscribe to updates of status of a swap. Pushes a notification each time the stage
updates, and a final push when the swap completes, containing its completion status.

Paramters:
- `offerID`: the swap ID.

Returns:
- `status`: the swap's status.

Example:
```bash
wscat -c ws://localhost:5001/ws
# Connected (press CTRL+C to quit)

# > {"jsonrpc":"2.0", "method":"swap_subscribeStatus", "params": {"offerID": "0x6610ef5ba1c093a5c88eb0c2b21be22aa92e68943ac88da1cd45b3e58f8f3166"}, "id": 0}

# < {"jsonrpc":"2.0","result":{"status":"XMRLocked"},"error":null,"id":null}
# < {"jsonrpc":"2.0","result":{"status":"Success"},"error":null,"id":null}
```

### `net_makeOfferAndSubscribe`

Make a swap offer and subscribe to updates on it. A notification will be pushed with the
swap ID when the offer is taken, as well as status updates after that, until the swap has
completed.

Parameters:
- `minAmount`: minimum amount to swap, in XMR.
- `maxAmount`: maximum amount to swap, in XMR.
- `exchangeRate`: exchange rate of ETH-XMR for the swap, expressed in a fraction of
  XMR/ETH. For example, if you wish to trade 10 XMR for 1 ETH, the exchange rate would be
  0.1.
- `ethAsset`: (optional) Ethereum asset to trade, either an ERC-20 token address or the
  zero address for regular ETH. default: regular ETH

Returns:
- `offerID`: ID of the offer which will become the ID of the swap when taken.
- `peerID`: Your peer ID which needs to be specified by the party taking the offer.
- `status`: the swap's status.

Example (including notifications when swap is taken):
```
wscat -c ws://localhost:5000/ws
Connected (press CTRL+C to quit)

> {"jsonrpc":"2.0", "method":"net_makeOfferAndSubscribe", "params": {"minAmount": "0.1", "maxAmount": "1", "exchangeRate": "0.05"}, "id": 0}

< {"jsonrpc":"2.0","result":{"peerID":"12D3KooWNseb7Ei8Xx1aBKjSFoZ9PGfdxN9MwQxfSRxsBAyA8op4","offerID":"0x64f49193dc5e8d70893331498b76a156e33ed8cdf46a1f901c7fab59a827e840"},"error":null,"id":null}
< {"jsonrpc":"2.0","result":{"status":"KeysExchanged"},"error":null,"id":null}
< {"jsonrpc":"2.0","result":{"status":"XMRLocked"},"error":null,"id":null}
< {"jsonrpc":"2.0","result":{"status":"Success"},"error":null,"id":null}
```

### `net_takeOfferAndSubscribe`

Take an advertised swap offer and subscribe to updates on it. This call will initiate and
execute an atomic swap.

Parameters:
- `peerID`: Peer ID of the XMR maker, the party that created the offer.
- `offerID`: ID of the swap offer.
- `providesAmount`: amount of ETH you will be providing. Must be between the offer's
  `minAmount * exchangeRate` and `maxAmount * exchangeRate`. For example, if the
  offer has a minimum of 1 XMR and a maximum of 5 XMR and an exchange rate of 0.1, you
  must provide between 0.1 ETH and 0.5 ETH.

Returns:
- `offerID`: ID of the initiated swap.
- `status`: the swap's status.

Example:
```
wscat -c ws://localhost:5001/ws
Connected (press CTRL+C to quit)

> {"jsonrpc":"2.0", "method":"net_takeOfferAndSubscribe", "params": {"peerID": "12D3KooWNseb7Ei8Xx1aBKjSFoZ9PGfdxN9MwQxfSRxsBAyA8op4", "offerID": "0x64f49193dc5e8d70893331498b76a156e33ed8cdf46a1f901c7fab59a827e840", "providesAmount": "0.025"}, "id": 0}

< {"jsonrpc":"2.0","result":{"status":"ExpectingKeys"},"error":null,"id":null}
< {"jsonrpc":"2.0","result":{"status":"ETHLocked"},"error":null,"id":null}
< {"jsonrpc":"2.0","result":{"status":"ContractReady"},"error":null,"id":null}
< {"jsonrpc":"2.0","result":{"status":"Success"},"error":null,"id":null}
```
