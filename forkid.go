package forkid
import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"math"
	"math/big"
	"reflect"
	"strings"
	"github.com/Cryptochain-VON/common"
	"github.com/Cryptochain-VON/core/types"
	"github.com/Cryptochain-VON/log"
	"github.com/Cryptochain-VON/params"
)
var (
	ErrRemoteStale = errors.New("remote needs update")
	ErrLocalIncompatibleOrStale = errors.New("local incompatible or needs update")
)
type Blockchain interface {
	Config() *params.ChainConfig
	Genesis() *types.Block
	CurrentHeader() *types.Header
}
type ID struct {
	Hash [4]byte 
	Next uint64  
}
type Filter func(id ID) error
func NewID(chain Blockchain) ID {
	return newID(
		chain.Config(),
		chain.Genesis().Hash(),
		chain.CurrentHeader().Number.Uint64(),
	)
}
func newID(config *params.ChainConfig, genesis common.Hash, head uint64) ID {
	hash := crc32.ChecksumIEEE(genesis[:])
	var next uint64
	for _, fork := range gatherForks(config) {
		if fork <= head {
			hash = checksumUpdate(hash, fork)
			continue
		}
		next = fork
		break
	}
	return ID{Hash: checksumToBytes(hash), Next: next}
}
func NewFilter(chain Blockchain) Filter {
	return newFilter(
		chain.Config(),
		chain.Genesis().Hash(),
		func() uint64 {
			return chain.CurrentHeader().Number.Uint64()
		},
	)
}
func NewStaticFilter(config *params.ChainConfig, genesis common.Hash) Filter {
	head := func() uint64 { return 0 }
	return newFilter(config, genesis, head)
}
func newFilter(config *params.ChainConfig, genesis common.Hash, headfn func() uint64) Filter {
	var (
		forks = gatherForks(config)
		sums  = make([][4]byte, len(forks)+1) 
	)
	hash := crc32.ChecksumIEEE(genesis[:])
	sums[0] = checksumToBytes(hash)
	for i, fork := range forks {
		hash = checksumUpdate(hash, fork)
		sums[i+1] = checksumToBytes(hash)
	}
	forks = append(forks, math.MaxUint64) 
	return func(id ID) error {
		head := headfn()
		for i, fork := range forks {
			if head > fork {
				continue
			}
			if sums[i] == id.Hash {
				if id.Next > 0 && head >= id.Next {
					return ErrLocalIncompatibleOrStale
				}
				return nil
			}
			for j := 0; j < i; j++ {
				if sums[j] == id.Hash {
					if forks[j] != id.Next {
						return ErrRemoteStale
					}
					return nil
				}
			}
			for j := i + 1; j < len(sums); j++ {
				if sums[j] == id.Hash {
					return nil
				}
			}
			return ErrLocalIncompatibleOrStale
		}
		log.Error("Impossible fork ID validation", "id", id)
		return nil 
	}
}
func checksumUpdate(hash uint32, fork uint64) uint32 {
	var blob [8]byte
	binary.BigEndian.PutUint64(blob[:], fork)
	return crc32.Update(hash, crc32.IEEETable, blob[:])
}
func checksumToBytes(hash uint32) [4]byte {
	var blob [4]byte
	binary.BigEndian.PutUint32(blob[:], hash)
	return blob
}
func gatherForks(config *params.ChainConfig) []uint64 {
	kind := reflect.TypeOf(params.ChainConfig{})
	conf := reflect.ValueOf(config).Elem()
	var forks []uint64
	for i := 0; i < kind.NumField(); i++ {
		field := kind.Field(i)
		if !strings.HasSuffix(field.Name, "Block") {
			continue
		}
		if field.Type != reflect.TypeOf(new(big.Int)) {
			continue
		}
		rule := conf.Field(i).Interface().(*big.Int)
		if rule != nil {
			forks = append(forks, rule.Uint64())
		}
	}
	for i := 0; i < len(forks); i++ {
		for j := i + 1; j < len(forks); j++ {
			if forks[i] > forks[j] {
				forks[i], forks[j] = forks[j], forks[i]
			}
		}
	}
	for i := 1; i < len(forks); i++ {
		if forks[i] == forks[i-1] {
			forks = append(forks[:i], forks[i+1:]...)
			i--
		}
	}
	if len(forks) > 0 && forks[0] == 0 {
		forks = forks[1:]
	}
	return forks
}
