package spatial

import (
	"iter"

	"github.com/bavix/spatial/internal/record"
	"github.com/bavix/spatial/internal/search"
)

type Searcher struct {
	engine  search.Engine
	metric  Metric
	within  radiusSearch
	box     boundsSearch
	nearest search.Collector
}

func (s *Searcher) Within(q WithinQuery) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		s.visitWithin(q, yield)
	}
}

func (s *Searcher) Nearest(q NearestQuery) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		visitResults(s.collect(q), yield)
	}
}

func (s *Searcher) Bounds(q BoundsQuery) iter.Seq[Item] {
	return func(yield func(Item) bool) {
		s.visitBounds(q, yield)
	}
}

func (s *Searcher) AppendWithin(dst []Result, q WithinQuery) []Result {
	if !q.Valid() {
		return dst
	}

	s.within.Values, s.within.query = dst, q
	defer s.within.clear()

	s.engine.Visit(s.within.prepare(s.metric), s.within.accept)

	return s.within.Values
}

func (s *Searcher) AppendNearest(dst []Result, q NearestQuery) []Result {
	return append(dst, s.collect(q)...)
}

func (s *Searcher) AppendBounds(dst []Item, q BoundsQuery) []Item {
	if !q.Valid() {
		return dst
	}

	s.box.Values, s.box.filter = dst, q.Filter
	defer s.box.clear()

	s.box.prepare()
	s.engine.Visit(q.Bounds, s.box.accept)

	return s.box.Values
}

func (s *Searcher) visitWithin(q WithinQuery, fn func(Result) bool) {
	if fn == nil || !q.Valid() {
		return
	}

	s.within.Visitor, s.within.query = fn, q
	defer s.within.clear()

	s.engine.Visit(s.within.prepare(s.metric), s.within.accept)
}

func visitResults(results []record.Result, fn func(Result) bool) {
	for _, result := range results {
		if !fn(result) {
			return
		}
	}
}

func (s *Searcher) visitBounds(q BoundsQuery, fn func(Item) bool) {
	if fn == nil || !q.Valid() {
		return
	}

	s.box.Visitor, s.box.filter = fn, q.Filter
	defer s.box.clear()

	s.box.prepare()
	s.engine.Visit(q.Bounds, s.box.accept)
}

func (s *Searcher) collect(q NearestQuery) []record.Result {
	if !q.Valid() || q.Limit == 0 || s.engine.Len() == 0 {
		return nil
	}

	s.nearest.Filter = q.Filter
	defer func() {
		s.nearest.Filter = nil
	}()

	s.nearest.Reset(s.metric, q.Center, min(q.Limit, s.engine.Len()))
	s.engine.Nearest(&s.nearest)

	return s.nearest.Results()
}

func accepts(filter Filter, item Item) bool {
	return filter == nil || filter(item)
}
