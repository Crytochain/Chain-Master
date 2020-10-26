package core
import (
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/core/types"
)
type NewTxsEvent struct{ Txs []*types.Transaction }
type NewMinedBlockEvent struct{ Block *types.Block }
type RemovedLogsEvent struct{ Logs []*types.Log }
type ChainEvent struct {
	Block *types.Block
	Hash  common.Hash
	Logs  []*types.Log
}
type ChainSideEvent struct {
	Block *types.Block
}
type ChainHeadEvent struct{ Block *types.Block }
