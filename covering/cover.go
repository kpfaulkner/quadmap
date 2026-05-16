package covering

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/kpfaulkner/quadmap/quadmap"
	"github.com/peterstace/simplefeatures/geom"
)

type coveringTile struct {
	qk quadmap.QuadKey
	// area of the tile that lies outside the geometry
	outsideArea float64
}

type priorityQueue []coveringTile

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool {
	// Tiles with more outside area should be processed first,
	// so reverse the order.
	return pq[i].outsideArea > pq[j].outsideArea
}
func (pq priorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x any)   { *pq = append(*pq, x.(coveringTile)) }
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}

func intersection(qk quadmap.QuadKey, g geom.Geometry) (coveringTile, bool, error) {
	tileEnv, err := qk.Envelope()
	if err != nil {
		return coveringTile{}, false, err
	}
	tileGeom := tileEnv.AsGeometry()

	// Fast path: if g fully covers the tile, outsideArea is exactly 0. This
	// avoids both the cost of geom.Intersection and the float-precision drift
	// in `tileEnv.Area() - inter.Area()` that would otherwise leave a
	// fully-covered tile with a tiny non-zero outsideArea and cause the heap
	// to subdivide it needlessly.
	covers, err := geom.Covers(g, tileGeom)
	if err != nil {
		return coveringTile{}, false, err
	}
	if covers {
		return coveringTile{qk, 0}, true, nil
	}

	inter, err := geom.Intersection(g, tileGeom)
	if err != nil {
		return coveringTile{}, false, err
	}
	if inter.IsEmpty() {
		return coveringTile{}, false, nil
	}
	// TODO: correct area for web mercator distortion? Won't matter in most cases.
	return coveringTile{qk, tileEnv.Area() - inter.Area()}, true, nil
}

// ExteriorCovering returns a set of QuadKeys that approximates a Geometry
// with no more than maxTiles keys. The covering fully covers the geometry,
// but may also include some area outside it.
func ExteriorCovering(g geom.Geometry, maxTiles int) ([]quadmap.QuadKey, error) {
	return exteriorCovering(g, maxTiles, quadmap.MaxZoom)
}

// ExteriorCoveringNoMax returns a covering with no cap on the number of tiles,
// subdividing down to (but not past) maxZoom. The covering fully covers the
// geometry, but may also include some area outside it.
func ExteriorCoveringNoMax(g geom.Geometry, maxZoom byte) ([]quadmap.QuadKey, error) {
	return exteriorCovering(g, math.MaxInt, maxZoom)
}

// exteriorCovering is the shared engine. It repeatedly subdivides the cell
// with the largest area outside g (priority-queue order) until one of:
//   - the worst remaining cell has outsideArea == 0 (covering is fully inside g);
//   - the worst remaining cell is at maxZoom (no further precision available);
//   - subdividing would push the cell count past maxTiles.
//
// pq is ordered by outsideArea descending, so when the top cell can't be
// improved no other cell can either — that's why pushing it back and breaking
// is sufficient to terminate.
func exteriorCovering(g geom.Geometry, maxTiles int, maxZoom byte) ([]quadmap.QuadKey, error) {
	score, ok, err := intersection(0, g)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	pq := priorityQueue{score}
	for len(pq) > 0 {
		cell := heap.Pop(&pq).(coveringTile)
		if cell.outsideArea == 0 {
			heap.Push(&pq, cell)
			break
		}
		if cell.qk.Zoom() >= maxZoom {
			heap.Push(&pq, cell)
			break
		}

		var next []coveringTile
		for _, ch := range cell.qk.Children() {
			score, overlap, err := intersection(ch, g)
			if err != nil {
				return nil, err
			}
			if overlap {
				next = append(next, score)
			}
		}
		if len(pq)+len(next) > maxTiles {
			heap.Push(&pq, cell)
			break
		}
		for _, c := range next {
			heap.Push(&pq, c)
		}
	}
	cover := make([]quadmap.QuadKey, len(pq))
	for i, c := range pq {
		cover[i] = c.qk
	}
	return cover, nil
}

