package quadtree

import "fmt"

// TileType is the type of a given tile
// Given we'll have very few tile types we could just use a byte, but
// may eventually use this as a bitmask to help filtering. So will
// keep it as a uint16 for now. Can change later if space becomes an issue
type TileType uint16

// GroupDetails has bare information about the "group" that created the tile this is associated with.
// Split into first 32 bits are groupID, next 16 bits are tiletype, last 16 bits full flag.
type GroupDetails uint64

// NewGroupDetails make new GroupDetails (uint64)
func NewGroupDetails(groupID uint32, tileType TileType, full bool) GroupDetails {
	var gd GroupDetails

	gd = GroupDetails(groupID)
	gd = gd << 16
	gd |= GroupDetails(tileType)
	gd = gd << 8

	b := byte(0)
	if full {
		b = 1
	}
	gd |= GroupDetails(b)
	return gd
}

func (gd GroupDetails) Details() (uint32, TileType, bool) {
	full := (gd&0xFF != 0)
	gd = gd >> 8
	tt := TileType(gd & 0xFFFF)
	gd = gd >> 16
	id := uint32(gd)
	return id, tt, full
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
func (t *Tile) SetTileType(groupID uint32, tt TileType) error {

	if t.HasTileType(groupID, tt) {
		return fmt.Errorf("GroupID + TileType combination already exists")
	}

	gd := NewGroupDetails(groupID, tt, false)
	t.groups = append(t.groups, gd)
	return nil
}

// HasTileType... needs to loop through array. Hopefully wouldn't be too many
// FIXME(kpfaulkner) measure and confirm
func (t *Tile) HasTileType(groupID uint32, tt TileType) bool {

	for _, g := range t.groups {
		gd, t, _ := g.Details()
		if gd == groupID && t == tt {
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
func (t *Tile) SetFullForGroupIDAndTileType(groupID uint32, tileType TileType, full bool) error {

	// loop through groups... see if already have groupid + type match.
	for i, g := range t.groups {
		gd, ty, _ := g.Details()
		if gd == groupID && ty == tileType {
			ggg := NewGroupDetails(gd, ty, full)
			t.groups[i] = ggg
			return nil
		}
	}

	ggg := NewGroupDetails(groupID, tileType, full)
	t.groups = append(t.groups, ggg)
	return nil
}

// GetFullForGroupIDAndTileType gets the full flag for a given tile type. Defaults to false if no flag found
func (t *Tile) GetFullForGroupIDAndTileType(groupID uint32, tileType TileType) bool {

	// BAD BAD BAD, need to loop through on a tile to find if its full or not.
	// BUT... not expecting that many overlaps for a single tile.
	// FIXME(kpfaulkner) measure perf and make judgement

	for _, g := range t.groups {
		gd, ty, full := g.Details()
		if gd == groupID && ty == tileType {
			return full
		}
	}

	return false
}
