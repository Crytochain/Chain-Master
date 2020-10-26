package rawdb
import (
	"encoding/binary"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/metrics"
)
var (
	databaseVerisionKey = []byte("DatabaseVersion")
	headHeaderKey = []byte("LastHeader")
	headBlockKey = []byte("LastBlock")
	headFastBlockKey = []byte("LastFast")
	fastTrieProgressKey = []byte("TrieSync")
	snapshotRootKey = []byte("SnapshotRoot")
	snapshotJournalKey = []byte("SnapshotJournal")
	txIndexTailKey = []byte("TransactionIndexTail")
	fastTxLookupLimitKey = []byte("FastTransactionLookupLimit")
	headerPrefix       = []byte("h") 
	headerTDSuffix     = []byte("t") 
	headerHashSuffix   = []byte("n") 
	headerNumberPrefix = []byte("H") 
	blockBodyPrefix     = []byte("b") 
	blockReceiptsPrefix = []byte("r") 
	txLookupPrefix        = []byte("l") 
	bloomBitsPrefix       = []byte("B") 
	SnapshotAccountPrefix = []byte("a") 
	SnapshotStoragePrefix = []byte("o") 
	preimagePrefix = []byte("secure-key-")      
	configPrefix   = []byte("ethereum-config-") 
	BloomBitsIndexPrefix = []byte("iB") 
	preimageCounter    = metrics.NewRegisteredCounter("db/preimage/total", nil)
	preimageHitCounter = metrics.NewRegisteredCounter("db/preimage/hits", nil)
)
const (
	freezerHeaderTable = "headers"
	freezerHashTable = "hashes"
	freezerBodiesTable = "bodies"
	freezerReceiptTable = "receipts"
	freezerDifficultyTable = "diffs"
)
var freezerNoSnappy = map[string]bool{
	freezerHeaderTable:     false,
	freezerHashTable:       true,
	freezerBodiesTable:     false,
	freezerReceiptTable:    false,
	freezerDifficultyTable: true,
}
type LegacyTxLookupEntry struct {
	BlockHash  common.Hash
	BlockIndex uint64
	Index      uint64
}
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}
func headerKeyPrefix(number uint64) []byte {
	return append(headerPrefix, encodeBlockNumber(number)...)
}
func headerKey(number uint64, hash common.Hash) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}
func headerTDKey(number uint64, hash common.Hash) []byte {
	return append(headerKey(number, hash), headerTDSuffix...)
}
func headerHashKey(number uint64) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), headerHashSuffix...)
}
func headerNumberKey(hash common.Hash) []byte {
	return append(headerNumberPrefix, hash.Bytes()...)
}
func blockBodyKey(number uint64, hash common.Hash) []byte {
	return append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}
func blockReceiptsKey(number uint64, hash common.Hash) []byte {
	return append(append(blockReceiptsPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}
func txLookupKey(hash common.Hash) []byte {
	return append(txLookupPrefix, hash.Bytes()...)
}
func accountSnapshotKey(hash common.Hash) []byte {
	return append(SnapshotAccountPrefix, hash.Bytes()...)
}
func storageSnapshotKey(accountHash, storageHash common.Hash) []byte {
	return append(append(SnapshotStoragePrefix, accountHash.Bytes()...), storageHash.Bytes()...)
}
func storageSnapshotsKey(accountHash common.Hash) []byte {
	return append(SnapshotStoragePrefix, accountHash.Bytes()...)
}
func bloomBitsKey(bit uint, section uint64, hash common.Hash) []byte {
	key := append(append(bloomBitsPrefix, make([]byte, 10)...), hash.Bytes()...)
	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint64(key[3:], section)
	return key
}
func preimageKey(hash common.Hash) []byte {
	return append(preimagePrefix, hash.Bytes()...)
}
func configKey(hash common.Hash) []byte {
	return append(configPrefix, hash.Bytes()...)
}