// AllAncestors returns the deduplicated set of QuadKeys reachable by walking
// up from each key in quadKeys until reaching a key whose zoom is <= minZoom
// (which itself is excluded). The result is sorted ascending so callers get a
// deterministic order independent of Go's map-iteration randomisation.
func AllAncestors(quadKeys []quadmap.QuadKey, minZoom byte) ([]quadmap.QuadKey, error) {
	seen := make(map[quadmap.QuadKey]bool)
	for _, qk := range quadKeys {
		for {
			if qk.Zoom() <= minZoom {
				break
			}
			var err error
			qk, err = qk.Parent()
			if err != nil {
				return nil, err
			}
			if seen[qk] {
				break
			}
			seen[qk] = true
		}
	}
	ancestors := make([]quadmap.QuadKey, 0, len(seen))
	for c := range seen {
		ancestors = append(ancestors, c)
	}
	sort.Slice(ancestors, func(i, j int) bool { return ancestors[i] < ancestors[j] })
	return ancestors, nil
}

// SearchRanges takes a set of QuadKeys and returns a sorted list of ranges
// that contain all of those tiles and any of their descendants.
//
// Note: these ranges also find a small number of keys outside those tiles (see
// the note in QuadKey.Range()). Use QuadKey.IsAncestorOf() to filter those out
// of the results.
func SearchRanges(quadKeys []quadmap.QuadKey, minZoom byte) ([]quadmap.QuadKeyRange, error) {
	ancestors, err := AllAncestors(quadKeys, minZoom)
	if err != nil {
		return nil, err
	}

	ranges := make([]quadmap.QuadKeyRange, 0, len(quadKeys)+len(ancestors))
	for _, qk := range quadKeys {
		ranges = append(ranges, qk.Range())
	}
	for _, a := range ancestors {
		ranges = append(ranges, a.SingleRange())
	}
	if len(ranges) == 0 {
		return nil, nil
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].Start < ranges[j].Start })

	// Two-pointer compaction: i is the write head into the merged prefix,
	// j scans forward. ranges[i] absorbs ranges[j] when they overlap or are
	// contiguous; otherwise ranges[j] is written to ranges[i+1].
	i := 0
	for j := 1; j < len(ranges); j++ {
		ri := &ranges[i]
		rj := &ranges[j]
		if ri.End >= rj.Start || // ranges overlap
			ri.End == rj.Start-1 { // ranges are contiguous (beware overflow)
			if ri.End < rj.End {
				ri.End = rj.End
			}
			continue
		}

		i++
		ranges[i] = ranges[j]
	}
	return ranges[:i+1], nil
}

// ToGeoJSON renders a coverage of QuadKeys as a GeoJSON
// FeatureCollection — one Polygon feature per tile, with quadkey / slippy
// coords attached as properties. Intended for piping into tools like
// geojson.io to eyeball a covering visually.
func ToGeoJSON(cover []quadmap.QuadKey) (string, error) {
	type polygonGeom struct {
		Type        string         `json:"type"`
		Coordinates [][][2]float64 `json:"coordinates"`
	}
	type feature struct {
		Type       string         `json:"type"`
		Geometry   polygonGeom    `json:"geometry"`
		Properties map[string]any `json:"properties"`
	}
	type featureCollection struct {
		Type     string    `json:"type"`
		Features []feature `json:"features"`
	}

	fc := featureCollection{
		Type:     "FeatureCollection",
		Features: make([]feature, 0, len(cover)),
	}
	for _, qk := range cover {
		env, err := qk.Envelope()
		if err != nil {
			return "", err
		}
		min, max, ok := env.MinMaxXYs()
		if !ok {
			continue
		}
		x, y, z := qk.SlippyCoords()
		fc.Features = append(fc.Features, feature{
			Type: "Feature",
			Geometry: polygonGeom{
				Type: "Polygon",
				// GeoJSON exterior ring: CCW, closed (first == last).
				Coordinates: [][][2]float64{{
					{min.X, min.Y},
					{max.X, min.Y},
					{max.X, max.Y},
					{min.X, max.Y},
					{min.X, min.Y},
				}},
			},
			Properties: map[string]any{
				"quadkey": fmt.Sprintf("0x%016x", uint64(qk)),
				"x":       x,
				"y":       y,
				"z":       z,
			},
		})
	}

	b, err := json.Marshal(fc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
