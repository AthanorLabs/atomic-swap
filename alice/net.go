package alice

import (
	"errors"
	"fmt"
	"time"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
)

func (a *alice) Provides() net.ProvidesCoin {
	return net.ProvidesETH
}

func (a *alice) SendKeysMessage() (*net.SendKeysMessage, error) {
	kp, err := a.generateKeys()
	if err != nil {
		return nil, err
	}

	return &net.SendKeysMessage{
		PublicSpendKey: kp.SpendKey().Hex(),
		PublicViewKey:  kp.ViewKey().Hex(),
	}, nil
}

func (a *alice) InitiateProtocol(providesAmount, desiredAmount uint64) error {
	if err := a.initiate(providesAmount, desiredAmount); err != nil {
		return err
	}

	a.setNextExpectedMessage(&net.SendKeysMessage{})
	return nil
}

func (a *alice) initiate(providesAmount, desiredAmount uint64) error {
	if a.initiated {
		return errors.New("protocol already in progress")
	}

	balance, err := a.ethClient.BalanceAt(a.ctx, a.auth.From, nil)
	if err != nil {
		return err
	}

	// check user's balance and that they actualy have what they will provide
	if balance.Uint64() <= a.providesAmount {
		return errors.New("balance lower than amount to be provided")
	}

	a.initiated = true
	a.providesAmount = providesAmount
	a.desiredAmount = desiredAmount
	return nil
}

func (a *alice) ProtocolComplete() {
	a.initiated = false
	a.setNextExpectedMessage(&net.InitiateMessage{})
}

func (a *alice) HandleProtocolMessage(msg net.Message) (net.Message, bool, error) {
	if err := a.checkMessageType(msg); err != nil {
		return nil, true, err
	}

	xmrLockedCh := make(chan struct{})

	switch msg := msg.(type) {
	case *net.InitiateMessage:
		if msg.Provides != net.ProvidesXMR {
			return nil, true, errors.New("peer does not provide XMR")
		}

		// TODO: notify the user via the CLI/websockets that someone wishes to initiate a swap with them.

		// the other party initiated, saying what they will provide and what they desire.
		// we initiate our protocol, saying we will provide what they desire and vice versa.
		if err := a.initiate(msg.DesiredAmount, msg.ProvidesAmount); err != nil {
			return nil, true, err
		}

		resp, err := a.handleSendKeysMessage(msg.SendKeysMessage, xmrLockedCh)
		if err != nil {
			return nil, true, err
		}

		return resp, false, nil
	case *net.SendKeysMessage:
		resp, err := a.handleSendKeysMessage(msg, xmrLockedCh)
		if err != nil {
			return nil, true, err
		}

		return resp, false, nil
	case *net.NotifyXMRLock:
		if msg.Address == "" {
			return nil, true, errors.New("got empty address for locked XMR")
		}

		// TODO: check that XMR was locked in expected account, and confirm amount
		a.setNextExpectedMessage(&net.NotifyClaimed{})
		close(xmrLockedCh)

		if err := a.ready(); err != nil {
			return nil, true, fmt.Errorf("failed to call Ready: %w", err)
		}

		log.Debug("set swap.IsReady == true")

		out := &net.NotifyReady{}
		return out, false, nil
	case *net.NotifyClaimed:
		address, err := a.handleNotifyClaimed(msg.TxHash)
		if err != nil {
			log.Error("failed to create monero address: err=", err)
			return nil, true, err
		}

		log.Info("successfully created monero wallet from our secrets: address=", address)
		return nil, true, nil
	default:
		return nil, false, errors.New("unexpected message type")
	}
}

func (a *alice) handleSendKeysMessage(msg *net.SendKeysMessage, xmrLockedCh <-chan struct{}) (net.Message, error) {
	if msg.PublicSpendKey == "" || msg.PrivateViewKey == "" {
		return nil, errors.New("did not receive Bob's public spend or private view key")
	}

	log.Debug("got Bob's keys")
	a.setNextExpectedMessage(&net.NotifyXMRLock{})

	sk, err := monero.NewPublicKeyFromHex(msg.PublicSpendKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Bob's public spend key: %w", err)
	}

	vk, err := monero.NewPrivateViewKeyFromHex(msg.PrivateViewKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Bob's private view keys: %w", err)
	}

	a.setBobKeys(sk, vk)
	address, err := a.deployAndLockETH(a.providesAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %w", err)
	}

	log.Info("deployed Swap contract, waiting for XMR to be locked: address=", address)

	// start goroutine to check that Bob locks before t_0
	go func() {
		const timeoutBuffer = time.Minute * 5

		st0, err := a.contract.Timeout0(a.callOpts)
		if err != nil {
			log.Errorf("failed to get timeout0 from contract: err=%s", err)
			return
		}

		t0 := time.Unix(st0.Int64(), 0)
		until := time.Until(t0)

		select {
		case <-a.ctx.Done():
			return
		case <-time.After(until - timeoutBuffer):
			// Bob hasn't locked yet, let's call refund
		case <-xmrLockedCh: 
			return
		}

	}()

	out := &net.NotifyContractDeployed{
		Address: address.String(),
	}

	return out, nil
}

func (a *alice) checkMessageType(msg net.Message) error {
	if msg.Type() != a.nextExpectedMessage.Type() {
		return errors.New("received unexpected message")
	}

	return nil
}
