package rawdb
import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/ethdb"
	"github.com/Cryptochain-VON/ethdb/leveldb"
	"github.com/Cryptochain-VON/ethdb/memorydb"
	"github.com/Cryptochain-VON/log"
	"github.com/olekukonko/tablewriter"
)
type freezerdb struct {
	ethdb.KeyValueStore
	ethdb.AncientStore
}
func (frdb *freezerdb) Close() error {
	var errs []error
	if err := frdb.AncientStore.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := frdb.KeyValueStore.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
type nofreezedb struct {
	ethdb.KeyValueStore
}
func (db *nofreezedb) HasAncient(kind string, number uint64) (bool, error) {
	return false, errNotSupported
}
func (db *nofreezedb) Ancient(kind string, number uint64) ([]byte, error) {
	return nil, errNotSupported
}
func (db *nofreezedb) Ancients() (uint64, error) {
	return 0, errNotSupported
}
func (db *nofreezedb) AncientSize(kind string) (uint64, error) {
	return 0, errNotSupported
}
func (db *nofreezedb) AppendAncient(number uint64, hash, header, body, receipts, td []byte) error {
	return errNotSupported
}
func (db *nofreezedb) TruncateAncients(items uint64) error {
	return errNotSupported
}
func (db *nofreezedb) Sync() error {
	return errNotSupported
}
func NewDatabase(db ethdb.KeyValueStore) ethdb.Database {
	return &nofreezedb{
		KeyValueStore: db,
	}
}
func NewDatabaseWithFreezer(db ethdb.KeyValueStore, freezer string, namespace string) (ethdb.Database, error) {
	frdb, err := newFreezer(freezer, namespace)
	if err != nil {
		return nil, err
	}
	if kvgenesis, _ := db.Get(headerHashKey(0)); len(kvgenesis) > 0 {
		if frozen, _ := frdb.Ancients(); frozen > 0 {
			if frgenesis, _ := frdb.Ancient(freezerHashTable, 0); !bytes.Equal(kvgenesis, frgenesis) {
				return nil, fmt.Errorf("genesis mismatch: %#x (leveldb) != %#x (ancients)", kvgenesis, frgenesis)
			}
			if kvhash, _ := db.Get(headerHashKey(frozen)); len(kvhash) == 0 {
				if *ReadHeaderNumber(db, ReadHeadHeaderHash(db)) > frozen-1 {
					return nil, fmt.Errorf("gap (#%d) in the chain between ancients and leveldb", frozen)
				}
			}
		} else {
			if ReadHeadHeaderHash(db) != common.BytesToHash(kvgenesis) {
				if kvblob, _ := db.Get(headerHashKey(1)); len(kvblob) == 0 {
					return nil, errors.New("ancient chain segments already extracted, please set --datadir.ancient to the correct path")
				}
			}
		}
	}
	go frdb.freeze(db)
	return &freezerdb{
		KeyValueStore: db,
		AncientStore:  frdb,
	}, nil
}
func NewMemoryDatabase() ethdb.Database {
	return NewDatabase(memorydb.New())
}
func NewMemoryDatabaseWithCap(size int) ethdb.Database {
	return NewDatabase(memorydb.NewWithCap(size))
}
func NewLevelDBDatabase(file string, cache int, handles int, namespace string) (ethdb.Database, error) {
	db, err := leveldb.New(file, cache, handles, namespace)
	if err != nil {
		return nil, err
	}
	return NewDatabase(db), nil
}
func NewLevelDBDatabaseWithFreezer(file string, cache int, handles int, freezer string, namespace string) (ethdb.Database, error) {
	kvdb, err := leveldb.New(file, cache, handles, namespace)
	if err != nil {
		return nil, err
	}
	frdb, err := NewDatabaseWithFreezer(kvdb, freezer, namespace)
	if err != nil {
		kvdb.Close()
		return nil, err
	}
	return frdb, nil
}
func InspectDatabase(db ethdb.Database) error {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	var (
		count  int64
		start  = time.Now()
		logged = time.Now()
		total           common.StorageSize
		headerSize      common.StorageSize
		bodySize        common.StorageSize
		receiptSize     common.StorageSize
		tdSize          common.StorageSize
		numHashPairing  common.StorageSize
		hashNumPairing  common.StorageSize
		trieSize        common.StorageSize
		txlookupSize    common.StorageSize
		accountSnapSize common.StorageSize
		storageSnapSize common.StorageSize
		preimageSize    common.StorageSize
		bloomBitsSize   common.StorageSize
		cliqueSnapsSize common.StorageSize
		ancientHeaders  common.StorageSize
		ancientBodies   common.StorageSize
		ancientReceipts common.StorageSize
		ancientHashes   common.StorageSize
		ancientTds      common.StorageSize
		chtTrieNodes   common.StorageSize
		bloomTrieNodes common.StorageSize
		metadata    common.StorageSize
		unaccounted common.StorageSize
	)
	for it.Next() {
		var (
			key  = it.Key()
			size = common.StorageSize(len(key) + len(it.Value()))
		)
		total += size
		switch {
		case bytes.HasPrefix(key, headerPrefix) && bytes.HasSuffix(key, headerTDSuffix):
			tdSize += size
		case bytes.HasPrefix(key, headerPrefix) && bytes.HasSuffix(key, headerHashSuffix):
			numHashPairing += size
		case bytes.HasPrefix(key, headerPrefix) && len(key) == (len(headerPrefix)+8+common.HashLength):
			headerSize += size
		case bytes.HasPrefix(key, headerNumberPrefix) && len(key) == (len(headerNumberPrefix)+common.HashLength):
			hashNumPairing += size
		case bytes.HasPrefix(key, blockBodyPrefix) && len(key) == (len(blockBodyPrefix)+8+common.HashLength):
			bodySize += size
		case bytes.HasPrefix(key, blockReceiptsPrefix) && len(key) == (len(blockReceiptsPrefix)+8+common.HashLength):
			receiptSize += size
		case bytes.HasPrefix(key, txLookupPrefix) && len(key) == (len(txLookupPrefix)+common.HashLength):
			txlookupSize += size
		case bytes.HasPrefix(key, SnapshotAccountPrefix) && len(key) == (len(SnapshotAccountPrefix)+common.HashLength):
			accountSnapSize += size
		case bytes.HasPrefix(key, SnapshotStoragePrefix) && len(key) == (len(SnapshotStoragePrefix)+2*common.HashLength):
			storageSnapSize += size
		case bytes.HasPrefix(key, preimagePrefix) && len(key) == (len(preimagePrefix)+common.HashLength):
			preimageSize += size
		case bytes.HasPrefix(key, bloomBitsPrefix) && len(key) == (len(bloomBitsPrefix)+10+common.HashLength):
			bloomBitsSize += size
		case bytes.HasPrefix(key, []byte("clique-")) && len(key) == 7+common.HashLength:
			cliqueSnapsSize += size
		case bytes.HasPrefix(key, []byte("cht-")) && len(key) == 4+common.HashLength:
			chtTrieNodes += size
		case bytes.HasPrefix(key, []byte("blt-")) && len(key) == 4+common.HashLength:
			bloomTrieNodes += size
		case len(key) == common.HashLength:
			trieSize += size
		default:
			var accounted bool
			for _, meta := range [][]byte{databaseVerisionKey, headHeaderKey, headBlockKey, headFastBlockKey, fastTrieProgressKey} {
				if bytes.Equal(key, meta) {
					metadata += size
					accounted = true
					break
				}
			}
			if !accounted {
				unaccounted += size
			}
		}
		count += 1
		if count%1000 == 0 && time.Since(logged) > 8*time.Second {
			log.Info("Inspecting database", "count", count, "elapsed", common.PrettyDuration(time.Since(start)))
			logged = time.Now()
		}
	}
	ancients := []*common.StorageSize{&ancientHeaders, &ancientBodies, &ancientReceipts, &ancientHashes, &ancientTds}
	for i, category := range []string{freezerHeaderTable, freezerBodiesTable, freezerReceiptTable, freezerHashTable, freezerDifficultyTable} {
		if size, err := db.AncientSize(category); err == nil {
			*ancients[i] += common.StorageSize(size)
			total += common.StorageSize(size)
		}
	}
	stats := [][]string{
		{"Key-Value store", "Headers", headerSize.String()},
		{"Key-Value store", "Bodies", bodySize.String()},
		{"Key-Value store", "Receipts", receiptSize.String()},
		{"Key-Value store", "Difficulties", tdSize.String()},
		{"Key-Value store", "Block number->hash", numHashPairing.String()},
		{"Key-Value store", "Block hash->number", hashNumPairing.String()},
		{"Key-Value store", "Transaction index", txlookupSize.String()},
		{"Key-Value store", "Bloombit index", bloomBitsSize.String()},
		{"Key-Value store", "Trie nodes", trieSize.String()},
		{"Key-Value store", "Trie preimages", preimageSize.String()},
		{"Key-Value store", "Account snapshot", accountSnapSize.String()},
		{"Key-Value store", "Storage snapshot", storageSnapSize.String()},
		{"Key-Value store", "Clique snapshots", cliqueSnapsSize.String()},
		{"Key-Value store", "Singleton metadata", metadata.String()},
		{"Ancient store", "Headers", ancientHeaders.String()},
		{"Ancient store", "Bodies", ancientBodies.String()},
		{"Ancient store", "Receipts", ancientReceipts.String()},
		{"Ancient store", "Difficulties", ancientTds.String()},
		{"Ancient store", "Block number->hash", ancientHashes.String()},
		{"Light client", "CHT trie nodes", chtTrieNodes.String()},
		{"Light client", "Bloom trie nodes", bloomTrieNodes.String()},
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Database", "Category", "Size"})
	table.SetFooter([]string{"", "Total", total.String()})
	table.AppendBulk(stats)
	table.Render()
	if unaccounted > 0 {
		log.Error("Database contains unaccounted data", "size", unaccounted)
	}
	return nil
}
