package record

import "github.com/bavix/spatial/geo"

type Item struct {
	ID    uint64
	Point geo.Point
}
type Result struct {
	ID       uint64
	Point    geo.Point
	Distance float64
}
type Source interface {
	Len() int
	Item(row int) Item
}
