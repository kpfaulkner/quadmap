package quadtree

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"sort"

	log "github.com/sirupsen/logrus"
)

// QuadMap is a quadtree in disguise...
type QuadMap struct {

	// map of quadkey to tile
	quadKeyMap map[QuadKey]Tile

	// persist to storage while creating
	persistWhileCreating bool

	//storage *PebbleStorage
}

var (
	TileLUT map[TileType]uint32
)

// NewQuadMap create a new quadmap
// Should provide a large initialCapacity when dealing with large quadtree structures
func NewQuadMap(initialCapacity int, persistWhileCreating bool) *QuadMap {
	return &QuadMap{
		quadKeyMap:           make(map[QuadKey]Tile, initialCapacity),
		persistWhileCreating: persistWhileCreating,
		//storage:              NewPebbleStorage(),
	}
}

func SetupTileLUT(lut map[TileType]uint32) {
	TileLUT = lut
}

// should not really exits
func (qm *QuadMap) GetUnderlyingMap() map[QuadKey]Tile {
	return qm.quadKeyMap
}

func (qm *QuadMap) DisplayStats() {
	fmt.Printf("QuadMap len %d\n", len(qm.quadKeyMap))

	groupSize := make(map[int]int)
	// group sizes
	groupTotal := 0

	var groupSizes []int
	for k, _ := range groupSize {
		groupSizes = append(groupSizes, k)
	}

	sort.Ints(groupSizes)
	for _, k := range groupSizes {
		fmt.Printf("groupsize %d : count %d\n", k, groupSize[k])
	}

	fmt.Printf("total groups (ie total tiles) %d\n", groupTotal)
}

// GetParentTile returns parent tile of passed in tile t
func (qm *QuadMap) GetParentTile(quadKey QuadKey) (Tile, error) {
	parentKey, err := quadKey.Parent()
	if err != nil {
		return false, err
	}
	parentTile, ok := qm.quadKeyMap[parentKey]
	if !ok {
		return false, errors.New("parent tile not found")
	}
	return parentTile, nil
}

// GetChildInPos returns child tile of passed in tile t which is in position pos
// pos is a number between 0 and 3, where 0 is top left, 1 is top right, 2 is bottom left and 3 is bottom right
func (qm *QuadMap) GetChildInPos(quadKey QuadKey, pos int) (Tile, error) {
	childKey, err := quadKey.ChildAtPos(pos)
	if err != nil {
		return false, err
	}
	childTile, ok := qm.quadKeyMap[childKey]
	if !ok {
		return false, errors.New(fmt.Sprintf("child tile in pos %d not found", pos))
	}
	return childTile, nil
}

// GetExactTileForSlippy returns tile for slippy co-ord match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForSlippy(x uint32, y uint32, z byte) (Tile, error) {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	return qm.GetExactTileForQuadKey(quadKey)
}

// GetExactTileForQuadKey returns tile for quadkey match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForQuadKey(quadKey QuadKey) (Tile, error) {

	// if actual quadkey exists, return tile.
	if t, ok := qm.quadKeyMap[quadKey]; ok {
		return t, nil
	}
	return false, errors.New("no tile found")
}

// NumberOfTilesForZoom returns number of tiles for a given zoom level.
// It will NOT include parents that may be used when querying (and the parents
// are marked as full)
// Given we don't keep track of zoom levels separately, we need to traverse the
// entire quadmap. If this is a common operation we'll need to track/cache this
// information somewhere. Although for the limited test cases so far it's pretty much instant
func (qm *QuadMap) NumberOfTilesForZoom(zoom byte) int {
	count := 0
	for k, _ := range qm.quadKeyMap {
		if k.Zoom() == zoom {
			count++
		}
	}
	return count
}

// GetTilesForTypeAndZoom gets tiles for a given tile type and zoom level
//func (qm *QuadMap) GetTilesForTypeAndZoom(tt TileType, zoom byte) []Tile {
//	tiles := []Tile{}
//
//	for k, t := range qm.quadKeyMap {
//		if k.Zoom() == zoom {
//			if t.HasTileType(tt) {
//				tiles = append(tiles, t)
//			}
//		}
//	}
//	return tiles
//}

// NumberOfTiles returns number of tiles in quadmap
func (qm *QuadMap) NumberOfTiles() int {
	return len(qm.quadKeyMap)
}

// AddTile adds a pre-generated tile
func (qm *QuadMap) AddTile(x uint32, y uint32, z byte, t Tile) (QuadKey, error) {
	qk := GenerateQuadKeyIndexFromSlippy(x, y, z)
	qm.quadKeyMap[qk] = t
	return qk, nil
}

// CreateTileAtSlippyCoords creates a tile to the quadmap at slippy coords and details stored in storage
// If tile already exists at coords, then tile is modified with groupID/tiletype information
// Tile is returned
func (qm *QuadMap) CreateTileAtSlippyCoords(x uint32, y uint32, z byte, groupID uint32, tileType TileType, full bool) (QuadKey, error) {

	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)

	//var child Tile
	//var ok bool

	qm.quadKeyMap[quadKey] = true
	//// check if child exists.
	//if child, ok = qm.quadKeyMap[quadKey]; ok {
	//	child.SetTileTypeForGroupID(groupID, tileType, full)
	//} else {
	//	child.SetTileTypeForGroupID(groupID, tileType, full)
	//}
	//qm.quadKeyMap[quadKey] = child

	//err := qm.storage.SetTileDetail(quadKey, TileDetail{GroupID: groupID, TileType: tileType, Scale: z})
	return quadKey, nil

}

