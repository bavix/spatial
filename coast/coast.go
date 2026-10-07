package coast

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/coast"
	"github.com/bavix/spatial/internal/fault"
)

type (
	Result = coast.Result
	Index  struct {
		engine *coast.Index
	}
	Searcher struct {
		search coast.Searcher
	}
)

var (
	ErrInvalidPoint  = fault.ErrInvalidPoint
	ErrInvalidConfig = fault.ErrInvalidConfig
	ErrInvalidLine   = coast.ErrInvalidLine
	ErrEmpty         = coast.ErrEmpty
)

func Build(lines []geo.Polyline) (*Index, error) {
	engine, err := coast.Build(lines)
	if err != nil {
		return nil, err
	}

	return &Index{
		engine: engine,
	}, nil
}

func (i *Index) Len() int {
	return i.engine.Len()
}

func (i *Index) NewSearcher() *Searcher {
	return &Searcher{
		search: coast.Searcher{
			Index: i.engine,
		},
	}
}

func (i *Index) Nearest(point geo.Point, tolerance float64) (Result, error) {
	s := coast.Searcher{
		Index: i.engine,
	}

	return s.Nearest(point, tolerance)
}

func (s *Searcher) Nearest(point geo.Point, tolerance float64) (Result, error) {
	return s.search.Nearest(point, tolerance)
}

func (i *Index) StorageBytes() uint64 {
	return i.engine.StorageBytes()
}
