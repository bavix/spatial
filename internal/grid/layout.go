package grid

import (
	"math"
	"math/bits"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/sphere"
	"github.com/bavix/spatial/internal/store"
)

type layout struct {
	size       float64
	rows, cols uint64
	depth      uint
}

func newLayout(size float64) layout {
	rows, cols := uint64(math.Ceil(180/size)), uint64(math.Ceil(360/size))

	return layout{
		size:  size,
		rows:  rows,
		cols:  cols,
		depth: uint(bits.Len64(max(rows, cols) - 1)),
	}
}

func spread(v uint64) uint64 {
	const (
		mask32 = 0xffffffff
		mask16 = 0x0000ffff0000ffff
		mask8  = 0x00ff00ff00ff00ff
		mask4  = 0x0f0f0f0f0f0f0f0f
		mask2  = 0x3333333333333333
		mask1  = 0x5555555555555555
	)

	v &= mask32
	v = (v | v<<16) & mask16
	v = (v | v<<8) & mask8
	v = (v | v<<4) & mask4
	v = (v | v<<2) & mask2

	return (v | v<<1) & mask1
}

func (l layout) key(p geo.Point) uint64 {
	x := min(uint64((sphere.NormalizeLongitude(p.Lon)+180)/l.size), l.cols-1)
	y := min(uint64((p.Lat+90)/l.size), l.rows-1)

	return spread(x) | spread(y)<<1
}

func (l layout) bounds(x, y, width uint64) geo.Bounds {
	return store.Expand(geo.Bounds{
		South: -90 + float64(y)*l.size,
		West:  -180 + float64(x)*l.size,
		North: min(90, -90+float64(y+width)*l.size),
		East:  min(180, -180+float64(x+width)*l.size),
	}, sphere.BoundsMargin)
}
