package quadtree

import (
	"errors"
	"fmt"
)

// TileDetailsGroup is same as TileDetails but we also want
// the quadkey that gave us the match.
type TileDetailsGroup struct {
	GroupDetails
	QuadKey uint64
}

// TileDetails information about a tile, groups its associated with,
// tiletypes etc etc.
type TileDetails struct {
	Groups []TileDetailsGroup
}

// QuadMap is a quadtree in disguise...
type QuadMap struct {

	// map of quadkey to tile
	quadKeyMap map[uint64]*Tile

	// Slice of all groups stored in the Quadmap.
	// Used for when wanting to know if we've processed all groups later on.
	groupIDs []string
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

// GetExactTileForSlippy returns tile for slippy co-ord match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForSlippy(x int32, y int32, z byte) (*Tile, error) {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	return qm.GetExactTileForQuadKey(quadKey)
}

// GetExactTileForQuadKey returns tile for quadkey match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForQuadKey(quadKey uint64) (*Tile, error) {

	// if actual quadkey exists, return tile.
	if t, ok := qm.quadKeyMap[quadKey]; ok {
		return t, nil
	}
	return nil, errors.New("no tile found")
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

func (qm *QuadMap) NumberOfTiles() int {
	return len(qm.quadKeyMap)
}

// AddTile adds a pre-generated tile (which has its quadkey already)
func (qm *QuadMap) AddTile(t *Tile) error {
	qm.quadKeyMap[t.QuadKey] = t
	return nil
}

// AddChild adds a child to tile t, in position pos.
// Returns created (and registered in quadmap) child tile
func (qm *QuadMap) AddChild(t *Tile, pos int, groupID uint32, tileType TileType) (*Tile, error) {
	quadKey, err := GetChildQuadKeyForPos(t.QuadKey, pos)
	if err != nil {
		return nil, err
	}

	// check if child exists.
	if child, ok := qm.quadKeyMap[quadKey]; ok {
		child.SetTileType(groupID, tileType)
		return child, nil
	}

	child, err := createChildForPos(t, pos)
	if err != nil {
		return nil, err
	}
	child.SetTileType(groupID, tileType)
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
	child := &Tile{QuadKey: quadKey}
	return child, nil
}

// HaveTileForSlippyTileTypeAndGroupID returns bool indicating if we have details for a tile at the provided
// slippy co-ords but also matching the tiletype and groupID.
func (qm *QuadMap) HaveTileForSlippyTileTypeAndGroupID(x int32, y int32, z byte, tileType TileType, groupID uint32) (bool, error) {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	return qm.HaveTileForTypeAndGroupID(quadKey, tileType, groupID, true)
}

// HaveTileForTypeAndGroupID returns bool indicating if we have details for a tile at the provided
// quadKey provided but also matching the tiletype and groupID.
// It will return true if tile requested is found OR an ancestor that is FULL
// The process is:
//
//   - if quadkey exists, groupID exists and tiletype exists, return true
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
func (qm *QuadMap) HaveTileForTypeAndGroupID(quadKey uint64, tileType TileType, groupID uint32, actualTile bool) (bool, error) {

	// if actual quadkey exists, check tiletype and groupID
	if t, ok := qm.quadKeyMap[quadKey]; ok {
		for _, g := range t.groups {
			if g.GroupID == groupID && g.Type == tileType {
				if g.Full || actualTile {
					return true, nil
				}
			}
		}
	}

	parentQuadKey, err := GetParentQuadKey(quadKey)
	if err != nil {
		return false, err
	}

	// check parents and upwards. actualTile is false since we're querying ancestors
	found, err := qm.HaveTileForTypeAndGroupID(parentQuadKey, tileType, groupID, false)
	if err != nil {
		return false, err
	}

	// return whether found or not
	return found, nil
}

// GetTileDetailsForSlippyCoords returns details for the tile at slippy coord x,y,z.
// This may involve multiple groups (ie multiple data sets loaded into single quadmap) but
// also different tiletypes as well.
func (qm *QuadMap) GetTileDetailsForSlippyCoords(x int32, y int32, z byte, tileDetails *TileDetails) error {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	return qm.GetTileDetailsForQuadkey(quadKey, tileDetails, true)
}

// GetTileDetailsForQuadkey returns details for the tile for quadkey
// This may involve multiple groups (ie multiple data sets loaded into single quadmap) but
// also different tiletypes as well.
func (qm *QuadMap) GetTileDetailsForQuadkey(quadKey uint64, tileDetails *TileDetails, isTargetLevel bool) error {

	// high as we can go... cant do any more, so return nil
	if quadKey == 0 {
		return nil
	}

	if t, ok := qm.quadKeyMap[quadKey]; ok {

		// whatever groups are in tile t....  add the details to tileDetails but only if full (if we're processing parent)
		for _, g := range t.groups {

			// correct target level... so store all the tiles at this level... no?
			if isTargetLevel || g.Full {
				tileDetails.Groups = append(tileDetails.Groups, TileDetailsGroup{GroupDetails: GroupDetails{Full: g.Full, Type: g.Type, GroupID: g.GroupID}, QuadKey: quadKey})
			}
		}
	}

	parentQuadKey, err := GetParentQuadKey(quadKey)
	if err != nil {

		// cant go any higher... stop the iteration.
		return nil
	}

	// isTargetLevel false due to we're processing an ancestor now.
	return qm.GetTileDetailsForQuadkey(parentQuadKey, tileDetails, false)
}
