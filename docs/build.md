# Building the project

1. Install Golang

On Ubuntu, the easiest way to keep up-to-date with the latest stable version of
Go is with snap:
```bash
sudo snap install go --classic
```
On other systems or in docker, see https://go.dev/doc/install

2. Clone the repo:
```bash
git clone https://github.com/noot/atomic-swap.git
cd atomic-swap
```

3. Finally, build the repo:
```bash
make build
```

This creates the binaries `swapd` and `swapcli`.
