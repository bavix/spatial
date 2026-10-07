package search

import (
	"cmp"
	"math"
	"slices"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/metric"
	"github.com/bavix/spatial/internal/record"
)

type Collector struct {
	Filter   func(record.Item) bool
	results  []record.Result
	limit    int
	distance metric.Evaluator
	Accept   func(record.Item) bool
}

func (c *Collector) Reset(m metric.Metric, center geo.Point, limit int) {
	c.Clear()
	c.limit = limit
	c.distance.Reset(m, center)

	if c.Accept == nil {
		c.Accept = c.accept
	}
}

func (c *Collector) Clear() {
	c.results = c.results[:0]
}

func (c *Collector) Results() []record.Result {
	slices.SortFunc(c.results, compare)

	return c.results
}

func (c *Collector) Worst() float64 {
	if len(c.results) < c.limit {
		return math.Inf(1)
	}

	return c.results[0].Distance
}

func (c *Collector) LowerBound(bounds geo.Bounds) float64 {
	return c.distance.LowerBound(bounds)
}

func (c *Collector) BoundsAround(radius float64) geo.Bounds {
	return c.distance.BoundsAround(radius)
}

func compare(a, b record.Result) int {
	if order := cmp.Compare(a.Distance, b.Distance); order != 0 {
		return order
	}

	return cmp.Compare(a.ID, b.ID)
}

func (c *Collector) Offer(item record.Item) {
	if c.Filter != nil && !c.Filter(item) {
		return
	}

	d := c.distance.DistanceWithin(item.Point, c.Worst())
	if !(d >= 0) || math.IsInf(d, 0) {
		return
	}

	r := record.Result{
		ID:       item.ID,
		Point:    item.Point,
		Distance: d,
	}
	if len(c.results) < c.limit {
		c.results = append(c.results, r)
		c.siftUp(len(c.results) - 1)
	} else if compare(r, c.results[0]) < 0 {
		c.results[0] = r
		c.siftDown()
	}
}

func (c *Collector) siftUp(child int) {
	for child > 0 {
		parent := (child - 1) / 2
		if compare(c.results[parent], c.results[child]) >= 0 {
			return
		}

		c.results[parent], c.results[child] = c.results[child], c.results[parent]
		child = parent
	}
}

func (c *Collector) siftDown() {
	for parent := 0; ; {
		child := 2*parent + 1
		if child >= len(c.results) {
			return
		}

		if child+1 < len(c.results) && compare(c.results[child+1], c.results[child]) > 0 {
			child++
		}

		if compare(c.results[parent], c.results[child]) >= 0 {
			return
		}

		c.results[parent], c.results[child] = c.results[child], c.results[parent]
		parent = child
	}
}

func (c *Collector) accept(item record.Item) bool {
	c.Offer(item)

	return true
}
