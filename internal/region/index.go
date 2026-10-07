package region

import (
	"cmp"
	"slices"
	"unsafe"

	"github.com/bavix/spatial/internal/geometry"
	"github.com/bavix/spatial/internal/sphere"
	"github.com/bavix/spatial/internal/traversal"
)

const regionLeafSize = 4

type (
	Area struct {
		geometry.Prepared

		ID uint64
	}

	areaNode struct {
		bounds     geometry.Envelope
		start, end int
		next       int
	}
)

type Index struct {
	polygons []Area
	nodes    []areaNode
}

func Build(polygons []Area) Index {
	t := Index{
		polygons: polygons,
		nodes:    make([]areaNode, 0, traversal.NodeCount(len(polygons), regionLeafSize)),
	}
	if len(polygons) > 0 {
		t.build(0, len(polygons), 0)
	}

	return t
}

func (t *Index) Len() int {
	return len(t.polygons)
}

func (t *Index) StorageBytes() uint64 {
	size := uint64(cap(t.polygons))*uint64(unsafe.Sizeof(Area{})) + uint64(cap(t.nodes))*uint64(unsafe.Sizeof(areaNode{}))
	for _, polygon := range t.polygons {
		size += polygon.StorageBytes()
	}

	return size
}

func (t *Index) Contains(point geometry.Point) bool {
	return !t.Visit(point, func(uint64) bool {
		return false
	})
}

func (t *Index) Visit(point geometry.Point, fn func(uint64) bool) bool {
	if len(t.nodes) == 0 {
		return true
	}

	point.Lon = sphere.NormalizeLongitude(point.Lon)

	return t.visit(point, fn)
}

func (t *Index) build(start, end, depth int) int {
	row := len(t.nodes)

	t.nodes = append(t.nodes, areaNode{
		start: start,
		end:   end,
		next:  row + 1,
	})
	if end-start <= regionLeafSize {
		bounds := t.polygons[start].Bounds()
		for _, p := range t.polygons[start+1 : end] {
			bounds = bounds.Merge(p.Bounds())
		}

		t.nodes[row].bounds = bounds

		return row
	}

	slices.SortFunc(t.polygons[start:end], func(a, b Area) int {
		return cmp.Compare(a.AxisKey(1-depth%2), b.AxisKey(1-depth%2))
	})
	middle := start + (end-start)/2
	left := t.build(start, middle, depth+1)
	right := t.build(middle, end, depth+1)
	t.nodes[row].next = len(t.nodes)
	t.nodes[row].bounds = t.nodes[left].bounds.Merge(t.nodes[right].bounds)

	return row
}

func (t *Index) visit(point geometry.Point, fn func(uint64) bool) bool {
	for row := 0; row < len(t.nodes); {
		n := t.nodes[row]
		if !n.bounds.Contains(point) {
			row = n.next

			continue
		}

		row++

		if n.end-n.start > regionLeafSize {
			continue
		}

		for _, p := range t.polygons[n.start:n.end] {
			if p.Contains(point) && !fn(p.ID) {
				return false
			}
		}
	}

	return true
}
