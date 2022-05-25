# Building the project

1. Install go [here](https://go.dev/doc/install).

For Linux 64-bit:
```bash
wget https://go.dev/dl/go1.18.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> .profile
source .profile
```

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