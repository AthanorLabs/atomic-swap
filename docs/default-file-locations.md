# Default File Locations

While almost all `swapd` file locations can be configured via the command line, it is
usually less burdensome to use convention over configuration with the default file
locations discussed here.

### {HOME}/.atomicswap/{ENV}

The base folder for the default location of `swapd` files is called the "data dir". You
can set it with `--data-dir`. When using the default, `HOME` is the user's home directory
and `ENV` is the value from the `--env` flag (e.g. `stagenet`, `mainnet`).

From here forward, we will use `{DATA_DIR}` to refer to this default value or the value
passed with the `--data-dir` flag.

### {DATA_DIR}/db

This is the location of swapd's (BadgerDB) database. Currently, it stores offers made so that 
they can be reloaded on restart. The database can be safely deleted if you don't have any offers
made, or don't care about saving offers made.

### {DATA_DIR}/wallet/swap-wallet

This is the default location for your monero wallet file. You can change the location
using `--wallet-file`, but place the wallet file in a dedicated wallet directory, as
`swapd` will create temporary, `xmrtaker-*` swap wallets in this same directory.
Internally, `swapd` launches a `monero-wallet-rpc` instance whose log file is located at
`{DATA_DIR}/moner-wallet-rpc.log` or one directory above the wallet file's directory when
using the `--wallet-file` flag.

Note: Monero wallets actually consist of 3 files. Using the default wallet file path, the
files are:
* `{DATA_DIR}/wallet/swap-wallet`
* `{DATA_DIR}/wallet/swap-wallet.keys`
* `{DATA_DIR}/wallet/swap-wallet.address.txt`

When passing the `--wallet-file` flag to `swapd`, we only specify the path to the first
file above. More information on what the individual files contain can be
[found here](https://monero.stackexchange.com/a/2804/3691).

### {DATA_DIR}/eth.key

This is the default location of your Ethereum private key used by swaps. Alternate
locations can be configured with `--eth-privkey`. If the file does not
exist, a new random key will be created and placed in this location.

### {DATA_DIR}/net.key

This is the private key that forms your libp2p identity. If the file does not exist, a new
random key will be generated and placed in this location. Alternate locations can be
configured with `--libp2p-key`.

### {DATA_DIR}/libp2p-datastore

Cache data from libp2p. The directory location is always relative to `DATA_DIR`.
It is safe to delete this directory if `swapd` is not running.

### {DATA_DIR}/info-{DATE}.json

Stores information on a swap when it reaches the stage where ethereum is locked.

### {DATA_DIR}/contract-addresses.json

Only written when `--deploy` is passed to swapd. This file stores the address
that the contract was deployed to along with other data.

## Bootnode default file locations

### {DATA_DIR}/bootnode

By default, all bootnode-related files will be placed in the `bootnode` directory within the data dir.

### {DATA_DIR}/bootnode/net.key

The private key to the bootnode's libp2p identity. If the file does not exist, a new
random key will be generated and placed in this location. Alternate locations can be
configured with `--libp2p-key`. It does not necessarily need to be a different key than that used by swapd.
