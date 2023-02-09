package quadtree

type TileType uint16

const (
	TileTypeNone      TileType = 0b00000000000000000
	TileTypeVert      TileType = 0b00000000000000001
	TileTypeEast      TileType = 0b00000000000000010
	TileTypeNorth     TileType = 0b00000000000000100
	TileTypeSouth     TileType = 0b00000000000001000
	TileTypeWest     TileType = 0b00000000000010000
	TileTypeDsm      TileType = 0b00000000000100000
	TileTypeTrueOrthoTileType = 0b00000000001000000
)

// TODO(kpfaulkner) change this to be a uint64...  we should be able to encode
// groupid, type and full into a single uint64
type GroupDetails struct {

	// GroupID... probably not required since it should be used as a map key to access this.
	// TODO(kpfaulkner) confirm if this is required.
	GroupID uint32

	// This tile is used for various tile types
	// use as bitmask. Assumption that will not have more than 16 tile types.
	Type TileType

	// This tile and all children/grandchildren/second cousins once removed etc... are present.
	// Full is tiletype specific.
	Full bool
}

// Tile is a node within a quadtree.
// Should shuffle these around for byte padding.... but not yet.
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
	QuadKey uint64

	// groups that have information for this tile. The IDs listed here can be used elsewhere to look up data.
	// Not convinced that the groupdata *has* to be stored actually IN the tree. TODO(kpfaulkner) investigate
	// if data should be in tree or just IDs?
	groups []GroupDetails
}

// NewTile creates a new tile at slippy co-ords x,y,z
// Probably should only be used for root tile. FIXME(kpfaulkner) confirm?
func NewTile(x int32, y int32, z byte) *Tile {
	t := &Tile{}
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	t.QuadKey = quadKey
	return t
}

// SetTileType for a given groupID and tiletype.
func (t *Tile) SetTileType(groupID uint32, tt TileType) {
	g := GroupDetails{
		GroupID: groupID,
		Type:    tt,
	}

	t.groups = append(t.groups, g)
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

func (t *Tile) GetTileZoomLevel() byte {
	return GetTileZoomLevel(t.QuadKey)
}

// SetFullForTileType sets the full flag for a given tile type.
// Only creates Full map at this stage (saves us creating a potential mass of unused maps)
func (t *Tile) SetFullForTileType(groupID uint32, tileType TileType, full bool) error {

	// loop through groups... see if already have groupid + type match.
	for i,g := range t.groups {
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

// GetFullForTileType gets the full flag for a given tile type. Defaults to false if no flag found
func (t *Tile) GetFullForTileType(groupID uint32, tileType TileType) bool {

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
