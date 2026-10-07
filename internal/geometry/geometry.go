package geometry

import (
	"iter"

	"github.com/bavix/spatial/internal/sphere"
)

type (
	Point struct {
		Lat, Lon float64
	}
	Line     [2]Point
	Polyline []Point
	Ring     []Point
	Triangle [3]Point
	Polygon  struct {
		Exterior Ring
		Holes    []Ring
	}
)

func (p Point) Valid() bool {
	return p.Lat >= -90 && p.Lat <= 90 && p.Lon >= -180 && p.Lon <= 180
}

func (l Line) Valid() bool {
	return l[0].Valid() && l[1].Valid()
}

func (l Line) Length() float64 {
	return sphere.Prepare(l[0].Lat, l[0].Lon).Distance(l[1].Lat, l[1].Lon)
}

func (l Polyline) Valid() bool {
	if len(l) < 2 {
		return false
	}

	for _, p := range l {
		if !p.Valid() {
			return false
		}
	}

	return true
}

func (l Polyline) Segments() iter.Seq[Line] {
	return func(yield func(Line) bool) {
		for row := 1; row < len(l); row++ {
			if !yield(Line{
				l[row-1],
				l[row],
			}) {
				return
			}
		}
	}
}

func (r Ring) Segments() iter.Seq[Line] {
	return func(yield func(Line) bool) {
		for line := range Polyline(r).Segments() {
			if !yield(line) {
				return
			}
		}

		if len(r) > 2 && !same(r[0], r[len(r)-1]) {
			yield(Line{
				r[len(r)-1],
				r[0],
			})
		}
	}
}

func (t Triangle) Segments() iter.Seq[Line] {
	return Ring(t[:]).Segments()
}

func (t Triangle) Valid() bool {
	return Ring(t[:]).Valid()
}

func (t Triangle) Contains(p Point) bool {
	return Ring(t[:]).Contains(p)
}

func (r Ring) Valid() bool {
	_, err := prepareRing(r)

	return err == nil
}

func (r Ring) Contains(p Point) bool {
	return locate(r, p) >= inside
}

func (p Polygon) Valid() bool {
	_, err := Prepare(p)

	return err == nil
}

func (p Polygon) Contains(point Point) bool {
	outer := locate(p.Exterior, point)
	if outer == outside {
		return false
	}

	if outer == boundary {
		return true
	}

	for _, hole := range p.Holes {
		switch locate(hole, point) {
		case inside:
			return false
		case boundary:
			return true
		case outside:
		}
	}

	return true
}

func (p Polygon) Segments() iter.Seq[Line] {
	return func(yield func(Line) bool) {
		for line := range p.Exterior.Segments() {
			if !yield(line) {
				return
			}
		}

		for _, ring := range p.Holes {
			for line := range ring.Segments() {
				if !yield(line) {
					return
				}
			}
		}
	}
}

func same(a, b Point) bool {
	return a.Lat == b.Lat && sphere.LongitudeDifference(a.Lon, b.Lon) == 0
}
