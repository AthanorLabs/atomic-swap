package xmrmaker

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
)

func TestXMRMaker_HandleInitiateMessage(t *testing.T) {
	b, db := newTestInstanceAndDB(t)
	min := coins.StrToDecimal("0.001")
	max := coins.StrToDecimal("0.002")
	rate := coins.ToExchangeRate(coins.StrToDecimal("0.1"))
	offer := types.NewOffer(coins.ProvidesXMR, min, max, rate, types.EthAssetETH)
	db.EXPECT().PutOffer(offer)
	db.EXPECT().DeleteOffer(offer.ID)

	b.net.(*MockP2pHost).EXPECT().Advertise([]string{"XMR"})

	_, err := b.MakeOffer(offer, nil)
	require.NoError(t, err)

	msg, _ := newTestXMRTakerSendKeysMessage(t)
	msg.OfferID = offer.ID
	msg.ProvidedAmount, err = offer.ExchangeRate.ToETH(offer.MinAmount)
	require.NoError(t, err)

	_, resp, err := b.HandleInitiateMessage(msg)
	require.NoError(t, err)
	require.Equal(t, message.SendKeysType, resp.Type())
	require.NotNil(t, b.swapStates[offer.ID])
}
