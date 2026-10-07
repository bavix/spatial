package grid

import (
	"math"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/record"
	"github.com/bavix/spatial/internal/search"
	"github.com/bavix/spatial/internal/sphere"
	"github.com/bavix/spatial/internal/store"
)

type position struct {
	cell  uint64
	index int
}
type Mutable struct {
	layout    layout
	cells     map[uint64][]record.Item
	positions map[uint64]position
}

func New(cellSize float64) *Mutable {
	return &Mutable{
		layout:    newLayout(cellSize),
		cells:     make(map[uint64][]record.Item),
		positions: make(map[uint64]position),
	}
}

func (m *Mutable) Len() int {
	return len(m.positions)
}

func (m *Mutable) Insert(id uint64, p geo.Point) bool {
	if !p.Valid() {
		return false
	}

	if _, ok := m.positions[id]; ok {
		return false
	}

	m.insert(record.Item{
		ID:    id,
		Point: p,
	}, m.layout.key(p))

	return true
}

func (m *Mutable) Update(id uint64, p geo.Point) bool {
	if !p.Valid() {
		return false
	}

	pos, ok := m.positions[id]
	if !ok {
		return false
	}

	m.update(id, p, pos)

	return true
}

func (m *Mutable) Upsert(id uint64, p geo.Point) bool {
	if !p.Valid() {
		return false
	}

	m.upsert(record.Item{
		ID:    id,
		Point: p,
	})

	return true
}

func (m *Mutable) Remove(id uint64) bool {
	pos, ok := m.positions[id]
	if !ok {
		return false
	}

	m.remove(id, pos)

	return true
}

func (m *Mutable) Load(input record.Source) error {
	counts, n, err := store.CountBy(input, m.layout.key)
	if err != nil {
		return err
	}

	cells := make(map[uint64][]record.Item, len(counts))

	positions := make(map[uint64]position, n)

	for key, n := range counts {
		cells[key] = make([]record.Item, 0, n)
	}

	for row := range n {
		item := input.Item(row)
		key := m.layout.key(item.Point)
		positions[item.ID] = position{
			key,
			len(cells[key]),
		}
		cells[key] = append(cells[key], item)
	}

	m.cells, m.positions = cells, positions

	return nil
}

func (m *Mutable) Visit(bounds geo.Bounds, fn func(record.Item) bool) {
	if !bounds.Valid() {
		return
	}

	if bounds.West > bounds.East || bounds.West == -180 || bounds.East == 180 {
		m.visitAll(bounds, fn)

		return
	}

	x0 := min(uint64((bounds.West+180)/m.layout.size), m.layout.cols-1)
	x1 := min(uint64((bounds.East+180)/m.layout.size), m.layout.cols-1)
	y0 := min(uint64((bounds.South+90)/m.layout.size), m.layout.rows-1)
	y1 := min(uint64((bounds.North+90)/m.layout.size), m.layout.rows-1)

	width, height := x1-x0+1, y1-y0+1
	if float64(width)*float64(height) > float64(len(m.cells)) {
		m.visitAll(bounds, fn)

		return
	}

	for y := y0; y <= y1; y++ {
		row := spread(y) << 1
		for x := x0; x <= x1; x++ {
			if !search.VisitItems(m.cells[spread(x)|row], bounds, fn) {
				return
			}
		}
	}
}

func (m *Mutable) Nearest(c *search.Collector) {
	for radius := max(1, m.layout.size*sphere.MetersPerDegree); ; radius *= 2 {
		c.Clear()
		bounds := c.BoundsAround(radius)
		m.Visit(bounds, c.Accept)

		if bounds == geo.World() {
			return
		}

		worst := c.Worst()
		if math.IsInf(worst, 1) {
			continue
		}

		if worst <= radius {
			return
		}

		bounds = c.BoundsAround(worst)
		c.Clear()
		m.Visit(bounds, c.Accept)

		return
	}
}

func (m *Mutable) upsert(item record.Item) {
	pos, ok := m.positions[item.ID]
	if ok {
		m.update(item.ID, item.Point, pos)

		return
	}

	m.insert(item, m.layout.key(item.Point))
}

func (m *Mutable) update(id uint64, p geo.Point, pos position) {
	key := m.layout.key(p)
	if pos.cell == key {
		m.cells[pos.cell][pos.index].Point = p

		return
	}

	m.remove(id, pos)
	m.insert(record.Item{
		ID:    id,
		Point: p,
	}, key)
}

func (m *Mutable) visitAll(bounds geo.Bounds, fn func(record.Item) bool) {
	for _, items := range m.cells {
		if !search.VisitItems(items, bounds, fn) {
			return
		}
	}
}

func (m *Mutable) insert(item record.Item, key uint64) {
	items := m.cells[key]
	m.positions[item.ID] = position{
		key,
		len(items),
	}
	m.cells[key] = append(items, item)
}

func (m *Mutable) remove(id uint64, pos position) {
	items := m.cells[pos.cell]

	last := len(items) - 1

	if pos.index != last {
		items[pos.index] = items[last]
		m.positions[items[pos.index].ID] = pos
	}

	items[last] = record.Item{}

	if last == 0 {
		delete(m.cells, pos.cell)
	} else {
		m.cells[pos.cell] = items[:last]
	}

	delete(m.positions, id)
}
