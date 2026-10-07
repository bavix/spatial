package spatial

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/metric"
	"github.com/bavix/spatial/internal/record"
)

type (
	ID     = uint64
	Item   = record.Item
	Result = record.Result
	Source = record.Source
)
type Items []Item

func (s Items) Len() int {
	return len(s)
}

func (s Items) Item(row int) Item {
	return s[row]
}

type Filter func(Item) bool

type (
	Metric        = metric.Metric
	BoundedMetric = metric.BoundedMetric
)
type DistanceFunc func(geo.Point, geo.Point) float64

func (f DistanceFunc) Distance(a, b geo.Point) float64 {
	return f(a, b)
}
