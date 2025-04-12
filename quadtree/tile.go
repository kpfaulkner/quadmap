package quadtree

type TileType uint16

// Tile is a node within a quadtree.
// Although a Tile instance will only be in the quadmap once (for a given quadkey) it may
// contain a key used to look up specifics for the quadkey in SQLite.
type Tile struct {
	QuadKey QuadKey

	// details about tilestypes, full or not..
	Info uint64
}

// NewTile creates a new tile at slippy co-ords x,y,z
// Will probably only be used for root tile
func NewTile(x uint32, y uint32, z byte) (*Tile, error) {
	t := &Tile{}
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return nil, err
	}
	t.QuadKey = quadKey
	return t, nil
}

func NewTileWithQuadKey(quadKey QuadKey) *Tile {
	t := &Tile{}
	t.QuadKey = quadKey
	return t
}

// HasTileType checks if specific tiletype associated with tile.
// Need to make sure we look at tiletype section of quadkey
func (t *Tile) HasTileType(tt TileType) bool {

	// coarse grain locking for now
	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, g := range t.groups {
		if g.Details.GroupID() == groupID && g.Details.HasTileType(tt) {
			return true
		}
	}
	return false
}

// GetTileZoomLevel returns zoom level of tile
func (t *Tile) GetTileZoomLevel() byte {
	return t.QuadKey.Zoom()
}

// IsTileFullByGroupIDAndTileType gets the full flag for a given tile type. Defaults to false if no flag found
func (t *Tile) IsTileFullByGroupIDAndTileType(groupID GroupID, tileType TileType) bool {

	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, g := range t.groups {
		if g.Details.GroupID() == groupID {
			hasTileType, isFull := g.Details.HasTileTypeAndFull(tileType)
			if hasTileType {
				return isFull
			}

		}
	}
	return false
}

func (t *Tile) GetGroupDetailsByGroupIDAndTileType(groupID GroupID, tileType TileType) *GroupDetails {

	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, g := range t.groups {
		if g.Details.GroupID() == groupID {
			hasTileType, _ := g.Details.HasTileTypeAndFull(tileType)
			if hasTileType {
				return &g
			}
		}
	}
	return nil
}
