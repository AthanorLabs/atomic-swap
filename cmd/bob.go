package main

import (
	"errors"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
)

func (n *node) doProtocolBob() error {
	if err := n.host.Start(); err != nil {
		return err
	}
	defer n.host.Stop()

	outCh := make(chan *net.MessageInfo)
	n.host.SetOutgoingCh(outCh)
	n.outCh = outCh
	n.inCh = n.host.ReceivedMessageCh()

	// closed when we have received all the expected network messages, and we
	// can move on to just watching the contract
	setupDone := make(chan struct{})

	for {
		select {
		case <-n.done:
			return nil
		case msg := <-n.inCh:
			if err := n.handleMessageBob(msg.Who, msg.Message, setupDone); err != nil {
				fmt.Printf("failed to handle message: error=%s\n", err)
			}
		case <-setupDone:
			break
		}
	}

	ready, err := n.bob.WatchForReady()
	if err != nil {
		return err
	}

	refund, err := n.bob.WatchForRefund()
	if err != nil {
		return err
	}

	for {
		select {
		case <-n.done:
			return nil
		case <-ready:
			fmt.Println("Alice called Ready!")

			// contract ready, let's claim our ether
		case kp := <-refund:
			fmt.Println("Alice refunded, got monero account key", kp)
		}
	}

	n.wait()
	return nil
}

func (n *node) handleMessageBob(who peer.ID, msg net.Message, setupDone chan struct{}) error {
	switch msg := msg.(type) {
	case *net.WantMessage:
		if msg.Want != "XMR" {
			return errors.New("Bob has XMR, peer does not want XMR")
		}

		fmt.Println("found peer that wants XMR, initiating swap protocol...")
		n.host.SetNextExpectedMessage(&net.SendKeysMessage{})

		sk, vk, err := n.bob.GenerateKeys()
		if err != nil {
			return err
		}

		out := &net.SendKeysMessage{
			PublicSpendKey: sk.Hex(),
			PrivateViewKey: vk.Hex(),
		}

		n.outCh <- &net.MessageInfo{
			Message: out,
			Who:     who,
		}
	case *net.SendKeysMessage:
		if msg.PublicSpendKey == "" || msg.PublicViewKey == "" {
			return errors.New("did not receive Alice's public spend or view key")
		}

		fmt.Println("got Alice's public keys")
		n.host.SetNextExpectedMessage(&net.NotifyContractDeployed{})

		kp, err := monero.NewPublicKeyPairFromHex(msg.PublicSpendKey, msg.PublicViewKey)
		if err != nil {
			return fmt.Errorf("failed to generate Alice's public keys: %w", err)
		}

		n.bob.SetAlicePublicKeys(kp)
	case *net.NotifyContractDeployed:
		if msg.Address == "" {
			return errors.New("got empty contract address")
		}

		fmt.Println("got Swap contract address!")

		if err := n.bob.SetContract(ethcommon.HexToAddress(msg.Address)); err != nil {
			return fmt.Errorf("failed to instantiate contract instance: %w", err)
		}

		close(setupDone)
	default:
		return errors.New("unexpected message type")
	}

	return nil
}
