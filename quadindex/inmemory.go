package quadindex

import (
	"sort"

	"github.com/kpfaulkner/quadmap/quadtree"
)

type inMemoryQuadIndex struct {
	entries []Entry
}

type byKey []Entry

func (bk byKey) Len() int           { return len(bk) }
func (bk byKey) Less(i, j int) bool { return bk[i].Key < bk[j].Key }
func (bk byKey) Swap(i, j int)      { bk[i], bk[j] = bk[j], bk[i] }

func NewInMemoryQuadIndex(entries []Entry) Index {
	qi := inMemoryQuadIndex{
		entries: make([]Entry, len(entries)),
	}
	copy(qi.entries, entries)
	sort.Sort(byKey(qi.entries))
	return qi
}

func (qi inMemoryQuadIndex) Get(ranges []quadtree.QuadKeyRange) (<-chan Entry, error) {
	ch := make(chan Entry)
	go func() {
		for _, r := range ranges {
			start := sort.Search(len(qi.entries), func(i int) bool {
				return uint64(qi.entries[i].Key) >= r.Start
			})
			for i := start; i < len(qi.entries) && uint64(qi.entries[i].Key) <= r.End; i++ {
				ch <- qi.entries[i]
			}
		}
		close(ch)
	}()
	return ch, nil
}
