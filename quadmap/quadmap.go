package quadmap

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
)

var (
	TileNotFoundError        = errors.New("tile not found")
	TileWithTileTypeNotFound = errors.New("tile with tile type not found")
)

// DataReader function is provided by the consumer of the Quadmap.
// This will read a byte slice and scale and populate the Quadmap with
// the appropriate deserialised data.
// Will read as far as expandToLevel value (exclusive)
// Do not remove, this will be used by consumers of quadmap package
type DataReader func(qm *QuadMap, data *[]byte, tileType TileType) error

// QuadMap is a quadmap in disguise...
type QuadMap struct {

	// map of quadkey to tile
	QuadKeyMap map[QuadKey]*Tile

	// function able to take byte slices and populate Quadmap.
	dataReader DataReader

	lock sync.RWMutex
}

// NewQuadMap create a new quadmap
// Should provide a large initialCapacity when dealing with large quadmap structures
func NewQuadMap(initialCapacity int) *QuadMap {
	return &QuadMap{
		QuadKeyMap: make(map[QuadKey]*Tile, initialCapacity),
	}
}

// SetDataReader sets the data reader for the quadmap
// Ideally would like it to be as part of NewQuadMap but need to pass in the quadmap...  so catch 22
func (qm *QuadMap) SetDataReader(dr DataReader) {
	qm.dataReader = dr
}

func (qm *QuadMap) GetAllTiles(sorted bool) ([]*Tile, error) {
	qm.lock.RLock()
	allTiles := make([]*Tile, 0, len(qm.QuadKeyMap))
	for _, tile := range qm.QuadKeyMap {
		allTiles = append(allTiles, tile)
	}
	qm.lock.RUnlock()

	if sorted {
		sort.Slice(allTiles, func(i, j int) bool {
			return allTiles[i].QuadKey < allTiles[j].QuadKey
		})
	}

	return allTiles, nil
}

// GetParentTile returns parent tile of passed in tile t
func (qm *QuadMap) GetParentTile(t *Tile) (*Tile, error) {
	parentKey, err := t.QuadKey.Parent()
	if err != nil {
		return nil, err
	}

	qm.lock.RLock()
	parentTile, ok := qm.QuadKeyMap[parentKey]
	qm.lock.RUnlock()

	if !ok {
		return nil, errors.New("parent tile not found")
	}
	return parentTile, nil
}

// GetChildInPos returns child tile of passed in tile t which is in position pos
// pos is a number between 0 and 3, where 0 is top left, 1 is top right, 2 is bottom left and 3 is bottom right
func (qm *QuadMap) GetChildInPos(t *Tile, pos int) (*Tile, error) {
	childKey, err := t.QuadKey.ChildAtPos(pos)
	if err != nil {
		return nil, err
	}
	qm.lock.RLock()
	childTile, ok := qm.QuadKeyMap[childKey]
	qm.lock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("child tile in pos %d not found", pos)
	}
	return childTile, nil
}

// GetExactTileForSlippy returns tile for slippy co-ord match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForSlippy(x uint32, y uint32, z byte) (*Tile, error) {
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return nil, err
	}
	return qm.GetExactTileForQuadKey(quadKey)
}

func (qm *QuadMap) GetTileForSlippyAndTileType(x uint32, y uint32, z byte, tt TileType) (*Tile, error) {
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return nil, err
	}

	tile, err := qm.GetExactTileForQuadKey(quadKey)
	if err != nil {
		return nil, err
	}

	if tile.HasTileType(tt) {
		return tile, nil
	}

	return nil, TileWithTileTypeNotFound
}

// GetExactTileForQuadKey returns tile for quadkey match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForQuadKey(quadKey QuadKey) (*Tile, error) {

	qm.lock.RLock()
	defer qm.lock.RUnlock()

	// if actual quadkey exists, return tile.
	if t, ok := qm.QuadKeyMap[quadKey]; ok {
		return t, nil
	}
	return nil, TileNotFoundError
}

// NumberOfTilesForZoom returns number of tiles for a given zoom level.
// It will NOT include parents that may be used when querying (and the parents
// are marked as full)
// Given we don't keep track of zoom levels separately, we need to traverse the
// entire quadmap. If this is a common operation we'll need to track/cache this
// information somewhere. Although for the limited test cases so far it's pretty much instant
func (qm *QuadMap) NumberOfTilesForZoom(zoom byte) int {
	qm.lock.RLock()
	defer qm.lock.RUnlock()
	count := 0
	for _, t := range qm.QuadKeyMap {
		if t.QuadKey.Zoom() == zoom {
			count++
		}
	}
	return count
}

// GetTilesForTypeAndZoom gets tiles for a given tile type and zoom level
func (qm *QuadMap) GetTilesForTypeAndZoom(tt TileType, zoom byte) []*Tile {
	tiles := []*Tile{}

	qm.lock.RLock()
	defer qm.lock.RUnlock()
	for _, t := range qm.QuadKeyMap {
		if t.QuadKey.Zoom() == zoom && t.HasTileType(tt) {
			tiles = append(tiles, t)
		}
	}
	return tiles
}

// NumberOfTiles returns number of tiles in quadmap
func (qm *QuadMap) NumberOfTiles() int {
	qm.lock.RLock()
	defer qm.lock.RUnlock()
	return len(qm.QuadKeyMap)
}

// AddTile adds a pre-generated tile (which has its quadkey already)
func (qm *QuadMap) AddTile(t *Tile) error {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	qm.QuadKeyMap[t.QuadKey] = t
	return nil
}

