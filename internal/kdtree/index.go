package kdtree

import (
	"math/bits"
	"sort"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/record"
	"github.com/bavix/spatial/internal/search"
	"github.com/bavix/spatial/internal/store"
	"github.com/bavix/spatial/internal/traversal"
)

type Engine struct {
	data   *store.Store
	leaf   int
	bounds geo.Bounds
}

func Build(input record.Source, leafSize int, e7 bool) (*Engine, error) {
	data, err := store.Build(input, e7)
	if err != nil {
		return nil, err
	}

	return build(data, leafSize), nil
}

func BuildExternal(source record.Source, leafSize int) (*Engine, error) {
	data, err := store.External(source)
	if err != nil {
		return nil, err
	}

	return build(data, leafSize), nil
}

func build(data *store.Store, leafSize int) *Engine {
	e := &Engine{
		data:   data,
		leaf:   leafSize,
		bounds: tightBounds(data),
	}
	e.partition(0, data.Len(), 0)

	return e
}

func (e *Engine) StorageBytes() uint64 {
	return e.data.StorageBytes()
}

func (e *Engine) Len() int {
	return e.data.Len()
}

type axisRange struct {
	e            *Engine
	lo, hi, axis int
}

func (r axisRange) Len() int {
	return r.hi - r.lo
}

func (r axisRange) Less(i, j int) bool {
	return r.e.data.AxisKey(r.lo+i, r.axis) < r.e.data.AxisKey(r.lo+j, r.axis)
}

func (r axisRange) Swap(i, j int) {
	r.e.data.Swap(r.lo+i, r.lo+j)
}

type frame struct {
	lo, hi, axis int
	bounds       geo.Bounds
	lower        float64
}

func (e *Engine) Visit(bounds geo.Bounds, fn func(record.Item) bool) {
	if !bounds.Valid() || e.Len() == 0 {
		return
	}

	query := e.data.Expanded(bounds)

	var stack traversal.Stack[frame]

	stack.Push(frame{
		lo:     0,
		hi:     e.Len(),
		bounds: e.bounds,
	})

	for stack.Len() > 0 {
		f := stack.Pop()
		if !search.Intersects(query, f.bounds) {
			continue
		}

		if f.hi-f.lo <= e.leaf {
			if !e.data.VisitKeys(f.lo, f.hi, bounds, query, fn) {
				return
			}

			continue
		}

		a, b, m := e.children(f)
		if search.Contains(query, e.data.Key(m)) {
			item := e.data.Item(m)
			if search.Contains(bounds, item.Point) && !fn(item) {
				return
			}
		}

		stack.Push(a, b)
	}
}

func (e *Engine) Nearest(c *search.Collector) {
	if e.Len() == 0 {
		return
	}

	var stack traversal.Stack[frame]

	stack.Push(frame{
		lo:     0,
		hi:     e.Len(),
		bounds: e.bounds,
	})

	for stack.Len() > 0 {
		f := stack.Pop()
		if f.lower > c.Worst() {
			continue
		}

		if f.hi-f.lo <= e.leaf {
			for i := f.lo; i < f.hi; i++ {
				c.Offer(e.data.Item(i))
			}

			continue
		}

		a, b, m := e.children(f)
		c.Offer(e.data.Item(m))

		a.lower, b.lower = c.LowerBound(e.data.Expanded(a.bounds)), c.LowerBound(e.data.Expanded(b.bounds))
		if a.lower < b.lower {
			a, b = b, a
		}

		stack.Push(a, b)
	}
}

func (e *Engine) partition(lo, hi, axis int) {
	if hi-lo <= e.leaf {
		return
	}

	mid := lo + (hi-lo)/2
	e.selectMedian(lo, hi, mid, axis)
	e.partition(lo, mid, 1-axis)
	e.partition(mid+1, hi, 1-axis)
}

func (e *Engine) selectMedian(lo, hi, k, axis int) {
	budget := 2 * bits.Len(uint(hi-lo))
	for hi-lo > 1 {
		if budget == 0 {
			sort.Sort(axisRange{
				e,
				lo,
				hi,
				axis,
			})

			return
		}

		budget--
		l, r := e.split(lo, hi, axis)

		switch {
		case k < l:
			hi = l
		case k >= r:
			lo = r
		default:
			return
		}
	}
}

func (e *Engine) children(f frame) (frame, frame, int) {
	m := f.lo + (f.hi-f.lo)/2
	p := e.data.Key(m)

	a, b := frame{
		lo:     f.lo,
		hi:     m,
		axis:   1 - f.axis,
		bounds: f.bounds,
	}, frame{
		lo:     m + 1,
		hi:     f.hi,
		axis:   1 - f.axis,
		bounds: f.bounds,
	}

	if f.axis == 0 {
		a.bounds.East = p.Lon
		b.bounds.West = p.Lon
	} else {
		a.bounds.North = p.Lat
		b.bounds.South = p.Lat
	}

	return a, b, m
}

func (e *Engine) split(lo, hi, axis int) (int, int) {
	a, b, c := e.data.AxisKey(lo, axis), e.data.AxisKey(lo+(hi-lo)/2, axis), e.data.AxisKey(hi-1, axis)
	pivot := max(min(a, b), min(max(a, b), c))

	l, i, r := lo, lo, hi
	for i < r {
		x := e.data.AxisKey(i, axis)
		switch {
		case x < pivot:
			e.data.Swap(l, i)

			l++
			i++
		case x > pivot:
			r--
			e.data.Swap(i, r)
		default:
			i++
		}
	}

	return l, r
}
