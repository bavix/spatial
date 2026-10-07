package store

import "math/bits"

type cellRange struct {
	lo, hi int
	shift  uint
}

func sortCells(data *Store, keys []uint64) {
	if len(keys) < 2 {
		return
	}

	var difference uint64
	for _, key := range keys {
		difference |= keys[0] ^ key
	}

	if difference == 0 {
		return
	}

	shift := uint((bits.Len64(difference)-1)/cellByteBits) * cellByteBits
	pending := make([]cellRange, 1, 1+cellByteMask*int(shift/cellByteBits))

	pending[0] = cellRange{
		hi:    len(keys),
		shift: shift,
	}
	for len(pending) > 0 {
		part := pending[len(pending)-1]
		pending = pending[:len(pending)-1]

		if part.hi-part.lo <= cellInsertionLimit {
			insertCells(data, keys, part.lo, part.hi)

			continue
		}

		pending = partitionCells(data, keys, part, pending)
	}
}

func partitionCells(data *Store, keys []uint64, part cellRange, pending []cellRange) []cellRange {
	var next, ends [256]int
	for _, key := range keys[part.lo:part.hi] {
		ends[byte((key>>part.shift)&cellByteMask)]++
	}

	start := part.lo
	for bucket, count := range ends {
		next[bucket] = start
		start += count
		ends[bucket] = start
	}

	start = part.lo

	for bucket, end := range ends {
		for row := next[bucket]; row < end; {
			actual := byte((keys[row] >> part.shift) & cellByteMask)
			if int(actual) == bucket {
				row++

				continue
			}

			target := next[actual]
			keys[row], keys[target] = keys[target], keys[row]
			data.Swap(row, target)

			next[actual]++
		}

		next[bucket] = end
		if part.shift > 0 && end-start > 1 {
			pending = append(pending, cellRange{
				lo:    start,
				hi:    end,
				shift: part.shift - cellByteBits,
			})
		}

		start = end
	}

	return pending
}

func insertCells(data *Store, keys []uint64, lo, hi int) {
	for row := lo + 1; row < hi; row++ {
		for target := row; target > lo && keys[target] < keys[target-1]; target-- {
			keys[target], keys[target-1] = keys[target-1], keys[target]
			data.Swap(target, target-1)
		}
	}
}
