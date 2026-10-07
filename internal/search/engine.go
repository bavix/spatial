package search

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/record"
)

type Engine interface {
	Len() int
	Visit(bounds geo.Bounds, fn func(record.Item) bool)
	Nearest(collector *Collector)
}

type StaticEngine interface {
	Engine
	StorageBytes() uint64
}
