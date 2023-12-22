package quadtree

import (
	"errors"
	"fmt"
	"math"
	"sort"

	log "github.com/sirupsen/logrus"
)

// TileDetailsGroup is same as TileDetails but we also want
// the quadkey that gave us the match.
// TileDetailsGroup is used when returning query results and NOT
// actually part of the quadmap itself.
type TileDetail struct {
	GroupID  uint32
	TileType TileType
	Scale    byte
	Full     bool
}

// TileDetails information about a tile, groups its associated with,
// tiletypes etc etc.
// A lot of this may be stored out of memory on storage, so is NOT required
// for querying the quadmap itself, but more for when you know you are interested
// in a specific tile and want more details
type TileDetails struct {
	Details []TileDetail
}

// QuadMap is a quadtree in disguise...
type QuadMap struct {

	// map of quadkey to tile
	quadKeyMap map[QuadKey]Tile

	// persist to storage while creating
	persistWhileCreating bool

	storage Storage
}

// NewQuadMap create a new quadmap
// Should provide a large initialCapacity when dealing with large quadtree structures
func NewQuadMap(initialCapacity int, persistWhileCreating bool) *QuadMap {
	return &QuadMap{
		quadKeyMap:           make(map[QuadKey]Tile, initialCapacity),
		persistWhileCreating: persistWhileCreating,
	}
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
		return 0, err
	}
	parentTile, ok := qm.quadKeyMap[parentKey]
	if !ok {
		return 0, errors.New("parent tile not found")
	}
	return parentTile, nil
}

// GetChildInPos returns child tile of passed in tile t which is in position pos
// pos is a number between 0 and 3, where 0 is top left, 1 is top right, 2 is bottom left and 3 is bottom right
func (qm *QuadMap) GetChildInPos(quadKey QuadKey, pos int) (Tile, error) {
	childKey, err := quadKey.ChildAtPos(pos)
	if err != nil {
		return 0, err
	}
	childTile, ok := qm.quadKeyMap[childKey]
	if !ok {
		return 0, errors.New(fmt.Sprintf("child tile in pos %d not found", pos))
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
	return 0, errors.New("no tile found")
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
func (qm *QuadMap) GetTilesForTypeAndZoom(tt TileType, zoom byte) []Tile {
	tiles := []Tile{}

	for k, t := range qm.quadKeyMap {
		if k.Zoom() == zoom {
			if t.HasTileType(tt) {
				tiles = append(tiles, t)
			}
		}
	}
	return tiles
}

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

// CreateTileAtSlippyCoords creates a tile to the quadmap at slippy coords
// If tile already exists at coords, then tile is modified with groupID/tiletype information
// Tile is returned
func (qm *QuadMap) CreateTileAtSlippyCoords(x uint32, y uint32, z byte, groupID uint32, tileType TileType, full bool) (QuadKey, error) {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)

	var child Tile
	var ok bool
	// check if child exists.
	if child, ok = qm.quadKeyMap[quadKey]; ok {
		child.SetTileType(tileType, full)
	} else {
		// create new tile
		child = NewTile()
		child.SetTileType(tileType, full)
	}
	qm.quadKeyMap[quadKey] = child

	err := qm.storage.SetTileDetail(quadKey, TileDetail{GroupID: groupID, TileType: tileType, Scale: z, Full: full})
	return quadKey, err

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

	// if actual quadkey exists, check tiletype and groupID
	if tile, ok := qm.quadKeyMap[quadKey]; ok {
		if tile.HasTileType(tileType) {
			tileDetails, err := qm.storage.GetTileDetailsByTileType(quadKey, tileType)
			if err != nil {
				return false, err
			}

			for _, detail := range tileDetails.Details {
				if detail.GroupID == groupID {
					if detail.Full || actualTile {
						return true, nil
					}
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

// GetTileDetailsForSlippyCoords returns details for the tile at slippy coord x,y,z.
// This may involve multiple groups (ie multiple data sets loaded into single quadmap) but
// also different tiletypes as well.
func (qm *QuadMap) GetTileDetailsForSlippyCoords(x uint32, y uint32, z byte) (TileDetails, error) {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	var existingDetails TileDetails
	return qm.GetTileDetailsForQuadkey(quadKey, existingDetails, true)
}

// GetTileDetailsForQuadkey returns details for the tile for quadkey
// This may involve multiple groups (ie multiple data sets loaded into single quadmap) but
// also different tiletypes as well.
// Goes through the ancestry if we do not want to target the level.
func (qm *QuadMap) GetTileDetailsForQuadkey(quadKey QuadKey, existingDetails TileDetails, isTargetLevel bool) (TileDetails, error) {

	// high as we can go... cant do any more, so return nil
	if quadKey == 0 {
		return existingDetails, nil
	}

	var details TileDetails
	var err error
	if _, ok := qm.quadKeyMap[quadKey]; ok {

		// check details in storage
		details, err = qm.storage.GetTileDetails(quadKey)
		if err != nil {
			return TileDetails{}, err
		}
	}

	// if we only want the target level, then return details
	if isTargetLevel {

		// append to the details we already have.
		existingDetails.Details = append(existingDetails.Details, details.Details...)
		return existingDetails, nil
	}

	// if not target level, then loop through details and see if any are full
	// if full, then add to details
	for _, detail := range details.Details {
		if detail.Full {
			existingDetails.Details = append(existingDetails.Details, detail)
		}
	}

	// go through parents
	parentQuadKey, err := quadKey.Parent()
	if err != nil {
		// cant go any higher... stop the iteration.
		return details, nil
	}

	// isTargetLevel false due to we're processing an ancestor now.
	return qm.GetTileDetailsForQuadkey(parentQuadKey, existingDetails, false)
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

		// this is not the tile you're looking for.
		if !tile.HasTileType(tileType) {
			continue
		}

		tileDetails, err := qm.storage.GetTileDetails(quadKey)
		if err != nil {
			return 0, 0, 0, 0, err
		}

		// only get tiletype and groupID that we want. Also needs to be either equal zoom OR is full.
		for _, detail := range tileDetails.Details {
			if detail.GroupID == groupID && detail.TileType == tileType {

				// only continue if precise zoom level OR this tile is considered full.
				if z == zoom || detail.Full {
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
