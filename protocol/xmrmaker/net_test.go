package xmrmaker

import (
	"testing"

	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net/message"

	"github.com/stretchr/testify/require"
)

func TestXMRMaker_HandleInitiateMessage(t *testing.T) {
	b := newTestXMRMaker(t)

	offer := types.NewOffer(types.ProvidesXMR, 0.001, 0.002, 0.1)
	_, err := b.MakeOffer(offer)
	require.NoError(t, err)

	msg, _ := newTestXMRTakerSendKeysMessage(t)
	msg.OfferID = offer.GetID().String()
	msg.ProvidedAmount = offer.MinimumAmount * float64(offer.ExchangeRate)

	_, resp, err := b.HandleInitiateMessage(msg)
	require.NoError(t, err)
	require.Equal(t, message.SendKeysType, resp.Type())
	require.NotNil(t, b.swapStates[offer.GetID()])
}
