package rawdb
import (
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/ethdb"
	"github.com/Cryptochain-VON/log"
)
func ReadSnapshotRoot(db ethdb.KeyValueReader) common.Hash {
	data, _ := db.Get(snapshotRootKey)
	if len(data) != common.HashLength {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}
func WriteSnapshotRoot(db ethdb.KeyValueWriter, root common.Hash) {
	if err := db.Put(snapshotRootKey, root[:]); err != nil {
		log.Crit("Failed to store snapshot root", "err", err)
	}
}
func DeleteSnapshotRoot(db ethdb.KeyValueWriter) {
	if err := db.Delete(snapshotRootKey); err != nil {
		log.Crit("Failed to remove snapshot root", "err", err)
	}
}
func ReadAccountSnapshot(db ethdb.KeyValueReader, hash common.Hash) []byte {
	data, _ := db.Get(accountSnapshotKey(hash))
	return data
}
func WriteAccountSnapshot(db ethdb.KeyValueWriter, hash common.Hash, entry []byte) {
	if err := db.Put(accountSnapshotKey(hash), entry); err != nil {
		log.Crit("Failed to store account snapshot", "err", err)
	}
}
func DeleteAccountSnapshot(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(accountSnapshotKey(hash)); err != nil {
		log.Crit("Failed to delete account snapshot", "err", err)
	}
}
func ReadStorageSnapshot(db ethdb.KeyValueReader, accountHash, storageHash common.Hash) []byte {
	data, _ := db.Get(storageSnapshotKey(accountHash, storageHash))
	return data
}
func WriteStorageSnapshot(db ethdb.KeyValueWriter, accountHash, storageHash common.Hash, entry []byte) {
	if err := db.Put(storageSnapshotKey(accountHash, storageHash), entry); err != nil {
		log.Crit("Failed to store storage snapshot", "err", err)
	}
}
func DeleteStorageSnapshot(db ethdb.KeyValueWriter, accountHash, storageHash common.Hash) {
	if err := db.Delete(storageSnapshotKey(accountHash, storageHash)); err != nil {
		log.Crit("Failed to delete storage snapshot", "err", err)
	}
}
func IterateStorageSnapshots(db ethdb.Iteratee, accountHash common.Hash) ethdb.Iterator {
	return db.NewIterator(storageSnapshotsKey(accountHash), nil)
}
func ReadSnapshotJournal(db ethdb.KeyValueReader) []byte {
	data, _ := db.Get(snapshotJournalKey)
	return data
}
func WriteSnapshotJournal(db ethdb.KeyValueWriter, journal []byte) {
	if err := db.Put(snapshotJournalKey, journal); err != nil {
		log.Crit("Failed to store snapshot journal", "err", err)
	}
}
func DeleteSnapshotJournal(db ethdb.KeyValueWriter) {
	if err := db.Delete(snapshotJournalKey); err != nil {
		log.Crit("Failed to remove snapshot journal", "err", err)
	}
}
