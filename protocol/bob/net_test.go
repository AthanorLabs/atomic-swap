package bob

import (
	"testing"

	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net/message"

	"github.com/stretchr/testify/require"
)

func TestBob_HandleInitiateMessage(t *testing.T) {
	b := newTestBob(t)

	offer := &types.Offer{
		Provides:      types.ProvidesXMR,
		MinimumAmount: 0.1,
		MaximumAmount: 0.2,
		ExchangeRate:  0.1,
	}
	b.MakeOffer(offer)
	msg, _ := newTestAliceSendKeysMessage(t)
	msg.OfferID = offer.GetID().String()
	msg.ProvidedAmount = offer.MinimumAmount * float64(offer.ExchangeRate)

	_, resp, err := b.HandleInitiateMessage(msg)
	require.NoError(t, err)
	require.Equal(t, message.SendKeysType, resp.Type())
	require.NotNil(t, b.swapState)
}
