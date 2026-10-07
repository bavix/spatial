package store

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/fault"
	"github.com/bavix/spatial/internal/record"
)

const (
	cellByteBits       = 8
	cellByteMask       = 255
	cellInsertionLimit = 32
)

type Offset interface {
	uint32 | int
}

type Groups[T Offset] struct {
	Keys    []uint64
	Offsets []T
}

func Group[T Offset](data *Store, key func(geo.Point) uint64) Groups[T] {
	keys := make([]uint64, data.Len())
	for row := range keys {
		keys[row] = key(data.Key(row))
	}

	sortCells(data, keys)

	count := 0

	for row, cell := range keys {
		if row == 0 || cell != keys[row-1] {
			count++
		}
	}

	groups := Groups[T]{
		Keys:    make([]uint64, 0, count),
		Offsets: make([]T, 0, count+1),
	}

	for row, cell := range keys {
		if row == 0 || cell != keys[row-1] {
			groups.Keys = append(groups.Keys, cell)
			groups.Offsets = append(groups.Offsets, T(row))
		}
	}

	groups.Offsets = append(groups.Offsets, T(len(keys)))

	return groups
}

func CountBy(input record.Source, key func(geo.Point) uint64) (map[uint64]int, int, error) {
	n, err := sourceLen(input)
	if err != nil {
		return nil, 0, err
	}

	counts := make(map[uint64]int)

	ids := make([]uint64, n)
	for i := range n {
		item := input.Item(i)
		if !item.Point.Valid() {
			return nil, 0, fault.Point(i, item.ID)
		}

		ids[i] = item.ID
		counts[key(item.Point)]++
	}

	err = validateIDs(ids)
	if err != nil {
		return nil, 0, err
	}

	return counts, n, nil
}
