package rawdb
import (
	"github.com/Cryptochain-VON/ethdb"
)
type table struct {
	db     ethdb.Database
	prefix string
}
func NewTable(db ethdb.Database, prefix string) ethdb.Database {
	return &table{
		db:     db,
		prefix: prefix,
	}
}
func (t *table) Close() error {
	return nil
}
func (t *table) Has(key []byte) (bool, error) {
	return t.db.Has(append([]byte(t.prefix), key...))
}
func (t *table) Get(key []byte) ([]byte, error) {
	return t.db.Get(append([]byte(t.prefix), key...))
}
func (t *table) HasAncient(kind string, number uint64) (bool, error) {
	return t.db.HasAncient(kind, number)
}
func (t *table) Ancient(kind string, number uint64) ([]byte, error) {
	return t.db.Ancient(kind, number)
}
func (t *table) Ancients() (uint64, error) {
	return t.db.Ancients()
}
func (t *table) AncientSize(kind string) (uint64, error) {
	return t.db.AncientSize(kind)
}
func (t *table) AppendAncient(number uint64, hash, header, body, receipts, td []byte) error {
	return t.db.AppendAncient(number, hash, header, body, receipts, td)
}
func (t *table) TruncateAncients(items uint64) error {
	return t.db.TruncateAncients(items)
}
func (t *table) Sync() error {
	return t.db.Sync()
}
func (t *table) Put(key []byte, value []byte) error {
	return t.db.Put(append([]byte(t.prefix), key...), value)
}
func (t *table) Delete(key []byte) error {
	return t.db.Delete(append([]byte(t.prefix), key...))
}
func (t *table) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	innerPrefix := append([]byte(t.prefix), prefix...)
	iter := t.db.NewIterator(innerPrefix, start)
	return &tableIterator{
		iter:   iter,
		prefix: t.prefix,
	}
}
func (t *table) Stat(property string) (string, error) {
	return t.db.Stat(property)
}
func (t *table) Compact(start []byte, limit []byte) error {
	if start == nil {
		start = []byte(t.prefix)
	}
	if limit == nil {
		limit = []byte(t.prefix)
		for i := len(limit) - 1; i >= 0; i-- {
			limit[i]++
			if limit[i] > 0 {
				break
			}
			if i == 0 {
				limit = nil
			}
		}
	}
	return t.db.Compact(start, limit)
}
func (t *table) NewBatch() ethdb.Batch {
	return &tableBatch{t.db.NewBatch(), t.prefix}
}
type tableBatch struct {
	batch  ethdb.Batch
	prefix string
}
func (b *tableBatch) Put(key, value []byte) error {
	return b.batch.Put(append([]byte(b.prefix), key...), value)
}
func (b *tableBatch) Delete(key []byte) error {
	return b.batch.Delete(append([]byte(b.prefix), key...))
}
func (b *tableBatch) ValueSize() int {
	return b.batch.ValueSize()
}
func (b *tableBatch) Write() error {
	return b.batch.Write()
}
func (b *tableBatch) Reset() {
	b.batch.Reset()
}
type tableReplayer struct {
	w      ethdb.KeyValueWriter
	prefix string
}
func (r *tableReplayer) Put(key []byte, value []byte) error {
	trimmed := key[len(r.prefix):]
	return r.w.Put(trimmed, value)
}
func (r *tableReplayer) Delete(key []byte) error {
	trimmed := key[len(r.prefix):]
	return r.w.Delete(trimmed)
}
func (b *tableBatch) Replay(w ethdb.KeyValueWriter) error {
	return b.batch.Replay(&tableReplayer{w: w, prefix: b.prefix})
}
type tableIterator struct {
	iter   ethdb.Iterator
	prefix string
}
func (iter *tableIterator) Next() bool {
	return iter.iter.Next()
}
func (iter *tableIterator) Error() error {
	return iter.iter.Error()
}
func (iter *tableIterator) Key() []byte {
	key := iter.iter.Key()
	if key == nil {
		return nil
	}
	return key[len(iter.prefix):]
}
func (iter *tableIterator) Value() []byte {
	return iter.iter.Value()
}
func (iter *tableIterator) Release() {
	iter.iter.Release()
}
