package spatial

import (
	"math"

	"github.com/bavix/spatial/internal/grid"
	"github.com/bavix/spatial/internal/kdtree"
	"github.com/bavix/spatial/internal/search"
	"github.com/bavix/spatial/internal/store"
)

type Index struct {
	reader[search.StaticEngine]
}

func (i *Index) StorageBytes() uint64 {
	return i.engine.StorageBytes()
}

func Build(items []Item, options ...Option) (*Index, error) {
	c, err := configure(false, options)
	if err != nil {
		return nil, err
	}

	return c.build(Items(items))
}

func (c config) build(input Source) (*Index, error) {
	if c.grid {
		if size := input.Len(); size >= 0 && uint64(size) <= math.MaxUint32 {
			return buildGrid[uint32](input, c)
		}

		return buildGrid[int](input, c)
	}

	engine, err := kdtree.Build(input, c.leafSize, c.e7)
	if err != nil {
		return nil, err
	}

	return newIndex(engine, c.metric), nil
}

func buildGrid[T store.Offset](input Source, c config) (*Index, error) {
	engine, err := grid.Build[T](input, c.cellSize, c.e7)
	if err != nil {
		return nil, err
	}

	return newIndex(engine, c.metric), nil
}

func BuildExternal(input Source, options ...Option) (*Index, error) {
	c, err := configure(false, options)
	if err != nil {
		return nil, err
	}

	if c.grid || c.e7 {
		return nil, ErrInvalidConfig
	}

	engine, err := kdtree.BuildExternal(input, c.leafSize)
	if err != nil {
		return nil, err
	}

	return newIndex(engine, c.metric), nil
}

func newIndex(engine search.StaticEngine, distance Metric) *Index {
	return &Index{
		reader: reader[search.StaticEngine]{
			engine: engine,
			metric: distance,
		},
	}
}
