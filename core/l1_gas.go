package core

import (
	gomath "math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var big10 = big.NewInt(10)

var (
	// TODO: make configurable
	L1BaseFeeSlot = common.BigToHash(big.NewInt(2))
	OverheadSlot  = common.BigToHash(big.NewInt(3))
	ScalarSlot    = common.BigToHash(big.NewInt(4))
	DecimalsSlot  = common.BigToHash(big.NewInt(5))
)

// calculateL1GasUsed returns the gas used to include the transaction data in
// the calldata on L1
func calculateL1GasUsed(data []byte, overhead *big.Int) *big.Int {
	var zeroes uint64
	var ones uint64
	for _, byt := range data {
		if byt == 0 {
			zeroes++
		} else {
			ones++
		}
	}

	zeroesGas := zeroes * params.TxDataZeroGas
	onesGas := (ones + 68) * params.TxDataNonZeroGasEIP2028
	l1Gas := new(big.Int).SetUint64(zeroesGas + onesGas)
	return new(big.Int).Add(l1Gas, overhead)
}

// mulByFloat multiplies a big.Int by a float and returns the
// big.Int rounded upwards
func mulByFloat(num *big.Int, float *big.Float) *big.Int {
	n := new(big.Float).SetUint64(num.Uint64())
	product := n.Mul(n, float)
	pfloat, _ := product.Float64()
	rounded := gomath.Ceil(pfloat)
	return new(big.Int).SetUint64(uint64(rounded))
}

// L1FeeContext includes all the context necessary to calculate the cost of
// including the transaction in L1
type L1FeeContext struct {
	BaseFee  *big.Int
	Overhead *big.Int
	Scalar   *big.Float
}

// NewL1FeeContext returns a context for calculating L1 fee cst
func NewL1FeeContext(cfg *params.ChainConfig, statedb *state.StateDB) *L1FeeContext {
	if cfg.OptimismFee == nil || !cfg.OptimismFee.Enabled {
		return &L1FeeContext{
			BaseFee:  big.NewInt(0),
			Overhead: big.NewInt(0),
			Scalar:   big.NewFloat(0.0),
		}
	}

	// TODO: these need to be typecasted into big.Ints
	// also L1BaseFee is packed in the slot, see
	// https://github.com/ethereum-optimism/optimism/pull/2596
	// unit test - statedb - interface

	l1BaseFee := statedb.GetState(cfg.OptimismFee.L1Block, L1BaseFeeSlot)
	overhead := statedb.GetState(cfg.OptimismFee.GasPriceOracle, OverheadSlot)
	scalar := statedb.GetState(cfg.OptimismFee.GasPriceOracle, ScalarSlot)
	decimals := statedb.GetState(cfg.OptimismFee.GasPriceOracle, DecimalsSlot)

	scaled := ScaleDecimals(scalar.Big(), decimals.Big())

	return &L1FeeContext{
		BaseFee:  l1BaseFee.Big(),
		Overhead: overhead.Big(),
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

// L1Cost returns the L1 fee cost.
// This depends on the chainconfig because gas costs
// can change over time
func L1Cost(tx *types.Transaction, ctx *L1FeeContext) *big.Int {
	rlp := tx.Data()
	l1GasUsed := calculateL1GasUsed(rlp, ctx.Overhead)
	l1Cost := new(big.Int).Mul(l1GasUsed, ctx.BaseFee)
	return mulByFloat(l1Cost, ctx.Scalar)
}
