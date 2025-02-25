package save

import (
	"fmt"
	"math"
)

const indexOutOfBounds = "index out of bounds"
const valueOutOfBounds = "value out of bounds"

// BitStorage implement the compacted data array used in chunk storage and heightmaps.
// You can think of this as a []intN whose N is called "bits" in NewBitStorage.
// For more info, see: https://wiki.vg/Chunk_Format
// This is implementation of the format since Minecraft 1.16
type BitStorage struct {
	data []uint64
	mask uint64

	bits, length  int
	valuesPerLong int
}

// NewBitStorage create a new BitStorage. Return nil if bits == 0.
//
// The "bits" is the number of bits per value, which can be calculated by math/bits.Len()
// The "length" is the number of values.
// The "data" is optional for initializing. Panic if data != nil && len(data) != BitStorageSize(bits, length).
func NewBitStorage(bits, length int, data []uint64) (b *BitStorage) {
	if bits == 0 {
		return nil
	}

	b = &BitStorage{
		mask:          1<<bits - 1,
		bits:          bits,
		length:        length,
		valuesPerLong: 64 / bits,
	}
	dataLen := BitStorageSize(bits, length)
	if data != nil {
		if len(data) != dataLen {
			panic(newBitStorageErr{ArrlLen: len(data), WantLen: dataLen})
		}
		b.data = data
	} else {
		b.data = make([]uint64, dataLen)
	}
	return
}

// BitStorageSize calculate how many uint64 is needed for given bits and length.
func BitStorageSize(bits, length int) (size int) {
	if bits == 0 {
		return 0
	}
	valuesPerLong := 64 / bits
	return (length + valuesPerLong - 1) / valuesPerLong
}

type newBitStorageErr struct {
	ArrlLen int
	WantLen int
}

func (i newBitStorageErr) Error() string {
	return fmt.Sprintf("invalid length given for storage, got: %d but expected: %d", i.ArrlLen, i.WantLen)
}

func (b *BitStorage) calcIndex(n int) (c, o int) {
	c = n / b.valuesPerLong
	o = (n - c*b.valuesPerLong) * b.bits
	return
}

// Swap sets v into [i], and return the previous [i] value.
func (b *BitStorage) Swap(i, v int) (old int) {
	if b == nil || i < 0 || i > b.length-1 {
		panic(indexOutOfBounds)
	}
	if v < 0 || uint64(v) > b.mask {
		panic(valueOutOfBounds)
	}

	c, offset := b.calcIndex(i)
	l := b.data[c]
	old = int(l >> offset & b.mask)
	b.data[c] = l&(b.mask<<offset^math.MaxUint64) | (uint64(v)&b.mask)<<offset
	return
}

// Set sets v into [i].
func (b *BitStorage) Set(i, v int) {
	if b == nil || i < 0 || i > b.length-1 {
		panic(indexOutOfBounds)
	}
	if v < 0 || uint64(v) > b.mask {
		panic(valueOutOfBounds)
	}

	c, offset := b.calcIndex(i)
	l := b.data[c]
	b.data[c] = l&(b.mask<<offset^math.MaxUint64) | (uint64(v)&b.mask)<<offset
}

// Get gets [i] value.
func (b *BitStorage) Get(i int) int {
	if b == nil || i < 0 || i > b.length-1 {
		panic(indexOutOfBounds)
	}

	c, offset := b.calcIndex(i)
	l := b.data[c]
	return int(l >> offset & b.mask)
}

// Len is the number of stored values.
func (b *BitStorage) Len() int {
	if b == nil {
		return 0
	}
	return b.length
}

// Longs return the underling array of uint64 for encoding/decoding.
func (b *BitStorage) Longs() []uint64 {
	if b == nil {
		return []uint64{}
	}
	return b.data
}
