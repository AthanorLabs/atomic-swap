# Building the project

1. Install Golang

On Ubuntu, the easiest way to keep up-to-date with the latest stable version of
Go is with snap:
```bash
sudo snap install go --classic
```
On other systems or in docker, use the directions here: https://go.dev/doc/install.
Summary for X86-64 Linux (update GO_VERSION below to the latest stable release):
```bash
GO_VERSION=1.20.3
wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
rm -rf /usr/local/go && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> .profile
source .profile
```

2. Clone the repo:
```bash
git clone https://github.com/athanorlabs/atomic-swap.git
cd atomic-swap
```

3. Finally, build the repo:
```bash
make build
```

This creates `swapd`, `swapcli` and `relayer` binaries in the `bin` folder.
