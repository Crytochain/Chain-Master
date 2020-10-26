package core
import (
	"sync"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/core/state"
)
type txNoncer struct {
	fallback *state.StateDB
	nonces   map[common.Address]uint64
	lock     sync.Mutex
}
func newTxNoncer(statedb *state.StateDB) *txNoncer {
	return &txNoncer{
		fallback: statedb.Copy(),
		nonces:   make(map[common.Address]uint64),
	}
}
func (txn *txNoncer) get(addr common.Address) uint64 {
	txn.lock.Lock()
	defer txn.lock.Unlock()
	if _, ok := txn.nonces[addr]; !ok {
		txn.nonces[addr] = txn.fallback.GetNonce(addr)
	}
	return txn.nonces[addr]
}
func (txn *txNoncer) set(addr common.Address, nonce uint64) {
	txn.lock.Lock()
	defer txn.lock.Unlock()
	txn.nonces[addr] = nonce
}
func (txn *txNoncer) setIfLower(addr common.Address, nonce uint64) {
	txn.lock.Lock()
	defer txn.lock.Unlock()
	if _, ok := txn.nonces[addr]; !ok {
		txn.nonces[addr] = txn.fallback.GetNonce(addr)
	}
	if txn.nonces[addr] <= nonce {
		return
	}
	txn.nonces[addr] = nonce
}
