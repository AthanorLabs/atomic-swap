package xmrmaker

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func TestXMRMaker_HandleInitiateMessage(t *testing.T) {
	b, db := newTestInstanceAndDB(t)
	min := tests.Str2Decimal("0.001")
	max := tests.Str2Decimal("0.002")
	rate := types.ToExchangeRate(tests.Str2Decimal("0.1"))
	offer := types.NewOffer(types.ProvidesXMR, min, max, rate, types.EthAssetETH)
	db.EXPECT().PutOffer(offer)
	db.EXPECT().DeleteOffer(offer.ID)

	b.net.(*MockHost).EXPECT().Advertise()

	_, err := b.MakeOffer(offer, "", nil)
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
