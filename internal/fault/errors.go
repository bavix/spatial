package fault

import "github.com/pkg/errors"

var (
	ErrInvalidPoint  = errors.New("invalid point")
	ErrDuplicateID   = errors.New("duplicate ID")
	ErrInvalidConfig = errors.New("invalid configuration")
	ErrInvalidSource = errors.New("invalid source")
	ErrNotFound      = errors.New("ID not found")
)

func Point(row int, id uint64) error {
	return errors.Wrapf(ErrInvalidPoint, "item %d, ID %d", row, id)
}
