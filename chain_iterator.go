package rawdb
import (
	"runtime"
	"sync/atomic"
	"time"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/common/prque"
	"github.com/Cryptochain-VON/ethdb"
	"github.com/Cryptochain-VON/log"
	"github.com/Cryptochain-VON/rlp"
	"golang.org/x/crypto/sha3"
)
func InitDatabaseFromFreezer(db ethdb.Database) {
	frozen, err := db.Ancients()
	if err != nil || frozen == 0 {
		return
	}
	var (
		batch  = db.NewBatch()
		start  = time.Now()
		logged = start.Add(-7 * time.Second) 
		hash   common.Hash
	)
	for i := uint64(0); i < frozen; i++ {
		if h, err := db.Ancient(freezerHashTable, i); err != nil {
			log.Crit("Failed to init database from freezer", "err", err)
		} else {
			hash = common.BytesToHash(h)
		}
		WriteHeaderNumber(batch, hash, i)
		if batch.ValueSize() > ethdb.IdealBatchSize {
			if err := batch.Write(); err != nil {
				log.Crit("Failed to write data to db", "err", err)
			}
			batch.Reset()
		}
		if time.Since(logged) > 8*time.Second {
			log.Info("Initializing database from freezer", "total", frozen, "number", i, "hash", hash, "elapsed", common.PrettyDuration(time.Since(start)))
			logged = time.Now()
		}
	}
	if err := batch.Write(); err != nil {
		log.Crit("Failed to write data to db", "err", err)
	}
	batch.Reset()
	WriteHeadHeaderHash(db, hash)
	WriteHeadFastBlockHash(db, hash)
	log.Info("Initialized database from freezer", "blocks", frozen, "elapsed", common.PrettyDuration(time.Since(start)))
}
type blockTxHashes struct {
	number uint64
	hashes []common.Hash
}
func iterateTransactions(db ethdb.Database, from uint64, to uint64, reverse bool) (chan *blockTxHashes, chan struct{}) {
	type numberRlp struct {
		number uint64
		rlp    rlp.RawValue
	}
	if to == from {
		return nil, nil
	}
	threads := to - from
	if cpus := runtime.NumCPU(); threads > uint64(cpus) {
		threads = uint64(cpus)
	}
	var (
		rlpCh    = make(chan *numberRlp, threads*2)     
		hashesCh = make(chan *blockTxHashes, threads*2) 
		abortCh  = make(chan struct{})
	)
	lookup := func() {
		n, end := from, to
		if reverse {
			n, end = to-1, from-1
		}
		defer close(rlpCh)
		for n != end {
			data := ReadCanonicalBodyRLP(db, n)
			select {
			case rlpCh <- &numberRlp{n, data}:
			case <-abortCh:
				return
			}
			if reverse {
				n--
			} else {
				n++
			}
		}
	}
	nThreadsAlive := int32(threads)
	process := func() {
		defer func() {
			if atomic.AddInt32(&nThreadsAlive, -1) == 0 {
				close(hashesCh)
			}
		}()
		var hasher = sha3.NewLegacyKeccak256()
		for data := range rlpCh {
			it, err := rlp.NewListIterator(data.rlp)
			if err != nil {
				log.Warn("tx iteration error", "error", err)
				return
			}
			it.Next()
			txs := it.Value()
			txIt, err := rlp.NewListIterator(txs)
			if err != nil {
				log.Warn("tx iteration error", "error", err)
				return
			}
			var hashes []common.Hash
			for txIt.Next() {
				if err := txIt.Err(); err != nil {
					log.Warn("tx iteration error", "error", err)
					return
				}
				var txHash common.Hash
				hasher.Reset()
				hasher.Write(txIt.Value())
				hasher.Sum(txHash[:0])
				hashes = append(hashes, txHash)
			}
			result := &blockTxHashes{
				hashes: hashes,
				number: data.number,
			}
			select {
			case hashesCh <- result:
			case <-abortCh:
				return
			}
		}
	}
	go lookup() 
	for i := 0; i < int(threads); i++ {
		go process()
	}
	return hashesCh, abortCh
}
func IndexTransactions(db ethdb.Database, from uint64, to uint64) {
	if from >= to {
		return
	}
	var (
		hashesCh, abortCh = iterateTransactions(db, from, to, true)
		batch             = db.NewBatch()
		start             = time.Now()
		logged            = start.Add(-7 * time.Second)
		lastNum = to
		queue   = prque.New(nil)
		blocks, txs = 0, 0
	)
	defer close(abortCh)
	for chanDelivery := range hashesCh {
		queue.Push(chanDelivery, int64(chanDelivery.number))
		for !queue.Empty() {
			if _, priority := queue.Peek(); priority != int64(lastNum-1) {
				break
			}
			delivery := queue.PopItem().(*blockTxHashes)
			lastNum = delivery.number
			WriteTxLookupEntriesByHash(batch, delivery.number, delivery.hashes)
			blocks++
			txs += len(delivery.hashes)
			if batch.ValueSize() > ethdb.IdealBatchSize {
				WriteTxIndexTail(batch, lastNum)
				if err := batch.Write(); err != nil {
					log.Crit("Failed writing batch to db", "error", err)
					return
				}
				batch.Reset()
			}
			if time.Since(logged) > 8*time.Second {
				log.Info("Indexing transactions", "blocks", blocks, "txs", txs, "tail", lastNum, "total", to-from, "elapsed", common.PrettyDuration(time.Since(start)))
				logged = time.Now()
			}
		}
	}
	if lastNum < to {
		WriteTxIndexTail(batch, lastNum)
		if err := batch.Write(); err != nil {
			log.Crit("Failed writing batch to db", "error", err)
			return
		}
	}
	log.Info("Indexed transactions", "blocks", blocks, "txs", txs, "tail", lastNum, "elapsed", common.PrettyDuration(time.Since(start)))
}
func UnindexTransactions(db ethdb.Database, from uint64, to uint64) {
	if from >= to {
		return
	}
	WriteTxIndexTail(db, to)
	var (
		hashesCh, abortCh = iterateTransactions(db, from, to, false)
		batch             = db.NewBatch()
		start             = time.Now()
		logged            = start.Add(-7 * time.Second)
	)
	defer close(abortCh)
	blocks, txs := 0, 0
	for delivery := range hashesCh {
		DeleteTxLookupEntriesByHash(batch, delivery.hashes)
		txs += len(delivery.hashes)
		blocks++
		if blocks%1000 == 0 {
			if err := batch.Write(); err != nil {
				log.Crit("Failed writing batch to db", "error", err)
				return
			}
			batch.Reset()
		}
		if time.Since(logged) > 8*time.Second {
			log.Info("Unindexing transactions", "blocks", blocks, "txs", txs, "total", to-from, "elapsed", common.PrettyDuration(time.Since(start)))
			logged = time.Now()
		}
	}
	if err := batch.Write(); err != nil {
		log.Crit("Failed writing batch to db", "error", err)
		return
	}
	log.Info("Unindexed transactions", "blocks", blocks, "txs", txs, "tail", to, "elapsed", common.PrettyDuration(time.Since(start)))
}
