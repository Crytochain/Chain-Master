package core
import (
	"errors"
	"fmt"
	"io"
	"math/big"
	mrand "math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/common/mclock"
	"github.com/Cryptochain-VON/common/prque"
	"github.com/Cryptochain-VON/consensus"
	"github.com/Cryptochain-VON/core/rawdb"
	"github.com/Cryptochain-VON/core/state"
	"github.com/Cryptochain-VON/core/state/snapshot"
	"github.com/Cryptochain-VON/core/types"
	"github.com/Cryptochain-VON/core/vm"
	"github.com/Cryptochain-VON/ethdb"
	"github.com/Cryptochain-VON/event"
	"github.com/Cryptochain-VON/log"
	"github.com/Cryptochain-VON/metrics"
	"github.com/Cryptochain-VON/params"
	"github.com/Cryptochain-VON/rlp"
	"github.com/Cryptochain-VON/trie"
	lru "github.com/hashicorp/golang-lru"
)
var (
	headBlockGauge     = metrics.NewRegisteredGauge("chain/head/block", nil)
	headHeaderGauge    = metrics.NewRegisteredGauge("chain/head/header", nil)
	headFastBlockGauge = metrics.NewRegisteredGauge("chain/head/receipt", nil)
	accountReadTimer   = metrics.NewRegisteredTimer("chain/account/reads", nil)
	accountHashTimer   = metrics.NewRegisteredTimer("chain/account/hashes", nil)
	accountUpdateTimer = metrics.NewRegisteredTimer("chain/account/updates", nil)
	accountCommitTimer = metrics.NewRegisteredTimer("chain/account/commits", nil)
	storageReadTimer   = metrics.NewRegisteredTimer("chain/storage/reads", nil)
	storageHashTimer   = metrics.NewRegisteredTimer("chain/storage/hashes", nil)
	storageUpdateTimer = metrics.NewRegisteredTimer("chain/storage/updates", nil)
	storageCommitTimer = metrics.NewRegisteredTimer("chain/storage/commits", nil)
	snapshotAccountReadTimer = metrics.NewRegisteredTimer("chain/snapshot/account/reads", nil)
	snapshotStorageReadTimer = metrics.NewRegisteredTimer("chain/snapshot/storage/reads", nil)
	snapshotCommitTimer      = metrics.NewRegisteredTimer("chain/snapshot/commits", nil)
	blockInsertTimer     = metrics.NewRegisteredTimer("chain/inserts", nil)
	blockValidationTimer = metrics.NewRegisteredTimer("chain/validation", nil)
	blockExecutionTimer  = metrics.NewRegisteredTimer("chain/execution", nil)
	blockWriteTimer      = metrics.NewRegisteredTimer("chain/write", nil)
	blockReorgAddMeter   = metrics.NewRegisteredMeter("chain/reorg/drop", nil)
	blockReorgDropMeter  = metrics.NewRegisteredMeter("chain/reorg/add", nil)
	blockPrefetchExecuteTimer   = metrics.NewRegisteredTimer("chain/prefetch/executes", nil)
	blockPrefetchInterruptMeter = metrics.NewRegisteredMeter("chain/prefetch/interrupts", nil)
	errInsertionInterrupted = errors.New("insertion is interrupted")
)
const (
	bodyCacheLimit      = 256
	blockCacheLimit     = 256
	receiptsCacheLimit  = 32
	txLookupCacheLimit  = 1024
	maxFutureBlocks     = 256
	maxTimeFutureBlocks = 30
	badBlockLimit       = 10
	TriesInMemory       = 128
	BlockChainVersion uint64 = 7
)
type CacheConfig struct {
	TrieCleanLimit      int           
	TrieCleanNoPrefetch bool          
	TrieDirtyLimit      int           
	TrieDirtyDisabled   bool          
	TrieTimeLimit       time.Duration 
	SnapshotLimit       int           
	SnapshotWait bool 
}
type BlockChain struct {
	chainConfig *params.ChainConfig 
	cacheConfig *CacheConfig        
	db     ethdb.Database 
	snaps  *snapshot.Tree 
	triegc *prque.Prque   
	gcproc time.Duration  
	txLookupLimit uint64
	hc            *HeaderChain
	rmLogsFeed    event.Feed
	chainFeed     event.Feed
	chainSideFeed event.Feed
	chainHeadFeed event.Feed
	logsFeed      event.Feed
	blockProcFeed event.Feed
	scope         event.SubscriptionScope
	genesisBlock  *types.Block
	chainmu sync.RWMutex 
	currentBlock     atomic.Value 
	currentFastBlock atomic.Value 
	stateCache    state.Database 
	bodyCache     *lru.Cache     
	bodyRLPCache  *lru.Cache     
	receiptsCache *lru.Cache     
	blockCache    *lru.Cache     
	txLookupCache *lru.Cache     
	futureBlocks  *lru.Cache     
	quit          chan struct{}  
	wg            sync.WaitGroup 
	running       int32          
	procInterrupt int32          
	engine     consensus.Engine
	validator  Validator  
	prefetcher Prefetcher 
	processor  Processor  
	vmConfig   vm.Config
	badBlocks       *lru.Cache                     
	shouldPreserve  func(*types.Block) bool        
	terminateInsert func(common.Hash, uint64) bool 
}
func NewBlockChain(db ethdb.Database, cacheConfig *CacheConfig, chainConfig *params.ChainConfig, engine consensus.Engine, vmConfig vm.Config, shouldPreserve func(block *types.Block) bool, txLookupLimit *uint64) (*BlockChain, error) {
	if cacheConfig == nil {
		cacheConfig = &CacheConfig{
			TrieCleanLimit: 256,
			TrieDirtyLimit: 256,
			TrieTimeLimit:  5 * time.Minute,
			SnapshotLimit:  256,
			SnapshotWait:   true,
		}
	}
	bodyCache, _ := lru.New(bodyCacheLimit)
	bodyRLPCache, _ := lru.New(bodyCacheLimit)
	receiptsCache, _ := lru.New(receiptsCacheLimit)
	blockCache, _ := lru.New(blockCacheLimit)
	txLookupCache, _ := lru.New(txLookupCacheLimit)
	futureBlocks, _ := lru.New(maxFutureBlocks)
	badBlocks, _ := lru.New(badBlockLimit)
	bc := &BlockChain{
		chainConfig:    chainConfig,
		cacheConfig:    cacheConfig,
		db:             db,
		triegc:         prque.New(nil),
		stateCache:     state.NewDatabaseWithCache(db, cacheConfig.TrieCleanLimit),
		quit:           make(chan struct{}),
		shouldPreserve: shouldPreserve,
		bodyCache:      bodyCache,
		bodyRLPCache:   bodyRLPCache,
		receiptsCache:  receiptsCache,
		blockCache:     blockCache,
		txLookupCache:  txLookupCache,
		futureBlocks:   futureBlocks,
		engine:         engine,
		vmConfig:       vmConfig,
		badBlocks:      badBlocks,
	}
	bc.validator = NewBlockValidator(chainConfig, bc, engine)
	bc.prefetcher = newStatePrefetcher(chainConfig, bc, engine)
	bc.processor = NewStateProcessor(chainConfig, bc, engine)
	var err error
	bc.hc, err = NewHeaderChain(db, chainConfig, engine, bc.insertStopped)
	if err != nil {
		return nil, err
	}
	bc.genesisBlock = bc.GetBlockByNumber(0)
	if bc.genesisBlock == nil {
		return nil, ErrNoGenesis
	}
	var nilBlock *types.Block
	bc.currentBlock.Store(nilBlock)
	bc.currentFastBlock.Store(nilBlock)
	var txIndexBlock uint64
	if bc.empty() {
		rawdb.InitDatabaseFromFreezer(bc.db)
		frozen, _ := bc.db.Ancients()
		if frozen > 0 {
			txIndexBlock = frozen
		}
	}
	if err := bc.loadLastState(); err != nil {
		return nil, err
	}
	bc.engine.VerifyHeader(bc, bc.CurrentHeader(), true)
	if frozen, err := bc.db.Ancients(); err == nil && frozen > 0 {
		var (
			needRewind bool
			low        uint64
		)
		fullBlock := bc.CurrentBlock()
		if fullBlock != nil && fullBlock != bc.genesisBlock && fullBlock.NumberU64() < frozen-1 {
			needRewind = true
			low = fullBlock.NumberU64()
		}
		fastBlock := bc.CurrentFastBlock()
		if fastBlock != nil && fastBlock.NumberU64() < frozen-1 {
			needRewind = true
			if fastBlock.NumberU64() < low || low == 0 {
				low = fastBlock.NumberU64()
			}
		}
		if needRewind {
			var hashes []common.Hash
			previous := bc.CurrentHeader().Number.Uint64()
			for i := low + 1; i <= bc.CurrentHeader().Number.Uint64(); i++ {
				hashes = append(hashes, rawdb.ReadCanonicalHash(bc.db, i))
			}
			bc.Rollback(hashes)
			log.Warn("Truncate ancient chain", "from", previous, "to", low)
		}
	}
	for hash := range BadHashes {
		if header := bc.GetHeaderByHash(hash); header != nil {
			headerByNumber := bc.GetHeaderByNumber(header.Number.Uint64())
			if headerByNumber != nil && headerByNumber.Hash() == header.Hash() {
				log.Error("Found bad hash, rewinding chain", "number", header.Number, "hash", header.ParentHash)
				bc.SetHead(header.Number.Uint64() - 1)
				log.Error("Chain rewind was successful, resuming normal operation")
			}
		}
	}
	if bc.cacheConfig.SnapshotLimit > 0 {
		bc.snaps = snapshot.New(bc.db, bc.stateCache.TrieDB(), bc.cacheConfig.SnapshotLimit, bc.CurrentBlock().Root(), !bc.cacheConfig.SnapshotWait)
	}
	go bc.update()
	if txLookupLimit != nil {
		bc.txLookupLimit = *txLookupLimit
		go bc.maintainTxIndex(txIndexBlock)
	}
	return bc, nil
}
func (bc *BlockChain) GetVMConfig() *vm.Config {
	return &bc.vmConfig
}
func (bc *BlockChain) empty() bool {
	genesis := bc.genesisBlock.Hash()
	for _, hash := range []common.Hash{rawdb.ReadHeadBlockHash(bc.db), rawdb.ReadHeadHeaderHash(bc.db), rawdb.ReadHeadFastBlockHash(bc.db)} {
		if hash != genesis {
			return false
		}
	}
	return true
}
func (bc *BlockChain) loadLastState() error {
	head := rawdb.ReadHeadBlockHash(bc.db)
	if head == (common.Hash{}) {
		log.Warn("Empty database, resetting chain")
		return bc.Reset()
	}
	currentBlock := bc.GetBlockByHash(head)
	if currentBlock == nil {
		log.Warn("Head block missing, resetting chain", "hash", head)
		return bc.Reset()
	}
	if _, err := state.New(currentBlock.Root(), bc.stateCache, bc.snaps); err != nil {
		log.Warn("Head state missing, repairing chain", "number", currentBlock.Number(), "hash", currentBlock.Hash())
		if err := bc.repair(&currentBlock); err != nil {
			return err
		}
		rawdb.WriteHeadBlockHash(bc.db, currentBlock.Hash())
	}
	bc.currentBlock.Store(currentBlock)
	headBlockGauge.Update(int64(currentBlock.NumberU64()))
	currentHeader := currentBlock.Header()
	if head := rawdb.ReadHeadHeaderHash(bc.db); head != (common.Hash{}) {
		if header := bc.GetHeaderByHash(head); header != nil {
			currentHeader = header
		}
	}
	bc.hc.SetCurrentHeader(currentHeader)
	bc.currentFastBlock.Store(currentBlock)
	headFastBlockGauge.Update(int64(currentBlock.NumberU64()))
	if head := rawdb.ReadHeadFastBlockHash(bc.db); head != (common.Hash{}) {
		if block := bc.GetBlockByHash(head); block != nil {
			bc.currentFastBlock.Store(block)
			headFastBlockGauge.Update(int64(block.NumberU64()))
		}
	}
	currentFastBlock := bc.CurrentFastBlock()
	headerTd := bc.GetTd(currentHeader.Hash(), currentHeader.Number.Uint64())
	blockTd := bc.GetTd(currentBlock.Hash(), currentBlock.NumberU64())
	fastTd := bc.GetTd(currentFastBlock.Hash(), currentFastBlock.NumberU64())
	log.Info("Loaded most recent local header", "number", currentHeader.Number, "hash", currentHeader.Hash(), "td", headerTd, "age", common.PrettyAge(time.Unix(int64(currentHeader.Time), 0)))
	log.Info("Loaded most recent local full block", "number", currentBlock.Number(), "hash", currentBlock.Hash(), "td", blockTd, "age", common.PrettyAge(time.Unix(int64(currentBlock.Time()), 0)))
	log.Info("Loaded most recent local fast block", "number", currentFastBlock.Number(), "hash", currentFastBlock.Hash(), "td", fastTd, "age", common.PrettyAge(time.Unix(int64(currentFastBlock.Time()), 0)))
	return nil
}
func (bc *BlockChain) SetHead(head uint64) error {
	log.Warn("Rewinding blockchain", "target", head)
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()
	updateFn := func(db ethdb.KeyValueWriter, header *types.Header) {
		if currentBlock := bc.CurrentBlock(); currentBlock != nil && header.Number.Uint64() < currentBlock.NumberU64() {
			newHeadBlock := bc.GetBlock(header.Hash(), header.Number.Uint64())
			if newHeadBlock == nil {
				newHeadBlock = bc.genesisBlock
			} else {
				if _, err := state.New(newHeadBlock.Root(), bc.stateCache, bc.snaps); err != nil {
					newHeadBlock = bc.genesisBlock
				}
			}
			rawdb.WriteHeadBlockHash(db, newHeadBlock.Hash())
			bc.currentBlock.Store(newHeadBlock)
			headBlockGauge.Update(int64(newHeadBlock.NumberU64()))
		}
		if currentFastBlock := bc.CurrentFastBlock(); currentFastBlock != nil && header.Number.Uint64() < currentFastBlock.NumberU64() {
			newHeadFastBlock := bc.GetBlock(header.Hash(), header.Number.Uint64())
			if newHeadFastBlock == nil {
				newHeadFastBlock = bc.genesisBlock
			}
			rawdb.WriteHeadFastBlockHash(db, newHeadFastBlock.Hash())
			bc.currentFastBlock.Store(newHeadFastBlock)
			headFastBlockGauge.Update(int64(newHeadFastBlock.NumberU64()))
		}
	}
	delFn := func(db ethdb.KeyValueWriter, hash common.Hash, num uint64) {
		frozen, _ := bc.db.Ancients()
		if num+1 <= frozen {
			if err := bc.db.TruncateAncients(num + 1); err != nil {
				log.Crit("Failed to truncate ancient data", "number", num, "err", err)
			}
			rawdb.DeleteHeaderNumber(db, hash)
		} else {
			rawdb.DeleteBody(db, hash, num)
			rawdb.DeleteReceipts(db, hash, num)
		}
	}
	bc.hc.SetHead(head, updateFn, delFn)
	bc.bodyCache.Purge()
	bc.bodyRLPCache.Purge()
	bc.receiptsCache.Purge()
	bc.blockCache.Purge()
	bc.txLookupCache.Purge()
	bc.futureBlocks.Purge()
	return bc.loadLastState()
}
func (bc *BlockChain) FastSyncCommitHead(hash common.Hash) error {
	block := bc.GetBlockByHash(hash)
	if block == nil {
		return fmt.Errorf("non existent block [%x…]", hash[:4])
	}
	if _, err := trie.NewSecure(block.Root(), bc.stateCache.TrieDB()); err != nil {
		return err
	}
	bc.chainmu.Lock()
	bc.currentBlock.Store(block)
	headBlockGauge.Update(int64(block.NumberU64()))
	bc.chainmu.Unlock()
	if bc.snaps != nil {
		bc.snaps.Rebuild(block.Root())
	}
	log.Info("Committed new head block", "number", block.Number(), "hash", hash)
	return nil
}
func (bc *BlockChain) GasLimit() uint64 {
	return bc.CurrentBlock().GasLimit()
}
func (bc *BlockChain) CurrentBlock() *types.Block {
	return bc.currentBlock.Load().(*types.Block)
}
func (bc *BlockChain) Snapshot() *snapshot.Tree {
	return bc.snaps
}
func (bc *BlockChain) CurrentFastBlock() *types.Block {
	return bc.currentFastBlock.Load().(*types.Block)
}
func (bc *BlockChain) Validator() Validator {
	return bc.validator
}
func (bc *BlockChain) Processor() Processor {
	return bc.processor
}
func (bc *BlockChain) State() (*state.StateDB, error) {
	return bc.StateAt(bc.CurrentBlock().Root())
}
func (bc *BlockChain) StateAt(root common.Hash) (*state.StateDB, error) {
	return state.New(root, bc.stateCache, bc.snaps)
}
func (bc *BlockChain) StateCache() state.Database {
	return bc.stateCache
}
func (bc *BlockChain) Reset() error {
	return bc.ResetWithGenesisBlock(bc.genesisBlock)
}
func (bc *BlockChain) ResetWithGenesisBlock(genesis *types.Block) error {
	if err := bc.SetHead(0); err != nil {
		return err
	}
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()
	batch := bc.db.NewBatch()
	rawdb.WriteTd(batch, genesis.Hash(), genesis.NumberU64(), genesis.Difficulty())
	rawdb.WriteBlock(batch, genesis)
	if err := batch.Write(); err != nil {
		log.Crit("Failed to write genesis block", "err", err)
	}
	bc.writeHeadBlock(genesis)
	bc.genesisBlock = genesis
	bc.currentBlock.Store(bc.genesisBlock)
	headBlockGauge.Update(int64(bc.genesisBlock.NumberU64()))
	bc.hc.SetGenesis(bc.genesisBlock.Header())
	bc.hc.SetCurrentHeader(bc.genesisBlock.Header())
	bc.currentFastBlock.Store(bc.genesisBlock)
	headFastBlockGauge.Update(int64(bc.genesisBlock.NumberU64()))
	return nil
}
func (bc *BlockChain) repair(head **types.Block) error {
	for {
		if _, err := state.New((*head).Root(), bc.stateCache, bc.snaps); err == nil {
			log.Info("Rewound blockchain to past state", "number", (*head).Number(), "hash", (*head).Hash())
			return nil
		}
		block := bc.GetBlock((*head).ParentHash(), (*head).NumberU64()-1)
		if block == nil {
			return fmt.Errorf("missing block %d [%x]", (*head).NumberU64()-1, (*head).ParentHash())
		}
		*head = block
	}
}
func (bc *BlockChain) Export(w io.Writer) error {
	return bc.ExportN(w, uint64(0), bc.CurrentBlock().NumberU64())
}
func (bc *BlockChain) ExportN(w io.Writer, first uint64, last uint64) error {
	bc.chainmu.RLock()
	defer bc.chainmu.RUnlock()
	if first > last {
		return fmt.Errorf("export failed: first (%d) is greater than last (%d)", first, last)
	}
	log.Info("Exporting batch of blocks", "count", last-first+1)
	start, reported := time.Now(), time.Now()
	for nr := first; nr <= last; nr++ {
		block := bc.GetBlockByNumber(nr)
		if block == nil {
			return fmt.Errorf("export failed on #%d: not found", nr)
		}
		if err := block.EncodeRLP(w); err != nil {
			return err
		}
		if time.Since(reported) >= statsReportLimit {
			log.Info("Exporting blocks", "exported", block.NumberU64()-first, "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
	}
	return nil
}
func (bc *BlockChain) writeHeadBlock(block *types.Block) {
	updateHeads := rawdb.ReadCanonicalHash(bc.db, block.NumberU64()) != block.Hash()
	batch := bc.db.NewBatch()
	rawdb.WriteCanonicalHash(batch, block.Hash(), block.NumberU64())
	rawdb.WriteTxLookupEntries(batch, block)
	rawdb.WriteHeadBlockHash(batch, block.Hash())
	if updateHeads {
		rawdb.WriteHeadHeaderHash(batch, block.Hash())
		rawdb.WriteHeadFastBlockHash(batch, block.Hash())
	}
	if err := batch.Write(); err != nil {
		log.Crit("Failed to update chain indexes and markers", "err", err)
	}
	if updateHeads {
		bc.hc.SetCurrentHeader(block.Header())
		bc.currentFastBlock.Store(block)
		headFastBlockGauge.Update(int64(block.NumberU64()))
	}
	bc.currentBlock.Store(block)
	headBlockGauge.Update(int64(block.NumberU64()))
}
func (bc *BlockChain) Genesis() *types.Block {
	return bc.genesisBlock
}
func (bc *BlockChain) GetBody(hash common.Hash) *types.Body {
	if cached, ok := bc.bodyCache.Get(hash); ok {
		body := cached.(*types.Body)
		return body
	}
	number := bc.hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	body := rawdb.ReadBody(bc.db, hash, *number)
	if body == nil {
		return nil
	}
	bc.bodyCache.Add(hash, body)
	return body
}
func (bc *BlockChain) GetBodyRLP(hash common.Hash) rlp.RawValue {
	if cached, ok := bc.bodyRLPCache.Get(hash); ok {
		return cached.(rlp.RawValue)
	}
	number := bc.hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	body := rawdb.ReadBodyRLP(bc.db, hash, *number)
	if len(body) == 0 {
		return nil
	}
	bc.bodyRLPCache.Add(hash, body)
	return body
}
func (bc *BlockChain) HasBlock(hash common.Hash, number uint64) bool {
	if bc.blockCache.Contains(hash) {
		return true
	}
	return rawdb.HasBody(bc.db, hash, number)
}
func (bc *BlockChain) HasFastBlock(hash common.Hash, number uint64) bool {
	if !bc.HasBlock(hash, number) {
		return false
	}
	if bc.receiptsCache.Contains(hash) {
		return true
	}
	return rawdb.HasReceipts(bc.db, hash, number)
}
func (bc *BlockChain) HasState(hash common.Hash) bool {
	_, err := bc.stateCache.OpenTrie(hash)
	return err == nil
}
func (bc *BlockChain) HasBlockAndState(hash common.Hash, number uint64) bool {
	block := bc.GetBlock(hash, number)
	if block == nil {
		return false
	}
	return bc.HasState(block.Root())
}
func (bc *BlockChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	if block, ok := bc.blockCache.Get(hash); ok {
		return block.(*types.Block)
	}
	block := rawdb.ReadBlock(bc.db, hash, number)
	if block == nil {
		return nil
	}
	bc.blockCache.Add(block.Hash(), block)
	return block
}
func (bc *BlockChain) GetBlockByHash(hash common.Hash) *types.Block {
	number := bc.hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	return bc.GetBlock(hash, *number)
}
func (bc *BlockChain) GetBlockByNumber(number uint64) *types.Block {
	hash := rawdb.ReadCanonicalHash(bc.db, number)
	if hash == (common.Hash{}) {
		return nil
	}
	return bc.GetBlock(hash, number)
}
func (bc *BlockChain) GetReceiptsByHash(hash common.Hash) types.Receipts {
	if receipts, ok := bc.receiptsCache.Get(hash); ok {
		return receipts.(types.Receipts)
	}
	number := rawdb.ReadHeaderNumber(bc.db, hash)
	if number == nil {
		return nil
	}
	receipts := rawdb.ReadReceipts(bc.db, hash, *number, bc.chainConfig)
	if receipts == nil {
		return nil
	}
	bc.receiptsCache.Add(hash, receipts)
	return receipts
}
func (bc *BlockChain) GetBlocksFromHash(hash common.Hash, n int) (blocks []*types.Block) {
	number := bc.hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	for i := 0; i < n; i++ {
		block := bc.GetBlock(hash, *number)
		if block == nil {
			break
		}
		blocks = append(blocks, block)
		hash = block.ParentHash()
		*number--
	}
	return
}
func (bc *BlockChain) GetUnclesInChain(block *types.Block, length int) []*types.Header {
	uncles := []*types.Header{}
	for i := 0; block != nil && i < length; i++ {
		uncles = append(uncles, block.Uncles()...)
		block = bc.GetBlock(block.ParentHash(), block.NumberU64()-1)
	}
	return uncles
}
func (bc *BlockChain) TrieNode(hash common.Hash) ([]byte, error) {
	return bc.stateCache.TrieDB().Node(hash)
}
func (bc *BlockChain) Stop() {
	if !atomic.CompareAndSwapInt32(&bc.running, 0, 1) {
		return
	}
	bc.scope.Close()
	close(bc.quit)
	bc.StopInsert()
	bc.wg.Wait()
	var snapBase common.Hash
	if bc.snaps != nil {
		var err error
		if snapBase, err = bc.snaps.Journal(bc.CurrentBlock().Root()); err != nil {
			log.Error("Failed to journal state snapshot", "err", err)
		}
	}
	if !bc.cacheConfig.TrieDirtyDisabled {
		triedb := bc.stateCache.TrieDB()
		for _, offset := range []uint64{0, 1, TriesInMemory - 1} {
			if number := bc.CurrentBlock().NumberU64(); number > offset {
				recent := bc.GetBlockByNumber(number - offset)
				log.Info("Writing cached state to disk", "block", recent.Number(), "hash", recent.Hash(), "root", recent.Root())
				if err := triedb.Commit(recent.Root(), true); err != nil {
					log.Error("Failed to commit recent state trie", "err", err)
				}
			}
		}
		if snapBase != (common.Hash{}) {
			log.Info("Writing snapshot state to disk", "root", snapBase)
			if err := triedb.Commit(snapBase, true); err != nil {
				log.Error("Failed to commit recent state trie", "err", err)
			}
		}
		for !bc.triegc.Empty() {
			triedb.Dereference(bc.triegc.PopItem().(common.Hash))
		}
		if size, _ := triedb.Size(); size != 0 {
			log.Error("Dangling trie nodes after full cleanup")
		}
	}
	log.Info("Blockchain stopped")
}
func (bc *BlockChain) StopInsert() {
	atomic.StoreInt32(&bc.procInterrupt, 1)
}
func (bc *BlockChain) insertStopped() bool {
	return atomic.LoadInt32(&bc.procInterrupt) == 1
}
func (bc *BlockChain) procFutureBlocks() {
	blocks := make([]*types.Block, 0, bc.futureBlocks.Len())
	for _, hash := range bc.futureBlocks.Keys() {
		if block, exist := bc.futureBlocks.Peek(hash); exist {
			blocks = append(blocks, block.(*types.Block))
		}
	}
	if len(blocks) > 0 {
		sort.Slice(blocks, func(i, j int) bool {
			return blocks[i].NumberU64() < blocks[j].NumberU64()
		})
		for i := range blocks {
			bc.InsertChain(blocks[i : i+1])
		}
	}
}
type WriteStatus byte
const (
	NonStatTy WriteStatus = iota
	CanonStatTy
	SideStatTy
)
func (bc *BlockChain) Rollback(chain []common.Hash) {
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()
	batch := bc.db.NewBatch()
	for i := len(chain) - 1; i >= 0; i-- {
		hash := chain[i]
		currentHeader := bc.hc.CurrentHeader()
		if currentHeader.Hash() == hash {
			newHeadHeader := bc.GetHeader(currentHeader.ParentHash, currentHeader.Number.Uint64()-1)
			rawdb.WriteHeadHeaderHash(batch, currentHeader.ParentHash)
			bc.hc.SetCurrentHeader(newHeadHeader)
		}
		if currentFastBlock := bc.CurrentFastBlock(); currentFastBlock.Hash() == hash {
			newFastBlock := bc.GetBlock(currentFastBlock.ParentHash(), currentFastBlock.NumberU64()-1)
			rawdb.WriteHeadFastBlockHash(batch, currentFastBlock.ParentHash())
			bc.currentFastBlock.Store(newFastBlock)
			headFastBlockGauge.Update(int64(newFastBlock.NumberU64()))
		}
		if currentBlock := bc.CurrentBlock(); currentBlock.Hash() == hash {
			newBlock := bc.GetBlock(currentBlock.ParentHash(), currentBlock.NumberU64()-1)
			rawdb.WriteHeadBlockHash(batch, currentBlock.ParentHash())
			bc.currentBlock.Store(newBlock)
			headBlockGauge.Update(int64(newBlock.NumberU64()))
		}
	}
	if err := batch.Write(); err != nil {
		log.Crit("Failed to rollback chain markers", "err", err)
	}
	if err := bc.truncateAncient(bc.hc.CurrentHeader().Number.Uint64()); err != nil {
		log.Crit("Truncate ancient store failed", "err", err)
	}
}
func (bc *BlockChain) truncateAncient(head uint64) error {
	frozen, err := bc.db.Ancients()
	if err != nil {
		return err
	}
	if frozen <= head+1 {
		return nil
	}
	if err := bc.db.TruncateAncients(head + 1); err != nil {
		return err
	}
	bc.hc.headerCache.Purge()
	bc.hc.tdCache.Purge()
	bc.hc.numberCache.Purge()
	bc.bodyCache.Purge()
	bc.bodyRLPCache.Purge()
	bc.receiptsCache.Purge()
	bc.blockCache.Purge()
	bc.txLookupCache.Purge()
	bc.futureBlocks.Purge()
	log.Info("Rewind ancient data", "number", head)
	return nil
}
type numberHash struct {
	number uint64
	hash   common.Hash
}
func (bc *BlockChain) InsertReceiptChain(blockChain types.Blocks, receiptChain []types.Receipts, ancientLimit uint64) (int, error) {
	bc.wg.Add(1)
	defer bc.wg.Done()
	var (
		ancientBlocks, liveBlocks     types.Blocks
		ancientReceipts, liveReceipts []types.Receipts
	)
	for i := 0; i < len(blockChain); i++ {
		if i != 0 {
			if blockChain[i].NumberU64() != blockChain[i-1].NumberU64()+1 || blockChain[i].ParentHash() != blockChain[i-1].Hash() {
				log.Error("Non contiguous receipt insert", "number", blockChain[i].Number(), "hash", blockChain[i].Hash(), "parent", blockChain[i].ParentHash(),
					"prevnumber", blockChain[i-1].Number(), "prevhash", blockChain[i-1].Hash())
				return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x…], item %d is #%d [%x…] (parent [%x…])", i-1, blockChain[i-1].NumberU64(),
					blockChain[i-1].Hash().Bytes()[:4], i, blockChain[i].NumberU64(), blockChain[i].Hash().Bytes()[:4], blockChain[i].ParentHash().Bytes()[:4])
			}
		}
		if blockChain[i].NumberU64() <= ancientLimit {
			ancientBlocks, ancientReceipts = append(ancientBlocks, blockChain[i]), append(ancientReceipts, receiptChain[i])
		} else {
			liveBlocks, liveReceipts = append(liveBlocks, blockChain[i]), append(liveReceipts, receiptChain[i])
		}
	}
	var (
		stats = struct{ processed, ignored int32 }{}
		start = time.Now()
		size  = 0
	)
	updateHead := func(head *types.Block) bool {
		bc.chainmu.Lock()
		if bc.CurrentHeader().Number.Cmp(head.Number()) >= 0 {
			currentFastBlock, td := bc.CurrentFastBlock(), bc.GetTd(head.Hash(), head.NumberU64())
			if bc.GetTd(currentFastBlock.Hash(), currentFastBlock.NumberU64()).Cmp(td) < 0 {
				rawdb.WriteHeadFastBlockHash(bc.db, head.Hash())
				bc.currentFastBlock.Store(head)
				headFastBlockGauge.Update(int64(head.NumberU64()))
				bc.chainmu.Unlock()
				return true
			}
		}
		bc.chainmu.Unlock()
		return false
	}
	writeAncient := func(blockChain types.Blocks, receiptChain []types.Receipts) (int, error) {
		var (
			previous = bc.CurrentFastBlock()
			batch    = bc.db.NewBatch()
		)
		defer func() {
			if previous != nil {
				if err := bc.truncateAncient(previous.NumberU64()); err != nil {
					log.Crit("Truncate ancient store failed", "err", err)
				}
			}
		}()
		var deleted []*numberHash
		for i, block := range blockChain {
			if bc.insertStopped() {
				return 0, errInsertionInterrupted
			}
			if bc.terminateInsert != nil && bc.terminateInsert(block.Hash(), block.NumberU64()) {
				return i, errors.New("insertion is terminated for testing purpose")
			}
			if !bc.HasHeader(block.Hash(), block.NumberU64()) {
				return i, fmt.Errorf("containing header #%d [%x…] unknown", block.Number(), block.Hash().Bytes()[:4])
			}
			var (
				start  = time.Now()
				logged = time.Now()
				count  int
			)
			for {
				frozen, _ := bc.db.Ancients()
				if frozen >= block.NumberU64() {
					break
				}
				h := rawdb.ReadCanonicalHash(bc.db, frozen)
				b := rawdb.ReadBlock(bc.db, h, frozen)
				size += rawdb.WriteAncientBlock(bc.db, b, rawdb.ReadReceipts(bc.db, h, frozen, bc.chainConfig), rawdb.ReadTd(bc.db, h, frozen))
				count += 1
				if b.NumberU64() != 0 {
					deleted = append(deleted, &numberHash{b.NumberU64(), b.Hash()})
				}
				if time.Since(logged) > 8*time.Second {
					log.Info("Migrating ancient blocks", "count", count, "elapsed", common.PrettyDuration(time.Since(start)))
					logged = time.Now()
				}
				if len(deleted) > 100000 {
					if err := bc.db.Sync(); err != nil {
						return 0, err
					}
					for _, nh := range deleted {
						rawdb.DeleteBlockWithoutNumber(batch, nh.hash, nh.number)
						rawdb.DeleteCanonicalHash(batch, nh.number)
					}
					if err := batch.Write(); err != nil {
						return 0, err
					}
					batch.Reset()
					for _, nh := range deleted {
						for _, hash := range rawdb.ReadAllHashes(bc.db, nh.number) {
							rawdb.DeleteBlock(batch, hash, nh.number)
						}
					}
					if err := batch.Write(); err != nil {
						return 0, err
					}
					batch.Reset()
					deleted = deleted[0:]
				}
			}
			if count > 0 {
				log.Info("Migrated ancient blocks", "count", count, "elapsed", common.PrettyDuration(time.Since(start)))
			}
			size += rawdb.WriteAncientBlock(bc.db, block, receiptChain[i], bc.GetTd(block.Hash(), block.NumberU64()))
			if bc.txLookupLimit == 0 || ancientLimit <= bc.txLookupLimit || block.NumberU64() >= ancientLimit-bc.txLookupLimit {
				rawdb.WriteTxLookupEntries(batch, block)
			} else if rawdb.ReadTxIndexTail(bc.db) != nil {
				rawdb.WriteTxLookupEntries(batch, block)
			}
			stats.processed++
		}
		size += batch.ValueSize()
		if err := batch.Write(); err != nil {
			return 0, err
		}
		batch.Reset()
		if err := bc.db.Sync(); err != nil {
			return 0, err
		}
		if !updateHead(blockChain[len(blockChain)-1]) {
			return 0, errors.New("side blocks can't be accepted as the ancient chain data")
		}
		previous = nil 
		for _, nh := range deleted {
			rawdb.DeleteBlockWithoutNumber(batch, nh.hash, nh.number)
			rawdb.DeleteCanonicalHash(batch, nh.number)
		}
		for _, block := range blockChain {
			if block.NumberU64() != 0 {
				rawdb.DeleteBlockWithoutNumber(batch, block.Hash(), block.NumberU64())
				rawdb.DeleteCanonicalHash(batch, block.NumberU64())
			}
		}
		if err := batch.Write(); err != nil {
			return 0, err
		}
		batch.Reset()
		for _, nh := range deleted {
			for _, hash := range rawdb.ReadAllHashes(bc.db, nh.number) {
				rawdb.DeleteBlock(batch, hash, nh.number)
			}
		}
		for _, block := range blockChain {
			if block.NumberU64() != 0 {
				for _, hash := range rawdb.ReadAllHashes(bc.db, block.NumberU64()) {
					rawdb.DeleteBlock(batch, hash, block.NumberU64())
				}
			}
		}
		if err := batch.Write(); err != nil {
			return 0, err
		}
		return 0, nil
	}
	writeLive := func(blockChain types.Blocks, receiptChain []types.Receipts) (int, error) {
		batch := bc.db.NewBatch()
		for i, block := range blockChain {
			if bc.insertStopped() {
				return 0, errInsertionInterrupted
			}
			if !bc.HasHeader(block.Hash(), block.NumberU64()) {
				return i, fmt.Errorf("containing header #%d [%x…] unknown", block.Number(), block.Hash().Bytes()[:4])
			}
			if bc.HasBlock(block.Hash(), block.NumberU64()) {
				stats.ignored++
				continue
			}
			rawdb.WriteBody(batch, block.Hash(), block.NumberU64(), block.Body())
			rawdb.WriteReceipts(batch, block.Hash(), block.NumberU64(), receiptChain[i])
			rawdb.WriteTxLookupEntries(batch, block) 
			if batch.ValueSize() >= ethdb.IdealBatchSize {
				if err := batch.Write(); err != nil {
					return 0, err
				}
				size += batch.ValueSize()
				batch.Reset()
			}
			stats.processed++
		}
		if batch.ValueSize() > 0 {
			size += batch.ValueSize()
			if err := batch.Write(); err != nil {
				return 0, err
			}
		}
		updateHead(blockChain[len(blockChain)-1])
		return 0, nil
	}
	if len(ancientBlocks) > 0 {
		if n, err := writeAncient(ancientBlocks, ancientReceipts); err != nil {
			if err == errInsertionInterrupted {
				return 0, nil
			}
			return n, err
		}
	}
	if len(liveBlocks) > 0 && liveBlocks[0].NumberU64() == ancientLimit+1 {
		if tail := rawdb.ReadTxIndexTail(bc.db); tail == nil {
			if bc.txLookupLimit == 0 || ancientLimit <= bc.txLookupLimit {
				rawdb.WriteTxIndexTail(bc.db, 0)
			} else {
				rawdb.WriteTxIndexTail(bc.db, ancientLimit-bc.txLookupLimit)
			}
		}
	}
	if len(liveBlocks) > 0 {
		if n, err := writeLive(liveBlocks, liveReceipts); err != nil {
			if err == errInsertionInterrupted {
				return 0, nil
			}
			return n, err
		}
	}
	head := blockChain[len(blockChain)-1]
	context := []interface{}{
		"count", stats.processed, "elapsed", common.PrettyDuration(time.Since(start)),
		"number", head.Number(), "hash", head.Hash(), "age", common.PrettyAge(time.Unix(int64(head.Time()), 0)),
		"size", common.StorageSize(size),
	}
	if stats.ignored > 0 {
		context = append(context, []interface{}{"ignored", stats.ignored}...)
	}
	log.Info("Imported new block receipts", context...)
	return 0, nil
}
func (bc *BlockChain) SetTxLookupLimit(limit uint64) {
	bc.txLookupLimit = limit
}
func (bc *BlockChain) TxLookupLimit() uint64 {
	return bc.txLookupLimit
}
var lastWrite uint64
func (bc *BlockChain) writeBlockWithoutState(block *types.Block, td *big.Int) (err error) {
	bc.wg.Add(1)
	defer bc.wg.Done()
	batch := bc.db.NewBatch()
	rawdb.WriteTd(batch, block.Hash(), block.NumberU64(), td)
	rawdb.WriteBlock(batch, block)
	if err := batch.Write(); err != nil {
		log.Crit("Failed to write block into disk", "err", err)
	}
	return nil
}
func (bc *BlockChain) writeKnownBlock(block *types.Block) error {
	bc.wg.Add(1)
	defer bc.wg.Done()
	current := bc.CurrentBlock()
	if block.ParentHash() != current.Hash() {
		if err := bc.reorg(current, block); err != nil {
			return err
		}
	}
	bc.writeHeadBlock(block)
	return nil
}
func (bc *BlockChain) WriteBlockWithState(block *types.Block, receipts []*types.Receipt, logs []*types.Log, state *state.StateDB, emitHeadEvent bool) (status WriteStatus, err error) {
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()
	return bc.writeBlockWithState(block, receipts, logs, state, emitHeadEvent)
}
func (bc *BlockChain) writeBlockWithState(block *types.Block, receipts []*types.Receipt, logs []*types.Log, state *state.StateDB, emitHeadEvent bool) (status WriteStatus, err error) {
	bc.wg.Add(1)
	defer bc.wg.Done()
	ptd := bc.GetTd(block.ParentHash(), block.NumberU64()-1)
	if ptd == nil {
		return NonStatTy, consensus.ErrUnknownAncestor
	}
	currentBlock := bc.CurrentBlock()
	localTd := bc.GetTd(currentBlock.Hash(), currentBlock.NumberU64())
	externTd := new(big.Int).Add(block.Difficulty(), ptd)
	blockBatch := bc.db.NewBatch()
	rawdb.WriteTd(blockBatch, block.Hash(), block.NumberU64(), externTd)
	rawdb.WriteBlock(blockBatch, block)
	rawdb.WriteReceipts(blockBatch, block.Hash(), block.NumberU64(), receipts)
	rawdb.WritePreimages(blockBatch, state.Preimages())
	if err := blockBatch.Write(); err != nil {
		log.Crit("Failed to write block into disk", "err", err)
	}
	root, err := state.Commit(bc.chainConfig.IsEIP158(block.Number()))
	if err != nil {
		return NonStatTy, err
	}
	triedb := bc.stateCache.TrieDB()
	if bc.cacheConfig.TrieDirtyDisabled {
		if err := triedb.Commit(root, false); err != nil {
			return NonStatTy, err
		}
	} else {
		triedb.Reference(root, common.Hash{}) 
		bc.triegc.Push(root, -int64(block.NumberU64()))
		if current := block.NumberU64(); current > TriesInMemory {
			var (
				nodes, imgs = triedb.Size()
				limit       = common.StorageSize(bc.cacheConfig.TrieDirtyLimit) * 1024 * 1024
			)
			if nodes > limit || imgs > 4*1024*1024 {
				triedb.Cap(limit - ethdb.IdealBatchSize)
			}
			chosen := current - TriesInMemory
			if bc.gcproc > bc.cacheConfig.TrieTimeLimit {
				header := bc.GetHeaderByNumber(chosen)
				if header == nil {
					log.Warn("Reorg in progress, trie commit postponed", "number", chosen)
				} else {
					if chosen < lastWrite+TriesInMemory && bc.gcproc >= 2*bc.cacheConfig.TrieTimeLimit {
						log.Info("State in memory for too long, committing", "time", bc.gcproc, "allowance", bc.cacheConfig.TrieTimeLimit, "optimum", float64(chosen-lastWrite)/TriesInMemory)
					}
					triedb.Commit(header.Root, true)
					lastWrite = chosen
					bc.gcproc = 0
				}
			}
			for !bc.triegc.Empty() {
				root, number := bc.triegc.Pop()
				if uint64(-number) > chosen {
					bc.triegc.Push(root, number)
					break
				}
				triedb.Dereference(root.(common.Hash))
			}
		}
	}
	reorg := externTd.Cmp(localTd) > 0
	currentBlock = bc.CurrentBlock()
	if !reorg && externTd.Cmp(localTd) == 0 {
		if block.NumberU64() < currentBlock.NumberU64() {
			reorg = true
		} else if block.NumberU64() == currentBlock.NumberU64() {
			var currentPreserve, blockPreserve bool
			if bc.shouldPreserve != nil {
				currentPreserve, blockPreserve = bc.shouldPreserve(currentBlock), bc.shouldPreserve(block)
			}
			reorg = !currentPreserve && (blockPreserve || mrand.Float64() < 0.5)
		}
	}
	if reorg {
		if block.ParentHash() != currentBlock.Hash() {
			if err := bc.reorg(currentBlock, block); err != nil {
				return NonStatTy, err
			}
		}
		status = CanonStatTy
	} else {
		status = SideStatTy
	}
	if status == CanonStatTy {
		bc.writeHeadBlock(block)
	}
	bc.futureBlocks.Remove(block.Hash())
	if status == CanonStatTy {
		bc.chainFeed.Send(ChainEvent{Block: block, Hash: block.Hash(), Logs: logs})
		if len(logs) > 0 {
			bc.logsFeed.Send(logs)
		}
		if emitHeadEvent {
			bc.chainHeadFeed.Send(ChainHeadEvent{Block: block})
		}
	} else {
		bc.chainSideFeed.Send(ChainSideEvent{Block: block})
	}
	return status, nil
}
func (bc *BlockChain) addFutureBlock(block *types.Block) error {
	max := uint64(time.Now().Unix() + maxTimeFutureBlocks)
	if block.Time() > max {
		return fmt.Errorf("future block timestamp %v > allowed %v", block.Time(), max)
	}
	bc.futureBlocks.Add(block.Hash(), block)
	return nil
}
func (bc *BlockChain) InsertChain(chain types.Blocks) (int, error) {
	if len(chain) == 0 {
		return 0, nil
	}
	bc.blockProcFeed.Send(true)
	defer bc.blockProcFeed.Send(false)
	var (
		block, prev *types.Block
	)
	for i := 1; i < len(chain); i++ {
		block = chain[i]
		prev = chain[i-1]
		if block.NumberU64() != prev.NumberU64()+1 || block.ParentHash() != prev.Hash() {
			log.Error("Non contiguous block insert", "number", block.Number(), "hash", block.Hash(),
				"parent", block.ParentHash(), "prevnumber", prev.Number(), "prevhash", prev.Hash())
			return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x…], item %d is #%d [%x…] (parent [%x…])", i-1, prev.NumberU64(),
				prev.Hash().Bytes()[:4], i, block.NumberU64(), block.Hash().Bytes()[:4], block.ParentHash().Bytes()[:4])
		}
	}
	bc.wg.Add(1)
	bc.chainmu.Lock()
	n, err := bc.insertChain(chain, true)
	bc.chainmu.Unlock()
	bc.wg.Done()
	return n, err
}
func (bc *BlockChain) insertChain(chain types.Blocks, verifySeals bool) (int, error) {
	if atomic.LoadInt32(&bc.procInterrupt) == 1 {
		return 0, nil
	}
	senderCacher.recoverFromBlocks(types.MakeSigner(bc.chainConfig, chain[0].Number()), chain)
	var (
		stats     = insertStats{startTime: mclock.Now()}
		lastCanon *types.Block
	)
	defer func() {
		if lastCanon != nil && bc.CurrentBlock().Hash() == lastCanon.Hash() {
			bc.chainHeadFeed.Send(ChainHeadEvent{lastCanon})
		}
	}()
	headers := make([]*types.Header, len(chain))
	seals := make([]bool, len(chain))
	for i, block := range chain {
		headers[i] = block.Header()
		seals[i] = verifySeals
	}
	abort, results := bc.engine.VerifyHeaders(bc, headers, seals)
	defer close(abort)
	it := newInsertIterator(chain, results, bc.validator)
	block, err := it.next()
	if err == ErrKnownBlock {
		var (
			current  = bc.CurrentBlock()
			localTd  = bc.GetTd(current.Hash(), current.NumberU64())
			externTd = bc.GetTd(block.ParentHash(), block.NumberU64()-1) 
		)
		for block != nil && err == ErrKnownBlock {
			externTd = new(big.Int).Add(externTd, block.Difficulty())
			if localTd.Cmp(externTd) < 0 {
				break
			}
			log.Debug("Ignoring already known block", "number", block.Number(), "hash", block.Hash())
			stats.ignored++
			block, err = it.next()
		}
		for block != nil && err == ErrKnownBlock {
			log.Debug("Writing previously known block", "number", block.Number(), "hash", block.Hash())
			if err := bc.writeKnownBlock(block); err != nil {
				return it.index, err
			}
			lastCanon = block
			block, err = it.next()
		}
	}
	switch {
	case err == consensus.ErrPrunedAncestor:
		log.Debug("Pruned ancestor, inserting as sidechain", "number", block.Number(), "hash", block.Hash())
		return bc.insertSideChain(block, it)
	case err == consensus.ErrFutureBlock || (err == consensus.ErrUnknownAncestor && bc.futureBlocks.Contains(it.first().ParentHash())):
		for block != nil && (it.index == 0 || err == consensus.ErrUnknownAncestor) {
			log.Debug("Future block, postponing import", "number", block.Number(), "hash", block.Hash())
			if err := bc.addFutureBlock(block); err != nil {
				return it.index, err
			}
			block, err = it.next()
		}
		stats.queued += it.processed()
		stats.ignored += it.remaining()
		return it.index, err
	case err != nil:
		bc.futureBlocks.Remove(block.Hash())
		stats.ignored += len(it.chain)
		bc.reportBlock(block, nil, err)
		return it.index, err
	}
	for ; block != nil && err == nil || err == ErrKnownBlock; block, err = it.next() {
		if bc.insertStopped() {
			log.Debug("Abort during block processing")
			break
		}
		if BadHashes[block.Hash()] {
			bc.reportBlock(block, nil, ErrBlacklistedHash)
			return it.index, ErrBlacklistedHash
		}
		if err == ErrKnownBlock {
			logger := log.Debug
			if bc.chainConfig.Clique == nil {
				logger = log.Warn
			}
			logger("Inserted known block", "number", block.Number(), "hash", block.Hash(),
				"uncles", len(block.Uncles()), "txs", len(block.Transactions()), "gas", block.GasUsed(),
				"root", block.Root())
			if len(block.Transactions()) == 0 {
				rawdb.WriteReceipts(bc.db, block.Hash(), block.NumberU64(), nil)
			} else {
				log.Error("Please file an issue, skip known block execution without receipt",
					"hash", block.Hash(), "number", block.NumberU64())
			}
			if err := bc.writeKnownBlock(block); err != nil {
				return it.index, err
			}
			stats.processed++
			lastCanon = block
			continue
		}
		start := time.Now()
		parent := it.previous()
		if parent == nil {
			parent = bc.GetHeader(block.ParentHash(), block.NumberU64()-1)
		}
		statedb, err := state.New(parent.Root, bc.stateCache, bc.snaps)
		if err != nil {
			return it.index, err
		}
		var followupInterrupt uint32
		if !bc.cacheConfig.TrieCleanNoPrefetch {
			if followup, err := it.peek(); followup != nil && err == nil {
				throwaway, _ := state.New(parent.Root, bc.stateCache, bc.snaps)
				go func(start time.Time, followup *types.Block, throwaway *state.StateDB, interrupt *uint32) {
					bc.prefetcher.Prefetch(followup, throwaway, bc.vmConfig, &followupInterrupt)
					blockPrefetchExecuteTimer.Update(time.Since(start))
					if atomic.LoadUint32(interrupt) == 1 {
						blockPrefetchInterruptMeter.Mark(1)
					}
				}(time.Now(), followup, throwaway, &followupInterrupt)
			}
		}
		substart := time.Now()
		receipts, logs, usedGas, err := bc.processor.Process(block, statedb, bc.vmConfig)
		if err != nil {
			bc.reportBlock(block, receipts, err)
			atomic.StoreUint32(&followupInterrupt, 1)
			return it.index, err
		}
		accountReadTimer.Update(statedb.AccountReads)                 
		storageReadTimer.Update(statedb.StorageReads)                 
		accountUpdateTimer.Update(statedb.AccountUpdates)             
		storageUpdateTimer.Update(statedb.StorageUpdates)             
		snapshotAccountReadTimer.Update(statedb.SnapshotAccountReads) 
		snapshotStorageReadTimer.Update(statedb.SnapshotStorageReads) 
		triehash := statedb.AccountHashes + statedb.StorageHashes 
		trieproc := statedb.SnapshotAccountReads + statedb.AccountReads + statedb.AccountUpdates
		trieproc += statedb.SnapshotStorageReads + statedb.StorageReads + statedb.StorageUpdates
		blockExecutionTimer.Update(time.Since(substart) - trieproc - triehash)
		substart = time.Now()
		if err := bc.validator.ValidateState(block, statedb, receipts, usedGas); err != nil {
			bc.reportBlock(block, receipts, err)
			atomic.StoreUint32(&followupInterrupt, 1)
			return it.index, err
		}
		proctime := time.Since(start)
		accountHashTimer.Update(statedb.AccountHashes) 
		storageHashTimer.Update(statedb.StorageHashes) 
		blockValidationTimer.Update(time.Since(substart) - (statedb.AccountHashes + statedb.StorageHashes - triehash))
		substart = time.Now()
		status, err := bc.writeBlockWithState(block, receipts, logs, statedb, false)
		atomic.StoreUint32(&followupInterrupt, 1)
		if err != nil {
			return it.index, err
		}
		accountCommitTimer.Update(statedb.AccountCommits)   
		storageCommitTimer.Update(statedb.StorageCommits)   
		snapshotCommitTimer.Update(statedb.SnapshotCommits) 
		blockWriteTimer.Update(time.Since(substart) - statedb.AccountCommits - statedb.StorageCommits - statedb.SnapshotCommits)
		blockInsertTimer.UpdateSince(start)
		switch status {
		case CanonStatTy:
			log.Debug("Inserted new block", "number", block.Number(), "hash", block.Hash(),
				"uncles", len(block.Uncles()), "txs", len(block.Transactions()), "gas", block.GasUsed(),
				"elapsed", common.PrettyDuration(time.Since(start)),
				"root", block.Root())
			lastCanon = block
			bc.gcproc += proctime
		case SideStatTy:
			log.Debug("Inserted forked block", "number", block.Number(), "hash", block.Hash(),
				"diff", block.Difficulty(), "elapsed", common.PrettyDuration(time.Since(start)),
				"txs", len(block.Transactions()), "gas", block.GasUsed(), "uncles", len(block.Uncles()),
				"root", block.Root())
		default:
			log.Warn("Inserted block with unknown status", "number", block.Number(), "hash", block.Hash(),
				"diff", block.Difficulty(), "elapsed", common.PrettyDuration(time.Since(start)),
				"txs", len(block.Transactions()), "gas", block.GasUsed(), "uncles", len(block.Uncles()),
				"root", block.Root())
		}
		stats.processed++
		stats.usedGas += usedGas
		dirty, _ := bc.stateCache.TrieDB().Size()
		stats.report(chain, it.index, dirty)
	}
	if block != nil && err == consensus.ErrFutureBlock {
		if err := bc.addFutureBlock(block); err != nil {
			return it.index, err
		}
		block, err = it.next()
		for ; block != nil && err == consensus.ErrUnknownAncestor; block, err = it.next() {
			if err := bc.addFutureBlock(block); err != nil {
				return it.index, err
			}
			stats.queued++
		}
	}
	stats.ignored += it.remaining()
	return it.index, err
}
func (bc *BlockChain) insertSideChain(block *types.Block, it *insertIterator) (int, error) {
	var (
		externTd *big.Int
		current  = bc.CurrentBlock()
	)
	err := consensus.ErrPrunedAncestor
	for ; block != nil && (err == consensus.ErrPrunedAncestor); block, err = it.next() {
		if number := block.NumberU64(); current.NumberU64() >= number {
			canonical := bc.GetBlockByNumber(number)
			if canonical != nil && canonical.Hash() == block.Hash() {
				externTd = bc.GetTd(block.Hash(), block.NumberU64())
				continue
			}
			if canonical != nil && canonical.Root() == block.Root() {
				log.Warn("Sidechain ghost-state attack detected", "number", block.NumberU64(), "sideroot", block.Root(), "canonroot", canonical.Root())
				return it.index, errors.New("sidechain ghost-state attack")
			}
		}
		if externTd == nil {
			externTd = bc.GetTd(block.ParentHash(), block.NumberU64()-1)
		}
		externTd = new(big.Int).Add(externTd, block.Difficulty())
		if !bc.HasBlock(block.Hash(), block.NumberU64()) {
			start := time.Now()
			if err := bc.writeBlockWithoutState(block, externTd); err != nil {
				return it.index, err
			}
			log.Debug("Injected sidechain block", "number", block.Number(), "hash", block.Hash(),
				"diff", block.Difficulty(), "elapsed", common.PrettyDuration(time.Since(start)),
				"txs", len(block.Transactions()), "gas", block.GasUsed(), "uncles", len(block.Uncles()),
				"root", block.Root())
		}
	}
	localTd := bc.GetTd(current.Hash(), current.NumberU64())
	if localTd.Cmp(externTd) > 0 {
		log.Info("Sidechain written to disk", "start", it.first().NumberU64(), "end", it.previous().Number, "sidetd", externTd, "localtd", localTd)
		return it.index, err
	}
	var (
		hashes  []common.Hash
		numbers []uint64
	)
	parent := it.previous()
	for parent != nil && !bc.HasState(parent.Root) {
		hashes = append(hashes, parent.Hash())
		numbers = append(numbers, parent.Number.Uint64())
		parent = bc.GetHeader(parent.ParentHash, parent.Number.Uint64()-1)
	}
	if parent == nil {
		return it.index, errors.New("missing parent")
	}
	var (
		blocks []*types.Block
		memory common.StorageSize
	)
	for i := len(hashes) - 1; i >= 0; i-- {
		block := bc.GetBlock(hashes[i], numbers[i])
		blocks = append(blocks, block)
		memory += block.Size()
		if len(blocks) >= 2048 || memory > 64*1024*1024 {
			log.Info("Importing heavy sidechain segment", "blocks", len(blocks), "start", blocks[0].NumberU64(), "end", block.NumberU64())
			if _, err := bc.insertChain(blocks, false); err != nil {
				return 0, err
			}
			blocks, memory = blocks[:0], 0
			if bc.insertStopped() {
				log.Debug("Abort during blocks processing")
				return 0, nil
			}
		}
	}
	if len(blocks) > 0 {
		log.Info("Importing sidechain segment", "start", blocks[0].NumberU64(), "end", blocks[len(blocks)-1].NumberU64())
		return bc.insertChain(blocks, false)
	}
	return 0, nil
}
func (bc *BlockChain) reorg(oldBlock, newBlock *types.Block) error {
	var (
		newChain    types.Blocks
		oldChain    types.Blocks
		commonBlock *types.Block
		deletedTxs types.Transactions
		addedTxs   types.Transactions
		deletedLogs [][]*types.Log
		rebirthLogs [][]*types.Log
		collectLogs = func(hash common.Hash, removed bool) {
			number := bc.hc.GetBlockNumber(hash)
			if number == nil {
				return
			}
			receipts := rawdb.ReadReceipts(bc.db, hash, *number, bc.chainConfig)
			var logs []*types.Log
			for _, receipt := range receipts {
				for _, log := range receipt.Logs {
					l := *log
					if removed {
						l.Removed = true
					} else {
					}
					logs = append(logs, &l)
				}
			}
			if len(logs) > 0 {
				if removed {
					deletedLogs = append(deletedLogs, logs)
				} else {
					rebirthLogs = append(rebirthLogs, logs)
				}
			}
		}
		mergeLogs = func(logs [][]*types.Log, reverse bool) []*types.Log {
			var ret []*types.Log
			if reverse {
				for i := len(logs) - 1; i >= 0; i-- {
					ret = append(ret, logs[i]...)
				}
			} else {
				for i := 0; i < len(logs); i++ {
					ret = append(ret, logs[i]...)
				}
			}
			return ret
		}
	)
	if oldBlock.NumberU64() > newBlock.NumberU64() {
		for ; oldBlock != nil && oldBlock.NumberU64() != newBlock.NumberU64(); oldBlock = bc.GetBlock(oldBlock.ParentHash(), oldBlock.NumberU64()-1) {
			oldChain = append(oldChain, oldBlock)
			deletedTxs = append(deletedTxs, oldBlock.Transactions()...)
			collectLogs(oldBlock.Hash(), true)
		}
	} else {
		for ; newBlock != nil && newBlock.NumberU64() != oldBlock.NumberU64(); newBlock = bc.GetBlock(newBlock.ParentHash(), newBlock.NumberU64()-1) {
			newChain = append(newChain, newBlock)
		}
	}
	if oldBlock == nil {
		return fmt.Errorf("invalid old chain")
	}
	if newBlock == nil {
		return fmt.Errorf("invalid new chain")
	}
	for {
		if oldBlock.Hash() == newBlock.Hash() {
			commonBlock = oldBlock
			break
		}
		oldChain = append(oldChain, oldBlock)
		deletedTxs = append(deletedTxs, oldBlock.Transactions()...)
		collectLogs(oldBlock.Hash(), true)
		newChain = append(newChain, newBlock)
		oldBlock = bc.GetBlock(oldBlock.ParentHash(), oldBlock.NumberU64()-1)
		if oldBlock == nil {
			return fmt.Errorf("invalid old chain")
		}
		newBlock = bc.GetBlock(newBlock.ParentHash(), newBlock.NumberU64()-1)
		if newBlock == nil {
			return fmt.Errorf("invalid new chain")
		}
	}
	if len(oldChain) > 0 && len(newChain) > 0 {
		logFn := log.Info
		msg := "Chain reorg detected"
		if len(oldChain) > 63 {
			msg = "Large chain reorg detected"
			logFn = log.Warn
		}
		logFn(msg, "number", commonBlock.Number(), "hash", commonBlock.Hash(),
			"drop", len(oldChain), "dropfrom", oldChain[0].Hash(), "add", len(newChain), "addfrom", newChain[0].Hash())
		blockReorgAddMeter.Mark(int64(len(newChain)))
		blockReorgDropMeter.Mark(int64(len(oldChain)))
	} else {
		log.Error("Impossible reorg, please file an issue", "oldnum", oldBlock.Number(), "oldhash", oldBlock.Hash(), "newnum", newBlock.Number(), "newhash", newBlock.Hash())
	}
	for i := len(newChain) - 1; i >= 1; i-- {
		bc.writeHeadBlock(newChain[i])
		collectLogs(newChain[i].Hash(), false)
		addedTxs = append(addedTxs, newChain[i].Transactions()...)
	}
	indexesBatch := bc.db.NewBatch()
	for _, tx := range types.TxDifference(deletedTxs, addedTxs) {
		rawdb.DeleteTxLookupEntry(indexesBatch, tx.Hash())
	}
	number := bc.CurrentBlock().NumberU64()
	for i := number + 1; ; i++ {
		hash := rawdb.ReadCanonicalHash(bc.db, i)
		if hash == (common.Hash{}) {
			break
		}
		rawdb.DeleteCanonicalHash(indexesBatch, i)
	}
	if err := indexesBatch.Write(); err != nil {
		log.Crit("Failed to delete useless indexes", "err", err)
	}
	if len(deletedLogs) > 0 {
		bc.rmLogsFeed.Send(RemovedLogsEvent{mergeLogs(deletedLogs, true)})
	}
	if len(rebirthLogs) > 0 {
		bc.logsFeed.Send(mergeLogs(rebirthLogs, false))
	}
	if len(oldChain) > 0 {
		for i := len(oldChain) - 1; i >= 0; i-- {
			bc.chainSideFeed.Send(ChainSideEvent{Block: oldChain[i]})
		}
	}
	return nil
}
func (bc *BlockChain) update() {
	futureTimer := time.NewTicker(5 * time.Second)
	defer futureTimer.Stop()
	for {
		select {
		case <-futureTimer.C:
			bc.procFutureBlocks()
		case <-bc.quit:
			return
		}
	}
}
func (bc *BlockChain) maintainTxIndex(ancients uint64) {
	if ancients > 0 {
		var from = uint64(0)
		if bc.txLookupLimit != 0 && ancients > bc.txLookupLimit {
			from = ancients - bc.txLookupLimit
		}
		rawdb.IndexTransactions(bc.db, from, ancients)
	}
	indexBlocks := func(tail *uint64, head uint64, done chan struct{}) {
		defer func() { done <- struct{}{} }()
		if tail == nil {
			if bc.txLookupLimit == 0 || head < bc.txLookupLimit {
				rawdb.WriteTxIndexTail(bc.db, 0)
			} else {
				rawdb.UnindexTransactions(bc.db, 0, head-bc.txLookupLimit+1)
			}
			return
		}
		if bc.txLookupLimit == 0 || head < bc.txLookupLimit {
			if *tail > 0 {
				rawdb.IndexTransactions(bc.db, 0, *tail)
			}
			return
		}
		if head-bc.txLookupLimit+1 < *tail {
			rawdb.IndexTransactions(bc.db, head-bc.txLookupLimit+1, *tail)
		} else {
			rawdb.UnindexTransactions(bc.db, *tail, head-bc.txLookupLimit+1)
		}
	}
	var (
		done   chan struct{}                  
		headCh = make(chan ChainHeadEvent, 1) 
	)
	sub := bc.SubscribeChainHeadEvent(headCh)
	if sub == nil {
		return
	}
	defer sub.Unsubscribe()
	for {
		select {
		case head := <-headCh:
			if done == nil {
				done = make(chan struct{})
				go indexBlocks(rawdb.ReadTxIndexTail(bc.db), head.Block.NumberU64(), done)
			}
		case <-done:
			done = nil
		case <-bc.quit:
			return
		}
	}
}
func (bc *BlockChain) BadBlocks() []*types.Block {
	blocks := make([]*types.Block, 0, bc.badBlocks.Len())
	for _, hash := range bc.badBlocks.Keys() {
		if blk, exist := bc.badBlocks.Peek(hash); exist {
			block := blk.(*types.Block)
			blocks = append(blocks, block)
		}
	}
	return blocks
}
func (bc *BlockChain) addBadBlock(block *types.Block) {
	bc.badBlocks.Add(block.Hash(), block)
}
func (bc *BlockChain) reportBlock(block *types.Block, receipts types.Receipts, err error) {
	bc.addBadBlock(block)
	var receiptString string
	for i, receipt := range receipts {
		receiptString += fmt.Sprintf("\t %d: cumulative: %v gas: %v contract: %v status: %v tx: %v logs: %v bloom: %x state: %x\n",
			i, receipt.CumulativeGasUsed, receipt.GasUsed, receipt.ContractAddress.Hex(),
			receipt.Status, receipt.TxHash.Hex(), receipt.Logs, receipt.Bloom, receipt.PostState)
	}
	log.Error(fmt.Sprintf(`
########## BAD BLOCK #########
Chain config: %v
Number: %v
Hash: 0x%x
%v
Error: %v
##############################
`, bc.chainConfig, block.Number(), block.Hash(), receiptString, err))
}
func (bc *BlockChain) InsertHeaderChain(chain []*types.Header, checkFreq int) (int, error) {
	start := time.Now()
	if i, err := bc.hc.ValidateHeaderChain(chain, checkFreq); err != nil {
		return i, err
	}
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()
	bc.wg.Add(1)
	defer bc.wg.Done()
	whFunc := func(header *types.Header) error {
		_, err := bc.hc.WriteHeader(header)
		return err
	}
	return bc.hc.InsertHeaderChain(chain, whFunc, start)
}
func (bc *BlockChain) CurrentHeader() *types.Header {
	return bc.hc.CurrentHeader()
}
func (bc *BlockChain) GetTd(hash common.Hash, number uint64) *big.Int {
	return bc.hc.GetTd(hash, number)
}
func (bc *BlockChain) GetTdByHash(hash common.Hash) *big.Int {
	return bc.hc.GetTdByHash(hash)
}
func (bc *BlockChain) GetHeader(hash common.Hash, number uint64) *types.Header {
	return bc.hc.GetHeader(hash, number)
}
func (bc *BlockChain) GetHeaderByHash(hash common.Hash) *types.Header {
	return bc.hc.GetHeaderByHash(hash)
}
func (bc *BlockChain) HasHeader(hash common.Hash, number uint64) bool {
	return bc.hc.HasHeader(hash, number)
}
func (bc *BlockChain) GetCanonicalHash(number uint64) common.Hash {
	return bc.hc.GetCanonicalHash(number)
}
func (bc *BlockChain) GetBlockHashesFromHash(hash common.Hash, max uint64) []common.Hash {
	return bc.hc.GetBlockHashesFromHash(hash, max)
}
func (bc *BlockChain) GetAncestor(hash common.Hash, number, ancestor uint64, maxNonCanonical *uint64) (common.Hash, uint64) {
	return bc.hc.GetAncestor(hash, number, ancestor, maxNonCanonical)
}
func (bc *BlockChain) GetHeaderByNumber(number uint64) *types.Header {
	return bc.hc.GetHeaderByNumber(number)
}
func (bc *BlockChain) GetTransactionLookup(hash common.Hash) *rawdb.LegacyTxLookupEntry {
	if lookup, exist := bc.txLookupCache.Get(hash); exist {
		return lookup.(*rawdb.LegacyTxLookupEntry)
	}
	tx, blockHash, blockNumber, txIndex := rawdb.ReadTransaction(bc.db, hash)
	if tx == nil {
		return nil
	}
	lookup := &rawdb.LegacyTxLookupEntry{BlockHash: blockHash, BlockIndex: blockNumber, Index: txIndex}
	bc.txLookupCache.Add(hash, lookup)
	return lookup
}
func (bc *BlockChain) Config() *params.ChainConfig { return bc.chainConfig }
func (bc *BlockChain) Engine() consensus.Engine { return bc.engine }
func (bc *BlockChain) SubscribeRemovedLogsEvent(ch chan<- RemovedLogsEvent) event.Subscription {
	return bc.scope.Track(bc.rmLogsFeed.Subscribe(ch))
}
func (bc *BlockChain) SubscribeChainEvent(ch chan<- ChainEvent) event.Subscription {
	return bc.scope.Track(bc.chainFeed.Subscribe(ch))
}
func (bc *BlockChain) SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription {
	return bc.scope.Track(bc.chainHeadFeed.Subscribe(ch))
}
func (bc *BlockChain) SubscribeChainSideEvent(ch chan<- ChainSideEvent) event.Subscription {
	return bc.scope.Track(bc.chainSideFeed.Subscribe(ch))
}
func (bc *BlockChain) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return bc.scope.Track(bc.logsFeed.Subscribe(ch))
}
func (bc *BlockChain) SubscribeBlockProcessingEvent(ch chan<- bool) event.Subscription {
	return bc.scope.Track(bc.blockProcFeed.Subscribe(ch))
}
