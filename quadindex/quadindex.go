package quadindex

import "github.com/kpfaulkner/quadmap/quadtree"

type Entry struct {
	Key   quadtree.QuadKey
	Value Value
}

type Value interface {
	ID() string
	Value() any
}

type Index interface {
	Get(ranges []quadtree.QuadKeyRange) (<-chan Entry, error)
}
