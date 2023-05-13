package v2_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/nibiru/x/common/testutil/mock"
	v2 "github.com/NibiruChain/nibiru/x/perp/types/v2"
)

func TestSwapBaseAsset(t *testing.T) {
	tests := []struct {
		name                    string
		baseAssetAmt            sdk.Dec
		dir                     v2.Direction
		expectedQuoteAssetDelta sdk.Dec
		expectedBaseReserve     sdk.Dec
		expectedQuoteReserve    sdk.Dec
		expectedTotalLong       sdk.Dec
		expectedTotalShort      sdk.Dec
		expectedMarkPrice       sdk.Dec
		expectedErr             error
	}{
		{
			name:                    "long base asset",
			baseAssetAmt:            sdk.NewDec(1e11),
			dir:                     v2.Direction_LONG,
			expectedQuoteAssetDelta: sdk.MustNewDecFromStr("111111111111.111111111111111111"),
			expectedBaseReserve:     sdk.NewDec(900000000000),
			expectedQuoteReserve:    sdk.MustNewDecFromStr("1111111111111.111111111111111111"),
			expectedTotalLong:       sdk.NewDec(100000000000),
			expectedTotalShort:      sdk.ZeroDec(),
			expectedMarkPrice:       sdk.MustNewDecFromStr("1.234567901234567901"),
		},
		{
			name:                    "short base asset",
			baseAssetAmt:            sdk.NewDec(1e11),
			dir:                     v2.Direction_SHORT,
			expectedQuoteAssetDelta: sdk.MustNewDecFromStr("90909090909.090909090909090909"),
			expectedBaseReserve:     sdk.NewDec(1100000000000),
			expectedQuoteReserve:    sdk.MustNewDecFromStr("909090909090.909090909090909091"),
			expectedTotalLong:       sdk.ZeroDec(),
			expectedTotalShort:      sdk.NewDec(100000000000),
			expectedMarkPrice:       sdk.MustNewDecFromStr("0.826446280991735537"),
		},
		{
			name:                    "long zero base asset",
			baseAssetAmt:            sdk.ZeroDec(),
			dir:                     v2.Direction_LONG,
			expectedQuoteAssetDelta: sdk.ZeroDec(),
			expectedBaseReserve:     sdk.NewDec(1e12),
			expectedQuoteReserve:    sdk.NewDec(1e12),
			expectedTotalLong:       sdk.ZeroDec(),
			expectedTotalShort:      sdk.ZeroDec(),
			expectedMarkPrice:       sdk.OneDec(),
		},
		{
			name:                    "short zero base asset",
			baseAssetAmt:            sdk.ZeroDec(),
			dir:                     v2.Direction_SHORT,
			expectedQuoteAssetDelta: sdk.ZeroDec(),
			expectedBaseReserve:     sdk.NewDec(1e12),
			expectedQuoteReserve:    sdk.NewDec(1e12),
			expectedTotalLong:       sdk.ZeroDec(),
			expectedTotalShort:      sdk.ZeroDec(),
			expectedMarkPrice:       sdk.OneDec(),
		},
		{
			name:         "not enough base in reserves",
			baseAssetAmt: sdk.NewDec(1e13),
			dir:          v2.Direction_LONG,
			expectedErr:  v2.ErrBaseReserveAtZero,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			amm := mock.TestAMM(sdk.NewDec(1e12), sdk.OneDec())

			quoteAssetDelta, err := amm.SwapBaseAsset(tc.baseAssetAmt, tc.dir)

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedQuoteAssetDelta, quoteAssetDelta)
				assert.Equal(t, v2.AMM{
					Pair:            amm.Pair,
					BaseReserve:     tc.expectedBaseReserve,
					QuoteReserve:    tc.expectedQuoteReserve,
					SqrtDepth:       amm.SqrtDepth,
					PriceMultiplier: amm.PriceMultiplier,
					TotalLong:       tc.expectedTotalLong,
					TotalShort:      tc.expectedTotalShort,
				}, *amm)
				assert.Equal(t, tc.expectedMarkPrice, amm.MarkPrice())
			}
		})
	}
}

