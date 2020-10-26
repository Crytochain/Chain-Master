package core
import (
	"runtime"
	"github.com/Cryptochain-VON/core/types"
)
var senderCacher = newTxSenderCacher(runtime.NumCPU())
type txSenderCacherRequest struct {
	signer types.Signer
	txs    []*types.Transaction
	inc    int
}
type txSenderCacher struct {
	threads int
	tasks   chan *txSenderCacherRequest
}
func newTxSenderCacher(threads int) *txSenderCacher {
	cacher := &txSenderCacher{
		tasks:   make(chan *txSenderCacherRequest, threads),
		threads: threads,
	}
	for i := 0; i < threads; i++ {
		go cacher.cache()
	}
	return cacher
}
func (cacher *txSenderCacher) cache() {
	for task := range cacher.tasks {
		for i := 0; i < len(task.txs); i += task.inc {
			types.Sender(task.signer, task.txs[i])
		}
	}
}
func (cacher *txSenderCacher) recover(signer types.Signer, txs []*types.Transaction) {
	if len(txs) == 0 {
		return
	}
	tasks := cacher.threads
	if len(txs) < tasks*4 {
		tasks = (len(txs) + 3) / 4
	}
	for i := 0; i < tasks; i++ {
		cacher.tasks <- &txSenderCacherRequest{
			signer: signer,
			txs:    txs[i:],
			inc:    tasks,
		}
	}
}
func (cacher *txSenderCacher) recoverFromBlocks(signer types.Signer, blocks []*types.Block) {
	count := 0
	for _, block := range blocks {
		count += len(block.Transactions())
	}
	txs := make([]*types.Transaction, 0, count)
	for _, block := range blocks {
		txs = append(txs, block.Transactions()...)
	}
	cacher.recover(signer, txs)
}
