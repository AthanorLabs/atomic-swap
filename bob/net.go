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
	sk, vk, err := b.GenerateKeys()
	if err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		PublicSpendKey: sk.Hex(),
		PrivateViewKey: vk.Hex(),
	}, nil
}

func (b *bob) InitiateProtocol(providesAmount, desiredAmount uint64) error {
	if b.initiated {
		return errors.New("protocol already in progress")
	}

	// TODO: check user's balance and that they actualy have what they will provide
	b.initiated = true
	b.providesAmount = providesAmount
	b.desiredAmount = desiredAmount
	b.setNextExpectedMessage(&net.SendKeysMessage{})
	return nil
}

func (b *bob) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	switch msg := msg.(type) {
	// case *net.HelloMessage:
	// 	peerProvides := false
	// 	for _, provides := range msg.Provides {
	// 		if provides == net.ProvidesETH {
	// 			peerProvides = true
	// 			break
	// 		}
	// 	}

	// 	if !peerProvides {
	// 		return errors.New("peer does not provide ETH")
	// 	}

	// 	log.Debug("found peer that wants XMR, initiating swap protocol...")
	// 	n.host.SetNextExpectedMessage(&net.SendKeysMessage{})

	// 	sk, vk, err := n.bob.GenerateKeys()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	out := &net.SendKeysMessage{
	// 		PublicSpendKey: sk.Hex(),
	// 		PrivateViewKey: vk.Hex(),
	// 	}

	// 	n.outCh <- &net.MessageInfo{
	// 		Message: out,
	// 		Who:     who,
	// 	}
	case *net.InitiateMessage:
		// TODO: this and the below case are the same
		if msg.PublicSpendKey == "" || msg.PublicViewKey == "" {
			return nil, true, errors.New("did not receive Alice's public spend or view key")
		}

		log.Debug("got Alice's public keys")
		b.setNextExpectedMessage(&net.NotifyContractDeployed{})

		kp, err := monero.NewPublicKeyPairFromHex(msg.PublicSpendKey, msg.PublicViewKey)
		if err != nil {
			return nil, true, fmt.Errorf("failed to generate Alice's public keys: %w", err)
		}

		b.SetAlicePublicKeys(kp)

		resp, err := b.SendKeysMessage()
		if err != nil {
			return nil, true, err
		}

		// TODO: check amounts are ok, less than our max, and desired coin
		b.providesAmount = msg.DesiredAmount
		b.desiredAmount = msg.ProvidesAmount

		return resp, false, nil
	case *net.SendKeysMessage:
		if msg.PublicSpendKey == "" || msg.PublicViewKey == "" {
			return nil, true, errors.New("did not receive Alice's public spend or view key")
		}

		log.Debug("got Alice's public keys")
		b.setNextExpectedMessage(&net.NotifyContractDeployed{})

		kp, err := monero.NewPublicKeyPairFromHex(msg.PublicSpendKey, msg.PublicViewKey)
		if err != nil {
			return nil, true, fmt.Errorf("failed to generate Alice's public keys: %w", err)
		}

		b.SetAlicePublicKeys(kp)
	case *net.NotifyContractDeployed:
		if msg.Address == "" {
			return nil, true, errors.New("got empty contract address")
		}

		b.setNextExpectedMessage(&net.NotifyReady{})
		log.Info("got Swap contract address! address=%s\n", msg.Address)

		if err := b.SetContract(ethcommon.HexToAddress(msg.Address)); err != nil {
			return nil, true, fmt.Errorf("failed to instantiate contract instance: %w", err)
		}

		// ready, err := b.WatchForReady()
		// if err != nil {
		// 	return nil, true, err
		// }

		refund, err := b.WatchForRefund()
		if err != nil {
			return nil, true, err
		}

		go func() {
			for {
				// TODO: add t0 timeout case
				select {
				case <-b.ctx.Done():
					return
				// case <-ready:
				// 	time.Sleep(time.Second * 3)
				// 	log.Debug("Alice called Ready!")
				// 	log.Debug("attempting to claim funds...")

				// 	time.Sleep(time.Second)

				// 	// contract ready, let's claim our ether
				// 	_, err = b.ClaimFunds()
				// 	if err != nil {
				// 		log.Error("failed to redeem ether: %w", err)
				// 		continue
				// 	}

				// 	log.Debug("funds claimed!!")
				// 	// out := &net.NotifyClaimed{
				// 	// 	TxHash: txHash,
				// 	// }

				// 	//time.Sleep(time.Second)
				// 	return
				// TODO: fix events, or just use messages for now
				case kp := <-refund:
					if kp == nil {
						continue
					}

					log.Debug("Alice refunded, got monero account key", kp)
					// TODO: generate wallet

					time.Sleep(time.Second)
					return
				}
			}
		}()

		addrAB, err := b.LockFunds(b.providesAmount)
		if err != nil {
			return nil, true, fmt.Errorf("failed to lock funds: %w", err)
		}

		out := &net.NotifyXMRLock{
			Address: string(addrAB),
		}

		return out, false, nil
	case *net.NotifyReady:
		//	time.Sleep(time.Second * 3)
		log.Debug("Alice called Ready!")
		log.Debug("attempting to claim funds...")

		time.Sleep(time.Second)

		// contract ready, let's claim our ether
		txHash, err := b.ClaimFunds()
		if err != nil {
			return nil, true, fmt.Errorf("failed to redeem ether: %w", err)
		}

		log.Debug("funds claimed!!")
		out := &net.NotifyClaimed{
			TxHash: txHash,
		}

		return out, true, nil
	default:
		return nil, true, errors.New("unexpected message type")
	}

	return nil, false, nil
}
