package store

import (
	"math"
	"math/bits"
	"slices"

	"github.com/pkg/errors"

	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/fault"
	"github.com/bavix/spatial/internal/record"
	"github.com/bavix/spatial/internal/search"
	"github.com/bavix/spatial/internal/sphere"
)

const (
	scale          = 1e7
	step           = 1 / scale
	externalMargin = step/2 + sphere.BoundsMargin
)

type encodedPoint struct {
	Lat, Lon int32
}

func encode(p geo.Point) encodedPoint {
	return encodedPoint{
		int32(math.Round(p.Lat * scale)),
		int32(math.Round(p.Lon * scale)),
	}
}

func (p encodedPoint) decode() geo.Point {
	return geo.P(float64(p.Lat)/scale, float64(p.Lon)/scale)
}

type Store struct {
	ids     []uint64
	points  []geo.Point
	encoded []encodedPoint
	rows    []uint32
	wide    []int
	source  record.Source
}

func validateIDs(ids []uint64) error {
	slices.Sort(ids)

	for i := 1; i < len(ids); i++ {
		if ids[i] == ids[i-1] {
			return errors.Wrapf(fault.ErrDuplicateID, "ID %d", ids[i])
		}
	}

	return nil
}

func newStore(n int, e7 bool) *Store {
	s := &Store{
		ids: make([]uint64, n),
	}
	if e7 {
		s.encoded = make([]encodedPoint, n)
	} else {
		s.points = make([]geo.Point, n)
	}

	return s
}

func Build(input record.Source, e7 bool) (*Store, error) {
	n, err := sourceLen(input)
	if err != nil {
		return nil, err
	}

	s := newStore(n, e7)
	for i := range n {
		item := input.Item(i)
		if !item.Point.Valid() {
			return nil, fault.Point(i, item.ID)
		}

		s.set(i, item)
	}

	ids := s.ids
	if !slices.IsSorted(ids) {
		ids = slices.Clone(ids)
	}

	err = validateIDs(ids)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func sourceLen(source record.Source) (int, error) {
	if source == nil {
		return 0, fault.ErrInvalidSource
	}

	n := source.Len()
	if n < 0 {
		return 0, errors.Wrapf(fault.ErrInvalidSource, "length %d", n)
	}

	return n, nil
}

func External(source record.Source) (*Store, error) {
	n, err := sourceLen(source)
	if err != nil {
		return nil, err
	}

	s := &Store{
		source:  source,
		encoded: make([]encodedPoint, n),
	}
	if uint64(len(s.encoded)) <= uint64(math.MaxUint32)+1 {
		s.rows = make([]uint32, n)
	} else {
		s.wide = make([]int, n)
	}

	ids := make([]uint64, n)
	for i := range n {
		item := source.Item(i)
		if !item.Point.Valid() {
			return nil, fault.Point(i, item.ID)
		}

		ids[i] = item.ID

		s.encoded[i] = encode(item.Point)
		if s.rows != nil {
			s.rows[i] = uint32(i)
		} else {
			s.wide[i] = i
		}
	}

	err = validateIDs(ids)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) Len() int {
	return max(len(s.ids), len(s.encoded))
}

func (s *Store) VisitKeys(lo, hi int, bounds, candidates geo.Bounds, fn func(record.Item) bool) bool {
	if s.source != nil {
		return s.visitExternal(lo, hi, bounds, candidates, fn)
	}

	for i := lo; i < hi; i++ {
		point := s.Key(i)
		if !search.Contains(candidates, point) {
			continue
		}

		item := record.Item{
			ID:    s.ids[i],
			Point: point,
		}

		if !fn(item) {
			return false
		}
	}

	return true
}

func (s *Store) Key(i int) geo.Point {
	if s.encoded != nil {
		return s.encoded[i].decode()
	}

	return s.points[i]
}

func (s *Store) AxisKey(row, axis int) float64 {
	if axis == 0 {
		if s.encoded != nil {
			return float64(s.encoded[row].Lon)
		}

		return s.points[row].Lon
	}

	if s.encoded != nil {
		return float64(s.encoded[row].Lat)
	}

	return s.points[row].Lat
}

func (s *Store) Item(i int) record.Item {
	if s.source != nil {
		if s.rows != nil {
			return s.source.Item(int(s.rows[i]))
		}

		return s.source.Item(s.wide[i])
	}

	return record.Item{
		ID:    s.ids[i],
		Point: s.Key(i),
	}
}

func (s *Store) Swap(i, j int) {
	if s.encoded != nil {
		s.encoded[i], s.encoded[j] = s.encoded[j], s.encoded[i]
	} else {
		s.points[i], s.points[j] = s.points[j], s.points[i]
	}

	switch {
	case s.source == nil:
		s.ids[i], s.ids[j] = s.ids[j], s.ids[i]
	case s.rows != nil:
		s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
	default:
		s.wide[i], s.wide[j] = s.wide[j], s.wide[i]
	}
}

func (s *Store) Expanded(b geo.Bounds) geo.Bounds {
	if s.source == nil {
		return b
	}

	return Expand(b, externalMargin)
}

func Expand(b geo.Bounds, margin float64) geo.Bounds {
	b.South = max(-90, b.South-margin)
	b.North = min(90, b.North+margin)

	width := b.East - b.West
	if width < 0 {
		width += 360
	}

	if width+2*margin >= 360 {
		b.West, b.East = -180, 180

		return b
	}

	b.West = sphere.NormalizeLongitude(b.West - margin)
	b.East = sphere.NormalizeLongitude(b.East + margin)

	return b
}

func (s *Store) StorageBytes() uint64 {
	const intBytes = bits.UintSize / 8

	return uint64(cap(s.ids))*8 + uint64(cap(s.points))*16 + uint64(cap(s.encoded))*8 +
		uint64(cap(s.rows))*4 + uint64(cap(s.wide))*intBytes
}

func (s *Store) set(i int, item record.Item) {
	s.ids[i] = item.ID
	if s.encoded != nil {
		s.encoded[i] = encode(item.Point)
	} else {
		s.points[i] = item.Point
	}
}

func (s *Store) visitExternal(lo, hi int, bounds, candidates geo.Bounds, fn func(record.Item) bool) bool {
	for i := lo; i < hi; i++ {
		if !search.Contains(candidates, s.Key(i)) {
			continue
		}

		item := s.Item(i)
		if search.Contains(bounds, item.Point) && !fn(item) {
			return false
		}
	}

	return true
}