// CreateTileAtSlippyCoords creates a tile to the quadmap at slippy coords
func (qm *QuadMap) CreateTileAtSlippyCoords(x uint32, y uint32, z byte, tileType TileType, full bool) (*Tile, error) {

	// x,y,z are already child coords...  so no need to take pos into account
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return nil, err
	}
	qm.lock.Lock()
	defer qm.lock.Unlock()

	// check if child exists.
	if tile, ok := qm.QuadKeyMap[quadKey]; ok {
		tile.AddTileType(tileType, full)
		return tile, nil
	}

	t, err := NewTileWithTileTypeAndFull(x, y, z, tileType, full)
	if err != nil {
		return nil, err
	}

	qm.QuadKeyMap[t.QuadKey] = t
	return t, nil
}

// createChildForPos creates child tile for tile t in appropriate position
// Populates tile type and full flags based off parent.
// FIXME(kpfaulkner) confirm can delete
func createChildForPos(childQuadKey QuadKey, pos int) (*Tile, error) {
	//child := &Tile{QuadKey: childQuadKey}
	child := NewTileWithQuadKey(childQuadKey)
	return child, nil
}

// GetSlippyBoundsForTileTypeAndZoom returns minX, minY, maxX, maxY slippy coords for a given tiletype and
// zoom level
func (qm *QuadMap) GetSlippyBoundsForTileTypeAndZoom(tileType TileType, zoom byte) (uint32, uint32, uint32, uint32, error) {

	var minX uint32 = math.MaxUint32
	var minY uint32 = math.MaxUint32
	var maxX uint32 = 0
	var maxY uint32 = 0

	qm.lock.RLock()
	defer qm.lock.RUnlock()
	for quadKey, v := range qm.QuadKeyMap {

		if quadKey == 0 {
			continue // should this be in the quadMap at all?
		}

		z := quadKey.Zoom()
		if z > zoom {
			continue
		}

		hasTileType, isFull := v.HasTileTypeAndFull(tileType)

		if !hasTileType {
			continue
		}

		// only continue if precise zoom level OR this tile is considered full.
		if z == zoom || isFull {
			minChild, maxChild, err := quadKey.GetMinMaxEquivForZoomLevel(zoom)
			if err != nil {
				return 0, 0, 0, 0, err
			}

			x, y, _ := minChild.SlippyCoords()
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}

			x, y, _ = maxChild.SlippyCoords()
			if x > maxX {
				maxX = x
			}

			if y > maxY {
				maxY = y
			}
		}
	}

	return minX, minY, maxX, maxY, nil
}

// GetAllChildrenForQuadKeyAndZoom returns all quadkeys for a given zoom level including situations where a parent
// is marked as full.
func (qm *QuadMap) GetAllChildrenForQuadKeyAndZoom(qk QuadKey, tileType TileType, zoom byte) ([]QuadKey, error) {
	qm.lock.RLock()
	defer qm.lock.RUnlock()
	return qm.getAllChildrenForQuadKeyAndZoomLocked(qk, tileType, zoom)
}

// getAllChildrenForQuadKeyAndZoomLocked is the recursive body of
// GetAllChildrenForQuadKeyAndZoom. Callers must hold qm.lock (read or write).
func (qm *QuadMap) getAllChildrenForQuadKeyAndZoomLocked(qk QuadKey, tileType TileType, zoom byte) ([]QuadKey, error) {
	if qk.Zoom() == zoom {
		return []QuadKey{qk}, nil
	}

	allKeys := []QuadKey{}
	for _, child := range qk.Children() {
		childData, ok := qm.QuadKeyMap[child]
		if !ok {
			continue
		}
		hasTileType, isFull := childData.HasTileTypeAndFull(tileType)
		if hasTileType {
			if isFull {
				// get all descendants of this tile for correct zoom level and include them.
				allChildren := child.GetAllPossibleChildrenAtZoom(zoom)
				allKeys = append(allKeys, allChildren...)
				continue
			}

			// check children of this child
			childKeys, err := qm.getAllChildrenForQuadKeyAndZoomLocked(child, tileType, zoom)
			if err != nil {
				return nil, err
			}
			allKeys = append(allKeys, childKeys...)
		}
	}
	return allKeys, nil
}

// IsTileCoveredForSlippyCoordsAndTileTypeTopDown takes slippy coord, gets all ancestors to see if tile should exist
// (by checking ancestors + full flag)
// Also returns the quadkey that covers the co-ord... whether its the actual QK for the co-ordinates
// or an ancestor that is full
func (qm *QuadMap) IsTileCoveredForSlippyCoordsAndTileTypeTopDown(x uint32, y uint32, z byte, tileType TileType) (bool, QuadKey, error) {

	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return false, 0, err
	}

	//allAncestors := quadKey.GetAllAncestorsAndSelf()

	qm.lock.RLock()
	defer qm.lock.RUnlock()

	qk := quadKey
	for {
		if t, ok := qm.QuadKeyMap[qk]; ok {

			// if at target zoom level and match... then true
			if t.QuadKey.Zoom() == z {
				// have match... return true
				return true, qk, nil
			}

			hasTileType, isFull := t.HasTileTypeAndFull(tileType)
			if hasTileType && isFull {
				return true, qk, nil
			}
		}
		if qk.Zoom() == 0 {
			break
		}
		qk = qk.ParentUnchecked()
	}

	//for _, qk := range allAncestors {
	//	qm.lock.RLock()
	//	t, ok := qm.QuadKeyMap[qk]
	//	qm.lock.RUnlock()
	//	if ok {
	//
	//		// if at target zoom level and match... then true
	//		if t.QuadKey.Zoom() == z {
	//			// have match... return true
	//			return true, nil
	//		}
	//
	//		hasTileType, isFull := t.HasTileTypeAndFull(tileType)
	//		if hasTileType && isFull {
	//			return true, nil
	//		}
	//	}
	//
	//}

	return false, 0, nil
}
