package quadtree

import "fmt"

// TileType is the type of a given tile
// Given we'll have very few tile types we could just use a byte, but
// may eventually use this as a bitmask to help filtering. So will
// keep it as a uint16 for now. Can change later if space becomes an issue
type TileType uint16

// GroupDetails has bare information about the "group" that created the tile this is associated with.
// GroupID is the ID of the data feed (simply has to be unique)
// Type is the tiletype assocated with this group.
// Full determines if this tile+groupID+type is "full"... if true then
// all child tiles are also full/exist.
// TODO(kpfaulkner) investigate if this can simply be changed to a uint64.
type GroupDetails struct {
	GroupID uint32

	// This tile is used for various tile types
	// use as bitmask. Assumption that will not have more than 16 tile types.
	Type TileType

	// This tile and all children/grandchildren/second cousins once removed etc... are present.
	// Full is tiletype specific.
	Full bool
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

	g := GroupDetails{
		GroupID: groupID,
		Type:    tt,
	}

	t.groups = append(t.groups, g)
	return nil
}

// HasTileType... needs to loop through array. Hopefully wouldn't be too many
// FIXME(kpfaulkner) measure and confirm
func (t *Tile) HasTileType(groupID uint32, tt TileType) bool {

	for _, g := range t.groups {
		if g.GroupID == groupID && g.Type == tt {
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
		if g.GroupID == groupID && g.Type == tileType {
			g.Full = full
			t.groups[i] = g
			return nil
		}
	}

	g := GroupDetails{
		GroupID: groupID,
		Type:    tileType,
		Full:    full,
	}

	t.groups = append(t.groups, g)
	return nil
}

// GetFullForGroupIDAndTileType gets the full flag for a given tile type. Defaults to false if no flag found
func (t *Tile) GetFullForGroupIDAndTileType(groupID uint32, tileType TileType) bool {

	// BAD BAD BAD, need to loop through on a tile to find if its full or not.
	// BUT... not expecting that many overlaps for a single tile.
	// FIXME(kpfaulkner) measure perf and make judgement

	for _, g := range t.groups {
		if g.GroupID == groupID && g.Type == tileType {
			return g.Full
		}
	}

	return false
}
