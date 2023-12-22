package quadtree

var (
	TileTypeLUT map[TileType]int
)

// TileType is the type of a given tile
// Given we'll have very few tile types we could just use a byte, but
// may eventually use this as a bitmask to help filtering. So will
// keep it as a uint16 for now. Can change later if space becomes an issue
type TileType uint32

// Tile will now just be a bitmask indicating what type of tile (Vert, etc etc) is at this location
// as well as if the tile (for that type) is full or not.
// Of course we could have multiple groups that cover this tile location and we dont know which of
// those are full. This is where going off to PebbleDB would be required for more details.
type Tile uint32

func SetupTileLUT(lut map[TileType]int) {
	TileTypeLUT = lut
}

// NewTile creates a new tile at slippy co-ords x,y,z
// Will probably only be used for root tile
func NewTile() Tile {
	t := Tile(0)
	return t
}

// SetTileType set the tile type for a given tile.
// If filetype is not already full but parameter full is true then it will be set.
// If filetype is already full but parameter is NOT full, then we will leave it as true.
// Basically... *some* version of this tiletype is full... but will need to consult with
// PebbleDB to figure out which. I *think* that's fine for now.
func (t *Tile) SetTileType(tt TileType, full bool) error {

	posMask := Tile(TileTypeLUT[tt])

	var fullMask Tile
	if full {
		fullMask = Tile(posMask >> 1)
	}

	// set it regardless
	combined := posMask | fullMask
	*t = *t | combined

	return nil
}

func (t *Tile) HasTileType(tt TileType) bool {
	posMask := Tile(TileTypeLUT[tt])
	return *t&posMask != 0
}

// IsFullForTileType gets the full flag for a given tile type. Defaults to false if no flag found
// Is it NOT checking groupID... so is not strictly reliable to determine which groupID caused it
// to be full. This is probably more useful to just eliminate tiles that are not full.
func (t *Tile) IsFullForTileType(tileType TileType) bool {

	posMask := Tile(TileTypeLUT[tileType])
	fullMask := Tile(posMask >> 1)
	return *t&fullMask != 0
}
