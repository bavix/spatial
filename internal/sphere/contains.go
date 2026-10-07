package sphere

import "math"

func Contains(south, west, north, east, lat, lon float64) bool {
	if !(lat >= south && lat <= north && lon >= -180 && lon <= 180) {
		return false
	}

	return containsLongitude(west, east, lon) || math.Abs(lon) == 180 && containsLongitude(west, east, -lon)
}

func containsLongitude(west, east, lon float64) bool {
	if west <= east {
		return lon >= west && lon <= east
	}

	return lon >= west || lon <= east
}