func TestSwapQuoteAsset(t *testing.T) {
	tests := []struct {
		name                   string
		quoteAssetAmt          sdk.Dec
		dir                    v2.Direction
		expectedBaseAssetDelta sdk.Dec
		expectedBaseReserve    sdk.Dec
		expectedQuoteReserve   sdk.Dec
		expectedTotalLong      sdk.Dec
		expectedTotalShort     sdk.Dec
		expectedMarkPrice      sdk.Dec
		expectedErr            error
	}{
		{
			name:                   "long quote asset",
			quoteAssetAmt:          sdk.NewDec(1e11),
			dir:                    v2.Direction_LONG,
			expectedBaseAssetDelta: sdk.MustNewDecFromStr("47619047619.047619047619047619"),
			expectedBaseReserve:    sdk.MustNewDecFromStr("952380952380.952380952380952381"),
			expectedQuoteReserve:   sdk.NewDec(1050000000000),
			expectedTotalLong:      sdk.MustNewDecFromStr("47619047619.047619047619047619"),
			expectedTotalShort:     sdk.ZeroDec(),
			expectedMarkPrice:      sdk.MustNewDecFromStr("2.205"),
		},
		{
			name:                   "short base asset",
			quoteAssetAmt:          sdk.NewDec(1e11),
			dir:                    v2.Direction_SHORT,
			expectedBaseAssetDelta: sdk.MustNewDecFromStr("52631578947.368421052631578947"),
			expectedBaseReserve:    sdk.MustNewDecFromStr("1052631578947.368421052631578947"),
			expectedQuoteReserve:   sdk.NewDec(950000000000),
			expectedTotalLong:      sdk.ZeroDec(),
			expectedTotalShort:     sdk.MustNewDecFromStr("52631578947.368421052631578947"),
			expectedMarkPrice:      sdk.MustNewDecFromStr("1.805"),
		},
		{
			name:                   "long zero base asset",
			quoteAssetAmt:          sdk.ZeroDec(),
			dir:                    v2.Direction_LONG,
			expectedBaseAssetDelta: sdk.ZeroDec(),
			expectedBaseReserve:    sdk.NewDec(1e12),
			expectedQuoteReserve:   sdk.NewDec(1e12),
			expectedTotalLong:      sdk.ZeroDec(),
			expectedTotalShort:     sdk.ZeroDec(),
			expectedMarkPrice:      sdk.NewDec(2),
		},
		{
			name:                   "long zero base asset",
			quoteAssetAmt:          sdk.ZeroDec(),
			dir:                    v2.Direction_SHORT,
			expectedBaseAssetDelta: sdk.ZeroDec(),
			expectedBaseReserve:    sdk.NewDec(1e12),
			expectedQuoteReserve:   sdk.NewDec(1e12),
			expectedTotalLong:      sdk.ZeroDec(),
			expectedTotalShort:     sdk.ZeroDec(),
			expectedMarkPrice:      sdk.NewDec(2),
		},
		{
			name:          "not enough base in reserves",
			quoteAssetAmt: sdk.NewDec(1e13),
			dir:           v2.Direction_SHORT,
			expectedErr:   v2.ErrQuoteReserveAtZero,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			amm := mock.TestAMM(sdk.NewDec(1e12), sdk.NewDec(2))

			quoteAssetDelta, err := amm.SwapQuoteAsset(tc.quoteAssetAmt, tc.dir)

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedBaseAssetDelta, quoteAssetDelta)
				assert.Equal(t, v2.AMM{
					Pair:            amm.Pair,
					BaseReserve:     tc.expectedBaseReserve,
					QuoteReserve:    tc.expectedQuoteReserve,
					SqrtDepth:       amm.SqrtDepth,
					PriceMultiplier: amm.PriceMultiplier,
					TotalLong:       tc.expectedTotalLong,
					TotalShort:      tc.expectedTotalShort,
				}, *amm)
				assert.Equal(t, tc.expectedMarkPrice, amm.MarkPrice())
			}
		})
	}
}

