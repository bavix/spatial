package sphere

import "math"

const (
	SemiMajorAxis      = 6378137
	InverseFlattening  = 298.257223563
	Radius             = SemiMajorAxis * (1 - 1/(3*InverseFlattening))
	Radians            = math.Pi / 180
	MetersPerDegree    = Radians * Radius
	BoundsMargin       = 1e-10
	DistanceMargin     = 1e-5
	AntipodalThreshold = 0.99
)

type Query struct {
	lat, lon, sin, cos float64
}

func Coincident(aLat, aLon, bLat, bLon float64) bool {
	return aLat == bLat && (aLon == bLon || math.Abs(aLon-bLon) == 360 || math.Abs(aLat) == 90)
}

func NormalizeLongitude(lon float64) float64 {
	if lon >= -180 && lon < 180 {
		return lon
	}

	lon = math.Remainder(lon, 360)
	if lon == 180 {
		return -180
	}

	return lon
}

func LongitudeDifference(a, b float64) float64 {
	d := b - a
	if d >= -180 && d <= 180 {
		return d
	}

	return math.Remainder(d, 360)
}

func Prepare(lat, lon float64) Query {
	p := lat * Radians

	s, c := math.Sincos(p)
	if math.Abs(lat) == 90 {
		c = 0
	}

	return Query{
		lat: lat,
		lon: lon,
		sin: s,
		cos: c,
	}
}

func (q Query) Distance(lat, lon float64) float64 {
	delta := LongitudeDifference(q.lon, lon)
	if delta == 0 {
		return MetersPerDegree * math.Abs(lat-q.lat)
	}

	p := lat * Radians
	d := delta * Radians

	s, c := math.Sincos(p)
	if math.Abs(lat) == 90 {
		c = 0
	}

	u, v := math.Sin((lat-q.lat)*Radians/2), math.Sin(d/2)

	h := min(1, math.Hypot(u, math.Sqrt(q.cos*c)*v))
	if h*h <= AntipodalThreshold {
		return 2 * Radius * math.Atan2(h, math.Sqrt((1-h)*(1+h)))
	}

	sd, cd := math.Sincos(d)
	x := c * sd
	y := q.cos*s - q.sin*c*cd
	z := q.sin*s + q.cos*c*cd

	return Radius * math.Atan2(math.Hypot(x, y), z)
}
