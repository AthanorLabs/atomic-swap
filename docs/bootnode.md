# Bootnode

The swap daemon uses a p2p network to discover offer-makers to do a swap with, and also the 
run the actual swap protocol. A node must know addresses of nodes already in the network to 
join the network. These nodes are often publicly posted and referred to as bootnodes. 
Bootnodes act as an entry-point into the p2p network. 

This repo comes with a `bootnode` program that runs only the p2p components of a swap node, 
and thus can be used as a lightweight bootnode. 

## Requirements
- see [build instructions](./build.md) for installation requirements.

## Build and run

To build and run the bootnode binary:
```bash
make build-all
./bin/bootnode --env ENVIRONMENT
```

`ENVIRONMENT` is one of `mainnet`, `stagenet`, or `dev`.

To get the p2p addresses of the bootnode:
```bash
./bin/swapcli addresses
```

You can then distribute these addresses for other swap nodes to connect to.