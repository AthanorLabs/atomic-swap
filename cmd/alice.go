package main

import (
	"errors"
	"fmt"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"

	"github.com/libp2p/go-libp2p-core/peer"
)

func (n *node) doProtocolAlice() error {
	if err := n.host.Start(); err != nil {
		return err
	}
	defer n.host.Stop()

	outCh := make(chan *net.MessageInfo)
	n.host.SetOutgoingCh(outCh)
	n.outCh = outCh
	n.inCh = n.host.ReceivedMessageCh()

	for {
		select {
		case <-n.done:
		case msg := <-n.inCh:
			if err := n.handleMessageAlice(msg.Who, msg.Message); err != nil {
				fmt.Printf("failed to handle message: error=%s\n", err)
			}
		}
	}

	n.wait()
	return nil
}

func (n *node) handleMessageAlice(who peer.ID, msg net.Message) error {
	switch msg := msg.(type) {
	case *net.WantMessage:
		if msg.Want != "ETH" {
			return errors.New("Alice has ETH, peer does not want ETH")
		}

		fmt.Println("found peer that wants ETH, initiating swap protocol...")
		n.host.SetNextExpectedMessage(&net.SendKeysMessage{})

		kp, err := n.alice.GenerateKeys()
		if err != nil {
			return err
		}

		out := &net.SendKeysMessage{
			PublicSpendKey: kp.SpendKey().Hex(),
			PublicViewKey:  kp.ViewKey().Hex(),
		}

		n.outCh <- &net.MessageInfo{
			Message: out,
			Who:     who,
		}
	case *net.SendKeysMessage:
		if msg.PublicSpendKey == "" || msg.PrivateViewKey == "" {
			return errors.New("did not receive Bob's public spend or private view key")
		}

		fmt.Println("got Bob's keys")
		n.host.SetNextExpectedMessage(&net.NotifyXMRLock{})

		sk, err := monero.NewPublicKeyFromHex(msg.PublicSpendKey)
		if err != nil {
			return fmt.Errorf("failed to generate Bob's public spend key: %w", err)
		}

		vk, err := monero.NewPrivateViewKeyFromHex(msg.PrivateViewKey)
		if err != nil {
			return fmt.Errorf("failed to generate Bob's private view keys: %w", err)
		}

		n.alice.SetBobKeys(sk, vk)
		address, err := n.alice.DeployAndLockETH(n.amount)
		if err != nil {
			return fmt.Errorf("failed to deploy contract: %w", err)
		}

		fmt.Printf("deployed Swap contract: address=%s\n", address)
	case *net.NotifyXMRLock:
		if msg.Address == "" {
			return errors.New("got empty address for locked XMR")
		}
	default:
		return errors.New("unexpected message type")
	}

	return nil
}
