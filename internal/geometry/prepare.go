package geometry

import (
	"math"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/bavix/spatial/internal/fault"
	"github.com/bavix/spatial/internal/sphere"
)

var ErrInvalidPolygon = errors.New("invalid polygon")

type ring struct {
	vertices                 []vertex
	west, east, south, north float64
}

type Prepared struct {
	exterior ring
	holes    []ring
}

func prepareRing(input Ring) (ring, error) {
	if len(input) > 1 && same(input[0], input[len(input)-1]) {
		input = input[:len(input)-1]
	}

	if len(input) < 3 {
		return ring{}, ErrInvalidPolygon
	}

	r := ring{
		vertices: make([]vertex, len(input)),
		west:     math.Inf(1),
		east:     math.Inf(-1),
		south:    math.Inf(1),
		north:    math.Inf(-1),
	}
	for row, p := range input {
		if !p.Valid() {
			return ring{}, errors.Wrapf(fault.ErrInvalidPoint, "point %d", row)
		}

		if math.Abs(p.Lat) == 90 {
			return ring{}, ErrInvalidPolygon
		}

		x := p.Lon
		if row > 0 {
			x = r.vertices[row-1].x + sphere.LongitudeDifference(input[row-1].Lon, p.Lon)
		}

		r.vertices[row] = vertex{
			x,
			p.Lat,
		}
		r.west, r.east = min(r.west, x), max(r.east, x)
		r.south, r.north = min(r.south, p.Lat), max(r.north, p.Lat)
	}

	if r.east-r.west >= 180 {
		return ring{}, ErrInvalidPolygon
	}

	if !simple(r.vertices) {
		return ring{}, ErrInvalidPolygon
	}

	return r, nil
}

func simple(vertices []vertex) bool {
	n := len(vertices)
	nondegenerate := false

	for row, a := range vertices {
		b := vertices[(row+1)%n]
		if a == b {
			return false
		}

		previous := vertices[(row+n-1)%n]
		if orientation(previous, a, b) != 0 {
			nondegenerate = true
		} else if onSegment(a, b, previous) || onSegment(previous, a, b) {
			return false
		}

		for column := row + 2; column < n; column++ {
			if row == 0 && column == n-1 {
				continue
			}

			if intersects(a, b, vertices[column], vertices[(column+1)%n]) {
				return false
			}
		}
	}

	return nondegenerate
}

func (r ring) locate(p Point) int {
	if p.Lat < r.south || p.Lat > r.north {
		return outside
	}

	point := vertex{
		r.vertices[0].x + sphere.LongitudeDifference(r.vertices[0].x, p.Lon),
		p.Lat,
	}
	if point.x < r.west || point.x > r.east {
		return outside
	}

	location := outside
	for row, a := range r.vertices {
		location = cross(a, r.vertices[(row+1)%len(r.vertices)], point, location)
		if location == boundary {
			return boundary
		}
	}

	return location
}

func Prepare(p Polygon) (Prepared, error) {
	outer, err := prepareRing(p.Exterior)
	if err != nil {
		return Prepared{}, errors.Wrap(err, "exterior")
	}

	result := Prepared{
		exterior: outer,
		holes:    make([]ring, 0, len(p.Holes)),
	}

	for row, input := range p.Holes {
		hole, err := prepareRing(input)
		if err != nil {
			return Prepared{}, errors.Wrapf(err, "hole %d", row)
		}

		if !result.accepts(hole) {
			return Prepared{}, errors.Wrapf(ErrInvalidPolygon, "hole %d", row)
		}

		result.holes = append(result.holes, hole)
	}

	return result, nil
}

func (p Prepared) Contains(point Point) bool {
	outer := p.exterior.locate(point)
	if outer == outside {
		return false
	}

	if outer == boundary {
		return true
	}

	for _, hole := range p.holes {
		location := hole.locate(point)
		if location == inside {
			return false
		}

		if location == boundary {
			return true
		}
	}

	return true
}

func (p Prepared) AxisKey(axis int) float64 {
	if axis == 0 {
		return p.exterior.west + p.exterior.east
	}

	return p.exterior.south + p.exterior.north
}

func (p Prepared) StorageBytes() uint64 {
	size := uint64(cap(p.exterior.vertices))*uint64(unsafe.Sizeof(vertex{})) + uint64(cap(p.holes))*uint64(unsafe.Sizeof(ring{}))
	for _, hole := range p.holes {
		size += uint64(cap(hole.vertices)) * uint64(unsafe.Sizeof(vertex{}))
	}

	return size
}

func (p Prepared) accepts(hole ring) bool {
	first := hole.vertices[0]

	point := Point{
		Lat: first.y,
		Lon: sphere.NormalizeLongitude(first.x),
	}
	if p.exterior.locate(point) != inside || ringsIntersect(p.exterior, hole) {
		return false
	}

	for _, other := range p.holes {
		otherPoint := Point{
			Lat: other.vertices[0].y,
			Lon: sphere.NormalizeLongitude(other.vertices[0].x),
		}
		if other.locate(point) != outside || hole.locate(otherPoint) != outside || ringsIntersect(other, hole) {
			return false
		}
	}

	return true
}

func ringsIntersect(a, b ring) bool {
	shift := a.vertices[0].x + sphere.LongitudeDifference(a.vertices[0].x, b.vertices[0].x) - b.vertices[0].x
	for row, start := range a.vertices {
		end := a.vertices[(row+1)%len(a.vertices)]
		for column, c := range b.vertices {
			d := b.vertices[(column+1)%len(b.vertices)]
			c.x += shift

			d.x += shift
			if intersects(start, end, c, d) {
				return true
			}
		}
	}

	return false
}
