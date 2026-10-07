package metric

import (
	"math"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/sphere"
)

type (
	Metric        interface{ Distance(a, b geo.Point) float64 }
	BoundedMetric interface {
		Metric
		BoundsAround(center geo.Point, radius float64) geo.Bounds
		LowerBound(point geo.Point, bounds geo.Bounds) float64
	}
	scaledMetric interface{ SphericalLowerBoundScale() float64 }
)

type Evaluator struct {
	metric    Metric
	bounded   BoundedMetric
	scale     float64
	center    geo.Point
	prepared  sphere.Query
	spherical bool
}

func (e *Evaluator) Reset(m Metric, center geo.Point) {
	e.metric, e.center = m, center
	e.bounded, _ = m.(BoundedMetric)
	_, e.spherical = m.(Sphere)

	e.scale = 0
	if e.spherical {
		e.scale = 1
	} else if scaled, ok := m.(scaledMetric); ok {
		e.scale = scaled.SphericalLowerBoundScale()
	}

	if e.scale > 0 {
		e.prepared = sphere.Prepare(center.Lat, center.Lon)
	}
}

func (e *Evaluator) DistanceWithin(point geo.Point, limit float64) float64 {
	if e.scale > 0 {
		lower := math.Abs(point.Lat-e.center.Lat) * sphere.MetersPerDegree * e.scale
		if lower > limit+sphere.DistanceMargin {
			return math.Inf(1)
		}
	}

	if e.spherical {
		return e.prepared.Distance(point.Lat, point.Lon)
	}

	if e.scale > 0 && !math.IsInf(limit, 1) && e.prepared.Distance(point.Lat, point.Lon)*e.scale-sphere.DistanceMargin > limit {
		return math.Inf(1)
	}

	return e.metric.Distance(e.center, point)
}

func (e *Evaluator) BoundsAround(radius float64) geo.Bounds {
	if e.bounded == nil || math.IsInf(radius, 0) {
		return geo.World()
	}

	bounds := e.bounded.BoundsAround(e.center, radius)
	if !bounds.Valid() {
		return geo.World()
	}

	return bounds
}

func (e *Evaluator) LowerBound(bounds geo.Bounds) float64 {
	if e.bounded == nil {
		return 0
	}

	if e.scale > 0 {
		if !bounds.Valid() {
			return 0
		}

		return e.prepared.LowerBound(bounds.South, bounds.West, bounds.North, bounds.East) * e.scale
	}

	lower := e.bounded.LowerBound(e.center, bounds)
	if !(lower >= 0) || math.IsInf(lower, 0) {
		return 0
	}

	return lower
}
