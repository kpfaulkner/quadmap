package covering

import (
	"container/heap"
	"fmt"
	"log"
	"sort"

	"github.com/kpfaulkner/quadmap/quadtree"
	"github.com/peterstace/simplefeatures/geom"
)

type intersectionScore struct {
	key   quadtree.QuadKey
	score float64
}

type priorityQueue []intersectionScore

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].score < pq[j].score }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x any)        { *pq = append(*pq, x.(intersectionScore)) }
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}

func scoreCell(q quadtree.QuadKey, g geom.Geometry) (intersectionScore, bool) {
	keyGeom := q.Geom()
	intersection, err := geom.Intersection(g, keyGeom)
	if err != nil {
		log.Println(err)
		return intersectionScore{}, false
	}
	// TODO: use WGS84AreaSqM? The distortion won't matter in most cases.
	if intersection.IsEmpty() {
		return intersectionScore{}, false
	}
	ia := intersection.Area()
	return intersectionScore{c, -(cg.Area() - ia)}, true
}

func ExteriorCovering(g geom.Geometry, maxCells int) []CellID { // TODO: minZoom
	score, ok := scoreCell(ZoomZeroCell, g)
	if !ok {
		return nil
	}
	pq := priorityQueue{score}
	for len(pq) > 0 {
		cell := heap.Pop(&pq).(intersectionScore)
		if cell.score == 0 {
			heap.Push(&pq, cell)
			break
		}
		if _, _, z := cell.id.XYZ(); z >= MaxZoom {
			heap.Push(&pq, cell)
			break
		}

		var next []intersectionScore
		for _, ch := range cell.id.Children() {
			score, overlap := scoreCell(ch, g)
			if overlap {
				next = append(next, score)
			}
		}
		if len(pq)+len(next) > maxCells {
			heap.Push(&pq, cell)
			break
		}
		for _, c := range next {
			heap.Push(&pq, c)
		}
	}
	cover := make([]CellID, len(pq))
	for i, c := range pq {
		cover[i] = c.id
	}
	return cover
}

func AllAncestors(cells []CellID, minZoom int) []CellID {
	seen := make(map[CellID]bool)
	for _, c := range cells {
		for {
			if _, _, z := c.XYZ(); z <= minZoom {
				break
			}
			c = c.Parent()
			if seen[c] {
				break
			}
			seen[c] = true
		}
	}
	anc := make([]CellID, 0, len(seen))
	for c := range seen {
		anc = append(anc, c)
	}
	return anc
}

func SearchRanges(cells []CellID, minZoom int) []CellRange {
	anc := AllAncestors(cells, minZoom)

	ranges := make([]CellRange, 0, len(cells)+len(anc))
	for _, c := range cells {
		ranges = append(ranges, c.Range())
	}
	for _, a := range anc {
		ranges = append(ranges, a.SingleRange())
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].Start < ranges[j].Start })

	i := 0
	for j := range ranges {
		rj := &ranges[j]
		if i == j {
			fmt.Println(i, *rj)
			continue
		}

		ri := &ranges[i]
		fmt.Println(i, j, *ri, *rj)
		if ri.End >= rj.Start || // ranges overlap
			ri.End == rj.Start-1 { // ranges are contiguous (beware overflow)
			if ri.End < rj.End {
				// rj is bigger; extend ri
				ri.End = rj.End
			}
			continue
		}

		i++
		ranges[i] = ranges[j]
	}
	return ranges[:i+1]
}
