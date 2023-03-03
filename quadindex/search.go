package quadindex

import (
	"github.com/kpfaulkner/quadmap/covering"
	"github.com/kpfaulkner/quadmap/quadtree"
)

func Search(index Index, tiles []quadtree.QuadKey, minZoom byte) ([]any, error) {
	// TODO: should this return a chan quadtree.QuadKey or something to allow streaming?
	// TODO: accept filters

	cov, err := covering.SearchRanges(tiles, minZoom)
	if err != nil {
		return nil, err
	}

	entries, err := index.Get(cov)
	if err != nil {
		return nil, err
	}

	// Exclude false positives. See
	// TODO: can probably do this more cheaply than iterating through tiles.
	byID := make(map[string]any)
	for e := range entries {
		id := e.Value.ID()
		if _, seen := byID[id]; seen {
			continue
		}
		for _, t := range tiles {
			if e.Key.IsAncestorOf(t) || t.IsAncestorOf(e.Key) {
				byID[id] = e.Value.Value()
				break
			}
		}
	}
	results := make([]any, 0, len(byID))
	for _, val := range byID {
		results = append(results, val)
	}
	return results, nil
}