// baseReserves := base reserves if no one is trading
// bias := totalLong (bias) + totalShort (bias) := the net size of all positions together
// In the test cases you see,
// one is repegging bias of +25 with cost of 20,
// and the other has bias -20 with cost of -25
// The reason for this is that swapping in different directions actually results in different amounts.
// Here's the case named "new peg -> simple math":
// Given:
// y = 100, x = 100, bias = 25, peg = 1
// Do Repeg(peg=2)
// To get rid of the bias, we swap it away and see what that is in quote units:
// dy = k / (x + dx)  - y, where dx = bias
// dy = 100^2 / (100 + 25) - 100  = -20
// Here's the case named "new peg -> simple math but negative bias":
// Given:
// y = 100, x =100, bias = -20, peg=1
// Do Repeg(peg=2)
// To get rid of the bias, we swap it away and see what that is in quote units:
// dy = k / (x + dx)  - y, where dx = bias
// dy = 100^2 / (100 - 20) - 100  = +25
func TestRepegCost(t *testing.T) {
	pair := asset.Registry.Pair(denoms.BTC, denoms.NUSD)
	tests := []struct {
		name string

		amm                v2.AMM
		newPriceMultiplier sdk.Dec

		expectedCost sdk.Dec
		shouldErr    bool
	}{
		{
			name: "zero bias -> zero cost",
			amm: v2.AMM{
				Pair:            pair,
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.ZeroDec(),
			},
			newPriceMultiplier: sdk.NewDec(3),
			expectedCost:       sdk.ZeroDec(),
			shouldErr:          false,
		},
		{
			name: "same peg -> zero cost",
			amm: v2.AMM{
				Pair:            pair,
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.ZeroDec(),
			},
			newPriceMultiplier: sdk.OneDec(),
			expectedCost:       sdk.ZeroDec(),
			shouldErr:          false,
		},
		{
			name: "new peg -> simple math",
			amm: v2.AMM{
				Pair:            pair,
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.NewDec(25),
				TotalShort:      sdk.ZeroDec(),
			},
			newPriceMultiplier: sdk.NewDec(2),
			expectedCost:       sdk.NewDec(20), // 20 * (2 - 1)
			shouldErr:          false,
		},
		{
			name: "new peg -> simple math but negative bias",
			amm: v2.AMM{
				Pair:            pair,
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.NewDec(20),
			},
			newPriceMultiplier: sdk.NewDec(2),
			expectedCost:       sdk.NewDec(-25), // -20 * (2 - 1)
			shouldErr:          false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cost, err := tc.amm.GetRepegCost(tc.newPriceMultiplier)
			if tc.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.expectedCost, cost)
			}
		})
	}
}

func TestGetMarketTotalQuoteReserves(t *testing.T) {
	tests := []struct {
		name                 string
		amm                  v2.AMM
		expectedQuoteReserve sdk.Dec
	}{
		{
			name: "zero reserves",
			amm: v2.AMM{
				BaseReserve:     sdk.ZeroDec(),
				QuoteReserve:    sdk.ZeroDec(),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.ZeroDec(),
				PriceMultiplier: sdk.OneDec(),
			},
			expectedQuoteReserve: sdk.ZeroDec(),
		},
		{
			name: "zero bias",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(1e12),
				QuoteReserve:    sdk.NewDec(1e12),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.ZeroDec(),
				PriceMultiplier: sdk.OneDec(),
			},
			expectedQuoteReserve: sdk.ZeroDec(),
		},
		{
			name: "long only bias",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(1e12),
				QuoteReserve:    sdk.NewDec(1e12),
				TotalLong:       sdk.OneDec(),
				TotalShort:      sdk.ZeroDec(),
				PriceMultiplier: sdk.OneDec(),
			},
			expectedQuoteReserve: sdk.MustNewDecFromStr("0.999999999999"),
		},
		{
			name: "short only bias",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(1e12),
				QuoteReserve:    sdk.NewDec(1e12),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.OneDec(),
				PriceMultiplier: sdk.OneDec(),
			},
			expectedQuoteReserve: sdk.MustNewDecFromStr("1.000000000001"),
		},
		{
			name: "long and short bias",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(1e12),
				QuoteReserve:    sdk.NewDec(1e12),
				TotalLong:       sdk.NewDec(1234),
				TotalShort:      sdk.NewDec(4321),
				PriceMultiplier: sdk.OneDec(),
			},
			expectedQuoteReserve: sdk.MustNewDecFromStr("5555.000017148285082557"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			quoteReserves, err := tc.amm.GetMarketTotalQuoteReserves()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedQuoteReserve, quoteReserves)
		})
	}
}

