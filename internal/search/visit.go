package search

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/record"
)

func VisitItems(items []record.Item, bounds geo.Bounds, fn func(record.Item) bool) bool {
	for _, item := range items {
		if Contains(bounds, item.Point) && !fn(item) {
			return false
		}
	}

	return true
}
