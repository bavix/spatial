package coast

import (
	"cmp"
	"slices"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/pymaxion/geographiclib-go/v2/geodesic"
	"github.com/pymaxion/geographiclib-go/v2/geodesic/capabilities"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/fault"
	"github.com/bavix/spatial/internal/sphere"
	"github.com/bavix/spatial/internal/traversal"
	"github.com/bavix/spatial/internal/wgs84"
)

const (
	leafSize     = 16
	margin       = 1e-5
	positionCaps = capabilities.Latitude | capabilities.Longitude
)

var (
	ErrInvalidLine = errors.New("invalid coastline")
	ErrEmpty       = errors.New("empty coastline")
)

type Result struct {
	Point      geo.Point
	Distance   float64
	ErrorBound float64
	Line       int
	Segment    int
}

type segment struct {
	start   int
	line    int
	length  float64
	azimuth float64
}

type node struct {
	center geo.Point
	radius float64
	start  int
	end    int
	right  int
}

type Index struct {
	points   []geo.Point
	offsets  []int
	segments []segment
	nodes    []node
}

func (i *Index) Len() int {
	return len(i.segments)
}

func (i *Index) StorageBytes() uint64 {
	return uint64(cap(i.points))*uint64(unsafe.Sizeof(geo.Point{})) +
		uint64(cap(i.offsets))*uint64(unsafe.Sizeof(int(0))) +
		uint64(cap(i.segments))*uint64(unsafe.Sizeof(segment{})) +
		uint64(cap(i.nodes))*uint64(unsafe.Sizeof(node{}))
}

func Build(lines []geo.Polyline) (*Index, error) {
	i := newIndex(lines)
	for row, line := range lines {
		err := i.add(line, row)
		if err != nil {
			return nil, err
		}
	}

	if i.Len() == 0 {
		return nil, ErrEmpty
	}

	i.nodes = make([]node, 0, traversal.NodeCount(i.Len(), leafSize))

	centers := make([]geo.Point, len(i.points)-len(i.offsets))
	for _, segment := range i.segments {
		centers[segment.cacheRow()] = i.midpoint(segment)
	}

	i.build(0, i.Len(), 0, centers)

	return i, nil
}

func newIndex(lines []geo.Polyline) *Index {
	points := 0
	for _, line := range lines {
		points += len(line)
		if points < 0 {
			panic(ErrInvalidLine)
		}
	}

	return &Index{
		points:  make([]geo.Point, 0, points),
		offsets: make([]int, 0, len(lines)),
	}
}

func (i *Index) add(line []geo.Point, row int) error {
	if len(line) < 2 {
		return errors.Wrapf(ErrInvalidLine, "line %d", row)
	}

	for column, point := range line {
		if !point.Valid() {
			return errors.Wrapf(fault.ErrInvalidPoint, "line %d, point %d", row, column)
		}
	}

	offset := len(i.points)
	i.offsets = append(i.offsets, offset)
	i.points = append(i.points, line...)

	for column := 0; column+1 < len(line); column++ {
		err := i.addSegment(offset+column, row)
		if err != nil {
			return errors.Wrapf(err, "line %d, segment %d", row, column)
		}
	}

	return nil
}

func (i *Index) addSegment(start, row int) error {
	a, b := i.points[start], i.points[start+1]
	if sphere.Coincident(a.Lat, a.Lon, b.Lat, b.Lon) {
		return nil
	}

	d := geodesic.WGS84.InverseWithCapabilities(a.Lat, a.Lon, b.Lat, b.Lon, capabilities.Distance|capabilities.Azimuth)
	if a.Lat == -b.Lat && (a.Lat == 90 || a.Lat == -90 || d.Azi1 != d.Azi2) {
		return ErrInvalidLine
	}

	if d.S12 > 0 {
		i.segments = append(i.segments, segment{
			start:   start,
			line:    row,
			length:  d.S12,
			azimuth: d.Azi1,
		})
	}

	return nil
}

func (i *Index) midpoint(s segment) geo.Point {
	a := i.points[s.start]
	d := geodesic.WGS84.DirectWithCapabilities(a.Lat, a.Lon, s.azimuth, s.length/2, positionCaps)

	return geo.P(d.Lat2, d.Lon2)
}

func (s segment) cacheRow() int {
	return s.start - s.line
}

func (i *Index) build(start, end, depth int, centers []geo.Point) int {
	row := len(i.nodes)

	n := node{
		center: centers[i.segments[start].cacheRow()],
		radius: i.segments[start].length/2 + margin,
		start:  start,
		end:    end,
		right:  -1,
	}
	for _, s := range i.segments[start+1 : end] {
		n.radius = max(n.radius, distance(n.center, centers[s.cacheRow()])+s.length/2+margin)
	}

	i.nodes = append(i.nodes, n)

	if end-start <= leafSize {
		return row
	}

	slices.SortFunc(i.segments[start:end], func(a, b segment) int {
		if depth%2 == 0 {
			return cmp.Compare(i.points[a.start].Lat, i.points[b.start].Lat)
		}

		return cmp.Compare(i.points[a.start].Lon, i.points[b.start].Lon)
	})
	middle := start + (end-start)/2
	i.build(start, middle, depth+1, centers)
	right := i.build(middle, end, depth+1, centers)
	i.nodes[row].right = right

	return row
}

func distance(a, b geo.Point) float64 {
	return (wgs84.Metric{}).Distance(a, b)
}
