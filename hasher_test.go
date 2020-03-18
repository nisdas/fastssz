package ssz

import (
	"fmt"
	"testing"
)

func TestHasherUint64(t *testing.T) {
	h := DefaultHasherPool.Get()
	defer DefaultHasherPool.Put(h)

	h.PutUint64(1234)
	fmt.Println(h.buf)
}

func TestHasherPutBytes(t *testing.T) {
	h := DefaultHasherPool.Get()

	a := make([]byte, 35)
	h.PutBytes(a)

	fmt.Println(h.buf)
}
