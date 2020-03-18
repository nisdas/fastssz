package ssz

import (
	"fmt"
	"hash"
	"sync"

	"encoding/binary"
	"encoding/hex"

	"github.com/minio/sha256-simd"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	"github.com/prysmaticlabs/prysm/shared/hashutil"
)

type HashRoot interface {
	HashTreeRootWith(hh *Hasher) error
}

func HashWithDefaultHasher(v HashRoot) ([]byte, error) {
	hh := DefaultHasherPool.Get()
	if err := v.HashTreeRootWith(hh); err != nil {
		DefaultHasherPool.Put(hh)
		return nil, err
	}
	root, err := hh.HashRoot()
	DefaultHasherPool.Put(hh)
	return root, err
}

const bytesPerChunk = 32

var zeroBytes = make([]byte, 32)

// DefaultHasherPool is a default hasher pool
var DefaultHasherPool HasherPool

// Hasher is a utility tool to hash SSZ structs
type Hasher struct {
	buf    []byte
	bounds []int
	hash   hash.Hash
}

func (h *Hasher) reset() {
	h.buf = h.buf[:0]
	h.bounds = h.bounds[:0]
	h.hash.Reset()
}

func (h *Hasher) appendBytes32(b []byte) {
	rest := 32 - len(b)
	h.buf = append(h.buf, b...)
	if rest != 0 {
		// pad zero bytes to the left
		h.buf = append(h.buf, zeroBytes[:rest]...)
	}
}

// PutUint64 appends a uint64 in 32 bytes
func (h *Hasher) PutUint64(i uint64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	h.appendBytes32(buf)
}

var trueBytes, falseBytes []byte

func init() {
	falseBytes = make([]byte, 32)
	trueBytes = make([]byte, 32)
	trueBytes[31] = 1
}

// PutBool appends a boolean
func (h *Hasher) PutBool(b bool) {
	if b {
		h.buf = append(h.buf, falseBytes...)
	} else {
		h.buf = append(h.buf, trueBytes...)
	}
}

// PutRoot adds a root of 32 bytes
func (h *Hasher) PutRoot(b []byte) {
	_ = b[31]
	h.buf = append(h.buf, b...)
}

func (h *Hasher) PutFixedBytes(b []byte) {
	h.appendBytes32(b)
}

// PutBytes puts bytes higher than 32
func (h *Hasher) PutBytes(b []byte) error {
	if len(b) <= 32 {
		panic("BUG: cannot use it")
	}
	h.Bound()
	// add the item
	h.buf = append(h.buf, b...)
	// fill to make it a size of 32
	rest := 32 - len(b)%32
	if rest != 0 {
		h.buf = append(h.buf, zeroBytes[:rest]...)
	}
	return h.BitwiseMerkleize()
}

// Bound creates a bound to apply the merkleize
func (h *Hasher) Bound() {
	h.bounds = append(h.bounds, len(h.buf))
}

// BitwiseMerkleize is used to merkleize the last group of the hasher
func (h *Hasher) BitwiseMerkleize() error {
	if len(h.bounds) == 0 {
		panic("BUG")
	}
	var xxx int
	xxx, h.bounds = h.bounds[len(h.bounds)-1], h.bounds[:len(h.bounds)-1]

	input := h.buf[xxx:]

	hasher := htr.HashFn(hashutil.CustomSHA256Hasher())
	leafIndexer := func(i uint64) []byte {
		indx := i * 32
		return input[indx : indx+32]
	}
	res := merkle.Merkleize(hasher, 3, 3, leafIndexer)
	h.buf = append(h.buf[:xxx], res[:]...)
	return nil
}

// HashRoot creates the hash final hash root
func (h *Hasher) HashRoot() ([]byte, error) {
	fmt.Println("dd")
	fmt.Println(h.bounds)
	fmt.Println(h.buf)
	fmt.Println(hex.EncodeToString(h.buf))

	xx := []byte{}
	xx = append(xx, h.buf...)
	return xx, nil
}

// HasherPool may be used for pooling Hashers for similarly typed SSZs.
type HasherPool struct {
	pool sync.Pool
}

// Get acquires a Hasher from the pool.
func (hh *HasherPool) Get() *Hasher {
	h := hh.pool.Get()
	if h == nil {
		return &Hasher{hash: sha256.New()}
	}
	return h.(*Hasher)
}

// Put releases the Hasher to the pool.
func (hh *HasherPool) Put(h *Hasher) {
	h.reset()
	hh.pool.Put(h)
}
