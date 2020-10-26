package core
import (
	"errors"
	"math"
	"math/big"
	"sort"
	"sync"
	"time"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/common/prque"
	"github.com/Cryptochain-VON/core/state"
	"github.com/Cryptochain-VON/core/types"
	"github.com/Cryptochain-VON/event"
	"github.com/Cryptochain-VON/log"
	"github.com/Cryptochain-VON/metrics"
	"github.com/Cryptochain-VON/params"
)
const (
	chainHeadChanSize = 10
	txSlotSize = 32 * 1024
	txMaxSize = 4 * txSlotSize 
)
var (
	ErrAlreadyKnown = errors.New("already known")
	ErrInvalidSender = errors.New("invalid sender")
	ErrUnderpriced = errors.New("transaction underpriced")
	ErrReplaceUnderpriced = errors.New("replacement transaction underpriced")
	ErrGasLimit = errors.New("exceeds block gas limit")
	ErrNegativeValue = errors.New("negative value")
	ErrOversizedData = errors.New("oversized data")
)
var (
	evictionInterval    = time.Minute     
	statsReportInterval = 8 * time.Second 
)
var (
	pendingDiscardMeter   = metrics.NewRegisteredMeter("txpool/pending/discard", nil)
	pendingReplaceMeter   = metrics.NewRegisteredMeter("txpool/pending/replace", nil)
	pendingRateLimitMeter = metrics.NewRegisteredMeter("txpool/pending/ratelimit", nil) 
	pendingNofundsMeter   = metrics.NewRegisteredMeter("txpool/pending/nofunds", nil)   
	queuedDiscardMeter   = metrics.NewRegisteredMeter("txpool/queued/discard", nil)
	queuedReplaceMeter   = metrics.NewRegisteredMeter("txpool/queued/replace", nil)
	queuedRateLimitMeter = metrics.NewRegisteredMeter("txpool/queued/ratelimit", nil) 
	queuedNofundsMeter   = metrics.NewRegisteredMeter("txpool/queued/nofunds", nil)   
	knownTxMeter       = metrics.NewRegisteredMeter("txpool/known", nil)
	validTxMeter       = metrics.NewRegisteredMeter("txpool/valid", nil)
	invalidTxMeter     = metrics.NewRegisteredMeter("txpool/invalid", nil)
	underpricedTxMeter = metrics.NewRegisteredMeter("txpool/underpriced", nil)
	pendingGauge = metrics.NewRegisteredGauge("txpool/pending", nil)
	queuedGauge  = metrics.NewRegisteredGauge("txpool/queued", nil)
	localGauge   = metrics.NewRegisteredGauge("txpool/local", nil)
	slotsGauge   = metrics.NewRegisteredGauge("txpool/slots", nil)
)
type TxStatus uint
const (
	TxStatusUnknown TxStatus = iota
	TxStatusQueued
	TxStatusPending
	TxStatusIncluded
)
type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
	SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
}
type TxPoolConfig struct {
	Locals    []common.Address 
	NoLocals  bool             
	Journal   string           
	Rejournal time.Duration    
	PriceLimit uint64 
	PriceBump  uint64 
	AccountSlots uint64 
	GlobalSlots  uint64 
	AccountQueue uint64 
	GlobalQueue  uint64 
	Lifetime time.Duration 
}
var DefaultTxPoolConfig = TxPoolConfig{
	Journal:   "transactions.rlp",
	Rejournal: time.Hour,
	PriceLimit: 1,
	PriceBump:  10,
	AccountSlots: 16,
	GlobalSlots:  4096,
	AccountQueue: 64,
	GlobalQueue:  1024,
	Lifetime: 3 * time.Hour,
}
func (config *TxPoolConfig) sanitize() TxPoolConfig {
	conf := *config
	if conf.Rejournal < time.Second {
		log.Warn("Sanitizing invalid txpool journal time", "provided", conf.Rejournal, "updated", time.Second)
		conf.Rejournal = time.Second
	}
	if conf.PriceLimit < 1 {
		log.Warn("Sanitizing invalid txpool price limit", "provided", conf.PriceLimit, "updated", DefaultTxPoolConfig.PriceLimit)
		conf.PriceLimit = DefaultTxPoolConfig.PriceLimit
	}
	if conf.PriceBump < 1 {
		log.Warn("Sanitizing invalid txpool price bump", "provided", conf.PriceBump, "updated", DefaultTxPoolConfig.PriceBump)
		conf.PriceBump = DefaultTxPoolConfig.PriceBump
	}
	if conf.AccountSlots < 1 {
		log.Warn("Sanitizing invalid txpool account slots", "provided", conf.AccountSlots, "updated", DefaultTxPoolConfig.AccountSlots)
		conf.AccountSlots = DefaultTxPoolConfig.AccountSlots
	}
	if conf.GlobalSlots < 1 {
		log.Warn("Sanitizing invalid txpool global slots", "provided", conf.GlobalSlots, "updated", DefaultTxPoolConfig.GlobalSlots)
		conf.GlobalSlots = DefaultTxPoolConfig.GlobalSlots
	}
	if conf.AccountQueue < 1 {
		log.Warn("Sanitizing invalid txpool account queue", "provided", conf.AccountQueue, "updated", DefaultTxPoolConfig.AccountQueue)
		conf.AccountQueue = DefaultTxPoolConfig.AccountQueue
	}
	if conf.GlobalQueue < 1 {
		log.Warn("Sanitizing invalid txpool global queue", "provided", conf.GlobalQueue, "updated", DefaultTxPoolConfig.GlobalQueue)
		conf.GlobalQueue = DefaultTxPoolConfig.GlobalQueue
	}
	if conf.Lifetime < 1 {
		log.Warn("Sanitizing invalid txpool lifetime", "provided", conf.Lifetime, "updated", DefaultTxPoolConfig.Lifetime)
		conf.Lifetime = DefaultTxPoolConfig.Lifetime
	}
	return conf
}
type TxPool struct {
	config      TxPoolConfig
	chainconfig *params.ChainConfig
	chain       blockChain
	gasPrice    *big.Int
	txFeed      event.Feed
	scope       event.SubscriptionScope
	signer      types.Signer
	mu          sync.RWMutex
	istanbul bool 
	currentState  *state.StateDB 
	pendingNonces *txNoncer      
	currentMaxGas uint64         
	locals  *accountSet 
	journal *txJournal  
	pending map[common.Address]*txList   
	queue   map[common.Address]*txList   
	beats   map[common.Address]time.Time 
	all     *txLookup                    
	priced  *txPricedList                
	chainHeadCh     chan ChainHeadEvent
	chainHeadSub    event.Subscription
	reqResetCh      chan *txpoolResetRequest
	reqPromoteCh    chan *accountSet
	queueTxEventCh  chan *types.Transaction
	reorgDoneCh     chan chan struct{}
	reorgShutdownCh chan struct{}  
	wg              sync.WaitGroup 
}
type txpoolResetRequest struct {
	oldHead, newHead *types.Header
}
func NewTxPool(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain) *TxPool {
	config = (&config).sanitize()
	pool := &TxPool{
		config:          config,
		chainconfig:     chainconfig,
		chain:           chain,
		signer:          types.NewEIP155Signer(chainconfig.ChainID),
		pending:         make(map[common.Address]*txList),
		queue:           make(map[common.Address]*txList),
		beats:           make(map[common.Address]time.Time),
		all:             newTxLookup(),
		chainHeadCh:     make(chan ChainHeadEvent, chainHeadChanSize),
		reqResetCh:      make(chan *txpoolResetRequest),
		reqPromoteCh:    make(chan *accountSet),
		queueTxEventCh:  make(chan *types.Transaction),
		reorgDoneCh:     make(chan chan struct{}),
		reorgShutdownCh: make(chan struct{}),
		gasPrice:        new(big.Int).SetUint64(config.PriceLimit),
	}
	pool.locals = newAccountSet(pool.signer)
	for _, addr := range config.Locals {
		log.Info("Setting new local account", "address", addr)
		pool.locals.add(addr)
	}
	pool.priced = newTxPricedList(pool.all)
	pool.reset(nil, chain.CurrentBlock().Header())
	pool.wg.Add(1)
	go pool.scheduleReorgLoop()
	if !config.NoLocals && config.Journal != "" {
		pool.journal = newTxJournal(config.Journal)
		if err := pool.journal.load(pool.AddLocals); err != nil {
			log.Warn("Failed to load transaction journal", "err", err)
		}
		if err := pool.journal.rotate(pool.local()); err != nil {
			log.Warn("Failed to rotate transaction journal", "err", err)
		}
	}
	pool.chainHeadSub = pool.chain.SubscribeChainHeadEvent(pool.chainHeadCh)
	pool.wg.Add(1)
	go pool.loop()
	return pool
}
func (pool *TxPool) loop() {
	defer pool.wg.Done()
	var (
		prevPending, prevQueued, prevStales int
		report  = time.NewTicker(statsReportInterval)
		evict   = time.NewTicker(evictionInterval)
		journal = time.NewTicker(pool.config.Rejournal)
		head = pool.chain.CurrentBlock()
	)
	defer report.Stop()
	defer evict.Stop()
	defer journal.Stop()
	for {
		select {
		case ev := <-pool.chainHeadCh:
			if ev.Block != nil {
				pool.requestReset(head.Header(), ev.Block.Header())
				head = ev.Block
			}
		case <-pool.chainHeadSub.Err():
			close(pool.reorgShutdownCh)
			return
		case <-report.C:
			pool.mu.RLock()
			pending, queued := pool.stats()
			stales := pool.priced.stales
			pool.mu.RUnlock()
			if pending != prevPending || queued != prevQueued || stales != prevStales {
				log.Debug("Transaction pool status report", "executable", pending, "queued", queued, "stales", stales)
				prevPending, prevQueued, prevStales = pending, queued, stales
			}
		case <-evict.C:
			pool.mu.Lock()
			for addr := range pool.queue {
				if pool.locals.contains(addr) {
					continue
				}
				if time.Since(pool.beats[addr]) > pool.config.Lifetime {
					for _, tx := range pool.queue[addr].Flatten() {
						pool.removeTx(tx.Hash(), true)
					}
				}
			}
			pool.mu.Unlock()
		case <-journal.C:
			if pool.journal != nil {
				pool.mu.Lock()
				if err := pool.journal.rotate(pool.local()); err != nil {
					log.Warn("Failed to rotate local tx journal", "err", err)
				}
				pool.mu.Unlock()
			}
		}
	}
}
func (pool *TxPool) Stop() {
	pool.scope.Close()
	pool.chainHeadSub.Unsubscribe()
	pool.wg.Wait()
	if pool.journal != nil {
		pool.journal.close()
	}
	log.Info("Transaction pool stopped")
}
func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}
func (pool *TxPool) GasPrice() *big.Int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	return new(big.Int).Set(pool.gasPrice)
}
func (pool *TxPool) SetGasPrice(price *big.Int) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.gasPrice = price
	for _, tx := range pool.priced.Cap(price, pool.locals) {
		pool.removeTx(tx.Hash(), false)
	}
	log.Info("Transaction pool price threshold updated", "price", price)
}
func (pool *TxPool) Nonce(addr common.Address) uint64 {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	return pool.pendingNonces.get(addr)
}
func (pool *TxPool) Stats() (int, int) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	return pool.stats()
}
func (pool *TxPool) stats() (int, int) {
	pending := 0
	for _, list := range pool.pending {
		pending += list.Len()
	}
	queued := 0
	for _, list := range pool.queue {
		queued += list.Len()
	}
	return pending, queued
}
func (pool *TxPool) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		pending[addr] = list.Flatten()
	}
	queued := make(map[common.Address]types.Transactions)
	for addr, list := range pool.queue {
		queued[addr] = list.Flatten()
	}
	return pending, queued
}
func (pool *TxPool) Pending() (map[common.Address]types.Transactions, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		pending[addr] = list.Flatten()
	}
	return pending, nil
}
func (pool *TxPool) Locals() []common.Address {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.locals.flatten()
}
func (pool *TxPool) local() map[common.Address]types.Transactions {
	txs := make(map[common.Address]types.Transactions)
	for addr := range pool.locals.accounts {
		if pending := pool.pending[addr]; pending != nil {
			txs[addr] = append(txs[addr], pending.Flatten()...)
		}
		if queued := pool.queue[addr]; queued != nil {
			txs[addr] = append(txs[addr], queued.Flatten()...)
		}
	}
	return txs
}
func (pool *TxPool) validateTx(tx *types.Transaction, local bool) error {
	if uint64(tx.Size()) > txMaxSize {
		return ErrOversizedData
	}
	if tx.Value().Sign() < 0 {
		return ErrNegativeValue
	}
	if pool.currentMaxGas < tx.Gas() {
		return ErrGasLimit
	}
	from, err := types.Sender(pool.signer, tx)
	if err != nil {
		return ErrInvalidSender
	}
	local = local || pool.locals.contains(from) 
	if !local && pool.gasPrice.Cmp(tx.GasPrice()) > 0 {
		return ErrUnderpriced
	}
	if pool.currentState.GetNonce(from) > tx.Nonce() {
		return ErrNonceTooLow
	}
	if pool.currentState.GetBalance(from).Cmp(tx.Cost()) < 0 {
		return ErrInsufficientFunds
	}
	intrGas, err := IntrinsicGas(tx.Data(), tx.To() == nil, true, pool.istanbul)
	if err != nil {
		return err
	}
	if tx.Gas() < intrGas {
		return ErrIntrinsicGas
	}
	return nil
}
func (pool *TxPool) add(tx *types.Transaction, local bool) (replaced bool, err error) {
	hash := tx.Hash()
	if pool.all.Get(hash) != nil {
		log.Trace("Discarding already known transaction", "hash", hash)
		knownTxMeter.Mark(1)
		return false, ErrAlreadyKnown
	}
	if err := pool.validateTx(tx, local); err != nil {
		log.Trace("Discarding invalid transaction", "hash", hash, "err", err)
		invalidTxMeter.Mark(1)
		return false, err
	}
	if uint64(pool.all.Count()) >= pool.config.GlobalSlots+pool.config.GlobalQueue {
		if !local && pool.priced.Underpriced(tx, pool.locals) {
			log.Trace("Discarding underpriced transaction", "hash", hash, "price", tx.GasPrice())
			underpricedTxMeter.Mark(1)
			return false, ErrUnderpriced
		}
		drop := pool.priced.Discard(pool.all.Slots()-int(pool.config.GlobalSlots+pool.config.GlobalQueue)+numSlots(tx), pool.locals)
		for _, tx := range drop {
			log.Trace("Discarding freshly underpriced transaction", "hash", tx.Hash(), "price", tx.GasPrice())
			underpricedTxMeter.Mark(1)
			pool.removeTx(tx.Hash(), false)
		}
	}
	from, _ := types.Sender(pool.signer, tx) 
	if list := pool.pending[from]; list != nil && list.Overlaps(tx) {
		inserted, old := list.Add(tx, pool.config.PriceBump)
		if !inserted {
			pendingDiscardMeter.Mark(1)
			return false, ErrReplaceUnderpriced
		}
		if old != nil {
			pool.all.Remove(old.Hash())
			pool.priced.Removed(1)
			pendingReplaceMeter.Mark(1)
		}
		pool.all.Add(tx)
		pool.priced.Put(tx)
		pool.journalTx(from, tx)
		pool.queueTxEvent(tx)
		log.Trace("Pooled new executable transaction", "hash", hash, "from", from, "to", tx.To())
		return old != nil, nil
	}
	replaced, err = pool.enqueueTx(hash, tx)
	if err != nil {
		return false, err
	}
	if local {
		if !pool.locals.contains(from) {
			log.Info("Setting new local account", "address", from)
			pool.locals.add(from)
		}
	}
	if local || pool.locals.contains(from) {
		localGauge.Inc(1)
	}
	pool.journalTx(from, tx)
	log.Trace("Pooled new future transaction", "hash", hash, "from", from, "to", tx.To())
	return replaced, nil
}
func (pool *TxPool) enqueueTx(hash common.Hash, tx *types.Transaction) (bool, error) {
	from, _ := types.Sender(pool.signer, tx) 
	if pool.queue[from] == nil {
		pool.queue[from] = newTxList(false)
	}
	inserted, old := pool.queue[from].Add(tx, pool.config.PriceBump)
	if !inserted {
		queuedDiscardMeter.Mark(1)
		return false, ErrReplaceUnderpriced
	}
	if old != nil {
		pool.all.Remove(old.Hash())
		pool.priced.Removed(1)
		queuedReplaceMeter.Mark(1)
	} else {
		queuedGauge.Inc(1)
	}
	if pool.all.Get(hash) == nil {
		pool.all.Add(tx)
		pool.priced.Put(tx)
	}
	return old != nil, nil
}
func (pool *TxPool) journalTx(from common.Address, tx *types.Transaction) {
	if pool.journal == nil || !pool.locals.contains(from) {
		return
	}
	if err := pool.journal.insert(tx); err != nil {
		log.Warn("Failed to journal local transaction", "err", err)
	}
}
func (pool *TxPool) promoteTx(addr common.Address, hash common.Hash, tx *types.Transaction) bool {
	if pool.pending[addr] == nil {
		pool.pending[addr] = newTxList(true)
	}
	list := pool.pending[addr]
	inserted, old := list.Add(tx, pool.config.PriceBump)
	if !inserted {
		pool.all.Remove(hash)
		pool.priced.Removed(1)
		pendingDiscardMeter.Mark(1)
		return false
	}
	if old != nil {
		pool.all.Remove(old.Hash())
		pool.priced.Removed(1)
		pendingReplaceMeter.Mark(1)
	} else {
		pendingGauge.Inc(1)
	}
	if pool.all.Get(hash) == nil {
		pool.all.Add(tx)
		pool.priced.Put(tx)
	}
	pool.beats[addr] = time.Now()
	pool.pendingNonces.set(addr, tx.Nonce()+1)
	return true
}
func (pool *TxPool) AddLocals(txs []*types.Transaction) []error {
	return pool.addTxs(txs, !pool.config.NoLocals, true)
}
func (pool *TxPool) AddLocal(tx *types.Transaction) error {
	errs := pool.AddLocals([]*types.Transaction{tx})
	return errs[0]
}
func (pool *TxPool) AddRemotes(txs []*types.Transaction) []error {
	return pool.addTxs(txs, false, false)
}
func (pool *TxPool) AddRemotesSync(txs []*types.Transaction) []error {
	return pool.addTxs(txs, false, true)
}
func (pool *TxPool) addRemoteSync(tx *types.Transaction) error {
	errs := pool.AddRemotesSync([]*types.Transaction{tx})
	return errs[0]
}
func (pool *TxPool) AddRemote(tx *types.Transaction) error {
	errs := pool.AddRemotes([]*types.Transaction{tx})
	return errs[0]
}
func (pool *TxPool) addTxs(txs []*types.Transaction, local, sync bool) []error {
	var (
		errs = make([]error, len(txs))
		news = make([]*types.Transaction, 0, len(txs))
	)
	for i, tx := range txs {
		if pool.all.Get(tx.Hash()) != nil {
			errs[i] = ErrAlreadyKnown
			knownTxMeter.Mark(1)
			continue
		}
		news = append(news, tx)
	}
	if len(news) == 0 {
		return errs
	}
	for _, tx := range news {
		types.Sender(pool.signer, tx)
	}
	pool.mu.Lock()
	newErrs, dirtyAddrs := pool.addTxsLocked(news, local)
	pool.mu.Unlock()
	var nilSlot = 0
	for _, err := range newErrs {
		for errs[nilSlot] != nil {
			nilSlot++
		}
		errs[nilSlot] = err
	}
	done := pool.requestPromoteExecutables(dirtyAddrs)
	if sync {
		<-done
	}
	return errs
}
func (pool *TxPool) addTxsLocked(txs []*types.Transaction, local bool) ([]error, *accountSet) {
	dirty := newAccountSet(pool.signer)
	errs := make([]error, len(txs))
	for i, tx := range txs {
		replaced, err := pool.add(tx, local)
		errs[i] = err
		if err == nil && !replaced {
			dirty.addTx(tx)
		}
	}
	validTxMeter.Mark(int64(len(dirty.accounts)))
	return errs, dirty
}
func (pool *TxPool) Status(hashes []common.Hash) []TxStatus {
	status := make([]TxStatus, len(hashes))
	for i, hash := range hashes {
		tx := pool.Get(hash)
		if tx == nil {
			continue
		}
		from, _ := types.Sender(pool.signer, tx) 
		pool.mu.RLock()
		if txList := pool.pending[from]; txList != nil && txList.txs.items[tx.Nonce()] != nil {
			status[i] = TxStatusPending
		} else if txList := pool.queue[from]; txList != nil && txList.txs.items[tx.Nonce()] != nil {
			status[i] = TxStatusQueued
		}
		pool.mu.RUnlock()
	}
	return status
}
func (pool *TxPool) Get(hash common.Hash) *types.Transaction {
	return pool.all.Get(hash)
}
func (pool *TxPool) Has(hash common.Hash) bool {
	return pool.all.Get(hash) != nil
}
func (pool *TxPool) removeTx(hash common.Hash, outofbound bool) {
	tx := pool.all.Get(hash)
	if tx == nil {
		return
	}
	addr, _ := types.Sender(pool.signer, tx) 
	pool.all.Remove(hash)
	if outofbound {
		pool.priced.Removed(1)
	}
	if pool.locals.contains(addr) {
		localGauge.Dec(1)
	}
	if pending := pool.pending[addr]; pending != nil {
		if removed, invalids := pending.Remove(tx); removed {
			if pending.Empty() {
				delete(pool.pending, addr)
				delete(pool.beats, addr)
			}
			for _, tx := range invalids {
				pool.enqueueTx(tx.Hash(), tx)
			}
			pool.pendingNonces.setIfLower(addr, tx.Nonce())
			pendingGauge.Dec(int64(1 + len(invalids)))
			return
		}
	}
	if future := pool.queue[addr]; future != nil {
		if removed, _ := future.Remove(tx); removed {
			queuedGauge.Dec(1)
		}
		if future.Empty() {
			delete(pool.queue, addr)
		}
	}
}
func (pool *TxPool) requestReset(oldHead *types.Header, newHead *types.Header) chan struct{} {
	select {
	case pool.reqResetCh <- &txpoolResetRequest{oldHead, newHead}:
		return <-pool.reorgDoneCh
	case <-pool.reorgShutdownCh:
		return pool.reorgShutdownCh
	}
}
func (pool *TxPool) requestPromoteExecutables(set *accountSet) chan struct{} {
	select {
	case pool.reqPromoteCh <- set:
		return <-pool.reorgDoneCh
	case <-pool.reorgShutdownCh:
		return pool.reorgShutdownCh
	}
}
func (pool *TxPool) queueTxEvent(tx *types.Transaction) {
	select {
	case pool.queueTxEventCh <- tx:
	case <-pool.reorgShutdownCh:
	}
}
func (pool *TxPool) scheduleReorgLoop() {
	defer pool.wg.Done()
	var (
		curDone       chan struct{} 
		nextDone      = make(chan struct{})
		launchNextRun bool
		reset         *txpoolResetRequest
		dirtyAccounts *accountSet
		queuedEvents  = make(map[common.Address]*txSortedMap)
	)
	for {
		if curDone == nil && launchNextRun {
			go pool.runReorg(nextDone, reset, dirtyAccounts, queuedEvents)
			curDone, nextDone = nextDone, make(chan struct{})
			launchNextRun = false
			reset, dirtyAccounts = nil, nil
			queuedEvents = make(map[common.Address]*txSortedMap)
		}
		select {
		case req := <-pool.reqResetCh:
			if reset == nil {
				reset = req
			} else {
				reset.newHead = req.newHead
			}
			launchNextRun = true
			pool.reorgDoneCh <- nextDone
		case req := <-pool.reqPromoteCh:
			if dirtyAccounts == nil {
				dirtyAccounts = req
			} else {
				dirtyAccounts.merge(req)
			}
			launchNextRun = true
			pool.reorgDoneCh <- nextDone
		case tx := <-pool.queueTxEventCh:
			addr, _ := types.Sender(pool.signer, tx)
			if _, ok := queuedEvents[addr]; !ok {
				queuedEvents[addr] = newTxSortedMap()
			}
			queuedEvents[addr].Put(tx)
		case <-curDone:
			curDone = nil
		case <-pool.reorgShutdownCh:
			if curDone != nil {
				<-curDone
			}
			close(nextDone)
			return
		}
	}
}
func (pool *TxPool) runReorg(done chan struct{}, reset *txpoolResetRequest, dirtyAccounts *accountSet, events map[common.Address]*txSortedMap) {
	defer close(done)
	var promoteAddrs []common.Address
	if dirtyAccounts != nil {
		promoteAddrs = dirtyAccounts.flatten()
	}
	pool.mu.Lock()
	if reset != nil {
		pool.reset(reset.oldHead, reset.newHead)
		for addr := range events {
			events[addr].Forward(pool.pendingNonces.get(addr))
			if events[addr].Len() == 0 {
				delete(events, addr)
			}
		}
		promoteAddrs = promoteAddrs[:0]
		for addr := range pool.queue {
			promoteAddrs = append(promoteAddrs, addr)
		}
	}
	promoted := pool.promoteExecutables(promoteAddrs)
	if reset != nil {
		pool.demoteUnexecutables()
	}
	pool.truncatePending()
	pool.truncateQueue()
	for addr, list := range pool.pending {
		txs := list.Flatten() 
		pool.pendingNonces.set(addr, txs[len(txs)-1].Nonce()+1)
	}
	pool.mu.Unlock()
	for _, tx := range promoted {
		addr, _ := types.Sender(pool.signer, tx)
		if _, ok := events[addr]; !ok {
			events[addr] = newTxSortedMap()
		}
		events[addr].Put(tx)
	}
	if len(events) > 0 {
		var txs []*types.Transaction
		for _, set := range events {
			txs = append(txs, set.Flatten()...)
		}
		pool.txFeed.Send(NewTxsEvent{txs})
	}
}
func (pool *TxPool) reset(oldHead, newHead *types.Header) {
	var reinject types.Transactions
	if oldHead != nil && oldHead.Hash() != newHead.ParentHash {
		oldNum := oldHead.Number.Uint64()
		newNum := newHead.Number.Uint64()
		if depth := uint64(math.Abs(float64(oldNum) - float64(newNum))); depth > 64 {
			log.Debug("Skipping deep transaction reorg", "depth", depth)
		} else {
			var discarded, included types.Transactions
			var (
				rem = pool.chain.GetBlock(oldHead.Hash(), oldHead.Number.Uint64())
				add = pool.chain.GetBlock(newHead.Hash(), newHead.Number.Uint64())
			)
			if rem == nil {
				if newNum < oldNum {
					log.Debug("Skipping transaction reset caused by setHead",
						"old", oldHead.Hash(), "oldnum", oldNum, "new", newHead.Hash(), "newnum", newNum)
				} else {
					log.Warn("Transaction pool reset with missing oldhead",
						"old", oldHead.Hash(), "oldnum", oldNum, "new", newHead.Hash(), "newnum", newNum)
				}
				return
			}
			for rem.NumberU64() > add.NumberU64() {
				discarded = append(discarded, rem.Transactions()...)
				if rem = pool.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
			}
			for add.NumberU64() > rem.NumberU64() {
				included = append(included, add.Transactions()...)
				if add = pool.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			for rem.Hash() != add.Hash() {
				discarded = append(discarded, rem.Transactions()...)
				if rem = pool.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
				included = append(included, add.Transactions()...)
				if add = pool.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			reinject = types.TxDifference(discarded, included)
		}
	}
	if newHead == nil {
		newHead = pool.chain.CurrentBlock().Header() 
	}
	statedb, err := pool.chain.StateAt(newHead.Root)
	if err != nil {
		log.Error("Failed to reset txpool state", "err", err)
		return
	}
	pool.currentState = statedb
	pool.pendingNonces = newTxNoncer(statedb)
	pool.currentMaxGas = newHead.GasLimit
	log.Debug("Reinjecting stale transactions", "count", len(reinject))
	senderCacher.recover(pool.signer, reinject)
	pool.addTxsLocked(reinject, false)
	next := new(big.Int).Add(newHead.Number, big.NewInt(1))
	pool.istanbul = pool.chainconfig.IsIstanbul(next)
}
func (pool *TxPool) promoteExecutables(accounts []common.Address) []*types.Transaction {
	var promoted []*types.Transaction
	for _, addr := range accounts {
		list := pool.queue[addr]
		if list == nil {
			continue 
		}
		forwards := list.Forward(pool.currentState.GetNonce(addr))
		for _, tx := range forwards {
			hash := tx.Hash()
			pool.all.Remove(hash)
			log.Trace("Removed old queued transaction", "hash", hash)
		}
		drops, _ := list.Filter(pool.currentState.GetBalance(addr), pool.currentMaxGas)
		for _, tx := range drops {
			hash := tx.Hash()
			pool.all.Remove(hash)
			log.Trace("Removed unpayable queued transaction", "hash", hash)
		}
		queuedNofundsMeter.Mark(int64(len(drops)))
		readies := list.Ready(pool.pendingNonces.get(addr))
		for _, tx := range readies {
			hash := tx.Hash()
			if pool.promoteTx(addr, hash, tx) {
				log.Trace("Promoting queued transaction", "hash", hash)
				promoted = append(promoted, tx)
			}
		}
		queuedGauge.Dec(int64(len(readies)))
		var caps types.Transactions
		if !pool.locals.contains(addr) {
			caps = list.Cap(int(pool.config.AccountQueue))
			for _, tx := range caps {
				hash := tx.Hash()
				pool.all.Remove(hash)
				log.Trace("Removed cap-exceeding queued transaction", "hash", hash)
			}
			queuedRateLimitMeter.Mark(int64(len(caps)))
		}
		pool.priced.Removed(len(forwards) + len(drops) + len(caps))
		queuedGauge.Dec(int64(len(forwards) + len(drops) + len(caps)))
		if pool.locals.contains(addr) {
			localGauge.Dec(int64(len(forwards) + len(drops) + len(caps)))
		}
		if list.Empty() {
			delete(pool.queue, addr)
		}
	}
	return promoted
}
func (pool *TxPool) truncatePending() {
	pending := uint64(0)
	for _, list := range pool.pending {
		pending += uint64(list.Len())
	}
	if pending <= pool.config.GlobalSlots {
		return
	}
	pendingBeforeCap := pending
	spammers := prque.New(nil)
	for addr, list := range pool.pending {
		if !pool.locals.contains(addr) && uint64(list.Len()) > pool.config.AccountSlots {
			spammers.Push(addr, int64(list.Len()))
		}
	}
	offenders := []common.Address{}
	for pending > pool.config.GlobalSlots && !spammers.Empty() {
		offender, _ := spammers.Pop()
		offenders = append(offenders, offender.(common.Address))
		if len(offenders) > 1 {
			threshold := pool.pending[offender.(common.Address)].Len()
			for pending > pool.config.GlobalSlots && pool.pending[offenders[len(offenders)-2]].Len() > threshold {
				for i := 0; i < len(offenders)-1; i++ {
					list := pool.pending[offenders[i]]
					caps := list.Cap(list.Len() - 1)
					for _, tx := range caps {
						hash := tx.Hash()
						pool.all.Remove(hash)
						pool.pendingNonces.setIfLower(offenders[i], tx.Nonce())
						log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
					}
					pool.priced.Removed(len(caps))
					pendingGauge.Dec(int64(len(caps)))
					if pool.locals.contains(offenders[i]) {
						localGauge.Dec(int64(len(caps)))
					}
					pending--
				}
			}
		}
	}
	if pending > pool.config.GlobalSlots && len(offenders) > 0 {
		for pending > pool.config.GlobalSlots && uint64(pool.pending[offenders[len(offenders)-1]].Len()) > pool.config.AccountSlots {
			for _, addr := range offenders {
				list := pool.pending[addr]
				caps := list.Cap(list.Len() - 1)
				for _, tx := range caps {
					hash := tx.Hash()
					pool.all.Remove(hash)
					pool.pendingNonces.setIfLower(addr, tx.Nonce())
					log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
				}
				pool.priced.Removed(len(caps))
				pendingGauge.Dec(int64(len(caps)))
				if pool.locals.contains(addr) {
					localGauge.Dec(int64(len(caps)))
				}
				pending--
			}
		}
	}
	pendingRateLimitMeter.Mark(int64(pendingBeforeCap - pending))
}
func (pool *TxPool) truncateQueue() {
	queued := uint64(0)
	for _, list := range pool.queue {
		queued += uint64(list.Len())
	}
	if queued <= pool.config.GlobalQueue {
		return
	}
	addresses := make(addressesByHeartbeat, 0, len(pool.queue))
	for addr := range pool.queue {
		if !pool.locals.contains(addr) { 
			addresses = append(addresses, addressByHeartbeat{addr, pool.beats[addr]})
		}
	}
	sort.Sort(addresses)
	for drop := queued - pool.config.GlobalQueue; drop > 0 && len(addresses) > 0; {
		addr := addresses[len(addresses)-1]
		list := pool.queue[addr.address]
		addresses = addresses[:len(addresses)-1]
		if size := uint64(list.Len()); size <= drop {
			for _, tx := range list.Flatten() {
				pool.removeTx(tx.Hash(), true)
			}
			drop -= size
			queuedRateLimitMeter.Mark(int64(size))
			continue
		}
		txs := list.Flatten()
		for i := len(txs) - 1; i >= 0 && drop > 0; i-- {
			pool.removeTx(txs[i].Hash(), true)
			drop--
			queuedRateLimitMeter.Mark(1)
		}
	}
}
func (pool *TxPool) demoteUnexecutables() {
	for addr, list := range pool.pending {
		nonce := pool.currentState.GetNonce(addr)
		olds := list.Forward(nonce)
		for _, tx := range olds {
			hash := tx.Hash()
			pool.all.Remove(hash)
			log.Trace("Removed old pending transaction", "hash", hash)
		}
		drops, invalids := list.Filter(pool.currentState.GetBalance(addr), pool.currentMaxGas)
		for _, tx := range drops {
			hash := tx.Hash()
			log.Trace("Removed unpayable pending transaction", "hash", hash)
			pool.all.Remove(hash)
		}
		pool.priced.Removed(len(olds) + len(drops))
		pendingNofundsMeter.Mark(int64(len(drops)))
		for _, tx := range invalids {
			hash := tx.Hash()
			log.Trace("Demoting pending transaction", "hash", hash)
			pool.enqueueTx(hash, tx)
		}
		pendingGauge.Dec(int64(len(olds) + len(drops) + len(invalids)))
		if pool.locals.contains(addr) {
			localGauge.Dec(int64(len(olds) + len(drops) + len(invalids)))
		}
		if list.Len() > 0 && list.txs.Get(nonce) == nil {
			gapped := list.Cap(0)
			for _, tx := range gapped {
				hash := tx.Hash()
				log.Error("Demoting invalidated transaction", "hash", hash)
				pool.enqueueTx(hash, tx)
			}
			pendingGauge.Dec(int64(len(gapped)))
		}
		if list.Empty() {
			delete(pool.pending, addr)
			delete(pool.beats, addr)
		}
	}
}
type addressByHeartbeat struct {
	address   common.Address
	heartbeat time.Time
}
type addressesByHeartbeat []addressByHeartbeat
func (a addressesByHeartbeat) Len() int           { return len(a) }
func (a addressesByHeartbeat) Less(i, j int) bool { return a[i].heartbeat.Before(a[j].heartbeat) }
func (a addressesByHeartbeat) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
type accountSet struct {
	accounts map[common.Address]struct{}
	signer   types.Signer
	cache    *[]common.Address
}
func newAccountSet(signer types.Signer, addrs ...common.Address) *accountSet {
	as := &accountSet{
		accounts: make(map[common.Address]struct{}),
		signer:   signer,
	}
	for _, addr := range addrs {
		as.add(addr)
	}
	return as
}
func (as *accountSet) contains(addr common.Address) bool {
	_, exist := as.accounts[addr]
	return exist
}
func (as *accountSet) containsTx(tx *types.Transaction) bool {
	if addr, err := types.Sender(as.signer, tx); err == nil {
		return as.contains(addr)
	}
	return false
}
func (as *accountSet) add(addr common.Address) {
	as.accounts[addr] = struct{}{}
	as.cache = nil
}
func (as *accountSet) addTx(tx *types.Transaction) {
	if addr, err := types.Sender(as.signer, tx); err == nil {
		as.add(addr)
	}
}
func (as *accountSet) flatten() []common.Address {
	if as.cache == nil {
		accounts := make([]common.Address, 0, len(as.accounts))
		for account := range as.accounts {
			accounts = append(accounts, account)
		}
		as.cache = &accounts
	}
	return *as.cache
}
func (as *accountSet) merge(other *accountSet) {
	for addr := range other.accounts {
		as.accounts[addr] = struct{}{}
	}
	as.cache = nil
}
type txLookup struct {
	all   map[common.Hash]*types.Transaction
	slots int
	lock  sync.RWMutex
}
func newTxLookup() *txLookup {
	return &txLookup{
		all: make(map[common.Hash]*types.Transaction),
	}
}
func (t *txLookup) Range(f func(hash common.Hash, tx *types.Transaction) bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	for key, value := range t.all {
		if !f(key, value) {
			break
		}
	}
}
func (t *txLookup) Get(hash common.Hash) *types.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.all[hash]
}
func (t *txLookup) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return len(t.all)
}
func (t *txLookup) Slots() int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.slots
}
func (t *txLookup) Add(tx *types.Transaction) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.slots += numSlots(tx)
	slotsGauge.Update(int64(t.slots))
	t.all[tx.Hash()] = tx
}
func (t *txLookup) Remove(hash common.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.slots -= numSlots(t.all[hash])
	slotsGauge.Update(int64(t.slots))
	delete(t.all, hash)
}
func numSlots(tx *types.Transaction) int {
	return int((tx.Size() + txSlotSize - 1) / txSlotSize)
}
