# spatial

Geospatial calculations and spatial search for Go.

[PHP geometry](https://github.com/bavix/geo/blob/d146af1adf3157b19692edb1a072ab71e8724b36/src/Geometry/Geometrable.php),
[current exact conversions](https://www.nist.gov/pml/special-publication-811/nist-guide-si-appendix-b-conversion-factors/nist-guide-si-appendix-b8).

## New

Mutable index; synchronize concurrent access externally.

```go
mutable, err := spatial.New()
if err != nil {
	panic(err)
}

mutable.Insert(1, geo.P(55.7558, 37.6173))
mutable.Insert(2, geo.P(55.7512, 37.6184))
mutable.Update(2, geo.P(55.7520, 37.6190))
mutable.Remove(2)

fmt.Println(mutable.Len())
```

## Build

Build an immutable spatial index.

```go
items := spatial.Items{
	{ID: 1, Point: geo.P(55.7558, 37.6173)},
	{ID: 2, Point: geo.P(55.7512, 37.6184)},
	{ID: 3, Point: geo.P(59.9343, 30.3351)},
}

index, err := spatial.Build(items)
if err != nil {
	panic(err)
}
```

## Load

Replace all indexed points.

```go
points := []geo.Point{
	geo.P(55.7558, 37.6173),
	geo.P(55.7512, 37.6184),
}

items := make(spatial.Items, len(points))
for row, point := range points {
	items[row] = spatial.Item{ID: uint64(row), Point: point}
}
```

```go
if err := mutable.Load(items); err != nil {
	panic(err)
}
```

## Within

Find points within a radius.

```go
center := geo.P(55.76, 37.62)

for result := range index.Within(center, 20*geo.Kilometer) {
	fmt.Printf("%d: %.2f km\n", result.ID, result.Distance/geo.Kilometer)
}
```

## Batch

Validate completely before changing points.

```go
items := []spatial.Item{
	{ID: 1, Point: geo.P(55.7558, 37.6173)},
	{ID: 2, Point: geo.P(55.7512, 37.6184)},
}

if err := mutable.InsertBatch(items); err != nil {
	panic(err)
}
if err := mutable.UpdateBatch(items); err != nil {
	panic(err)
}
if err := mutable.UpsertBatch(items); err != nil {
	panic(err)
}
if err := mutable.RemoveBatch([]spatial.ID{1, 2}); err != nil {
	panic(err)
}
```

## Upsert

Insert or replace one point.

```go
ok := mutable.Upsert(1, geo.P(55.7558, 37.6173))
```

## Nearest

Find the closest points.

```go
for result := range index.Nearest(center, 10) {
	fmt.Println(result.ID, result.Distance)
}
```

## Bounds

Find points inside a viewport.

```go
viewport := geo.Bounds{
	South: 55.70,
	West:  37.50,
	North: 55.80,
	East:  37.70,
}

for item := range index.Bounds(viewport) {
	fmt.Println(item.ID, item.Point)
}
```

## Filter

Exclude points by custom criteria.

```go
hidden := map[spatial.ID]bool{2: true}
searcher := index.NewSearcher()

nearest := searcher.AppendNearest(nil, spatial.NearestQuery{
	Center: center,
	Limit:  10,
	Filter: func(item spatial.Item) bool {
		return !hidden[item.ID]
	},
})
```

## Collect

Collect results into a slice.

```go
nearby := slices.Collect(index.Within(center, 20*geo.Kilometer))
```

## Concurrent requests

Share immutable index; use a searcher per request.

```go
hidden := map[spatial.ID]bool{2: true}

http.HandleFunc("/nearby", func(w http.ResponseWriter, r *http.Request) {
	searcher := index.NewSearcher()
	results := searcher.AppendNearest(nil, spatial.NearestQuery{
		Center: center,
		Limit: 10,
		Filter: func(item spatial.Item) bool {
			return !hidden[item.ID]
		},
	})
	if err := json.NewEncoder(w).Encode(results); err != nil {
		return
	}
})
```

## Reuse buffers

Reuse one searcher sequentially; avoid nested calls.

```go
buffer := make([]spatial.Result, 0, 100)
query := spatial.WithinQuery{Center: center, Radius: 20 * geo.Kilometer}

buffer = searcher.AppendWithin(buffer[:0], query)
query.Center = geo.P(59.9343, 30.3351)
buffer = searcher.AppendWithin(buffer[:0], query)
```

## Iterator

Stream results and stop early.

```go
for item := range searcher.Bounds(spatial.BoundsQuery{Bounds: viewport}) {
	fmt.Println(item.ID, item.Point)
	break
}
```

## Options

Configure storage, algorithm and precision.

| Option | Effect | Supported by |
|---|---|---|
| `WithGrid(degrees)` | Select grid; set cell width and height in degrees. | `New`, `Build` |
| `WithKDTree(points)` | Select KD-tree; limit points per leaf. | `Build`, `BuildExternal` |
| `WithMetric(metric)` | Use custom distance calculations returning meters. | All constructors |
| `WithWGS84()` | Use more accurate, slower WGS84 ellipsoid distances. | All constructors |
| `WithE7Coordinates()` | Halve coordinate storage; round coordinates before distance calculations. | `Build` |

Defaults: `New` uses grid (0.25°); `Build` uses KD-tree (64).
Distance defaults to spherical. Later options override earlier options.

```go
index, err := spatial.Build(items,
	spatial.WithGrid(0.25),
	spatial.WithE7Coordinates(),
	spatial.WithWGS84(),
)
if err != nil {
	panic(err)
}
```

### WithGrid

Configure grid cells for viewport queries.

```go
index, err := spatial.Build(items, spatial.WithGrid(0.01))
if err != nil {
	panic(err)
}

for item := range index.Bounds(viewport) {
	fmt.Println(item.ID, item.Point)
}
```

### WithKDTree

Configure leaf capacity for nearest queries.

```go
index, err := spatial.Build(items, spatial.WithKDTree(32))
if err != nil {
	panic(err)
}

for result := range index.Nearest(center, 10) {
	fmt.Println(result.ID, result.Distance)
}
```

### WithMetric

Adapt a distance function to Metric.

```go
index, err := spatial.Build(items,
	spatial.WithMetric(spatial.DistanceFunc(geo.Distance)),
)
if err != nil {
	panic(err)
}

for result := range index.Within(center, 2*geo.Kilometer) {
	fmt.Println(result.ID, result.Distance)
}
```

### WithWGS84

Measure nearby distances on Earth's ellipsoid.

```go
index, err := spatial.Build(items, spatial.WithWGS84())
if err != nil {
	panic(err)
}

for result := range index.Within(center, 100*geo.Meter) {
	fmt.Printf("%d: %.3f m\n", result.ID, result.Distance)
}
```

### WithE7Coordinates

Store compact coordinates rounded to 0.0000001°.

```go
index, err := spatial.Build(items, spatial.WithE7Coordinates())
if err != nil {
	panic(err)
}

fmt.Println(index.StorageBytes())
for result := range index.Nearest(center, 1) {
	fmt.Println(result.Point)
}
```

## BuildExternal

Index immutable caller-owned coordinates.

```go
index, err := spatial.BuildExternal(items, spatial.WithKDTree(32))
if err != nil {
	panic(err)
}
```

## Errors

Handle errors with sentinel values.

```go
_, err := spatial.Build([]spatial.Item{{ID: 1}, {ID: 1}})
if errors.Is(err, spatial.ErrDuplicateID) {
	fmt.Println(err)
}
```

## Snapshot

Publish immutable snapshots across goroutines.

```go
var current atomic.Pointer[spatial.Index]
current.Store(index)

snapshot := current.Load()
for item := range snapshot.Bounds(viewport) {
	fmt.Println(item.ID)
}
```

## Geo

Calculate distances, bearings and destinations.

```go
a := geo.P(55.7558, 37.6173)
b := geo.P(59.9343, 30.3351)

kilometers := geo.Distance(a, b) / geo.Kilometer
bearing := geo.Bearing(a, b)
destination := geo.Destination(a, 20*geo.Kilometer, 90)
bounds := geo.BoundsAround(a, 20*geo.Kilometer)
```

## Geometry

Use straight longitude/latitude polygon edges.

```go
line := geo.Line{a, b}
polyline := geo.Polyline{a, b, geo.P(54.1931, 37.6173)}
triangle := geo.Triangle{a, b, geo.P(54.1931, 37.6173)}

polygon := geo.Polygon{
	Exterior: geo.Ring{geo.P(0, 0), geo.P(0, 10), geo.P(10, 10), geo.P(10, 0)},
	Holes: []geo.Ring{
		{geo.P(2, 2), geo.P(2, 4), geo.P(4, 4), geo.P(4, 2)},
	},
}

valid := polygon.Valid()
inside := polygon.Contains(geo.P(1, 1))
for segment := range polygon.Segments() {
	fmt.Println(segment)
}
```

## Regions

Prepare nonpolar polygons spanning under 180° longitude.

```go
include, err := region.Build([]region.Item{
	{
		ID:      1,
		Polygon: polygon,
	},
})
if err != nil {
	panic(err)
}

results := index.NewSearcher().AppendBounds(nil, spatial.BoundsQuery{
	Bounds: viewport,
	Filter: func(item spatial.Item) bool {
		return region.Allows(item.Point, include, nil)
	},
})
```

## Overlapping zones

Return every matching zone, including polygon boundaries.

```go
zones, err := region.Build([]region.Item{
	{
		ID:      101,
		Polygon: firstHexagon,
	},
	{
		ID:      102,
		Polygon: secondHexagon,
	},
})
if err != nil {
	panic(err)
}

for id := range zones.At(destination) {
	fmt.Println(id, geo.Distance(warehouses[id], destination))
}

buffer := make([]region.ID, 0, 16)
buffer = zones.AppendAt(buffer[:0], destination)
```

## Coast

Find nearest shoreline along WGS84 geodesic edges.

```go
shore, err := coast.Build([]geo.Polyline{
	{geo.P(0, -1), geo.P(0, 1)},
})
if err != nil {
	panic(err)
}

result, err := shore.Nearest(geo.P(0.001, 0), 0.1*geo.Meter)
if err != nil {
	panic(err)
}

fmt.Println(result.Distance, result.ErrorBound, result.Point)
```

## Coast searcher

Reuse prepared lines across sequential queries.

```go
searcher := shore.NewSearcher()
hotels := []geo.Point{
	geo.P(0.001, 0),
	geo.P(0.002, 0.5),
}

for _, hotel := range hotels {
	result, err := searcher.Nearest(hotel, 0.1*geo.Meter)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Distance, result.ErrorBound)
}
```

## Storage

Measure retained arrays, excluding object headers.

```go
fmt.Println(index.StorageBytes())
fmt.Println(zones.StorageBytes())
fmt.Println(shore.StorageBytes())
```

## Units

Convert meters using exact unit factors.

```go
meters := geo.Distance(a, b)
kilometers := meters / geo.Kilometer
landMiles := meters / geo.Mile
nauticalMiles := meters / geo.NauticalMile
yards := meters / geo.Yard
```
