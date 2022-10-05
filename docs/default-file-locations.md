# Default File Layout

While almost all file locations can be configured with command line options,
it is usually less burdensome to use configuration by convention with the
default directory locations.

## {HOME}/.atomic/{ENV} (also known as DATA_DIR)

The base folder is called the "data dir". You can set it with `--data-dir`, or
you can accept the default of `{HOME}/.atomic/{ENV}`, where `HOME` is the user's
home directory and `ENV` is the set or defaulted value to the `--env` flag.

## {DATA_DIR}/wallet/swap-wallet

The default location for your monero wallet file is `{DATA_DIR}/wallet/swap-wallet`.
You can change the location using the `--wallet-file` flag, but make sure to place
the wallet file in a dedicated wallet directory, as `swapd` will create temporary
swap wallets in the same directory. (TODO: Mention the naming scheme for the
temporary wallets.) Internally, `swapd` launches `monero-wallet-rpc` instance and
the log file for this instance is located in `{DATA_DIR}/wallet/moner-wallet-rpc.log`
or, alternatively one directory above the wallet file's directory when using the
`--wallet-file`.

Note: Monero wallets actually consist of 3 files. Using the default wallet name, the
files are:
* swap-wallet (encrypted with your password)
* swap-wallet.keys (encrypted with your password)
* swap-wallet.address.txt (unprotected convenience file with your address)


* https://monero.stackexchange.com/a/2804/3691


## {DATA_DIR}/net.key

This is the private key that forms your libp2p identity. If the file does not exist,
a new random key will be generated. The location can be configured with `--libp2p-key`.

## {DATA_DIR}/libp2p-datastore

Cache data from libp2p. The directory location is always relative to `DATA_DIR`.
It is safe to delete this directory if `swapd` is not running.
