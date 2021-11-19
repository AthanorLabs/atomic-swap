package bob

import (
	"errors"
	"fmt"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
)

func (b *bob) Provides() net.ProvidesCoin {
	return net.ProvidesXMR
}

func (b *bob) SendKeysMessage() (*net.SendKeysMessage, error) {
	sk, vk, err := b.generateKeys()
	if err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		PublicSpendKey: sk.Hex(),
		PrivateViewKey: vk.Hex(),
	}, nil
}

func (b *bob) InitiateProtocol(providesAmount, desiredAmount uint64) error {
	if err := b.initiate(providesAmount, desiredAmount); err != nil {
		return err
	}

	b.setNextExpectedMessage(&net.SendKeysMessage{})
	return nil
}

func (b *bob) initiate(providesAmount, desiredAmount uint64) error {
	if b.initiated {
		return errors.New("protocol already in progress")
	}

	balance, err := b.client.GetBalance(0)
	if err != nil {
		return err
	}

	// check user's balance and that they actualy have what they will provide
	if balance.UnlockedBalance <= float64(b.providesAmount) {
		return errors.New("balance lower than amount to be provided")
	}

	b.initiated = true
	b.providesAmount = providesAmount
	b.desiredAmount = desiredAmount
	return nil
}

// ProtocolComplete is called when the protocol is done, whether it finished successfully or not.
func (b *bob) ProtocolComplete() {
	b.initiated = false
	b.setNextExpectedMessage(&net.InitiateMessage{})
}

func (b *bob) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	if err := b.checkMessageType(msg); err != nil {
		return nil, true, err
	}

	// TODO: put this in *bob.swapState
	readyCh := make(chan struct{})

	switch msg := msg.(type) {
	case *net.InitiateMessage:
		if msg.Provides != net.ProvidesETH {
			return nil, true, errors.New("peer does not provide ETH")
		}

		if err := b.handleSendKeysMessage(msg.SendKeysMessage); err != nil {
			return nil, true, err
		}

		resp, err := b.SendKeysMessage()
		if err != nil {
			return nil, true, err
		}

		if err = b.initiate(msg.DesiredAmount, msg.ProvidesAmount); err != nil {
			return nil, true, err
		}

		return resp, false, nil
	case *net.SendKeysMessage:
		if err := b.handleSendKeysMessage(msg); err != nil {
			return nil, true, err
		}

		// we initiated, so we're now waiting for Alice to deploy the contract.
		return nil, false, nil
	case *net.NotifyContractDeployed:
		if msg.Address == "" {
			return nil, true, errors.New("got empty contract address")
		}

		b.setNextExpectedMessage(&net.NotifyReady{})
		log.Infof("got Swap contract address! address=%s", msg.Address)

		if err := b.setContract(ethcommon.HexToAddress(msg.Address)); err != nil {
			return nil, true, fmt.Errorf("failed to instantiate contract instance: %w", err)
		}

		// TODO: add t0 timeout case

		addrAB, err := b.lockFunds(b.providesAmount)
		if err != nil {
			return nil, true, fmt.Errorf("failed to lock funds: %w", err)
		}

		out := &net.NotifyXMRLock{
			Address: string(addrAB),
		}

		go func() {
			st0, err := b.contract.Timeout0(b.callOpts)
			if err != nil {
				log.Errorf("failed to get timeout0 from contract: err=%s", err)
				return
			}

			t0 := time.Unix(st0.Int64(), 0)
			until := time.Until(t0)

			select {
			case <-b.ctx.Done():
				return
			case <-time.After(until):
				// we can now call Claim()
				if _, err = b.claimFunds(); err != nil {
					log.Errorf("failed to claim: err=%s", err)
					return
				}

				// TODO: send *net.NotifyClaimed
			case <-readyCh:
				return
			}
		}()

		return out, false, nil
	case *net.NotifyReady:
		log.Debug("Alice called Ready(), attempting to claim funds...")
		close(readyCh)

		// contract ready, let's claim our ether
		txHash, err := b.claimFunds()
		if err != nil {
			return nil, true, fmt.Errorf("failed to redeem ether: %w", err)
		}

		log.Debug("funds claimed!!")
		out := &net.NotifyClaimed{
			TxHash: txHash,
		}

		return out, true, nil
	case *net.NotifyRefund:
		// TODO: generate wallet
		return nil, false, errors.New("unimplemented")
	default:
		return nil, false, errors.New("unexpected message type")
	}
}

func (b *bob) handleSendKeysMessage(msg *net.SendKeysMessage) error {
	if msg.PublicSpendKey == "" || msg.PublicViewKey == "" {
		return errors.New("did not receive Alice's public spend or view key")
	}

	log.Debug("got Alice's public keys")
	b.setNextExpectedMessage(&net.NotifyContractDeployed{})

	kp, err := monero.NewPublicKeyPairFromHex(msg.PublicSpendKey, msg.PublicViewKey)
	if err != nil {
		return fmt.Errorf("failed to generate Alice's public keys: %w", err)
	}

	b.setAlicePublicKeys(kp)
	return nil
}

func (b *bob) checkMessageType(msg net.Message) error {
	if msg.Type() != b.nextExpectedMessage.Type() {
		return errors.New("received unexpected message")
	}

	return nil
}