// CreateTileAtSlippyCoords simply records the fact that at a given quadkey, we have *some* data.... but no idea what.
// This will be used
func (qm *QuadMap) CreateTileAtSlippyCoordsOrig(x uint32, y uint32, z byte, groupID uint32, tileType TileType, full bool) (QuadKey, error) {

	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)

	var child Tile
	//var ok bool
	//// check if child exists.
	//if child, ok = qm.quadKeyMap[quadKey]; ok {
	//	child.SetTileTypeForGroupID(groupID, tileType, full)
	//} else {
	//	child.SetTileTypeForGroupID(groupID, tileType, full)
	//}
	qm.quadKeyMap[quadKey] = child

	//err := qm.storage.SetTileDetail(quadKey, TileDetail{GroupID: groupID, TileType: tileType, Scale: z})
	return quadKey, nil

}

// HaveTileForSlippyGroupIDAndTileType returns bool indicating if we have details for a tile at the provided
// slippy co-ords but also matching the tiletype and groupID.
func (qm *QuadMap) HaveTileForSlippyGroupIDAndTileType(x uint32, y uint32, z byte, groupID uint32, tileType TileType) (bool, error) {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	return qm.HaveTileForGroupIDAndTileType(quadKey, groupID, tileType, true)
}

// HaveTileForGroupIDAndTileType returns bool indicating if we have details for a tile at the provided
// quadKey provided but also matching the tiletype and groupID.
// It will return true if tile requested is found OR an ancestor that is FULL
// The process is:
//
//   - if quadkey exists, tiletype exists in QM then check storage for groupID. If has groupID return true
//
//   - else
//
//   - get parent quadkey
//
//   - if parent quadkey exists and is full, return parent details
//
//   - loop until no parent.
//
//     What happens if we hit a parent that is NOT full? No tile therefore return error?
//     Returns tile (actual or parent), bool indicating if actual (true == actual, false == ancestor) and error
func (qm *QuadMap) HaveTileForGroupIDAndTileType(quadKey QuadKey, groupID uint32, tileType TileType, actualTile bool) (bool, error) {

	// if QK exists and tileType exists...
	// return if is actual tile (QK match) or tile is full
	if tile, ok := qm.quadKeyMap[quadKey]; ok {
		for _, td := range tile.Details {
			if td.GetGroupID() == groupID {
				typeExists, isFull := td.HasTileType(tileType)
				if typeExists && (isFull || actualTile) {
					return true, nil
				}
			}
		}
	}

	// otherwise... we need to check parent
	parentQuadKey, err := quadKey.Parent()
	if err != nil {
		return false, err
	}

	// check parents and upwards. actualTile is false since we're querying ancestors
	found, err := qm.HaveTileForGroupIDAndTileType(parentQuadKey, groupID, tileType, false)
	if err != nil {
		return false, err
	}

	// return whether found or not
	return found, nil
}

// GetTileForSlippyCoords returns tile for the tile at slippy coord x,y,z.
// This may involve multiple groups (ie multiple data sets loaded into single quadmap) but
// also different tiletypes as well.
func (qm *QuadMap) GetTileForSlippyCoords(x uint32, y uint32, z byte) (Tile, error) {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if tile, ok := qm.quadKeyMap[quadKey]; ok {
		return tile, nil
	}

	return Tile{}, errors.New("no tile at coordinates")
}

// GetTileDetailForQuadkeyAnyGeneration returns details for the tile for quadkey
// This will go back through ancestry to collect TileDetails from ancestors that may cover the same
// quadkey but at a higher level with a full indicator.
func (qm *QuadMap) GetTileDetailsForQuadkeyAnyGeneration(quadKey QuadKey, existingDetails []TileDetail) ([]TileDetail, error) {

	// high as we can go... cant do any more, so return nil
	if quadKey == 0 {
		return existingDetails, nil
	}

	var details []TileDetail
	if tile, ok := qm.quadKeyMap[quadKey]; ok {
		details = append(existingDetails, tile.Details...)
	}

	parentQuadKey, err := quadKey.Parent()
	if err != nil {
		// cant go any higher... stop the iteration.
		return details, nil
	}

	return qm.GetTileDetailsForQuadkeyAnyGeneration(parentQuadKey, details)

}

// GetSlippyBoundsForGroupIDTileTypeAndZoom returns the minx,miny,maxx,maxy slippy coords for a given zoom level
// extracted from the quadmap. Brute forcing it for now.
func (qm *QuadMap) GetSlippyBoundsForGroupIDTileTypeAndZoom(groupID uint32, tileType TileType, zoom byte) (uint32, uint32, uint32, uint32, error) {

	var minX uint32 = math.MaxUint32
	var minY uint32 = math.MaxUint32
	var maxX uint32 = 0
	var maxY uint32 = 0

	for quadKey, tile := range qm.quadKeyMap {

		if quadKey == 0 {
			continue // should this be in the quadMap at all?
		}

		z := quadKey.Zoom()
		if z > zoom {
			continue
		}

		// only get tiletype and groupID that we want. Also needs to be either equal zoom OR is full.
		for _, detail := range tile.Details {
			if detail.GetGroupID() == groupID {
				hasTT, full := detail.HasTileType(tileType)
				if !hasTT {
					continue
				}

				// only continue if precise zoom level OR this tile is considered full.
				if z == zoom || full {
					minChild, maxChild, err := quadKey.GetMinMaxEquivForZoomLevel(zoom)
					if err != nil {
						log.Errorf("error while generating min/max for quadkey %s", err.Error())
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
		}

	}

	return minX, minY, maxX, maxY, nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
