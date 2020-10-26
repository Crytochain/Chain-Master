package core
import (
	"time"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/common/mclock"
	"github.com/Cryptochain-VON/core/types"
	"github.com/Cryptochain-VON/log"
)
type insertStats struct {
	queued, processed, ignored int
	usedGas                    uint64
	lastIndex                  int
	startTime                  mclock.AbsTime
}
const statsReportLimit = 8 * time.Second
func (st *insertStats) report(chain []*types.Block, index int, dirty common.StorageSize) {
	var (
		now     = mclock.Now()
		elapsed = time.Duration(now) - time.Duration(st.startTime)
	)
	if index == len(chain)-1 || elapsed >= statsReportLimit {
		var txs int
		for _, block := range chain[st.lastIndex : index+1] {
			txs += len(block.Transactions())
		}
		end := chain[index]
		context := []interface{}{
			"blocks", st.processed, "txs", txs, "mgas", float64(st.usedGas) / 1000000,
			"elapsed", common.PrettyDuration(elapsed), "mgasps", float64(st.usedGas) * 1000 / float64(elapsed),
			"number", end.Number(), "hash", end.Hash(),
		}
		if timestamp := time.Unix(int64(end.Time()), 0); time.Since(timestamp) > time.Minute {
			context = append(context, []interface{}{"age", common.PrettyAge(timestamp)}...)
		}
		context = append(context, []interface{}{"dirty", dirty}...)
		if st.queued > 0 {
			context = append(context, []interface{}{"queued", st.queued}...)
		}
		if st.ignored > 0 {
			context = append(context, []interface{}{"ignored", st.ignored}...)
		}
		log.Info("Imported new chain segment", context...)
		*st = insertStats{startTime: now, lastIndex: index + 1}
	}
}
type insertIterator struct {
	chain types.Blocks 
	results <-chan error 
	errors  []error      
	index     int       
	validator Validator 
}
func newInsertIterator(chain types.Blocks, results <-chan error, validator Validator) *insertIterator {
	return &insertIterator{
		chain:     chain,
		results:   results,
		errors:    make([]error, 0, len(chain)),
		index:     -1,
		validator: validator,
	}
}
func (it *insertIterator) next() (*types.Block, error) {
	if it.index+1 >= len(it.chain) {
		it.index = len(it.chain)
		return nil, nil
	}
	it.index++
	if len(it.errors) <= it.index {
		it.errors = append(it.errors, <-it.results)
	}
	if it.errors[it.index] != nil {
		return it.chain[it.index], it.errors[it.index]
	}
	return it.chain[it.index], it.validator.ValidateBody(it.chain[it.index])
}
func (it *insertIterator) peek() (*types.Block, error) {
	if it.index+1 >= len(it.chain) {
		return nil, nil
	}
	if len(it.errors) <= it.index+1 {
		it.errors = append(it.errors, <-it.results)
	}
	if it.errors[it.index+1] != nil {
		return it.chain[it.index+1], it.errors[it.index+1]
	}
	return it.chain[it.index+1], nil
}
func (it *insertIterator) previous() *types.Header {
	if it.index < 1 {
		return nil
	}
	return it.chain[it.index-1].Header()
}
func (it *insertIterator) first() *types.Block {
	return it.chain[0]
}
func (it *insertIterator) remaining() int {
	return len(it.chain) - it.index
}
func (it *insertIterator) processed() int {
	return it.index + 1
}
