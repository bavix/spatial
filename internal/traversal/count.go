package traversal

import "math/bits"

func NodeCount(items, leafSize int) int {
	if items == 0 {
		return 0
	}

	if items <= leafSize {
		return 1
	}

	parents := 1 << (bits.Len(uint((items-1)/leafSize)) - 1)

	leaves := 2 * parents
	if items/parents == leafSize {
		leaves = parents + items%parents
	}

	return (leaves-1)*2 + 1
}
