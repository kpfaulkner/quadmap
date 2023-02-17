package quadtree

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
)

// TileDetailsGroup is same as TileDetails but we also want
// the quadkey that gave us the match.
// TileDetailsGroup is used when returning query results and NOT
// actually part of the quadmap itself.
type TileDetailsGroup struct {
	GroupDetails
	QuadKey QuadKey
}

// TileDetails information about a tile, groups its associated with,
// tiletypes etc etc.
// TileDetails is used when returning query results and NOT
// actually part of the quadmap itself.
type TileDetails struct {
	Groups []TileDetailsGroup
}

// QuadMap is a quadtree in disguise...
type QuadMap struct {

	// map of quadkey to tile
	quadKeyMap map[QuadKey]*Tile
}

// NewQuadMap create a new quadmap
// Should provide a large initialCapacity when dealing with large quadtree structures
func NewQuadMap(initialCapacity int) *QuadMap {
	return &QuadMap{
		quadKeyMap: make(map[QuadKey]*Tile, initialCapacity),
	}
}

// GetParentTile returns parent tile of passed in tile t
func (qm *QuadMap) GetParentTile(t *Tile) (*Tile, error) {
	parentKey, err := t.QuadKey.Parent()
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
	childKey, err := t.QuadKey.ChildAtPos(pos)
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
func (qm *QuadMap) GetExactTileForSlippy(x uint32, y uint32, z byte) (*Tile, error) {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	return qm.GetExactTileForQuadKey(quadKey)
}

// GetExactTileForQuadKey returns tile for quadkey match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForQuadKey(quadKey QuadKey) (*Tile, error) {

	// if actual quadkey exists, return tile.
	if t, ok := qm.quadKeyMap[quadKey]; ok {
		return t, nil
	}
	return nil, errors.New("no tile found")
}

// NumberOfTilesForZoom returns number of tiles for a given zoom level.
// It will NOT include parents that may be used when querying (and the parents
// are marked as full)
// Given we don't keep track of zoom levels separately, we need to traverse the
// entire quadmap. If this is a common operation we'll need to track/cache this
// information somewhere. Although for the limited test cases so far it's pretty much instant
func (qm *QuadMap) NumberOfTilesForZoom(zoom byte) int {
	count := 0
	for _, t := range qm.quadKeyMap {
		if t.QuadKey.Zoom() == zoom {
			count++
		}
	}
	return count
}

// GetTilesForTypeAndZoom gets tiles for a given tile type and zoom level
func (qm *QuadMap) GetTilesForTypeAndZoom(tt TileType, zoom byte) []*Tile {
	tiles := []*Tile{}

	for _, t := range qm.quadKeyMap {
		if t.QuadKey.Zoom() == zoom {
			for _, g := range t.groups {
				if g.Type == tt {
					tiles = append(tiles, t)
				}
			}
		}
	}
	return tiles
}

// NumberOfTiles returns number of tiles in quadmap
func (qm *QuadMap) NumberOfTiles() int {
	return len(qm.quadKeyMap)
}

// AddTile adds a pre-generated tile (which has its quadkey already)
func (qm *QuadMap) AddTile(t *Tile) error {
	qm.quadKeyMap[t.QuadKey] = t
	return nil
}

// CreateTileAtSlippyCoords creates a tile to the quadmap at slippy coords
// If tile already exists at coords, then tile is modified with groupID/tiletype information
// Tile is returned
func (qm *QuadMap) CreateTileAtSlippyCoords(x uint32, y uint32, z uint32, groupID uint32, tileType TileType) (*Tile, error) {

	// x,y,z are already child coords...  so no need to take pos into account
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, byte(z))

	// check if child exists.
	if child, ok := qm.quadKeyMap[quadKey]; ok {
		child.SetTileType(groupID, tileType)
		return child, nil
	}

	t := &Tile{QuadKey: quadKey}
	t.SetTileType(groupID, tileType)
	qm.quadKeyMap[t.QuadKey] = t
	return t, nil
}

// createChildForPos creates child tile for tile t in appropriate position
// Populates tile type and full flags based off parent.
// FIXME(kpfaulkner) confirm can delete
func createChildForPos(childQuadKey QuadKey, pos int) (*Tile, error) {
	child := &Tile{QuadKey: childQuadKey}
	return child, nil
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
func (qm *QuadMap) HaveTileForGroupIDAndTileType(quadKey QuadKey, groupID uint32, tileType TileType, actualTile bool) (bool, error) {

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
func (qm *QuadMap) GetTileDetailsForSlippyCoords(x uint32, y uint32, z byte, tileDetails *TileDetails) error {
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	return qm.GetTileDetailsForQuadkey(quadKey, tileDetails, true)
}

// GetTileDetailsForQuadkey returns details for the tile for quadkey
// This may involve multiple groups (ie multiple data sets loaded into single quadmap) but
// also different tiletypes as well.
func (qm *QuadMap) GetTileDetailsForQuadkey(quadKey QuadKey, tileDetails *TileDetails, isTargetLevel bool) error {

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

	parentQuadKey, err := quadKey.Parent()
	if err != nil {
		// cant go any higher... stop the iteration.
		return nil
	}

	// isTargetLevel false due to we're processing an ancestor now.
	return qm.GetTileDetailsForQuadkey(parentQuadKey, tileDetails, false)
}

// GetBoundsForZoom returns the minx,miny,maxx,maxy slippy coords for a given zoom level
// extracted from the quadmap. Brute forcing it for now.
func (qm *QuadMap) GetBoundsForZoom(zoom byte) (int32, int32, int32, int32, error) {

	var minX int32 = math.MaxInt32
	var minY int32 = math.MaxInt32
	var maxX int32 = 0
	var maxY int32 = 0

	for quadKey, _ := range qm.quadKeyMap {
		z := quadKey.Zoom()

		if z > zoom {
			// skip it...
			continue
		}

		if quadKey == 0 {
			//fmt.Printf("snoop\n")
			continue // should this be in the quadMap at all?

		}
		minChild, maxChild, err := GenerateMinMaxQuadKeysForZoom(quadKey, z)
		if err != nil {
			log.Errorf("error while generating min/max for quadkey %s", err.Error())
			return 0, 0, 0, 0, err
		}

		x, y, _ := minChild.SlippyCoords()

		if x == 0 {
			fmt.Printf("snoop\n")
		}
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}

		x, y, _ = maxChild.SlippyCoords()

		if x == 1929934 {
			fmt.Printf("snoop\n")
		}
		if x > maxX {
			maxX = x
		}

		if y > maxY {
			maxY = y
		}
	}

	return minX, minY, maxX, maxY, nil
}
