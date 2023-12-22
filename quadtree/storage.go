package quadtree

type Storage interface {

	// GetTileDetails returns the details (groupID, tileType,  full status etc) from storage.
	GetTileDetails(quadKey QuadKey) (TileDetails, error)

	// SetTileDetails sets the details (groupID, tileType,  full status etc) in storage.
	// If groupID+scale+tileType already exists, then it will be overwritten.
	SetTileDetail(quadKey QuadKey, detail TileDetail) error

	// GetTileDetailsGroupByTileType returns the details for a given tileType at location.
	GetTileDetailsByTileType(quadKey QuadKey, tileType TileType) (TileDetails, error)
}

type PebbleStorage struct {
}

func NewPebbleStorage() *PebbleStorage {
	return &PebbleStorage{}
}

func (ps *PebbleStorage) GetTileDetails(quadKey QuadKey) (TileDetails, error) {
	return TileDetails{}, nil
}

func (ps *PebbleStorage) SetTileDetail(quadKey QuadKey, details TileDetail) error {
	return nil
}

func (ps *PebbleStorage) GetTileDetailsByTileType(quadKey QuadKey, tileType TileType) (TileDetails, error) {
	return TileDetails{}, nil
}
