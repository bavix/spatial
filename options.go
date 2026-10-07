package spatial

import "github.com/bavix/spatial/internal/wgs84"

// Option configures index construction; later options override earlier ones.
type Option func(*config)

// WithGrid selects a grid with cells sized in degrees of latitude and longitude.
func WithGrid(cellSizeDegrees float64) Option {
	return func(c *config) {
		c.grid = true
		c.cellSize = cellSizeDegrees
	}
}

// WithKDTree selects a static KD-tree; leafSize limits points per leaf.
func WithKDTree(leafSize int) Option {
	return func(c *config) {
		c.grid = false
		c.leafSize = leafSize
	}
}

// WithMetric replaces spherical distance with a custom metric returning meters.
func WithMetric(distance Metric) Option {
	return func(c *config) {
		c.metric = distance
	}
}

// WithWGS84 measures distance on the WGS84 ellipsoid for greater Earth accuracy.
func WithWGS84() Option {
	return WithMetric(wgs84.Metric{})
}

// WithE7Coordinates halves stored coordinate memory by rounding to 1e-7 degrees.
// It applies to owned static indexes and changes returned coordinates.
func WithE7Coordinates() Option {
	return func(c *config) {
		c.e7 = true
	}
}
