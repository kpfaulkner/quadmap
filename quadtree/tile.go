package quadtree

import "fmt"

type TileType uint16
type GroupID uint32

const (
	TileTypeNone      TileType = 0b000000000
	TileTypeVert      TileType = 0b000000000001
	TileTypeEast      TileType = 0b000000000010
	TileTypeNorth     TileType = 0b000000000100
	TileTypeSouth     TileType = 0b00000001000
	TileTypeWest      TileType = 0b0000010000
	TileTypeDsm       TileType = 0b0000100000
	TileTypeTrueOrtho TileType = 0b0001000000
	TileTypeNIR       TileType = 0b0010000000
)

// GroupDetails has bare information about the "group" that created the tile this is associated with.
// The first 32 bits is a unique identifier for the group
// The next 32 bits are a combination of tiletype (can be maximum of 16) and a full flag.
// eg.
//
//		|63-----------------32|31-----------------------------0|
//		| GroupID             | TileType + Full                |
//		|                     |00000000000000010000000000000001|
//	    In this case, bit 16 is set (TileTypeVert) and bit 0 is set (Full). So GroupID has a tiletype of Vert and is full.
//      where-as
//		|                     |00000000000000100000000000000000|
//      Has Type of East, but is not full.

type GroupDetails uint64

func (gd GroupDetails) GroupID() GroupID {
	return GroupID(gd >> 32)
}

// HasTileTypeAndFull returns if tiletype is set for the GroupDetails and if
// the tile is full.
func (gd GroupDetails) HasTileTypeAndFull(tileType TileType) (bool, bool) {
	tt := uint32(gd >> 32)
	hasTileType := uint16(tt>>16)&uint16(tileType) == 1
	isFull := false
	if hasTileType {
		if uint16(tt)&uint16(tileType) == 1 {
			isFull = true
		}
	}
	return hasTileType, isFull
}

func (gd GroupDetails) ReturnTileTypes() []TileType {
	var tileTypes []TileType
	if gd.HasTileType(TileTypeVert) {
		tileTypes = append(tileTypes, TileTypeVert)
	}
	if gd.HasTileType(TileTypeEast) {
		tileTypes = append(tileTypes, TileTypeEast)

	}
	if gd.HasTileType(TileTypeNorth) {
		tileTypes = append(tileTypes, TileTypeNorth)
	}
	if gd.HasTileType(TileTypeSouth) {
		tileTypes = append(tileTypes, TileTypeSouth)
	}
	if gd.HasTileType(TileTypeWest) {
		tileTypes = append(tileTypes, TileTypeWest)
	}
	if gd.HasTileType(TileTypeDsm) {
		tileTypes = append(tileTypes, TileTypeDsm)
	}
	if gd.HasTileType(TileTypeTrueOrtho) {
		tileTypes = append(tileTypes, TileTypeTrueOrtho)
	}
	if gd.HasTileType(TileTypeNIR) {
		tileTypes = append(tileTypes, TileTypeNIR)
	}
	return tileTypes
}

func (gd GroupDetails) HasTileType(tileType TileType) bool {
	tt := uint32(gd >> 32)
	hasTileType := uint16(tt>>16)&uint16(tileType) == 1
	return hasTileType
}

func (gd GroupDetails) SetTileTypeAndFull(tileType TileType, full bool) GroupDetails {

	gd = gd | GroupDetails(uint32(tileType)<<16)
	if full {
		gd = gd | GroupDetails(tileType)
	} else {
		gd = gd & GroupDetails(^tileType)
	}

	return gd
}

func NewGroupDetails(gid GroupID, tt TileType, full bool) GroupDetails {
	gd := GroupDetails(uint64(gid) << 32)
	gd.SetTileTypeAndFull(tt, full)
	return gd
}

// Tile is a node within a quadtree.
// Although a Tile instance will only be in the quadmap once (for a given quadkey) it may
// be the case that the same tile+quadkey is used by multiple "groups" and also for
// multiple tiletypes within a group.
//
// For example, if we populate the quadmap with one set of data (called group1) and group1
// in turn has information about tiletype1 and tiletype2, this means that we'll need to track at a per quadkey
// level if the files are full (and for which tiletypes).
// Once the caller then populates with a completely different group, then we'll need to store that information
// as well (in the same tile). This means that we'll need to store a map of groupID -> GroupDetails.
//
// We *could* skip all this if we wanted a separate quadmap per group, but given we need to search all groups
// at once, this would be incredibly inefficient
type Tile struct {
	// stupid to keep it in here?
	// Will also store the zoom level in the key.
	QuadKey QuadKey

	// groups that have information for this tile. The IDs listed here can be used elsewhere to look up data.
	// Not convinced that the groupdata *has* to be stored actually IN the tree.
	groups []GroupDetails
}

// NewTile creates a new tile at slippy co-ords x,y,z
// Will probably only be used for root tile
func NewTile(x uint32, y uint32, z byte) *Tile {
	t := &Tile{}
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	t.QuadKey = quadKey
	return t
}

// SetTileType for a given groupID and tiletype.
// Checks if groupID + tiletype combination already exists. Returns error if so.
func (t *Tile) SetTileType(groupID GroupID, tt TileType) error {

	if t.HasTileType(groupID, tt) {
		return fmt.Errorf("GroupID + TileType combination already exists")
	}

	gd := NewGroupDetails(groupID, tt, false)
	t.groups = append(t.groups, gd)
	return nil
}

// HasTileType... needs to loop through array. Hopefully wouldn't be too many
// FIXME(kpfaulkner) measure and confirm
func (t *Tile) HasTileType(groupID GroupID, tt TileType) bool {

	for _, g := range t.groups {
		if g.GroupID() == groupID && g.HasTileType(tt) {
			return true
		}
	}
	return false
}

// GetTileZoomLevel returns zoom level of tile
func (t *Tile) GetTileZoomLevel() byte {
	return t.QuadKey.Zoom()
}

// SetFullForGroupIDAndTileType sets the full flag for a given tile type.
// Only creates Full map at this stage (saves us creating a potential mass of unused maps)
func (t *Tile) SetFullForGroupIDAndTileType(groupID GroupID, tileType TileType, full bool) error {

	// loop through groups... see if already have groupid + type match.
	for _, g := range t.groups {
		if g.GroupID() == groupID && g.HasTileType(tileType) {
			g.SetTileTypeAndFull(tileType, full)
			return nil
		}
	}

	gd := NewGroupDetails(groupID, tileType, full)
	t.groups = append(t.groups, gd)
	return nil
}

// GetFullForGroupIDAndTileType gets the full flag for a given tile type. Defaults to false if no flag found
func (t *Tile) GetFullForGroupIDAndTileType(groupID GroupID, tileType TileType) bool {

	for _, g := range t.groups {
		if g.GroupID() == groupID {
			hasTileType, isFull := g.HasTileTypeAndFull(tileType)
			if hasTileType {
				return isFull
			}

		}
	}
	return false
}
