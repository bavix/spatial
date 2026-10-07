package geo

import "github.com/bavix/spatial/internal/sphere"

const (
	Centimeter  = Meter / 100
	EarthRadius = sphere.Radius
	Foot        = 12 * Inch
	Inch        = 0.0254 * Meter
	Kilometer   = 1000 * Meter
	Meter       = 1.0
	// MidWorldWheel measures distance in wheels from Stephen King's The Dark Tower.
	// PHP conversion: https://github.com/bavix/geo/blob/d146af1adf3157b19692edb1a072ab71e8724b36/src/Unit/Provider/Wheel.php#L34
	MidWorldWheel = 1.69 * Mile
	Mile          = 1609.344 * Meter
	Millimeter    = Meter / 1000
	NauticalMile  = 1852 * Meter
	Yard          = 3 * Foot
)
