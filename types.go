package core
import (
	"github.com/Cryptochain-VON/core/state"
	"github.com/Cryptochain-VON/core/types"
	"github.com/Cryptochain-VON/core/vm"
)
type Validator interface {
	ValidateBody(block *types.Block) error
	ValidateState(block *types.Block, state *state.StateDB, receipts types.Receipts, usedGas uint64) error
}
type Prefetcher interface {
	Prefetch(block *types.Block, statedb *state.StateDB, cfg vm.Config, interrupt *uint32)
}
type Processor interface {
	Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error)
}
