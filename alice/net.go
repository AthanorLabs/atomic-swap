package alice

import (
	"errors"
	"fmt"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
)

func (a *alice) Provides() net.ProvidesCoin {
	return net.ProvidesETH
}

func (a *alice) SendKeysMessage() (*net.SendKeysMessage, error) {
	kp, err := a.GenerateKeys()
	if err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		PublicSpendKey: kp.SpendKey().Hex(),
		PublicViewKey:  kp.ViewKey().Hex(),
	}, nil
}

func (a *alice) InitiateProtocol(providesAmount, desiredAmount uint64) error {
	if a.initiated {
		return errors.New("protocol already in progress")
	}

	// TODO: check user's balance and that they actualy have what they will provide
	a.initiated = true
	a.providesAmount = providesAmount
	a.desiredAmount = desiredAmount
	return nil
}

func (a *alice) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	switch msg := msg.(type) {
	// case *net.HelloMessage:
	// 	peerProvides := false
	// 	for _, provides := range msg.Provides {
	// 		if provides == net.ProvidesXMR {
	// 			peerProvides = true
	// 			break
	// 		}
	// 	}

	// 	if !peerProvides {
	// 		return nil, true, errors.New("peer does not provide XMR")
	// 	}

	// 	log.Info("found peer that wants ETH, initiating swap protocol...")
	// 	a.setNextExpectedMessage(&net.SendKeysMessage{})

	// 	kp, err := a.GenerateKeys()
	// 	if err != nil {
	// 		return nil, true, err
	// 	}

	// 	out := &net.SendKeysMessage{
	// 		PublicSpendKey: kp.SpendKey().Hex(),
	// 		PublicViewKey:  kp.ViewKey().Hex(),
	// 	}

	// 	return out, false, nil
	case *net.InitiateMessage:
		if msg.PublicSpendKey == "" || msg.PrivateViewKey == "" {
			return nil, true, errors.New("did not receive Bob's public spend or private view key")
		}

		log.Debug("got Bob's keys")
		a.setNextExpectedMessage(&net.NotifyXMRLock{})

		sk, err := monero.NewPublicKeyFromHex(msg.PublicSpendKey)
		if err != nil {
			return nil, true, fmt.Errorf("failed to generate Bob's public spend key: %w", err)
		}

		vk, err := monero.NewPrivateViewKeyFromHex(msg.PrivateViewKey)
		if err != nil {
			return nil, true, fmt.Errorf("failed to generate Bob's private view keys: %w", err)
		}

		a.SetBobKeys(sk, vk)
		address, err := a.DeployAndLockETH(a.providesAmount)
		if err != nil {
			return nil, true, fmt.Errorf("failed to deploy contract: %w", err)
		}

		log.Info("deployed Swap contract: address=", address)

		claim, err := a.WatchForClaim()
		if err != nil {
			return nil, true, err
		}

		go func() {
			for {
				// TODO: add t1 timeout case
				select {
				case <-a.ctx.Done():
					return
				case kp := <-claim:
					if kp == nil {
						continue
					}

					log.Info("Bob claimed ether! got secret: ", kp)
					address, err := a.CreateMoneroWallet(kp)
					if err != nil {
						log.Debug("failed to create monero address: %s", err)
						return
					}

					log.Info("successfully created monero wallet from our secrets: address=", address)
					// TODO: get and print balance
				}
			}
		}()

		out := &net.NotifyContractDeployed{
			Address: address.String(),
		}

		return out, false, nil
	case *net.NotifyXMRLock:
		if msg.Address == "" {
			return nil, true, errors.New("got empty address for locked XMR")
		}

		// check that XMR was locked in expected account, and confirm amount
		a.setNextExpectedMessage(&net.NotifyClaimed{})

		if err := a.Ready(); err != nil {
			return nil, true, fmt.Errorf("failed to call Ready: %w", err)
		}

		log.Debug("called set swap.IsReady == true")

		out := &net.NotifyReady{}
		return out, false, nil
	case *net.NotifyClaimed:
		address, err := a.NotifyClaimed(msg.TxHash)
		if err != nil {
			log.Error("failed to create monero address: err=", err)
			return nil, true, err
		}

		log.Info("successfully created monero wallet from our secrets: address=", address)
		// TODO: get and print balance

		return nil, true, nil
	default:
		return nil, false, errors.New("unexpected message type")
	}

	return nil, false, nil
}
