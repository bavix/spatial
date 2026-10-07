package coast

import (
	"math"

	"github.com/pymaxion/geographiclib-go/v2/geodesic"
	"github.com/pymaxion/geographiclib-go/v2/geodesic/capabilities"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/fault"
)

type work struct {
	lower   float64
	start   float64
	end     float64
	node    int
	segment int
	line    *geodesic.Line
}

type cachedLine struct {
	row  int
	line *geodesic.Line
}

type Searcher struct {
	Index *Index
	queue []work
	point geo.Point
	best  Result
	owner *Index
	lines [128]cachedLine
}

func (s *Searcher) Nearest(point geo.Point, tolerance float64) (Result, error) {
	if !point.Valid() {
		return Result{}, fault.ErrInvalidPoint
	}

	if tolerance < 2*margin || math.IsNaN(tolerance) || math.IsInf(tolerance, 0) {
		return Result{}, fault.ErrInvalidConfig
	}

	if s.Index == nil || s.Index.Len() == 0 {
		return Result{}, ErrEmpty
	}

	if s.owner != s.Index {
		clear(s.lines[:])
		s.owner = s.Index
	}

	s.point, s.best = point, Result{
		Distance: math.Inf(1),
	}
	defer s.clear()

	s.addNode(0)

	for len(s.queue) > 0 {
		w := s.pop()
		if s.best.Distance+margin-w.lower <= tolerance {
			s.best.ErrorBound = max(0, s.best.Distance+margin-w.lower)

			return s.best, nil
		}

		s.expand(w)
	}

	s.best.ErrorBound = 2 * margin

	return s.best, nil
}

func (s *Searcher) clear() {
	clear(s.queue[:cap(s.queue)])
	s.queue = s.queue[:0]
}

func (s *Searcher) addNode(row int) {
	n := s.Index.nodes[row]
	s.push(work{
		node:  row,
		lower: max(0, distance(s.point, n.center)-n.radius-margin),
	})
}

func (s *Searcher) expand(w work) {
	if w.line != nil {
		middle := (w.start + w.end) / 2
		s.addInterval(w.line, w.segment, w.start, middle)
		s.addInterval(w.line, w.segment, middle, w.end)

		return
	}

	n := s.Index.nodes[w.node]
	if n.right >= 0 {
		s.addNode(w.node + 1)
		s.addNode(n.right)

		return
	}

	for row := n.start; row < n.end; row++ {
		s.addSegment(row)
	}
}

func (s *Searcher) addSegment(row int) {
	edge := s.Index.segments[row]
	a, b := s.Index.points[edge.start], s.Index.points[edge.start+1]
	s.offer(a, row)
	s.offer(b, row)

	s.addInterval(s.line(row), row, 0, edge.length)
}

func (s *Searcher) line(row int) *geodesic.Line {
	entry := &s.lines[row%len(s.lines)]
	if entry.line == nil || entry.row != row {
		edge := s.Index.segments[row]
		point := s.Index.points[edge.start]
		entry.row = row
		entry.line = geodesic.WGS84.LineWithCapabilities(point.Lat, point.Lon, edge.azimuth, positionCaps|capabilities.DistanceIn)
	}

	return entry.line
}

func (s *Searcher) addInterval(line *geodesic.Line, row int, start, end float64) {
	d := line.PositionWithCapabilities((start+end)/2, positionCaps)
	lower := max(0, s.offer(geo.P(d.Lat2, d.Lon2), row)-(end-start)/2-margin)
	s.push(work{
		lower:   lower,
		start:   start,
		end:     end,
		segment: row,
		line:    line,
	})
}

func (s *Searcher) offer(point geo.Point, row int) float64 {
	d := distance(s.point, point)
	edge := s.Index.segments[row]

	ordinal := edge.start - s.Index.offsets[edge.line]
	if d < s.best.Distance || d == s.best.Distance && (edge.line < s.best.Line || edge.line == s.best.Line && ordinal < s.best.Segment) {
		s.best = Result{
			Point:    point,
			Distance: d,
			Line:     edge.line,
			Segment:  ordinal,
		}
	}

	return d
}
