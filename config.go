package spatial

import (
	"math"

	"github.com/bavix/spatial/internal/metric"
)

const (
	defaultCellSize = 0.25
	defaultLeafSize = 64
	maxLeafSize     = 65535
)

type config struct {
	grid     bool
	cellSize float64
	leafSize int
	e7       bool
	metric   Metric
}

func configure(grid bool, options []Option) (config, error) {
	if len(options) == 0 {
		return defaults(grid), nil
	}

	c := defaults(grid)

	for _, option := range options {
		if option == nil {
			return c, ErrInvalidConfig
		}

		option(&c)
	}

	if !c.valid() {
		return c, ErrInvalidConfig
	}

	if c.metric == nil {
		c.metric = metric.Sphere{}
	}

	return c, nil
}

func defaults(grid bool) config {
	return config{
		grid:     grid,
		cellSize: defaultCellSize,
		leafSize: defaultLeafSize,
		metric:   metric.Sphere{},
	}
}

func (c config) valid() bool {
	if f, ok := c.metric.(DistanceFunc); ok && f == nil {
		return false
	}

	if c.grid {
		if !(c.cellSize > 0 && c.cellSize <= 180) {
			return false
		}

		return math.Ceil(360/c.cellSize) <= float64(math.MaxUint32)+1
	}

	return c.leafSize >= 2 && c.leafSize <= maxLeafSize
}
