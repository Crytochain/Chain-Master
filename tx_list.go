package core
import (
	"container/heap"
	"math"
	"math/big"
	"sort"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/core/types"
	"github.com/Cryptochain-VON/log"
)
type nonceHeap []uint64
func (h nonceHeap) Len() int           { return len(h) }
func (h nonceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h nonceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *nonceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}
func (h *nonceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
type txSortedMap struct {
	items map[uint64]*types.Transaction 
	index *nonceHeap                    
	cache types.Transactions            
}
func newTxSortedMap() *txSortedMap {
	return &txSortedMap{
		items: make(map[uint64]*types.Transaction),
		index: new(nonceHeap),
	}
}
func (m *txSortedMap) Get(nonce uint64) *types.Transaction {
	return m.items[nonce]
}
func (m *txSortedMap) Put(tx *types.Transaction) {
	nonce := tx.Nonce()
	if m.items[nonce] == nil {
		heap.Push(m.index, nonce)
	}
	m.items[nonce], m.cache = tx, nil
}
func (m *txSortedMap) Forward(threshold uint64) types.Transactions {
	var removed types.Transactions
	for m.index.Len() > 0 && (*m.index)[0] < threshold {
		nonce := heap.Pop(m.index).(uint64)
		removed = append(removed, m.items[nonce])
		delete(m.items, nonce)
	}
	if m.cache != nil {
		m.cache = m.cache[len(removed):]
	}
	return removed
}
func (m *txSortedMap) Filter(filter func(*types.Transaction) bool) types.Transactions {
	var removed types.Transactions
	for nonce, tx := range m.items {
		if filter(tx) {
			removed = append(removed, tx)
			delete(m.items, nonce)
		}
	}
	if len(removed) > 0 {
		*m.index = make([]uint64, 0, len(m.items))
		for nonce := range m.items {
			*m.index = append(*m.index, nonce)
		}
		heap.Init(m.index)
		m.cache = nil
	}
	return removed
}
func (m *txSortedMap) Cap(threshold int) types.Transactions {
	if len(m.items) <= threshold {
		return nil
	}
	var drops types.Transactions
	sort.Sort(*m.index)
	for size := len(m.items); size > threshold; size-- {
		drops = append(drops, m.items[(*m.index)[size-1]])
		delete(m.items, (*m.index)[size-1])
	}
	*m.index = (*m.index)[:threshold]
	heap.Init(m.index)
	if m.cache != nil {
		m.cache = m.cache[:len(m.cache)-len(drops)]
	}
	return drops
}
func (m *txSortedMap) Remove(nonce uint64) bool {
	_, ok := m.items[nonce]
	if !ok {
		return false
	}
	for i := 0; i < m.index.Len(); i++ {
		if (*m.index)[i] == nonce {
			heap.Remove(m.index, i)
			break
		}
	}
	delete(m.items, nonce)
	m.cache = nil
	return true
}
func (m *txSortedMap) Ready(start uint64) types.Transactions {
	if m.index.Len() == 0 || (*m.index)[0] > start {
		return nil
	}
	var ready types.Transactions
	for next := (*m.index)[0]; m.index.Len() > 0 && (*m.index)[0] == next; next++ {
		ready = append(ready, m.items[next])
		delete(m.items, next)
		heap.Pop(m.index)
	}
	m.cache = nil
	return ready
}
func (m *txSortedMap) Len() int {
	return len(m.items)
}
func (m *txSortedMap) Flatten() types.Transactions {
	if m.cache == nil {
		m.cache = make(types.Transactions, 0, len(m.items))
		for _, tx := range m.items {
			m.cache = append(m.cache, tx)
		}
		sort.Sort(types.TxByNonce(m.cache))
	}
	txs := make(types.Transactions, len(m.cache))
	copy(txs, m.cache)
	return txs
}
type txList struct {
	strict bool         
	txs    *txSortedMap 
	costcap *big.Int 
	gascap  uint64   
}
func newTxList(strict bool) *txList {
	return &txList{
		strict:  strict,
		txs:     newTxSortedMap(),
		costcap: new(big.Int),
	}
}
func (l *txList) Overlaps(tx *types.Transaction) bool {
	return l.txs.Get(tx.Nonce()) != nil
}
func (l *txList) Add(tx *types.Transaction, priceBump uint64) (bool, *types.Transaction) {
	old := l.txs.Get(tx.Nonce())
	if old != nil {
		threshold := new(big.Int).Div(new(big.Int).Mul(old.GasPrice(), big.NewInt(100+int64(priceBump))), big.NewInt(100))
		if old.GasPrice().Cmp(tx.GasPrice()) >= 0 || threshold.Cmp(tx.GasPrice()) > 0 {
			return false, nil
		}
	}
	l.txs.Put(tx)
	if cost := tx.Cost(); l.costcap.Cmp(cost) < 0 {
		l.costcap = cost
	}
	if gas := tx.Gas(); l.gascap < gas {
		l.gascap = gas
	}
	return true, old
}
func (l *txList) Forward(threshold uint64) types.Transactions {
	return l.txs.Forward(threshold)
}
func (l *txList) Filter(costLimit *big.Int, gasLimit uint64) (types.Transactions, types.Transactions) {
	if l.costcap.Cmp(costLimit) <= 0 && l.gascap <= gasLimit {
		return nil, nil
	}
	l.costcap = new(big.Int).Set(costLimit) 
	l.gascap = gasLimit
	removed := l.txs.Filter(func(tx *types.Transaction) bool { return tx.Cost().Cmp(costLimit) > 0 || tx.Gas() > gasLimit })
	var invalids types.Transactions
	if l.strict && len(removed) > 0 {
		lowest := uint64(math.MaxUint64)
		for _, tx := range removed {
			if nonce := tx.Nonce(); lowest > nonce {
				lowest = nonce
			}
		}
		invalids = l.txs.Filter(func(tx *types.Transaction) bool { return tx.Nonce() > lowest })
	}
	return removed, invalids
}
func (l *txList) Cap(threshold int) types.Transactions {
	return l.txs.Cap(threshold)
}
func (l *txList) Remove(tx *types.Transaction) (bool, types.Transactions) {
	nonce := tx.Nonce()
	if removed := l.txs.Remove(nonce); !removed {
		return false, nil
	}
	if l.strict {
		return true, l.txs.Filter(func(tx *types.Transaction) bool { return tx.Nonce() > nonce })
	}
	return true, nil
}
func (l *txList) Ready(start uint64) types.Transactions {
	return l.txs.Ready(start)
}
func (l *txList) Len() int {
	return l.txs.Len()
}
func (l *txList) Empty() bool {
	return l.Len() == 0
}
func (l *txList) Flatten() types.Transactions {
	return l.txs.Flatten()
}
type priceHeap []*types.Transaction
func (h priceHeap) Len() int      { return len(h) }
func (h priceHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h priceHeap) Less(i, j int) bool {
	switch h[i].GasPrice().Cmp(h[j].GasPrice()) {
	case -1:
		return true
	case 1:
		return false
	}
	return h[i].Nonce() > h[j].Nonce()
}
func (h *priceHeap) Push(x interface{}) {
	*h = append(*h, x.(*types.Transaction))
}
func (h *priceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
type txPricedList struct {
	all    *txLookup  
	items  *priceHeap 
	stales int        
}
func newTxPricedList(all *txLookup) *txPricedList {
	return &txPricedList{
		all:   all,
		items: new(priceHeap),
	}
}
func (l *txPricedList) Put(tx *types.Transaction) {
	heap.Push(l.items, tx)
}
func (l *txPricedList) Removed(count int) {
	l.stales += count
	if l.stales <= len(*l.items)/4 {
		return
	}
	reheap := make(priceHeap, 0, l.all.Count())
	l.stales, l.items = 0, &reheap
	l.all.Range(func(hash common.Hash, tx *types.Transaction) bool {
		*l.items = append(*l.items, tx)
		return true
	})
	heap.Init(l.items)
}
func (l *txPricedList) Cap(threshold *big.Int, local *accountSet) types.Transactions {
	drop := make(types.Transactions, 0, 128) 
	save := make(types.Transactions, 0, 64)  
	for len(*l.items) > 0 {
		tx := heap.Pop(l.items).(*types.Transaction)
		if l.all.Get(tx.Hash()) == nil {
			l.stales--
			continue
		}
		if tx.GasPrice().Cmp(threshold) >= 0 {
			save = append(save, tx)
			break
		}
		if local.containsTx(tx) {
			save = append(save, tx)
		} else {
			drop = append(drop, tx)
		}
	}
	for _, tx := range save {
		heap.Push(l.items, tx)
	}
	return drop
}
func (l *txPricedList) Underpriced(tx *types.Transaction, local *accountSet) bool {
	if local.containsTx(tx) {
		return false
	}
	for len(*l.items) > 0 {
		head := []*types.Transaction(*l.items)[0]
		if l.all.Get(head.Hash()) == nil {
			l.stales--
			heap.Pop(l.items)
			continue
		}
		break
	}
	if len(*l.items) == 0 {
		log.Error("Pricing query for empty pool") 
		return false
	}
	cheapest := []*types.Transaction(*l.items)[0]
	return cheapest.GasPrice().Cmp(tx.GasPrice()) >= 0
}
func (l *txPricedList) Discard(slots int, local *accountSet) types.Transactions {
	drop := make(types.Transactions, 0, slots) 
	save := make(types.Transactions, 0, 64)    
	for len(*l.items) > 0 && slots > 0 {
		tx := heap.Pop(l.items).(*types.Transaction)
		if l.all.Get(tx.Hash()) == nil {
			l.stales--
			continue
		}
		if local.containsTx(tx) {
			save = append(save, tx)
		} else {
			drop = append(drop, tx)
			slots -= numSlots(tx)
		}
	}
	for _, tx := range save {
		heap.Push(l.items, tx)
	}
	return drop
}
