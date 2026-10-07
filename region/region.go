package region

import (
	"iter"

	"github.com/pkg/errors"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/fault"
	"github.com/bavix/spatial/internal/geometry"
	"github.com/bavix/spatial/internal/region"
)

var (
	ErrInvalidPolygon = geometry.ErrInvalidPolygon
	ErrInvalidPoint   = fault.ErrInvalidPoint
	ErrDuplicateID    = fault.ErrDuplicateID
)

type ID = uint64

type Item struct {
	ID      ID
	Polygon geo.Polygon
}

type Index struct {
	engine region.Index
}

func Build(items []Item) (*Index, error) {
	ids := make(map[ID]struct{}, len(items))

	areas := make([]region.Area, len(items))
	for row, item := range items {
		if _, exists := ids[item.ID]; exists {
			return nil, errors.Wrapf(ErrDuplicateID, "item %d, ID %d", row, item.ID)
		}

		prepared, err := geometry.Prepare(item.Polygon)
		if err != nil {
			return nil, errors.Wrapf(err, "item %d, ID %d", row, item.ID)
		}

		ids[item.ID] = struct{}{}
		areas[row] = region.Area{
			ID:       item.ID,
			Prepared: prepared,
		}
	}

	return &Index{
		engine: region.Build(areas),
	}, nil
}

func (i *Index) Len() int {
	return i.engine.Len()
}

func (i *Index) Contains(point geo.Point) bool {
	return point.Valid() && i.engine.Contains(point)
}

func (i *Index) At(point geo.Point) iter.Seq[ID] {
	return func(yield func(ID) bool) {
		if point.Valid() {
			i.engine.Visit(point, yield)
		}
	}
}

func (i *Index) AppendAt(dst []ID, point geo.Point) []ID {
	if !point.Valid() {
		return dst
	}

	i.engine.Visit(point, func(id ID) bool {
		dst = append(dst, id)

		return true
	})

	return dst
}

func Allows(point geo.Point, include, exclude *Index) bool {
	if !point.Valid() {
		return false
	}

	if include != nil && include.Len() > 0 && !include.engine.Contains(point) {
		return false
	}

	return exclude == nil || !exclude.engine.Contains(point)
}

func (i *Index) StorageBytes() uint64 {
	return i.engine.StorageBytes()
}
