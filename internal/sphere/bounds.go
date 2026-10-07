package sphere

import "math"

const lowerBoundSlack = 128 * 0x1p-52

func (q Query) LowerBound(south, west, north, east float64) float64 {
	if west > east {
		return min(q.LowerBound(south, west, north, 180), q.LowerBound(south, -180, north, east))
	}

	delta := 0.0
	if q.lon < west || q.lon > east {
		delta = min(math.Abs(LongitudeDifference(q.lon, west)), math.Abs(LongitudeDifference(q.lon, east))) * Radians
	} else if q.lat >= south && q.lat <= north {
		return 0
	}

	a, c := q.sin, q.cos*math.Cos(delta)
	lo, hi := south*Radians, north*Radians
	dot := func(x float64) float64 {
		s, cosine := math.Sincos(x)

		return a*s + c*cosine
	}
	best := max(dot(lo), dot(hi))

	x := math.Atan2(a, c)
	if x >= lo && x <= hi {
		best = max(best, math.Hypot(a, c))
	}

	best = min(1, best+lowerBoundSlack)

	return max(0, Radius*math.Acos(max(-1, best))-DistanceMargin)
}
