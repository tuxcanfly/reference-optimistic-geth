package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

var big10 = big.NewInt(10)

type L1FeeContext struct {
	BaseFee  *big.Int
	Overhead *big.Int
	Scalar   *big.Float
}

// NewL1FeeContext returns a context for calculating L1 fee cst
func NewL1FeeContext(cfg *params.ChainConfig, statedb interface{}) *L1FeeContext {
	if !cfg.OptimismFee.Enabled {
		return &L1FeeContext{}
	}

	// TODO: these need to be typecasted into big.Ints
	// also L1BaseFee is packed in the slot, see
	// https://github.com/ethereum-optimism/optimism/pull/2596
	// unit test - statedb - interface

    // TODO: use interface
	//l1BaseFee := statedb.GetState(cfg.OptimismFees.L1Block, config.OptimismFees.L1BaseFeeSlot)
	//overhead := statedb.GetState(cfg.OptimismFees.GasPriceOracle, config.OptimismFees.OverheadSlot)
	//scalar := statedb.GetState(cfg.OptimismFees.GasPriceOracle, config.OptimismFees.ScalarSlot)
	//decimals := statedb.GetState(cfg.OptimismFees.GasPriceOracle, config.OptimismFees.DecimalsSlot)

    var l1BaseFee, overhead, scalar, decimals *big.Int

	scaled := ScaleDecimals(scalar, decimals)

	return &L1FeeContext{
		BaseFee:  l1BaseFee,
		Overhead: overhead,
		Scalar:   scaled,
	}
}

func ScaleDecimals(scalar, decimals *big.Int) *big.Float {
	// 10**decimals
	divisor := new(big.Int).Exp(big10, decimals, nil)
	fscalar := new(big.Float).SetInt(scalar)
	fdivisor := new(big.Float).SetInt(divisor)
	// fscalar / fdivisor
	return new(big.Float).Quo(fscalar, fdivisor)
}