func TestUpdateSwapInvariant(t *testing.T) {
	tests := []struct {
		name       string
		amm        v2.AMM
		multiplier sdk.Dec

		expectedBaseReserve  sdk.Dec
		expectedQuoteReserve sdk.Dec
		expectedSqrtDepth    sdk.Dec
	}{
		{
			name: "one multiplier",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				SqrtDepth:       sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
			},
			multiplier:           sdk.OneDec(),
			expectedBaseReserve:  sdk.NewDec(100),
			expectedQuoteReserve: sdk.NewDec(100),
			expectedSqrtDepth:    sdk.NewDec(100),
		},
		{
			name: "four multiplier",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				SqrtDepth:       sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
			},
			multiplier:           sdk.NewDec(4),
			expectedBaseReserve:  sdk.NewDec(200),
			expectedQuoteReserve: sdk.NewDec(200),
			expectedSqrtDepth:    sdk.NewDec(200),
		},
		{
			name: "quarter multiplier",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				SqrtDepth:       sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
			},
			multiplier:           sdk.MustNewDecFromStr("0.25"),
			expectedBaseReserve:  sdk.NewDec(50),
			expectedQuoteReserve: sdk.NewDec(50),
			expectedSqrtDepth:    sdk.NewDec(50),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.amm.UpdateSwapInvariant(tc.multiplier)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedBaseReserve, tc.amm.BaseReserve)
			assert.Equal(t, tc.expectedQuoteReserve, tc.amm.QuoteReserve)
			assert.Equal(t, tc.expectedSqrtDepth, tc.amm.SqrtDepth)
		})
	}
}

func TestGetSwapInvariantUpdateCost(t *testing.T) {
	tests := []struct {
		name       string
		amm        v2.AMM
		multiplier sdk.Dec

		expectedCost sdk.Dec
	}{
		{
			name: "zero cost - due to single multiplier",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				SqrtDepth:       sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.NewDec(50),
				TotalShort:      sdk.NewDec(50),
			},
			multiplier:   sdk.OneDec(),
			expectedCost: sdk.ZeroDec(),
		},

		{
			name: "zero cost - due to zero bias",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				SqrtDepth:       sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.ZeroDec(),
			},
			multiplier:   sdk.NewDec(100),
			expectedCost: sdk.ZeroDec(),
		},

		{
			name: "long bias, increase swap invariant",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				SqrtDepth:       sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.NewDec(10),
				TotalShort:      sdk.ZeroDec(),
			},
			multiplier:   sdk.NewDec(100),
			expectedCost: sdk.MustNewDecFromStr("0.810081008100810081"),
		},

		{
			name: "long bias, decrease swap invariant",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(1000),
				QuoteReserve:    sdk.NewDec(1000),
				SqrtDepth:       sdk.NewDec(1000),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.NewDec(10),
				TotalShort:      sdk.ZeroDec(),
			},
			multiplier:   sdk.MustNewDecFromStr("0.01"),
			expectedCost: sdk.MustNewDecFromStr("-0.810081008100810081"),
		},

		{
			name: "short bias, increase swap invariant",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(100),
				QuoteReserve:    sdk.NewDec(100),
				SqrtDepth:       sdk.NewDec(100),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.NewDec(10),
			},
			multiplier:   sdk.NewDec(100),
			expectedCost: sdk.MustNewDecFromStr("1.010101010101010101"),
		},

		{
			name: "short bias, decrease swap invariant",
			amm: v2.AMM{
				BaseReserve:     sdk.NewDec(1000),
				QuoteReserve:    sdk.NewDec(1000),
				SqrtDepth:       sdk.NewDec(1000),
				PriceMultiplier: sdk.OneDec(),
				TotalLong:       sdk.ZeroDec(),
				TotalShort:      sdk.NewDec(10),
			},
			multiplier:   sdk.MustNewDecFromStr("0.01"),
			expectedCost: sdk.MustNewDecFromStr("-1.010101010101010101"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cost, err := tc.amm.GetSwapInvariantUpdateCost(tc.multiplier)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCost, cost)
		})
	}
}
