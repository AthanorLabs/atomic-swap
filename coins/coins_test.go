package coins

import (
	"encoding/json"
	"math"
	"math/big"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPiconeroAmount(t *testing.T) {
	const preciseAmount = "666.666666666666666666" // 18 sixes after the Decimal
	const moneroAmount = "666.666666666667"        // 12 digits after Decimal saved
	const piconeroAmount = "666666666666667"       // 15 digits rounded

	amount := StrToDecimal(preciseAmount)
	piconero := MoneroToPiconero(amount)
	assert.Equal(t, moneroAmount, piconero.AsMonero().String())
	assert.Equal(t, piconeroAmount, piconero.String())
}

func TestMoneroToPiconero(t *testing.T) {
	xrmAmount := StrToDecimal("2")
	const expectedPiconeros = "2000000000000"
	piconeroAmount := MoneroToPiconero(xrmAmount)
	assert.Equal(t, expectedPiconeros, piconeroAmount.String())
}

func TestMoneroToPiconero_roundUp(t *testing.T) {
	//
	// This test is merely demonstrating the current behavior. It is not
	// entirely clear if the ideal behavior is to round-half-up, truncate,
	// or just store fractional piconeros.
	//
	xrmAmount := StrToDecimal("1.0000000000005") // 12 zeros, then "5"
	const expectedPiconeros = "1000000000001"
	piconeroAmount := MoneroToPiconero(xrmAmount)
	assert.Equal(t, expectedPiconeros, piconeroAmount.String())
}

func TestMoneroToPiconero_roundDown(t *testing.T) {
	xrmAmount := StrToDecimal("1.00000000000049") // 12 zeros, then "49"
	const expectedPiconeros = "1000000000000"
	piconeroAmount := MoneroToPiconero(xrmAmount)
	assert.Equal(t, expectedPiconeros, piconeroAmount.String())
}

func TestNewPiconeroAmount(t *testing.T) {
	onePn := NewPiconeroAmount(1)
	oneU64, err := onePn.Uint64()
	require.NoError(t, err)
	assert.Equal(t, oneU64, uint64(1))
}

func TestPiconeroAmount_Uint64(t *testing.T) {
	// MaxUint64 should work
	piconeros := NewPiconeroAmount(math.MaxUint64)
	piconerosU64, err := piconeros.Uint64()
	require.NoError(t, err)
	assert.Equal(t, uint64(math.MaxUint64), piconerosU64)

	// MaxUint64+1 should return an error
	one := apd.New(1, 0)
	_, err = decimalCtx.Add(piconeros.Decimal(), piconeros.Decimal(), one)
	require.NoError(t, err)
	_, err = piconeros.Uint64()
	assert.ErrorContains(t, err, "value out of range")

	// Negative values, which we should never have, return an error
	piconeros.Decimal().Set(apd.New(-1, 0))
	_, err = piconeros.Uint64()
	assert.ErrorContains(t, err, "can not convert")
}

func TestWeiAmount(t *testing.T) {
	amount := StrToDecimal("33.3")
	wei := EtherToWei(amount)
	assert.Equal(t, "33300000000000000000", wei.String())
	assert.Equal(t, "33.3", wei.AsEther().String())
	assert.Equal(t, "33.3", wei.AsStandard().String()) // alias for AsEther

	amountUint := int64(8181)
	WeiAmount := IntToWei(amountUint)
	assert.Equal(t, amountUint, WeiAmount.BigInt().Int64())
}

func TestBigInt2Wei(t *testing.T) {
	bi := big.NewInt(4321)
	wei := NewWeiAmount(bi)
	assert.Equal(t, "4321", wei.String())
}

func TestWeiAmount_BigInt(t *testing.T) {
	amount := StrToDecimal("0.12345678901234567890") // 20 decimal points, after 18 is partial wei
	const expectedWei = "123456789012345679"         // 8 at the end rounded to 9
	wei := EtherToWei(amount)
	require.Equal(t, expectedWei, wei.String())

	// EtherToWei already rounded, reset the internal value so BigInt() needs to round
	wei.Decimal().Set(amount)                           // reset to ether value
	wei.Decimal().Exponent += NumEtherDecimals          // turn the ether into wei
	assert.Equal(t, expectedWei, wei.BigInt().String()) // BigInt() also rounds if needed
}

func TestERC20TokenAmount(t *testing.T) {
	amount := StrToDecimal("33.999999999")
	wei := NewERC20TokenAmountFromDecimals(amount, 9)
	assert.Equal(t, amount.String(), wei.AsStandard().String())

	amount = StrToDecimal("33.000000005")
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	assert.Equal(t, "33.000000005", wei.AsStandard().String())

	amount = StrToDecimal("33.0000000005")
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	assert.Equal(t, "33.000000001", wei.AsStandard().String())

	amount = StrToDecimal("999999999999999999.0000000005")
	wei = NewERC20TokenAmountFromDecimals(amount, 9)
	assert.Equal(t, "999999999999999999.000000001", wei.AsStandard().String())

	amountUint := int64(8181)
	tokenAmt := NewERC20TokenAmount(amountUint, 9)
	assert.Equal(t, amountUint, tokenAmt.BigInt().Int64())
}

func TestNewERC20TokenAmountFromBigInt(t *testing.T) {
	bi := big.NewInt(4321)
	token := NewERC20TokenAmountFromBigInt(bi, 2)
	assert.Equal(t, "4321", token.String())
	assert.Equal(t, "43.21", token.AsStandard().String())
}

func TestNewERC20TokenAmountFromDecimals(t *testing.T) {
	stdAmount := StrToDecimal("0.19")
	token := NewERC20TokenAmountFromDecimals(stdAmount, 1)

	// There's only one decimal place, so this is getting rounded to 2
	// under the current implementation. It's not entirely clear what
	// the ideal behavior is.
	assert.Equal(t, "2", token.String())
	assert.Equal(t, "0.2", token.AsStandard().String())
}

func TestJSONMarshal(t *testing.T) {
	// NOTE: At the current time, ERC20TokenAmount only has private members and
	// is not serializable.
	type TestTypes struct {
		Piconeros *PiconeroAmount `json:"piconeros"`
		Wei       *WeiAmount      `json:"wei"`
		Rate      *ExchangeRate   `json:"rate"`
	}
	tt := &TestTypes{
		Piconeros: NewPiconeroAmount(10),
		Wei:       IntToWei(20),
		Rate:      ToExchangeRate(StrToDecimal("0.4")),
	}
	const expectedJSON = `{
		"piconeros": "10",
		"wei": "20",
		"rate": "0.4"
	}`
	data, err := json.Marshal(tt)
	require.NoError(t, err)
	assert.JSONEq(t, expectedJSON, string(data))

	// Test Unmarshal
	tt2 := new(TestTypes)
	err = json.Unmarshal([]byte(expectedJSON), tt2)
	require.NoError(t, err)
	assert.Zero(t, tt.Piconeros.Cmp(tt2.Piconeros))
	assert.Zero(t, tt.Piconeros.CmpU64(10))
	assert.Zero(t, tt.Wei.BigInt().Cmp(tt2.Wei.BigInt()))
	assert.Equal(t, tt.Rate.String(), tt2.Rate.String())

	// Test Unmarshal missing fields produces nil
	tt = new(TestTypes)
	err = json.Unmarshal([]byte(`{}`), tt)
	require.NoError(t, err)
	assert.Nil(t, tt.Piconeros)
	assert.Nil(t, tt.Wei)
	assert.Nil(t, tt.Rate)

	// Test Unmarshal empty strings produces error
	tt = new(TestTypes)
	err = json.Unmarshal([]byte(`{ "piconeros": "" }`), tt)
	require.Error(t, err)
	err = json.Unmarshal([]byte(`{ "wei": "" }`), tt)
	require.Error(t, err)
	err = json.Unmarshal([]byte(`{ "rate": "" }`), tt)
	require.Error(t, err)

	// Test that Unmarshalling negative values produces an error. Note: In most
	// places we marshal/unmarshal apd.Decimal directly. In those cases, input
	// validation in the receiving method is required to prevent negative values.
	tt = new(TestTypes)
	err = json.Unmarshal([]byte(`{ "piconeros": "-2" }`), tt)
	require.ErrorIs(t, err, errNegativePiconeros)
	err = json.Unmarshal([]byte(`{ "wei": "-3" }`), tt)
	require.ErrorIs(t, err, errNegativeWei)
	err = json.Unmarshal([]byte(`{ "rate": "-0.1" }`), tt)
	require.ErrorIs(t, err, ErrNegativeRate)
}
