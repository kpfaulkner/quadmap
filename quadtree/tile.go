package quadtree

type TileType uint16

const (
	// only use vert for now
	TileTypeVert TileType = 0b00000000000000001
	TileTypeEast TileType = 0b00000000000000010
)

type GroupDetails struct {

	// GroupID... probably not required since it should be used as a map key to access this.
	// TODO(kpfaulkner) confirm if this is required.
	GroupID string

	// This tile is used for various tile types
	// use as bitmask. Assumption that will not have more than 16 tile types.
	Types uint16

	// This tile and all children/grandchildren/second cousins once removed etc... are present.
	// Full is tiletype specific.
	full map[TileType]bool
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

	// groupIDs that have information for this tile. The IDs listed here can be used elsewhere to look up data.
	// Not convinced that the groupdata *has* to be stored actually IN the tree. TODO(kpfaulkner) investigate
	// if data should be in tree or just IDs?
	groupIDs map[string]*GroupDetails
}

// NewTile creates a new tile at slippy co-ords x,y,z
// Probably should only be used for root tile. FIXME(kpfaulkner) confirm?
func NewTile(x int32, y int32, z byte) *Tile {
	t := &Tile{}
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	t.QuadKey = quadKey
	t.groupIDs = make(map[string]*GroupDetails)
	return t
}

func (t *Tile) SetTileType(groupID string, tt TileType) {
	var g *GroupDetails
	var ok bool
	if g, ok = t.groupIDs[groupID]; !ok {
		g = &GroupDetails{GroupID: groupID, Types: uint16(tt)}
	}
	g.Types |= uint16(tt)
	t.groupIDs[groupID] = g
}

func (t *Tile) HasTileType(groupID string, tt TileType) bool {
	return t.groupIDs[groupID].Types&(1<<uint(tt)) != 0
}

func (t *Tile) GetTileZoomLevel() byte {
	return GetTileZoomLevel(t.QuadKey)
}

// SetFullForTileType sets the full flag for a given tile type.
// Only creates Full map at this stage (saves us creating a potential mass of unused maps)
func (t *Tile) SetFullForTileType(groupID string, tileType TileType, full bool) error {
	var g *GroupDetails
	var ok bool
	if g, ok = t.groupIDs[groupID]; !ok {
		g = &GroupDetails{}
		t.groupIDs[groupID] = g
	}

	g.Types |= uint16(tileType)
	if g.full == nil {
		g.full = make(map[TileType]bool)
		t.groupIDs[groupID] = g
	}
	t.groupIDs[groupID].full[tileType] = full
	return nil
}

// GetFullForTileType gets the full flag for a given tile type. Defaults to false if no flag found
func (t *Tile) GetFullForTileType(groupID string, tileType TileType) bool {
	if full, ok := t.groupIDs[groupID].full[tileType]; ok {
		return full
	}

	return false
}
