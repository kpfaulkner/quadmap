package quadmap

type TileType uint32

const (
	TileTypeVert           TileType = 1 << iota // 0
	TileTypeDEM                                 // 1
	TileTypeNorth                               // 2
	TileTypeSouth                               // 3
	TileTypeEast                                // 4
	TileTypeWest                                // 5
	TileTypeVertDetail                          // 6
	TileTypeDSM                                 // 7
	TileTypeDTM                                 // 8
	TileTypeOrthoDEM                            // 9
	TileTypeOrthoDSM                            // 10
	TileTypeDetailDEM                           // 11
	TileTypeDetailDSM                           // 12
	TileTypeDetailDTM                           // 13
	TileTypeTrueOrtho                           // 14
	TileTypeNorthWest                           // 15
	TileTypeNorthEast                           // 16
	TileTypeSouthWest                           // 17
	TileTypeSouthEast                           // 18
	TileTypeDTMDEM                              // 19
	TileTypeDetailDTMDEM                        // 20
	TileTypeOverviewNIR                         // 21
	TileTypeDetailNIR                           // 22

	// TileTypeOffset must be >= number of TileType bits (23) so the
	// "full" flags (bits 0-22) and "present" flags (bits 23-45) don't overlap.
	TileTypeOffset = 23
)

// Tile is a node within a quadmap.
// Although a Tile instance will only be in the quadmap once (for a given quadkey) it may
// contain a key used to look up specifics for the quadkey in SQLite.
type Tile struct {
	QuadKey QuadKey

	// Details holds information about tiletypes, full/empty.. and potentially other info.
	//		|63-----------46|45-----------23|22----------------0|
	//		|     unused    |    TileType   |  Tiletype full     |
	//		|    (18 bits)  |   present     |      flags         |
	//		|               |   (23 bits)   |     (23 bits)      |
	//
	// Bits 22 -> 0 (23 bits) are used to indicate full for TileType.
	// Bits 45 -> 23 (23 bits) are used to indicate TileType present.
	Details uint64
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

// NewTileWithTileType creates a new tile at slippy co-ords x,y,z
// and also supplies initial tiletype
func NewTileWithTileTypeAndFull(x uint32, y uint32, z byte, tt TileType, full bool) (*Tile, error) {
	t := &Tile{}
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return nil, err
	}

	t.QuadKey = quadKey
	t.Details = uint64(0)
	t.AddTileType(tt, full)
	return t, nil
}

func NewTileWithQuadKey(quadKey QuadKey) *Tile {
	t := &Tile{}
	t.QuadKey = quadKey
	return t
}

// GetTileZoomLevel returns zoom level of tile
func (t *Tile) GetTileZoomLevel() byte {
	return t.QuadKey.Zoom()
}

// AddTileType Adds tiletype and full flag to tile
func (t *Tile) AddTileType(tileType TileType, full bool) {

	// set tiletype
	t.Details = t.Details | (uint64(tileType) << TileTypeOffset)

	if full {
		// set full flag
		t.Details = t.Details | uint64(tileType)
	} else {

		// clear full flag.
		t.Details &= ^uint64(tileType)
	}
}

// HasTileType checks if specific tiletype associated with tile.
// Returns if tiletype present and if full or not
func (t *Tile) HasTileTypeAndFull(tileType TileType) (bool, bool) {
	tileTypeShift := uint64(tileType) << TileTypeOffset
	tileTypePresent := t.Details & tileTypeShift
	tileTypeFull := t.Details & uint64(tileType)

	return tileTypePresent != 0, tileTypeFull != 0
}

func (t *Tile) HasTileType(tileType TileType) bool {
	tileTypeShift := uint64(tileType) << TileTypeOffset
	return (t.Details & tileTypeShift) != 0
}
