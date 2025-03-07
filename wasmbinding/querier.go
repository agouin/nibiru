package wasmbinding

import (
	"encoding/json"
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/NibiruChain/nibiru/wasmbinding/bindings"
	"github.com/NibiruChain/nibiru/x/common/asset"
	oraclekeeper "github.com/NibiruChain/nibiru/x/oracle/keeper"
	oracletypes "github.com/NibiruChain/nibiru/x/oracle/types"
	perpv2keeper "github.com/NibiruChain/nibiru/x/perp/v2/keeper"
	perpv2types "github.com/NibiruChain/nibiru/x/perp/v2/types"
)

type QueryPlugin struct {
	Perp   *PerpQuerier
	Oracle *OracleQuerier
}

// NewQueryPlugin returns a pointer to a new QueryPlugin
func NewQueryPlugin(perp perpv2keeper.Keeper, oracle oraclekeeper.Keeper) QueryPlugin {
	return QueryPlugin{
		Perp: &PerpQuerier{
			perp: perpv2keeper.NewQuerier(perp),
		},
		Oracle: &OracleQuerier{
			oracle: oraclekeeper.NewQuerier(oracle),
		},
	}
}

func (qp *QueryPlugin) ToBinary(
	cwResp any, err error, cwReq any,
) ([]byte, error) {
	if err != nil {
		return nil, sdkerrors.Wrapf(err,
			"failed to query: perp all markets: request: %v",
			cwReq)
	}
	bz, err := json.Marshal(cwResp)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to JSON marshal response: %v", cwResp)
	}
	return bz, nil
}

