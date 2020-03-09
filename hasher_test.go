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
