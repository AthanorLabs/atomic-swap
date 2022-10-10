package xmrmaker

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"

	"github.com/stretchr/testify/require"
)

func TestXMRMaker_HandleInitiateMessage(t *testing.T) {
	b, db := newTestXMRMakerAndDB(t)

	offer := types.NewOffer(types.ProvidesXMR, 0.001, 0.002, 0.1, types.EthAssetETH)
	db.EXPECT().PutOffer(offer)
	db.EXPECT().DeleteOffer(offer.GetID())

	b.net.(*MockHost).EXPECT().Advertise()

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
