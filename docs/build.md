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
GO_VERSION=1.21.4
wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
rm -rf /usr/local/go && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> .profile
source .profile
```

### 2a. Build without cloning the repo (option 1)

If this is your first time testing the software and you don't have an up-to-date
installation of `monero-wallet-rpc` in your path, you may want to skip to 2b
(option 2), as the repo has a script, `scripts/install-monero-linux.sh`, for
installing the latest monero tools to a `monero-bin` subfolder.

Install the atomic swap binaries to a subfolder named `bin`. If you adjust the
install directory to something else, make sure to also adjust documented sample
commands accordingly:
```bash
GOBIN=${PWD}/bin go install -tags=prod github.com/athanorlabs/atomic-swap/cmd/...@latest
```

### 2b. Build from a cloned repo (option 2)

Clone the repo, put it on the commit hash of the most recent release, and build
the binaries:
```bash
git clone https://github.com/athanorlabs/atomic-swap.git
cd atomic-swap

# Check out the exact source code of the latest release
git checkout "$(git describe --abbrev=0 --tags)"

make build-release
```

Note that `build-release` always builds the latest tagged release, not the
currently checked out code, so the `git checkout` command above is not required
for the correct binaries. If you want to build the checked out code as-is, use
`make build` or `make build-all` (the latter includes the `bootnode`
executable), as you'll see in the next example.

If you wish to build the bleeding edge code that is not always compatible with
the previous release, do:
```bash
git checkout master && git pull
make build
```

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
