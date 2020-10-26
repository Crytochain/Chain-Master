package rawdb
import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/log"
	"github.com/Cryptochain-VON/metrics"
	"github.com/golang/snappy"
)
var (
	errClosed = errors.New("closed")
	errOutOfBounds = errors.New("out of bounds")
	errNotSupported = errors.New("this operation is not supported")
)
type indexEntry struct {
	filenum uint32 
	offset  uint32 
}
const indexEntrySize = 6
func (i *indexEntry) unmarshalBinary(b []byte) error {
	i.filenum = uint32(binary.BigEndian.Uint16(b[:2]))
	i.offset = binary.BigEndian.Uint32(b[2:6])
	return nil
}
func (i *indexEntry) marshallBinary() []byte {
	b := make([]byte, indexEntrySize)
	binary.BigEndian.PutUint16(b[:2], uint16(i.filenum))
	binary.BigEndian.PutUint32(b[2:6], i.offset)
	return b
}
type freezerTable struct {
	items uint64 
	noCompression bool   
	maxFileSize   uint32 
	name          string
	path          string
	head   *os.File            
	files  map[uint32]*os.File 
	headId uint32              
	tailId uint32              
	index  *os.File            
	itemOffset uint32 
	headBytes  uint32        
	readMeter  metrics.Meter 
	writeMeter metrics.Meter 
	sizeGauge  metrics.Gauge 
	logger log.Logger   
	lock   sync.RWMutex 
}
func newTable(path string, name string, readMeter metrics.Meter, writeMeter metrics.Meter, sizeGauge metrics.Gauge, disableSnappy bool) (*freezerTable, error) {
	return newCustomTable(path, name, readMeter, writeMeter, sizeGauge, 2*1000*1000*1000, disableSnappy)
}
func openFreezerFileForAppend(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	if _, err = file.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}
	return file, nil
}
func openFreezerFileForReadOnly(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDONLY, 0644)
}
func openFreezerFileTruncated(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}
func truncateFreezerFile(file *os.File, size int64) error {
	if err := file.Truncate(size); err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	return nil
}
func newCustomTable(path string, name string, readMeter metrics.Meter, writeMeter metrics.Meter, sizeGauge metrics.Gauge, maxFilesize uint32, noCompression bool) (*freezerTable, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}
	var idxName string
	if noCompression {
		idxName = fmt.Sprintf("%s.ridx", name)
	} else {
		idxName = fmt.Sprintf("%s.cidx", name)
	}
	offsets, err := openFreezerFileForAppend(filepath.Join(path, idxName))
	if err != nil {
		return nil, err
	}
	tab := &freezerTable{
		index:         offsets,
		files:         make(map[uint32]*os.File),
		readMeter:     readMeter,
		writeMeter:    writeMeter,
		sizeGauge:     sizeGauge,
		name:          name,
		path:          path,
		logger:        log.New("database", path, "table", name),
		noCompression: noCompression,
		maxFileSize:   maxFilesize,
	}
	if err := tab.repair(); err != nil {
		tab.Close()
		return nil, err
	}
	size, err := tab.sizeNolock()
	if err != nil {
		tab.Close()
		return nil, err
	}
	tab.sizeGauge.Inc(int64(size))
	return tab, nil
}
func (t *freezerTable) repair() error {
	buffer := make([]byte, indexEntrySize)
	stat, err := t.index.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		if _, err := t.index.Write(buffer); err != nil {
			return err
		}
	}
	if overflow := stat.Size() % indexEntrySize; overflow != 0 {
		truncateFreezerFile(t.index, stat.Size()-overflow) 
	}
	if stat, err = t.index.Stat(); err != nil {
		return err
	}
	offsetsSize := stat.Size()
	var (
		firstIndex  indexEntry
		lastIndex   indexEntry
		contentSize int64
		contentExp  int64
	)
	t.index.ReadAt(buffer, 0)
	firstIndex.unmarshalBinary(buffer)
	t.tailId = firstIndex.offset
	t.itemOffset = firstIndex.filenum
	t.index.ReadAt(buffer, offsetsSize-indexEntrySize)
	lastIndex.unmarshalBinary(buffer)
	t.head, err = t.openFile(lastIndex.filenum, openFreezerFileForAppend)
	if err != nil {
		return err
	}
	if stat, err = t.head.Stat(); err != nil {
		return err
	}
	contentSize = stat.Size()
	contentExp = int64(lastIndex.offset)
	for contentExp != contentSize {
		if contentExp < contentSize {
			t.logger.Warn("Truncating dangling head", "indexed", common.StorageSize(contentExp), "stored", common.StorageSize(contentSize))
			if err := truncateFreezerFile(t.head, contentExp); err != nil {
				return err
			}
			contentSize = contentExp
		}
		if contentExp > contentSize {
			t.logger.Warn("Truncating dangling indexes", "indexed", common.StorageSize(contentExp), "stored", common.StorageSize(contentSize))
			if err := truncateFreezerFile(t.index, offsetsSize-indexEntrySize); err != nil {
				return err
			}
			offsetsSize -= indexEntrySize
			t.index.ReadAt(buffer, offsetsSize-indexEntrySize)
			var newLastIndex indexEntry
			newLastIndex.unmarshalBinary(buffer)
			if newLastIndex.filenum != lastIndex.filenum {
				t.releaseFile(lastIndex.filenum)
				if t.head, err = t.openFile(newLastIndex.filenum, openFreezerFileForAppend); err != nil {
					return err
				}
				if stat, err = t.head.Stat(); err != nil {
					return err
				}
				contentSize = stat.Size()
			}
			lastIndex = newLastIndex
			contentExp = int64(lastIndex.offset)
		}
	}
	if err := t.index.Sync(); err != nil {
		return err
	}
	if err := t.head.Sync(); err != nil {
		return err
	}
	t.items = uint64(t.itemOffset) + uint64(offsetsSize/indexEntrySize-1) 
	t.headBytes = uint32(contentSize)
	t.headId = lastIndex.filenum
	if err := t.preopen(); err != nil {
		return err
	}
	t.logger.Debug("Chain freezer table opened", "items", t.items, "size", common.StorageSize(t.headBytes))
	return nil
}
func (t *freezerTable) preopen() (err error) {
	t.releaseFilesAfter(0, false)
	for i := t.tailId; i < t.headId; i++ {
		if _, err = t.openFile(i, openFreezerFileForReadOnly); err != nil {
			return err
		}
	}
	t.head, err = t.openFile(t.headId, openFreezerFileForAppend)
	return err
}
func (t *freezerTable) truncate(items uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if atomic.LoadUint64(&t.items) <= items {
		return nil
	}
	oldSize, err := t.sizeNolock()
	if err != nil {
		return err
	}
	t.logger.Warn("Truncating freezer table", "items", t.items, "limit", items)
	if err := truncateFreezerFile(t.index, int64(items+1)*indexEntrySize); err != nil {
		return err
	}
	buffer := make([]byte, indexEntrySize)
	if _, err := t.index.ReadAt(buffer, int64(items*indexEntrySize)); err != nil {
		return err
	}
	var expected indexEntry
	expected.unmarshalBinary(buffer)
	if expected.filenum != t.headId {
		t.releaseFile(expected.filenum)
		newHead, err := t.openFile(expected.filenum, openFreezerFileForAppend)
		if err != nil {
			return err
		}
		t.releaseFilesAfter(expected.filenum, true)
		t.head = newHead
		atomic.StoreUint32(&t.headId, expected.filenum)
	}
	if err := truncateFreezerFile(t.head, int64(expected.offset)); err != nil {
		return err
	}
	atomic.StoreUint64(&t.items, items)
	atomic.StoreUint32(&t.headBytes, expected.offset)
	newSize, err := t.sizeNolock()
	if err != nil {
		return err
	}
	t.sizeGauge.Dec(int64(oldSize - newSize))
	return nil
}
func (t *freezerTable) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	var errs []error
	if err := t.index.Close(); err != nil {
		errs = append(errs, err)
	}
	t.index = nil
	for _, f := range t.files {
		if err := f.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	t.head = nil
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
func (t *freezerTable) openFile(num uint32, opener func(string) (*os.File, error)) (f *os.File, err error) {
	var exist bool
	if f, exist = t.files[num]; !exist {
		var name string
		if t.noCompression {
			name = fmt.Sprintf("%s.%04d.rdat", t.name, num)
		} else {
			name = fmt.Sprintf("%s.%04d.cdat", t.name, num)
		}
		f, err = opener(filepath.Join(t.path, name))
		if err != nil {
			return nil, err
		}
		t.files[num] = f
	}
	return f, err
}
func (t *freezerTable) releaseFile(num uint32) {
	if f, exist := t.files[num]; exist {
		delete(t.files, num)
		f.Close()
	}
}
func (t *freezerTable) releaseFilesAfter(num uint32, remove bool) {
	for fnum, f := range t.files {
		if fnum > num {
			delete(t.files, fnum)
			f.Close()
			if remove {
				os.Remove(f.Name())
			}
		}
	}
}
func (t *freezerTable) Append(item uint64, blob []byte) error {
	t.lock.RLock()
	if t.index == nil || t.head == nil {
		t.lock.RUnlock()
		return errClosed
	}
	if atomic.LoadUint64(&t.items) != item {
		t.lock.RUnlock()
		return fmt.Errorf("appending unexpected item: want %d, have %d", t.items, item)
	}
	if !t.noCompression {
		blob = snappy.Encode(nil, blob)
	}
	bLen := uint32(len(blob))
	if t.headBytes+bLen < bLen ||
		t.headBytes+bLen > t.maxFileSize {
		t.lock.RUnlock()
		t.lock.Lock()
		nextID := atomic.LoadUint32(&t.headId) + 1
		newHead, err := t.openFile(nextID, openFreezerFileTruncated)
		if err != nil {
			t.lock.Unlock()
			return err
		}
		t.releaseFile(t.headId)
		t.openFile(t.headId, openFreezerFileForReadOnly)
		t.head = newHead
		atomic.StoreUint32(&t.headBytes, 0)
		atomic.StoreUint32(&t.headId, nextID)
		t.lock.Unlock()
		t.lock.RLock()
	}
	defer t.lock.RUnlock()
	if _, err := t.head.Write(blob); err != nil {
		return err
	}
	newOffset := atomic.AddUint32(&t.headBytes, bLen)
	idx := indexEntry{
		filenum: atomic.LoadUint32(&t.headId),
		offset:  newOffset,
	}
	t.index.Write(idx.marshallBinary())
	t.writeMeter.Mark(int64(bLen + indexEntrySize))
	t.sizeGauge.Inc(int64(bLen + indexEntrySize))
	atomic.AddUint64(&t.items, 1)
	return nil
}
func (t *freezerTable) getBounds(item uint64) (uint32, uint32, uint32, error) {
	var startIdx, endIdx indexEntry
	buffer := make([]byte, indexEntrySize)
	if _, err := t.index.ReadAt(buffer, int64(item*indexEntrySize)); err != nil {
		return 0, 0, 0, err
	}
	startIdx.unmarshalBinary(buffer)
	if _, err := t.index.ReadAt(buffer, int64((item+1)*indexEntrySize)); err != nil {
		return 0, 0, 0, err
	}
	endIdx.unmarshalBinary(buffer)
	if startIdx.filenum != endIdx.filenum {
		return 0, endIdx.offset, endIdx.filenum, nil
	}
	return startIdx.offset, endIdx.offset, endIdx.filenum, nil
}
func (t *freezerTable) Retrieve(item uint64) ([]byte, error) {
	t.lock.RLock()
	if t.index == nil || t.head == nil {
		t.lock.RUnlock()
		return nil, errClosed
	}
	if atomic.LoadUint64(&t.items) <= item {
		t.lock.RUnlock()
		return nil, errOutOfBounds
	}
	if uint64(t.itemOffset) > item {
		t.lock.RUnlock()
		return nil, errOutOfBounds
	}
	startOffset, endOffset, filenum, err := t.getBounds(item - uint64(t.itemOffset))
	if err != nil {
		t.lock.RUnlock()
		return nil, err
	}
	dataFile, exist := t.files[filenum]
	if !exist {
		t.lock.RUnlock()
		return nil, fmt.Errorf("missing data file %d", filenum)
	}
	blob := make([]byte, endOffset-startOffset)
	if _, err := dataFile.ReadAt(blob, int64(startOffset)); err != nil {
		t.lock.RUnlock()
		return nil, err
	}
	t.lock.RUnlock()
	t.readMeter.Mark(int64(len(blob) + 2*indexEntrySize))
	if t.noCompression {
		return blob, nil
	}
	return snappy.Decode(nil, blob)
}
func (t *freezerTable) has(number uint64) bool {
	return atomic.LoadUint64(&t.items) > number
}
func (t *freezerTable) size() (uint64, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.sizeNolock()
}
func (t *freezerTable) sizeNolock() (uint64, error) {
	stat, err := t.index.Stat()
	if err != nil {
		return 0, err
	}
	total := uint64(t.maxFileSize)*uint64(t.headId-t.tailId) + uint64(t.headBytes) + uint64(stat.Size())
	return total, nil
}
func (t *freezerTable) Sync() error {
	if err := t.index.Sync(); err != nil {
		return err
	}
	return t.head.Sync()
}
func (t *freezerTable) printIndex() {
	buf := make([]byte, indexEntrySize)
	fmt.Printf("|-----------------|\n")
	fmt.Printf("| fileno | offset |\n")
	fmt.Printf("|--------+--------|\n")
	for i := uint64(0); ; i++ {
		if _, err := t.index.ReadAt(buf, int64(i*indexEntrySize)); err != nil {
			break
		}
		var entry indexEntry
		entry.unmarshalBinary(buf)
		fmt.Printf("|  %03d   |  %03d   | \n", entry.filenum, entry.offset)
		if i > 100 {
			fmt.Printf(" ... \n")
			break
		}
	}
	fmt.Printf("|-----------------|\n")
}
