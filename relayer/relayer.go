// Package relayer provides libraries for creating and validating relay requests and responses.
package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	rcommon "github.com/athanorlabs/go-relayer/common"
	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/net/message"
)

// DefaultRelayerFee is the default fee for swap relayers.
// It's set to 0.009 ETH.
var DefaultRelayerFee = big.NewInt(9e15)

func forwarderFromAddress(address ethcommon.Address, ec *ethclient.Client) (*gsnforwarder.IForwarder, error) {
	forwarder, err := gsnforwarder.NewIForwarder(address, ec)
	if err != nil {
		return nil, err
	}

	return forwarder, nil
}

// CreateRelayClaimRequest fills and returns a RelayClaimRequest ready for
// submission to a relayer.
func CreateRelayClaimRequest(
	ctx context.Context,
	sk *ecdsa.PrivateKey, // Used for signing and the claim address
	ec *ethclient.Client,
	swapFactoryAddress ethcommon.Address,
	forwarderAddress ethcommon.Address,
	calldata []byte,
) (*message.RelayClaimRequest, error) {
	claimAddress := ethcrypto.PubkeyToAddress(sk.PublicKey)

	forwarder, err := forwarderFromAddress(forwarderAddress, ec)
	if err != nil {
		return nil, err
	}

	chainID, err := ec.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: ctx}

	nonce, err := forwarder.GetNonce(callOpts, claimAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce from forwarder: %w", err)
	}

	domainSeparator, err := rcommon.GetEIP712DomainSeparator(gsnforwarder.DefaultName,
		gsnforwarder.DefaultVersion, chainID, forwarderAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get EIP712 domain separator: %w", err)
	}

	req := &gsnforwarder.IForwarderForwardRequest{
		From:           claimAddress,
		To:             swapFactoryAddress,
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

	sig, err := rcommon.NewKeyFromPrivateKey(sk).Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("failed to sign forward request digest: %w", err)
	}

	err = forwarder.Verify(callOpts, *req, domainSeparator, gsnforwarder.ForwardRequestTypehash, nil, sig)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	return &message.RelayClaimRequest{
		ClaimerAddress:    req.From,
		SFContractAddress: req.To,
		Gas:               req.Gas,
		Nonce:             req.Nonce,
		Data:              req.Data,
		Signature:         sig,
		ValidUntilTime:    req.ValidUntilTime,
		DomainSeparator:   domainSeparator,
		RequestTypeHash:   gsnforwarder.ForwardRequestTypehash,
	}, nil
}
