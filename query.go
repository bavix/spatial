package spatial

import (
	"math"

	"github.com/bavix/spatial/geo"
)

type WithinQuery struct {
	Center geo.Point
	Radius float64
	Filter Filter
}

func (q WithinQuery) Valid() bool {
	return q.Center.Valid() && q.Radius >= 0 && !math.IsInf(q.Radius, 1)
}

type NearestQuery struct {
	Center geo.Point
	Limit  int
	Filter Filter
}

func (q NearestQuery) Valid() bool {
	return q.Center.Valid() && q.Limit >= 0
}

type BoundsQuery struct {
	Bounds geo.Bounds
	Filter Filter
}

func (q BoundsQuery) Valid() bool {
	return q.Bounds.Valid()
}
