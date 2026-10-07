package geo

import "github.com/bavix/spatial/internal/geometry"

type (
	Point    = geometry.Point
	Line     = geometry.Line
	Polyline = geometry.Polyline
	Ring     = geometry.Ring
	Triangle = geometry.Triangle
	Polygon  = geometry.Polygon
)

func P(lat, lon float64) Point {
	return Point{
		Lat: lat,
		Lon: lon,
	}
}
