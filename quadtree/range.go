package quadtree

// QuadKeyRange is a range containing QuadKeys.
type QuadKeyRange struct {
	// Start and endpoints of the range, both inclusive.
	// Note: Start and End aren't, in general, valid QuadKeys themselves.
	Start, End uint64
}

// Range returns the range of QuadKeys that contains q and all of its descendants.
func (q QuadKey) Range() (r QuadKeyRange) {
	// The range calculated here also includes some false positives of keys at
	// a lower zoom level in the 0,0 corner for the intervening levels. This
	// isn't too bad, it won't be many and we just need to double check
	// intersections with QuadKey.IsAncestorOf.
	//
	// TODO: find a way to avoid this. Keeping the zoom level in the lower bits
	// might work, although that prevents combining adjacent ranges.
	z := q.Zoom()
	mask := ^uint64(0) >> (z * 2)
	r.Start = uint64(q) & ^mask
	r.End = r.Start | mask
	return
}

// SingleRange returns a range that only contains this QuadKey.
func (q QuadKey) SingleRange() QuadKeyRange {
	return QuadKeyRange{uint64(q), uint64(q)}
}

func (r QuadKeyRange) Contains(q QuadKey) bool {
	return r.Start <= uint64(q) && uint64(q) <= r.End
}
