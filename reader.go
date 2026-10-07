package spatial

import (
	"iter"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/search"
)

type reader[T search.Engine] struct {
	engine T
	metric Metric
}

func (r reader[T]) Len() int {
	return r.engine.Len()
}

func (r reader[T]) NewSearcher() *Searcher {
	return &Searcher{
		engine: r.engine,
		metric: r.metric,
	}
}

func (r reader[T]) Within(center geo.Point, radius float64) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		q := WithinQuery{
			Center: center,
			Radius: radius,
		}
		if !q.Valid() || r.Len() == 0 {
			return
		}

		s := radiusSearch{
			query: q,
		}
		s.Visitor = yield
		r.engine.Visit(s.prepare(r.metric), s.accept)
	}
}

func (r reader[T]) Nearest(center geo.Point, limit int) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		if !center.Valid() || limit <= 0 || r.Len() == 0 {
			return
		}

		c := search.Collector{}
		c.Reset(r.metric, center, min(limit, r.Len()))
		r.engine.Nearest(&c)
		visitResults(c.Results(), yield)
	}
}

func (r reader[T]) Bounds(bounds geo.Bounds) iter.Seq[Item] {
	return func(yield func(Item) bool) {
		if !bounds.Valid() || r.Len() == 0 {
			return
		}

		s := boundsSearch{}
		s.Visitor = yield
		r.engine.Visit(bounds, s.acceptItem)
	}
}
