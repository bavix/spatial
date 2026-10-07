package grid

import (
	"slices"
	"unsafe"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/record"
	"github.com/bavix/spatial/internal/search"
	"github.com/bavix/spatial/internal/store"
	"github.com/bavix/spatial/internal/traversal"
)

type Engine[T store.Offset] struct {
	data    *store.Store
	layout  layout
	keys    []uint64
	offsets []T
}

func Build[T store.Offset](input record.Source, cellSize float64, e7 bool) (*Engine[T], error) {
	data, err := store.Build(input, e7)
	if err != nil {
		return nil, err
	}

	layout := newLayout(cellSize)
	groups := store.Group[T](data, layout.key)

	return &Engine[T]{
		data:    data,
		layout:  layout,
		keys:    groups.Keys,
		offsets: groups.Offsets,
	}, nil
}

func (e *Engine[T]) StorageBytes() uint64 {
	return e.data.StorageBytes() + uint64(cap(e.keys))*8 + uint64(cap(e.offsets))*uint64(unsafe.Sizeof(T(0)))
}

func (e *Engine[T]) Len() int {
	return e.data.Len()
}

type cellFrame struct {
	lo, hi int
	x, y   uint64
	depth  uint
	prefix uint64
	lower  float64
}

func (e *Engine[T]) Visit(bounds geo.Bounds, fn func(record.Item) bool) {
	if !bounds.Valid() || e.Len() == 0 {
		return
	}

	var stack traversal.Stack[cellFrame]

	stack.Push(cellFrame{
		hi:    len(e.keys),
		depth: e.layout.depth,
	})

	for stack.Len() > 0 {
		f := stack.Pop()
		if !search.Intersects(e.layout.bounds(f.x, f.y, uint64(1)<<f.depth), bounds) {
			continue
		}

		if f.hi-f.lo == 1 {
			if !e.data.VisitKeys(int(e.offsets[f.lo]), int(e.offsets[f.hi]), bounds, bounds, fn) {
				return
			}

			continue
		}

		var children [4]cellFrame

		count := e.children(f, &children)
		stack.Push(children[:count]...)
	}
}

func (e *Engine[T]) Nearest(c *search.Collector) {
	if e.Len() == 0 {
		return
	}

	var stack traversal.Stack[cellFrame]

	stack.Push(cellFrame{
		hi:    len(e.keys),
		depth: e.layout.depth,
	})

	for stack.Len() > 0 {
		f := stack.Pop()
		if f.lower > c.Worst() {
			continue
		}

		if f.hi-f.lo == 1 {
			start, end := int(e.offsets[f.lo]), int(e.offsets[f.hi])
			for i := start; i < end; i++ {
				c.Offer(e.data.Item(i))
			}

			continue
		}

		var children [4]cellFrame

		count := e.children(f, &children)
		for j := range count {
			children[j].lower = c.LowerBound(e.layout.bounds(children[j].x, children[j].y, uint64(1)<<children[j].depth))
			for k := j; k > 0 && children[k].lower > children[k-1].lower; k-- {
				children[k], children[k-1] = children[k-1], children[k]
			}
		}

		stack.Push(children[:count]...)
	}
}

func (e *Engine[T]) children(f cellFrame, out *[4]cellFrame) int {
	depth := f.depth - 1
	width := uint64(1) << depth

	start, n := f.lo, 0

	for child := range 4 {
		if start == f.hi {
			break
		}

		end := f.hi

		if child < 3 {
			upper := f.prefix | uint64(child+1)<<(2*depth)
			offset, _ := slices.BinarySearch(e.keys[start:f.hi], upper)
			end = start + offset
		}

		x, y := f.x+uint64(child&1)*width, f.y+uint64(child>>1)*width
		if start < end && x < e.layout.cols && y < e.layout.rows {
			out[n] = cellFrame{
				lo:     start,
				hi:     end,
				x:      x,
				y:      y,
				depth:  depth,
				prefix: f.prefix | uint64(child)<<(2*depth),
			}
			n++
		}

		start = end
	}

	return n
}
