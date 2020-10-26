package core
import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	mrand "math/rand"
	"sync/atomic"
	"time"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/consensus"
	"github.com/Cryptochain-VON/core/rawdb"
	"github.com/Cryptochain-VON/core/types"
	"github.com/Cryptochain-VON/ethdb"
	"github.com/Cryptochain-VON/log"
	"github.com/Cryptochain-VON/params"
	lru "github.com/hashicorp/golang-lru"
)
const (
	headerCacheLimit = 512
	tdCacheLimit     = 1024
	numberCacheLimit = 2048
)
type HeaderChain struct {
	config *params.ChainConfig
	chainDb       ethdb.Database
	genesisHeader *types.Header
	currentHeader     atomic.Value 
	currentHeaderHash common.Hash  
	headerCache *lru.Cache 
	tdCache     *lru.Cache 
	numberCache *lru.Cache 
	procInterrupt func() bool
	rand   *mrand.Rand
	engine consensus.Engine
}
func NewHeaderChain(chainDb ethdb.Database, config *params.ChainConfig, engine consensus.Engine, procInterrupt func() bool) (*HeaderChain, error) {
	headerCache, _ := lru.New(headerCacheLimit)
	tdCache, _ := lru.New(tdCacheLimit)
	numberCache, _ := lru.New(numberCacheLimit)
	seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	hc := &HeaderChain{
		config:        config,
		chainDb:       chainDb,
		headerCache:   headerCache,
		tdCache:       tdCache,
		numberCache:   numberCache,
		procInterrupt: procInterrupt,
		rand:          mrand.New(mrand.NewSource(seed.Int64())),
		engine:        engine,
	}
	hc.genesisHeader = hc.GetHeaderByNumber(0)
	if hc.genesisHeader == nil {
		return nil, ErrNoGenesis
	}
	hc.currentHeader.Store(hc.genesisHeader)
	if head := rawdb.ReadHeadBlockHash(chainDb); head != (common.Hash{}) {
		if chead := hc.GetHeaderByHash(head); chead != nil {
			hc.currentHeader.Store(chead)
		}
	}
	hc.currentHeaderHash = hc.CurrentHeader().Hash()
	headHeaderGauge.Update(hc.CurrentHeader().Number.Int64())
	return hc, nil
}
func (hc *HeaderChain) GetBlockNumber(hash common.Hash) *uint64 {
	if cached, ok := hc.numberCache.Get(hash); ok {
		number := cached.(uint64)
		return &number
	}
	number := rawdb.ReadHeaderNumber(hc.chainDb, hash)
	if number != nil {
		hc.numberCache.Add(hash, *number)
	}
	return number
}
func (hc *HeaderChain) WriteHeader(header *types.Header) (status WriteStatus, err error) {
	var (
		hash   = header.Hash()
		number = header.Number.Uint64()
	)
	ptd := hc.GetTd(header.ParentHash, number-1)
	if ptd == nil {
		return NonStatTy, consensus.ErrUnknownAncestor
	}
	head := hc.CurrentHeader().Number.Uint64()
	localTd := hc.GetTd(hc.currentHeaderHash, head)
	externTd := new(big.Int).Add(header.Difficulty, ptd)
	headerBatch := hc.chainDb.NewBatch()
	rawdb.WriteTd(headerBatch, hash, number, externTd)
	rawdb.WriteHeader(headerBatch, header)
	if err := headerBatch.Write(); err != nil {
		log.Crit("Failed to write header into disk", "err", err)
	}
	reorg := externTd.Cmp(localTd) > 0
	if !reorg && externTd.Cmp(localTd) == 0 {
		if header.Number.Uint64() < head {
			reorg = true
		} else if header.Number.Uint64() == head {
			reorg = mrand.Float64() < 0.5
		}
	}
	if reorg {
		markerBatch := hc.chainDb.NewBatch()
		for i := number + 1; ; i++ {
			hash := rawdb.ReadCanonicalHash(hc.chainDb, i)
			if hash == (common.Hash{}) {
				break
			}
			rawdb.DeleteCanonicalHash(markerBatch, i)
		}
		var (
			headHash   = header.ParentHash
			headNumber = header.Number.Uint64() - 1
			headHeader = hc.GetHeader(headHash, headNumber)
		)
		for rawdb.ReadCanonicalHash(hc.chainDb, headNumber) != headHash {
			rawdb.WriteCanonicalHash(markerBatch, headHash, headNumber)
			headHash = headHeader.ParentHash
			headNumber = headHeader.Number.Uint64() - 1
			headHeader = hc.GetHeader(headHash, headNumber)
		}
		rawdb.WriteCanonicalHash(markerBatch, hash, number)
		rawdb.WriteHeadHeaderHash(markerBatch, hash)
		if err := markerBatch.Write(); err != nil {
			log.Crit("Failed to write header markers into disk", "err", err)
		}
		hc.currentHeaderHash = hash
		hc.currentHeader.Store(types.CopyHeader(header))
		headHeaderGauge.Update(header.Number.Int64())
		status = CanonStatTy
	} else {
		status = SideStatTy
	}
	hc.tdCache.Add(hash, externTd)
	hc.headerCache.Add(hash, header)
	hc.numberCache.Add(hash, number)
	return
}
type WhCallback func(*types.Header) error
func (hc *HeaderChain) ValidateHeaderChain(chain []*types.Header, checkFreq int) (int, error) {
	for i := 1; i < len(chain); i++ {
		if chain[i].Number.Uint64() != chain[i-1].Number.Uint64()+1 || chain[i].ParentHash != chain[i-1].Hash() {
			log.Error("Non contiguous header insert", "number", chain[i].Number, "hash", chain[i].Hash(),
				"parent", chain[i].ParentHash, "prevnumber", chain[i-1].Number, "prevhash", chain[i-1].Hash())
			return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x…], item %d is #%d [%x…] (parent [%x…])", i-1, chain[i-1].Number,
				chain[i-1].Hash().Bytes()[:4], i, chain[i].Number, chain[i].Hash().Bytes()[:4], chain[i].ParentHash[:4])
		}
	}
	seals := make([]bool, len(chain))
	if checkFreq != 0 {
		for i := 0; i < len(seals)/checkFreq; i++ {
			index := i*checkFreq + hc.rand.Intn(checkFreq)
			if index >= len(seals) {
				index = len(seals) - 1
			}
			seals[index] = true
		}
		seals[len(seals)-1] = true
	}
	abort, results := hc.engine.VerifyHeaders(hc, chain, seals)
	defer close(abort)
	for i, header := range chain {
		if hc.procInterrupt() {
			log.Debug("Premature abort during headers verification")
			return 0, errors.New("aborted")
		}
		if BadHashes[header.Hash()] {
			return i, ErrBlacklistedHash
		}
		if err := <-results; err != nil {
			return i, err
		}
	}
	return 0, nil
}
func (hc *HeaderChain) InsertHeaderChain(chain []*types.Header, writeHeader WhCallback, start time.Time) (int, error) {
	stats := struct{ processed, ignored int }{}
	for i, header := range chain {
		if hc.procInterrupt() {
			log.Debug("Premature abort during headers import")
			return i, errors.New("aborted")
		}
		hash := header.Hash()
		if hc.HasHeader(hash, header.Number.Uint64()) {
			externTd := hc.GetTd(hash, header.Number.Uint64())
			localTd := hc.GetTd(hc.currentHeaderHash, hc.CurrentHeader().Number.Uint64())
			if externTd == nil || externTd.Cmp(localTd) <= 0 {
				stats.ignored++
				continue
			}
		}
		if err := writeHeader(header); err != nil {
			return i, err
		}
		stats.processed++
	}
	last := chain[len(chain)-1]
	context := []interface{}{
		"count", stats.processed, "elapsed", common.PrettyDuration(time.Since(start)),
		"number", last.Number, "hash", last.Hash(),
	}
	if timestamp := time.Unix(int64(last.Time), 0); time.Since(timestamp) > time.Minute {
		context = append(context, []interface{}{"age", common.PrettyAge(timestamp)}...)
	}
	if stats.ignored > 0 {
		context = append(context, []interface{}{"ignored", stats.ignored}...)
	}
	log.Info("Imported new block headers", context...)
	return 0, nil
}
func (hc *HeaderChain) GetBlockHashesFromHash(hash common.Hash, max uint64) []common.Hash {
	header := hc.GetHeaderByHash(hash)
	if header == nil {
		return nil
	}
	chain := make([]common.Hash, 0, max)
	for i := uint64(0); i < max; i++ {
		next := header.ParentHash
		if header = hc.GetHeader(next, header.Number.Uint64()-1); header == nil {
			break
		}
		chain = append(chain, next)
		if header.Number.Sign() == 0 {
			break
		}
	}
	return chain
}
func (hc *HeaderChain) GetAncestor(hash common.Hash, number, ancestor uint64, maxNonCanonical *uint64) (common.Hash, uint64) {
	if ancestor > number {
		return common.Hash{}, 0
	}
	if ancestor == 1 {
		if header := hc.GetHeader(hash, number); header != nil {
			return header.ParentHash, number - 1
		} else {
			return common.Hash{}, 0
		}
	}
	for ancestor != 0 {
		if rawdb.ReadCanonicalHash(hc.chainDb, number) == hash {
			ancestorHash := rawdb.ReadCanonicalHash(hc.chainDb, number-ancestor)
			if rawdb.ReadCanonicalHash(hc.chainDb, number) == hash {
				number -= ancestor
				return ancestorHash, number
			}
		}
		if *maxNonCanonical == 0 {
			return common.Hash{}, 0
		}
		*maxNonCanonical--
		ancestor--
		header := hc.GetHeader(hash, number)
		if header == nil {
			return common.Hash{}, 0
		}
		hash = header.ParentHash
		number--
	}
	return hash, number
}
func (hc *HeaderChain) GetTd(hash common.Hash, number uint64) *big.Int {
	if cached, ok := hc.tdCache.Get(hash); ok {
		return cached.(*big.Int)
	}
	td := rawdb.ReadTd(hc.chainDb, hash, number)
	if td == nil {
		return nil
	}
	hc.tdCache.Add(hash, td)
	return td
}
func (hc *HeaderChain) GetTdByHash(hash common.Hash) *big.Int {
	number := hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	return hc.GetTd(hash, *number)
}
func (hc *HeaderChain) GetHeader(hash common.Hash, number uint64) *types.Header {
	if header, ok := hc.headerCache.Get(hash); ok {
		return header.(*types.Header)
	}
	header := rawdb.ReadHeader(hc.chainDb, hash, number)
	if header == nil {
		return nil
	}
	hc.headerCache.Add(hash, header)
	return header
}
func (hc *HeaderChain) GetHeaderByHash(hash common.Hash) *types.Header {
	number := hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	return hc.GetHeader(hash, *number)
}
func (hc *HeaderChain) HasHeader(hash common.Hash, number uint64) bool {
	if hc.numberCache.Contains(hash) || hc.headerCache.Contains(hash) {
		return true
	}
	return rawdb.HasHeader(hc.chainDb, hash, number)
}
func (hc *HeaderChain) GetHeaderByNumber(number uint64) *types.Header {
	hash := rawdb.ReadCanonicalHash(hc.chainDb, number)
	if hash == (common.Hash{}) {
		return nil
	}
	return hc.GetHeader(hash, number)
}
func (hc *HeaderChain) GetCanonicalHash(number uint64) common.Hash {
	return rawdb.ReadCanonicalHash(hc.chainDb, number)
}
func (hc *HeaderChain) CurrentHeader() *types.Header {
	return hc.currentHeader.Load().(*types.Header)
}
func (hc *HeaderChain) SetCurrentHeader(head *types.Header) {
	hc.currentHeader.Store(head)
	hc.currentHeaderHash = head.Hash()
	headHeaderGauge.Update(head.Number.Int64())
}
type (
	UpdateHeadBlocksCallback func(ethdb.KeyValueWriter, *types.Header)
	DeleteBlockContentCallback func(ethdb.KeyValueWriter, common.Hash, uint64)
)
func (hc *HeaderChain) SetHead(head uint64, updateFn UpdateHeadBlocksCallback, delFn DeleteBlockContentCallback) {
	var (
		parentHash common.Hash
		batch      = hc.chainDb.NewBatch()
	)
	for hdr := hc.CurrentHeader(); hdr != nil && hdr.Number.Uint64() > head; hdr = hc.CurrentHeader() {
		hash, num := hdr.Hash(), hdr.Number.Uint64()
		parent := hc.GetHeader(hdr.ParentHash, num-1)
		if parent == nil {
			parent = hc.genesisHeader
		}
		parentHash = hdr.ParentHash
		markerBatch := hc.chainDb.NewBatch()
		if updateFn != nil {
			updateFn(markerBatch, parent)
		}
		rawdb.WriteHeadHeaderHash(markerBatch, parentHash)
		if err := markerBatch.Write(); err != nil {
			log.Crit("Failed to update chain markers", "error", err)
		}
		hc.currentHeader.Store(parent)
		hc.currentHeaderHash = parentHash
		headHeaderGauge.Update(parent.Number.Int64())
		if delFn != nil {
			delFn(batch, hash, num)
		}
		rawdb.DeleteHeader(batch, hash, num)
		rawdb.DeleteTd(batch, hash, num)
		rawdb.DeleteCanonicalHash(batch, num)
	}
	if err := batch.Write(); err != nil {
		log.Crit("Failed to rewind block", "error", err)
	}
	hc.headerCache.Purge()
	hc.tdCache.Purge()
	hc.numberCache.Purge()
}
func (hc *HeaderChain) SetGenesis(head *types.Header) {
	hc.genesisHeader = head
}
func (hc *HeaderChain) Config() *params.ChainConfig { return hc.config }
func (hc *HeaderChain) Engine() consensus.Engine { return hc.engine }
func (hc *HeaderChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}
