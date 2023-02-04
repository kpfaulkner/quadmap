package quadtree

import (
	"errors"
	"fmt"
)

// QuadMap is a quadtree in disguise...
type QuadMap struct {

	// map of quadkey to tile
	quadKeyMap map[uint64]*Tile
}

func NewQuadMap() *QuadMap {
	return &QuadMap{
		quadKeyMap: make(map[uint64]*Tile),
	}
}

// GetParentTile returns parent tile of passed in tile t
func (qm *QuadMap) GetParentTile(t *Tile) (*Tile, error) {
	parentKey, err := GetParentQuadKey(t.QuadKey)
	if err != nil {
		return nil, err
	}
	parentTile, ok := qm.quadKeyMap[parentKey]
	if !ok {
		return nil, errors.New("parent tile not found")
	}
	return parentTile, nil
}

// GetChildInPos returns child tile of passed in tile t which is in position pos
// pos is a number between 0 and 3, where 0 is top left, 1 is top right, 2 is bottom left and 3 is bottom right
func (qm *QuadMap) GetChildInPos(t *Tile, pos int) (*Tile, error) {
	childKey, err := GetChildQuadKeyForPos(t.QuadKey, pos)
	if err != nil {
		return nil, err
	}
	childTile, ok := qm.quadKeyMap[childKey]
	if !ok {
		return nil, errors.New(fmt.Sprintf("child tile in pos %d not found", pos))
	}
	return childTile, nil
}

// GetTileForSlippy returns tile details for slippy co-ord OR error if none available.
// It will either return the tile requested OR an ancestor that is full
// The process is:
//
//   - convert slippy to quadkey
//
//   - if quadkey exists, return tile
//
//   - if quadkey does not exist, get parent quadkey
//
//   - if parent quadkey exists and is full, return parent details
//
//   - loop until no parent.
//
//     What happens if we hit a parent that is NOT full? No tile therefore return error?
//     Returns tile (actual or parent), bool indicating if actual (true == actual, false == ancestor) and error
func (qm *QuadMap) GetTileForSlippy(x int32, y int32, z byte) (*Tile, bool, error) {

	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)

	// if actual quadkey exists, return tile.
	if t, ok := qm.quadKeyMap[quadKey]; ok {
		return t, true, nil
	}

	// check parents and upwards
	ancestorTile, err := qm.traverseForTileOrFullParent(quadKey)
	if err != nil {
		return nil, false, err
	}

	// have a result...  but need to return indicating its an ancestor that is full
	return ancestorTile, false, nil
}

// traverseForTileOrParent returns tile for quadkey OR parent tile if quadkey does not exist
// AND the parent has the full flag set. Otherwise returns not found error
func (qm *QuadMap) traverseForTileOrFullParent(quadKey uint64) (*Tile, error) {

	// if quadkey has zoom of 0...  then we're out of luck and just return error
	if GetTileZoomLevel(quadKey) == 0 {
		return nil, errors.New("no tile found")
	}

	if t, ok := qm.quadKeyMap[quadKey]; ok {
		if t.Full {
			return t, nil
		}
		return nil, errors.New("no tile found")
	}

	// get parent key and recurse
	parentKey, err := GetParentQuadKey(quadKey)
	if err != nil {
		return nil, err
	}

	return qm.traverseForTileOrFullParent(parentKey)
}

// NumberOfTilesForLevel returns number of tiles for a given zoom level.
// It will NOT include parents that may be used when querying (and the parents
// are marked as full)
// Given we don't keep track of zoom levels separately, we need to traverse the
// entire quadmap. If this is a common operation we'll need to track/cache this
// information somewhere. Although for the limited test cases so far it's pretty much instant
func (qm *QuadMap) NumberOfTilesForLevel(level byte) int {
	count := 0
	for _, t := range qm.quadKeyMap {
		if GetTileZoomLevel(t.QuadKey) == level {
			count++
		}
	}
	return count
}

// AddTile adds a pre-generated tile (which has its quadkey already)
func (qm *QuadMap) AddTile(t *Tile) error {
	qm.quadKeyMap[t.QuadKey] = t
	return nil
}

// AddChild adds a child to tile t, in position pos.
// Returns created (and registered in quadmap) child tile
func (qm *QuadMap) AddChild(t *Tile, pos int) (*Tile, error) {
	child, err := createChildForPos(t, pos)
	if err != nil {
		return nil, err
	}
	qm.quadKeyMap[child.QuadKey] = child
	return child, nil
}

// createChildForPos creates child tile for tile t in appropriate position
// Populates tile type and full flags based off parent.
func createChildForPos(t *Tile, pos int) (*Tile, error) {
	quadKey, err := GetChildQuadKeyForPos(t.QuadKey, pos)
	if err != nil {
		return nil, err
	}
	child := &Tile{QuadKey: quadKey, Types: t.Types, Full: t.Full}
	return child, nil
}
