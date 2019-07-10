package exec

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/repository"
	"github.com/DSiSc/wasm/util"
	"math"
	"math/big"
)

//WasmChainContext chain context for wasm to execute
type WasmChainContext struct {
	// Message information
	Origin      *types.Address // Provides information for ORIGIN
	GasPrice    *big.Int       // Provides information for GASPRICE
	Coinbase    types.Address  // Provides information for COINBASE
	GasLimit    uint64         // Provides information for GASLIMIT
	BlockNumber *big.Int       // Provides information for NUMBER
	Time        *big.Int       // Provides information for TIME
}

// NewEVMContext creates a new chain context for use in the WASM.
func NewWasmChainContext(tx *types.Transaction, header *types.Header, chain *repository.Repository, author types.Address) *WasmChainContext {
	beneficiary := author
	if beneficiary == author {
		beneficiary = util.HexToAddress("0x0000000000000000000000000000000000000000")
	}
	return &WasmChainContext{
		Origin:   tx.Data.From,
		GasPrice: new(big.Int).Set(tx.Data.Price),
		Coinbase: beneficiary,
		// TODO: Initially we will not specify a precise gas limit
		GasLimit:    uint64(math.MaxInt64),
		BlockNumber: new(big.Int).SetUint64(header.Height),
		Time:        new(big.Int).SetUint64(header.Timestamp),
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db *repository.Repository, addr types.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db *repository.Repository, sender, recipient types.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}
