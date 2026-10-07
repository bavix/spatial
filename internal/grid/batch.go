package grid

import (
	"github.com/pkg/errors"

	"github.com/bavix/spatial/internal/fault"
	"github.com/bavix/spatial/internal/record"
)

type mutation uint8

const (
	insertOnly mutation = iota
	updateOnly
	upsert
)

func (m *Mutable) InsertBatch(items []record.Item) error {
	return m.batch(items, insertOnly)
}

func (m *Mutable) UpdateBatch(items []record.Item) error {
	return m.batch(items, updateOnly)
}

func (m *Mutable) UpsertBatch(items []record.Item) error {
	return m.batch(items, upsert)
}

func (m *Mutable) batch(items []record.Item, mode mutation) error {
	err := m.validateBatch(items, mode)
	if err != nil {
		return err
	}

	for _, item := range items {
		m.upsert(item)
	}

	return nil
}

func (m *Mutable) validateBatch(items []record.Item, mode mutation) error {
	for row, item := range items {
		if !item.Point.Valid() {
			return fault.Point(row, item.ID)
		}
	}

	seen := make(map[uint64]struct{}, len(items))
	for row, item := range items {
		if _, ok := seen[item.ID]; ok {
			return batchError(fault.ErrDuplicateID, row, item.ID)
		}

		seen[item.ID] = struct{}{}

		_, exists := m.positions[item.ID]
		if mode == insertOnly && exists {
			return batchError(fault.ErrDuplicateID, row, item.ID)
		}

		if mode == updateOnly && !exists {
			return batchError(fault.ErrNotFound, row, item.ID)
		}
	}

	return nil
}

func (m *Mutable) RemoveBatch(ids []uint64) error {
	seen := make(map[uint64]struct{}, len(ids))
	for row, id := range ids {
		if _, ok := seen[id]; ok {
			return batchError(fault.ErrDuplicateID, row, id)
		}

		if _, ok := m.positions[id]; !ok {
			return batchError(fault.ErrNotFound, row, id)
		}

		seen[id] = struct{}{}
	}

	for _, id := range ids {
		m.remove(id, m.positions[id])
	}

	return nil
}

func batchError(err error, row int, id uint64) error {
	return errors.Wrapf(err, "item %d, ID %d", row, id)
}
