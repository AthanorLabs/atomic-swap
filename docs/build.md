# Building the project

## From source

### 1. Install Golang

On Ubuntu, the easiest way to keep up-to-date with the latest stable version of
Go is with snap:
```bash
sudo snap install go --classic
```
On other systems or in docker, use the directions here: https://go.dev/doc/install.
Summary for X86-64 Linux (update GO_VERSION below to the latest stable release):
```bash
GO_VERSION=1.20.5
wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
rm -rf /usr/local/go && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> .profile
source .profile
```

### 2. Clone the repo
```bash
git clone https://github.com/athanorlabs/atomic-swap.git
cd atomic-swap

# Optional: Check out the exact source code of the latest release
git checkout "$(git describe --abbrev=0 --tags)"
```

### 3. Finally, build the repo

#### Option 1: Build the latest released/tagged version.
```bash
make build-release
```
This option, and the ones below, all create `swapd` and `swapcli` binaries in
a `bin` subfolder.

#### Option 2: Build release binaries without checking out the repo

The make-target in "Option 1" uses "go install", a command that can be run
without checking out the source code. If you just want the release binaries in a
subfolder named "bin", you can use:
```bash
GOBIN=${PWD}/bin go install -tags=prod github.com/athanorlabs/atomic-swap/cmd/...@latest
```
Note: `swapd` depends on `monero-wallet-rpc`. If you are not using
`scripts/install-monero-linux.sh` to install it, you'll need to ensure [that the
latest version of `monero-wallet-rpc`](https://www.getmonero.org/downloads/#cli)
is either in your path or in a folder named `monero-bin` of the directory that
`swapd` is started from.

#### Option 3: Use the latest, bleeding edge code

If you just want to run the latest, bleeding edge code that is not always compatible
with the previous release, you can do this:
```bash
git checkout master && git pull
make build
```

Note: If you wish to run a bootnode (see [here](./bootnode.md)), pass
`build-all` instead of `build` as the target for `make`.

## Docker

### 1. Ensure docker is installed on your machine.

For the purposes here, using `docker-ce` directly from Ubuntu's `apt`
repositories or from Docker's repositories will work equally well.

### 2. Build the docker image:
```bash
make docker-images
```

### 3. For an example of how to run `swapd` with docker on stagenet:
```bash
./scripts/docker-swapd/run-docker-image.sh
```

This runs `swapd` on stagenet. The container name is `swapd-stagenet`.

You can interact with it by running `swapcli` inside the container:
```bash
docker exec CONTAINER_NAME_OR_ID swapcli SUBCOMMAND ...
```

You can also set command line arguments with environment variables, eg. to run on mainnet:
```bash
SWAPD_ENV=mainnet SWAPD_ETH_ENDPOINT=your-eth-endpoint ./scripts/docker-swapd/run-docker-image.sh
```
