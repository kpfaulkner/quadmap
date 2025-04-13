package quadtree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAddTile create quadmap and adds tile
func TestAddTile(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(0, 0, 0)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")
}

// TestNumberOfTiles create quadmap and adds tile
func TestNumberOfTiles(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(0, 0, 0)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")

	numTiles := qm.NumberOfTiles()
	assert.EqualValues(t, 1, numTiles, "Should have 1 tile")
}

// TestCreateTileAtSlippyCoords create quadmap and adds tile
func TestCreateTileAtSlippyCoords(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(0, 0, 0)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.Nil(t, err, "Should not have error when adding tile")

	// quadindex for 1,1,1 is 0b1100000000000000000000000000000000000000000000000000000000000001
	tile, err = qm.CreateTileAtSlippyCoords(1, 1, 1, TileTypeVert, true)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")
	assert.Equal(t, QuadKey(0b1100000000000000000000000000000000000000000000000000000000000001), tile.QuadKey, "QuadKey incorrect")
}

// TestGetTileForSlippyAndTileType create quadmap and adds tile and check if exists
func TestGetTileForSlippyAndTileType(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(1, 1, 1)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")

	// quadindex for 5,5,5 is 0b110011000000000000000000000000000000000000000000000000000101
	tile, err = qm.CreateTileAtSlippyCoords(5, 5, 5, TileTypeNorth, true)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")
	assert.Equal(t, QuadKey(0b110011000000000000000000000000000000000000000000000000000101), tile.QuadKey, "QuadKey incorrect")

	tile, err = qm.GetTileForSlippyAndTileType(5, 5, 5, TileTypeNorth)
	assert.NoError(t, err, "Should not have error when getting tile")

	x, y, z := tile.QuadKey.SlippyCoords()
	assert.EqualValues(t, x, 5)
	assert.EqualValues(t, y, 5)
	assert.EqualValues(t, z, 5)

}
