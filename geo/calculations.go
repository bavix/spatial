package geo

import (
	"math"

	"github.com/bavix/spatial/internal/sphere"
)

func Distance(a, b Point) float64 {
	return sphere.Prepare(a.Lat, a.Lon).Distance(b.Lat, b.Lon)
}

func Equirectangular(a, b Point) float64 {
	x := sphere.LongitudeDifference(a.Lon, b.Lon) * sphere.Radians * math.Cos((a.Lat+b.Lat)*sphere.Radians/2)
	y := (b.Lat - a.Lat) * sphere.Radians

	return EarthRadius * math.Hypot(x, y)
}

func Bearing(a, b Point) float64 {
	if sphere.Coincident(a.Lat, a.Lon, b.Lat, b.Lon) {
		return 0
	}

	sa, ca := math.Sincos(a.Lat * sphere.Radians)
	sb, cb := math.Sincos(b.Lat * sphere.Radians)
	sd, cd := math.Sincos(sphere.LongitudeDifference(a.Lon, b.Lon) * sphere.Radians)
	y := sd * cb
	x := ca*sb - sa*cb*cd

	return math.Mod(math.Atan2(y, x)/sphere.Radians+360, 360)
}

func Destination(p Point, distance, bearing float64) Point {
	lat, lon, a, d := p.Lat*sphere.Radians, p.Lon*sphere.Radians, bearing*sphere.Radians, distance/EarthRadius
	s, c := math.Sincos(lat)
	sd, cd := math.Sincos(d)
	sa, ca := math.Sincos(a)

	if math.Abs(p.Lat) == 90 {
		c = 0
	}

	x := c*cd - s*sd*ca
	y := sd * sa
	z := s*cd + c*sd*ca
	q := math.Atan2(z, math.Hypot(x, y))
	l := lon + math.Atan2(y, x)

	return P(q/sphere.Radians, sphere.NormalizeLongitude(l/sphere.Radians))
}
