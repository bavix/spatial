package spatial

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/grid"
)

type Mutable struct {
	reader[*grid.Mutable]
}

func New(options ...Option) (*Mutable, error) {
	c, err := configure(true, options)
	if err != nil {
		return nil, err
	}

	if !c.grid || c.e7 {
		return nil, ErrInvalidConfig
	}

	engine := grid.New(c.cellSize)

	return &Mutable{
		reader: reader[*grid.Mutable]{
			engine: engine,
			metric: c.metric,
		},
	}, nil
}

func (m *Mutable) Insert(id ID, point geo.Point) bool {
	return m.engine.Insert(id, point)
}

func (m *Mutable) Update(id ID, point geo.Point) bool {
	return m.engine.Update(id, point)
}

func (m *Mutable) Remove(id ID) bool {
	return m.engine.Remove(id)
}

func (m *Mutable) Load(items []Item) error {
	return m.engine.Load(Items(items))
}

func (m *Mutable) Upsert(id ID, point geo.Point) bool {
	return m.engine.Upsert(id, point)
}

func (m *Mutable) InsertBatch(items []Item) error {
	return m.engine.InsertBatch(items)
}

func (m *Mutable) UpdateBatch(items []Item) error {
	return m.engine.UpdateBatch(items)
}

func (m *Mutable) UpsertBatch(items []Item) error {
	return m.engine.UpsertBatch(items)
}

func (m *Mutable) RemoveBatch(ids []ID) error {
	return m.engine.RemoveBatch(ids)
}
