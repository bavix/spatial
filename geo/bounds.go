package geo

import (
	"math"

	"github.com/bavix/spatial/internal/sphere"
)

type Bounds struct {
	South, West, North, East float64
}

func World() Bounds {
	return Bounds{
		-90,
		-180,
		90,
		180,
	}
}

func (b Bounds) Valid() bool {
	return P(b.South, b.West).Valid() && P(b.North, b.East).Valid() && b.South <= b.North
}

func (b Bounds) Contains(p Point) bool {
	return b.Valid() && sphere.Contains(b.South, b.West, b.North, b.East, p.Lat, p.Lon)
}

func BoundsAround(p Point, radius float64) Bounds {
	if !p.Valid() || radius < 0 || math.IsNaN(radius) || math.IsInf(radius, 0) {
		return Bounds{
			South: math.NaN(),
		}
	}

	if radius >= math.Pi*EarthRadius {
		return World()
	}

	d := radius / EarthRadius
	lat := p.Lat * sphere.Radians
	margin := sphere.BoundsMargin
	south := max(-90, (lat-d)/sphere.Radians-margin)

	north := min(90, (lat+d)/sphere.Radians+margin)

	if south <= -90 || north >= 90 {
		return Bounds{
			south,
			-180,
			north,
			180,
		}
	}

	w := math.Asin(min(1, math.Sin(d)/math.Cos(lat)))/sphere.Radians + margin

	return Bounds{
		south,
		sphere.NormalizeLongitude(p.Lon - w),
		north,
		sphere.NormalizeLongitude(p.Lon + w),
	}
}