// CustomQuerier returns a function that is an implementation of the custom
// querier mechanism for specific messages
func CustomQuerier(qp QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var wasmContractQuery bindings.BindingQuery
		if err := json.Unmarshal(request, &wasmContractQuery); err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to JSON unmarshal nibiru query: %v", err)
		}

		switch {
		case wasmContractQuery.AllMarkets != nil:
			cwReq := wasmContractQuery.AllMarkets
			cwResp, err := qp.Perp.AllMarkets(ctx)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.Reserves != nil:
			cwReq := wasmContractQuery.Reserves
			cwResp, err := qp.Perp.Reserves(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.BasePrice != nil:
			cwReq := wasmContractQuery.BasePrice
			cwResp, err := qp.Perp.BasePrice(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.Positions != nil:
			cwReq := wasmContractQuery.Positions
			cwResp, err := qp.Perp.Positions(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.Position != nil:
			cwReq := wasmContractQuery.Position
			cwResp, err := qp.Perp.Position(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.PremiumFraction != nil:
			cwReq := wasmContractQuery.PremiumFraction
			cwResp, err := qp.Perp.PremiumFraction(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.Metrics != nil:
			cwReq := wasmContractQuery.Metrics
			cwResp, err := qp.Perp.Metrics(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.ModuleAccounts != nil:
			cwReq := wasmContractQuery.ModuleAccounts
			cwResp, err := qp.Perp.ModuleAccounts(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.PerpParams != nil:
			cwReq := wasmContractQuery.PerpParams
			cwResp, err := qp.Perp.ModuleParams(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		case wasmContractQuery.OraclePrices != nil:
			cwReq := wasmContractQuery.OraclePrices
			cwResp, err := qp.Oracle.ExchangeRates(ctx, cwReq)
			return qp.ToBinary(cwResp, err, cwReq)

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown nibiru query variant"}
		}
	}
}

// ----------------------------------------------------------------------
// PerpQuerier
// ----------------------------------------------------------------------

type PerpQuerier struct {
	perp perpv2types.QueryServer
}

func (perpExt *PerpQuerier) Reserves(
	ctx sdk.Context, cwReq *bindings.ReservesRequest,
) (*bindings.ReservesResponse, error) {
	pair := asset.Pair(cwReq.Pair)
	sdkReq := &perpv2types.QueryMarketsRequest{}
	goCtx := sdk.WrapSDKContext(ctx)
	sdkResp, err := perpExt.perp.QueryMarkets(goCtx, sdkReq)
	if err != nil {
		return nil, err
	}

	for _, market := range sdkResp.AmmMarkets {
		if market.Amm.Pair.Equal(pair) {
			return &bindings.ReservesResponse{
				Pair:         pair.String(),
				BaseReserve:  market.Amm.BaseReserve,
				QuoteReserve: market.Amm.QuoteReserve,
			}, err
		}
	}

	return nil, fmt.Errorf("market not found for pair %s", pair)
}

func (perpExt *PerpQuerier) AllMarkets(
	ctx sdk.Context,
) (*bindings.AllMarketsResponse, error) {
	sdkReq := &perpv2types.QueryMarketsRequest{}
	goCtx := sdk.WrapSDKContext(ctx)
	sdkResp, err := perpExt.perp.QueryMarkets(goCtx, sdkReq)
	if err != nil {
		return nil, err
	}

	marketMap := make(map[string]bindings.Market)
	for _, pbMarket := range sdkResp.AmmMarkets {
		key := pbMarket.Amm.Pair.String()
		marketMap[key] = bindings.Market{
			Pair:         key,
			Version:      sdk.NewIntFromUint64(pbMarket.Market.Version),
			BaseReserve:  pbMarket.Amm.BaseReserve,
			QuoteReserve: pbMarket.Amm.QuoteReserve,
			SqrtDepth:    pbMarket.Amm.SqrtDepth,
			TotalLong:    pbMarket.Amm.TotalLong,
			TotalShort:   pbMarket.Amm.TotalShort,
			PegMult:      pbMarket.Amm.PriceMultiplier,
			Config: &bindings.MarketConfig{
				MaintenanceMarginRatio: pbMarket.Market.MaintenanceMarginRatio,
				MaxLeverage:            pbMarket.Market.MaxLeverage,
			},
			MarkPrice:   pbMarket.Amm.InstMarkPrice(),
			BlockNumber: sdk.NewInt(ctx.BlockHeight()),
		}
	}

	return &bindings.AllMarketsResponse{
		MarketMap: marketMap,
	}, err
}

func (perpExt *PerpQuerier) BasePrice(
	ctx sdk.Context, cwReq *bindings.BasePriceRequest,
) (*bindings.BasePriceResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (perpExt *PerpQuerier) PremiumFraction(
	ctx sdk.Context, cwReq *bindings.PremiumFractionRequest,
) (*bindings.PremiumFractionResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (perpExt *PerpQuerier) Metrics(
	ctx sdk.Context, cwReq *bindings.MetricsRequest,
) (*bindings.MetricsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (perpExt *PerpQuerier) ModuleAccounts(
	ctx sdk.Context, cwReq *bindings.ModuleAccountsRequest,
) (*bindings.ModuleAccountsResponse, error) {
	if cwReq == nil {
		return nil, errors.New("nil request")
	}

	sdkReq := &perpv2types.QueryModuleAccountsRequest{}
	goCtx := sdk.WrapSDKContext(ctx)
	sdkResp, err := perpExt.perp.ModuleAccounts(goCtx, sdkReq)
	if err != nil {
		return nil, err
	}

	moduleAccounts := make(map[string]bindings.ModuleAccountWithBalance)
	for _, acc := range sdkResp.Accounts {
		addr, err := sdk.AccAddressFromBech32(acc.Address)
		if err != nil {
			return nil, err
		}
		moduleAccounts[acc.Name] = bindings.ModuleAccountWithBalance{
			Name:    acc.Name,
			Addr:    addr,
			Balance: acc.Balance,
		}
	}

	return &bindings.ModuleAccountsResponse{
		ModuleAccounts: moduleAccounts,
	}, err
}

func (perpExt *PerpQuerier) ModuleParams(
	ctx sdk.Context, cwReq *bindings.PerpParamsRequest,
) (*bindings.PerpParamsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (perpExt *PerpQuerier) Position(
	ctx sdk.Context, cwReq *bindings.PositionRequest,
) (*bindings.PositionResponse, error) {
	pair, err := asset.TryNewPair(cwReq.Pair)
	if err != nil {
		return nil, err
	}
	sdkReq := &perpv2types.QueryPositionRequest{
		Pair:   pair,
		Trader: cwReq.Trader,
	}
	goCtx := sdk.WrapSDKContext(ctx)
	sdkResp, err := perpExt.perp.QueryPosition(goCtx, sdkReq)
	if err != nil {
		return nil, err
	}
	return &bindings.PositionResponse{
		Position: bindings.Position{
			TraderAddr:   sdkResp.Position.TraderAddress,
			Pair:         sdkResp.Position.Pair.String(),
			Size:         sdkResp.Position.Size_,
			Margin:       sdkResp.Position.Margin,
			OpenNotional: sdkResp.Position.OpenNotional,
			LatestCPF:    sdkResp.Position.LatestCumulativePremiumFraction,
			BlockNumber:  sdk.NewInt(sdkResp.Position.LastUpdatedBlockNumber),
		},
		Notional:          sdkResp.PositionNotional,
		Upnl:              sdkResp.UnrealizedPnl,
		Margin_ratio_mark: sdkResp.MarginRatio,
		// Margin_ratio_index: sdkResp.MarginRatioIndex,
		Block_number: sdk.NewInt(sdkResp.Position.LastUpdatedBlockNumber),
	}, err
}

func (perpExt *PerpQuerier) Positions(
	ctx sdk.Context, cwReq *bindings.PositionsRequest,
) (*bindings.PositionsResponse, error) {
	sdkReq := &perpv2types.QueryPositionsRequest{
		Trader: cwReq.Trader,
	}
	goCtx := sdk.WrapSDKContext(ctx)
	sdkResp, err := perpExt.perp.QueryPositions(goCtx, sdkReq)
	if err != nil {
		return nil, err
	}

	positionMap := make(map[string]bindings.Position)
	for _, posResp := range sdkResp.Positions {
		pair := posResp.Position.Pair.String()
		pos := posResp.Position
		positionMap[pair] = bindings.Position{
			TraderAddr:   pos.TraderAddress,
			Pair:         pair,
			Size:         pos.Size_,
			Margin:       pos.Margin,
			OpenNotional: pos.OpenNotional,
			LatestCPF:    pos.LatestCumulativePremiumFraction,
			BlockNumber:  sdk.NewInt(pos.LastUpdatedBlockNumber),
		}
	}

	return &bindings.PositionsResponse{
		Positions: positionMap,
	}, err
}

// ----------------------------------------------------------------------
// OracleQuerier
// ----------------------------------------------------------------------

type OracleQuerier struct {
	oracle oracletypes.QueryServer
}

func (oracleExt *OracleQuerier) ExchangeRates(
	ctx sdk.Context, cwReq *bindings.OraclePrices,
) (*bindings.OraclePricesResponse, error) {
	queryExchangeRatesRequest := oracletypes.QueryExchangeRatesRequest{}
	queryExchangeRates, err := oracleExt.oracle.ExchangeRates(ctx, &queryExchangeRatesRequest)

	// Transform Tuple to Map
	exchangeRates := make(map[string]sdk.Dec)
	for _, exchangeRate := range queryExchangeRates.ExchangeRates {
		exchangeRates[exchangeRate.Pair.String()] = exchangeRate.ExchangeRate
	}

	cwResp := new(bindings.OraclePricesResponse)
	*cwResp = exchangeRates
	return cwResp, err
}
