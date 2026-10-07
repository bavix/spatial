package kdtree

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/store"
)

func tightBounds(data *store.Store) geo.Bounds {
	if data.Len() == 0 {
		return geo.Bounds{}
	}

	p := data.Key(0)

	b := geo.Bounds{
		South: p.Lat,
		West:  p.Lon,
		North: p.Lat,
		East:  p.Lon,
	}
	for i := 1; i < data.Len(); i++ {
		p = data.Key(i)
		b.South, b.North = min(b.South, p.Lat), max(b.North, p.Lat)
		b.West, b.East = min(b.West, p.Lon), max(b.East, p.Lon)
	}

	return b
}
