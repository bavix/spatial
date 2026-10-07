package geometry

import "github.com/bavix/spatial/internal/sphere"

type Envelope struct {
	south, west, north, east float64
}

func (b Envelope) Merge(other Envelope) Envelope {
	return Envelope{
		south: min(b.south, other.south),
		west:  min(b.west, other.west),
		north: max(b.north, other.north),
		east:  max(b.east, other.east),
	}
}

func (b Envelope) Contains(p Point) bool {
	return p.Lat >= b.south && p.Lat <= b.north && p.Lon >= b.west && p.Lon <= b.east
}

func (p Prepared) Bounds() Envelope {
	r := p.exterior
	west := sphere.NormalizeLongitude(r.west)

	east := west + r.east - r.west
	if east >= 180 {
		west, east = -180, 180
	}

	return Envelope{
		r.south - sphere.BoundsMargin,
		west - sphere.BoundsMargin,
		r.north + sphere.BoundsMargin,
		east + sphere.BoundsMargin,
	}
}
