package quadtree

import "fmt"

type TileType int

const (
	// only use vert for now
	TileTypeVert TileType = iota
)

// Tile is a node within a quadtree.
// Should shuffle these around for byte padding.... but not yet.
type Tile struct {
	// stupid to keep it in here?
	// Will also store the zoom level in the key.
	QuadKey uint64

	// This tile is used for various tile types
	// use as bitmask. Assumption that will not have more than 16 tile types.
	Types uint16

	// This tile and all children/grandchildren/second cousins once removed etc... are present.
	Full bool

	// Misc data that is associated with the tile.
	// This could be anything that the caller/client wants it to be.
	// This data is TileType specific.
	Data map[TileType]interface{}
}

// NewTile creates a new tile at slippy co-ords x,y,z
// Probably should only be used for root tile. FIXME(kpfaulkner) confirm?
func NewTile(x int32, y int32, z byte) *Tile {
	t := &Tile{}
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	t.QuadKey = quadKey
	return t
}

func (t *Tile) SetTileType(tt TileType) {
	t.Types |= 1 << uint(tt)
}

func (t *Tile) HasTileType(tt TileType) bool {
	return t.Types&(1<<uint(tt)) != 0
}

func (t *Tile) GetTileZoomLevel() byte {
	return GetTileZoomLevel(t.QuadKey)
}

// SetDataForTileType sets the data for a given tile type.
// Only creates Data map at this stage (saves us creating a potential mass of unused maps)
func (t *Tile) SetDataForTileType(tileType TileType, data interface{}) error {
	if t.Data == nil {
		t.Data = make(map[TileType]interface{})
	}

	t.Data[tileType] = data
	return nil
}

func (t *Tile) GetDataForTileType(tileType TileType) (interface{}, error) {
	if data, ok := t.Data[tileType]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("no data for tile type %d", tileType)
}
