// Package relayer provides client libraries that Bob (an XMR maker) can use to interact
// with a relay server that will pay the Ethereum gas fees needed to receive an Ethereum
// asset in exchange for Monero.
package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	client "github.com/athanorlabs/go-relayer-client"
	rcommon "github.com/athanorlabs/go-relayer/common"
	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
)

// Client is a client that can submit transactions to an Ethereum meta-transaction relayer.
// It has a private key for signing transactions to be forwarded.
type Client struct {
	key              *rcommon.Key
	c                *client.Client
	chainID          *big.Int
	forwarder        *gsnforwarder.IForwarder
	forwarderAddress ethcommon.Address
}

// NewClient returns a new relayer client.
func NewClient(
	sk *ecdsa.PrivateKey,
	ec *ethclient.Client,
	relayerEndpoint string,
	forwarderAddress ethcommon.Address,
) (*Client, error) {
	forwarder, err := forwarderFromAddress(forwarderAddress, ec)
	if err != nil {
		return nil, err
	}

	chainID, err := ec.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{
		key:              rcommon.NewKeyFromPrivateKey(sk),
		c:                client.NewClient(relayerEndpoint),
		chainID:          chainID,
		forwarder:        forwarder,
		forwarderAddress: forwarderAddress,
	}, nil
}

func forwarderFromAddress(address ethcommon.Address, ec *ethclient.Client) (*gsnforwarder.IForwarder, error) {
	forwarder, err := gsnforwarder.NewIForwarder(address, ec)
	if err != nil {
		return nil, err
	}

	return forwarder, nil
}

// SubmitTransaction submits a transaction with the given calldata to the relayer.
func (c *Client) SubmitTransaction(
	to ethcommon.Address,
	calldata []byte,
) (ethcommon.Hash, error) {
	rpcReq, err := createSubmitTransactionRequest(
		c.key,
		c.forwarder,
		c.forwarderAddress,
		c.chainID,
		to,
		calldata,
	)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	// submit transaction to relayer
	resp, err := c.c.SubmitTransaction(rpcReq)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to submit transaction to relayer: %w", err)
	}

	return resp.TxHash, nil
}

// CreateSubmitTransactionRequest fills and returns a SubmitTransactionRequest ready for submission
// to a relayer.
func CreateSubmitTransactionRequest(
	sk *ecdsa.PrivateKey,
	ec *ethclient.Client,
	forwarderAddress ethcommon.Address,
	to ethcommon.Address,
	calldata []byte,
) (*rcommon.SubmitTransactionRequest, error) {
	forwarder, err := forwarderFromAddress(forwarderAddress, ec)
	if err != nil {
		return nil, err
	}

	chainID, err := ec.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	key := rcommon.NewKeyFromPrivateKey(sk)
	return createSubmitTransactionRequest(
		key,
		forwarder,
		forwarderAddress,
		chainID,
		to,
		calldata,
	)
}

func createSubmitTransactionRequest(
	key *rcommon.Key,
	forwarder *gsnforwarder.IForwarder,
	forwarderAddress ethcommon.Address,
	chainID *big.Int,
	to ethcommon.Address,
	calldata []byte,
) (*rcommon.SubmitTransactionRequest, error) {
	nonce, err := forwarder.GetNonce(&bind.CallOpts{}, key.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce from forwarder: %w", err)
	}

	domainSeparator, err := rcommon.GetEIP712DomainSeparator(gsnforwarder.DefaultName,
		gsnforwarder.DefaultVersion, chainID, forwarderAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get EIP712 domain separator: %w", err)
	}

	req := &gsnforwarder.IForwarderForwardRequest{
		From:           key.Address(),
		To:             to,
		Value:          big.NewInt(0),
		Gas:            big.NewInt(200000), // TODO: fetch from ethclient
		Nonce:          nonce,
		Data:           calldata,
		ValidUntilTime: big.NewInt(0),
	}

	digest, err := rcommon.GetForwardRequestDigestToSign(
		req,
		domainSeparator,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get forward request digest: %w", err)
	}

	sig, err := key.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("failed to sign forward request digest: %w", err)
	}

	err = forwarder.Verify(
		&bind.CallOpts{},
		*req,
		domainSeparator,
		gsnforwarder.ForwardRequestTypehash,
		nil,
		sig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	return &rcommon.SubmitTransactionRequest{
		From:            req.From,
		To:              req.To,
		Value:           req.Value,
		Gas:             req.Gas,
		Nonce:           req.Nonce,
		Data:            req.Data,
		Signature:       sig,
		ValidUntilTime:  big.NewInt(0),
		DomainSeparator: domainSeparator,
		RequestTypeHash: gsnforwarder.ForwardRequestTypehash,
	}, nil
}
