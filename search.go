package spatial

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/metric"
	"github.com/bavix/spatial/internal/record"
	"github.com/bavix/spatial/internal/search"
)

type radiusSearch struct {
	search.Output[Result]

	query    WithinQuery
	distance metric.Evaluator
	accept   func(record.Item) bool
}

func (s *radiusSearch) clear() {
	s.Values, s.Visitor, s.query.Filter = nil, nil, nil
}

func (s *radiusSearch) prepare(distance Metric) geo.Bounds {
	if s.accept == nil {
		s.accept = s.acceptItem
	}

	s.distance.Reset(distance, s.query.Center)

	return s.distance.BoundsAround(s.query.Radius)
}

func (s *radiusSearch) acceptItem(item record.Item) bool {
	if !accepts(s.query.Filter, item) {
		return true
	}

	distance := s.distance.DistanceWithin(item.Point, s.query.Radius)
	if !(distance >= 0 && distance <= s.query.Radius) {
		return true
	}

	return s.Emit(Result{
		ID:       item.ID,
		Point:    item.Point,
		Distance: distance,
	})
}

type boundsSearch struct {
	search.Output[Item]

	filter Filter
	accept func(record.Item) bool
}

func (s *boundsSearch) clear() {
	s.Values, s.Visitor, s.filter = nil, nil, nil
}

func (s *boundsSearch) prepare() {
	if s.accept == nil {
		s.accept = s.acceptItem
	}
}

func (s *boundsSearch) acceptItem(item record.Item) bool {
	if !accepts(s.filter, item) {
		return true
	}

	return s.Emit(item)
}
