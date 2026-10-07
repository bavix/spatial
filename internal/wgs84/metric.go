package wgs84

import (
	"math"

	"github.com/pymaxion/geographiclib-go/v2/geodesic"
	"github.com/pymaxion/geographiclib-go/v2/geodesic/capabilities"

	"github.com/bavix/spatial/geo"
	spherical "github.com/bavix/spatial/internal/metric"
	"github.com/bavix/spatial/internal/sphere"
)

const (
	minimumRadius       = geodesic.WGS84_a * (1 - geodesic.WGS84_f) * (1 - geodesic.WGS84_f)
	lowerBoundScale     = minimumRadius / geo.EarthRadius
	boundingRadiusScale = geo.EarthRadius / minimumRadius
	maximumRadius       = math.Pi * minimumRadius
)

type Metric struct{}

func (Metric) Distance(a, b geo.Point) float64 {
	if sphere.Coincident(a.Lat, a.Lon, b.Lat, b.Lon) {
		return 0
	}

	return geodesic.WGS84.InverseWithCapabilities(a.Lat, a.Lon, b.Lat, b.Lon, capabilities.Distance).S12
}

func (Metric) SphericalLowerBoundScale() float64 {
	return lowerBoundScale
}

func (Metric) BoundsAround(p geo.Point, radius float64) geo.Bounds {
	if p.Valid() && radius >= maximumRadius && !math.IsInf(radius, 0) {
		return geo.World()
	}

	return geo.BoundsAround(p, radius*boundingRadiusScale)
}

func (Metric) LowerBound(p geo.Point, bounds geo.Bounds) float64 {
	return spherical.LowerBound(p, bounds) * lowerBoundScale
}
