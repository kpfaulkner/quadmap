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
	tile, err = qm.CreateTileAtSlippyCoords(1, 1, 1, TileTypeVert)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")
	assert.Equal(t, QuadKey(0b1100000000000000000000000000000000000000000000000000000000000001), tile.QuadKey, "QuadKey incorrect")
}

// TestHaveTileForSlippyGroupIDAndTileType create quadmap and adds tile and check if exists
func TestHaveTileForSlippyGroupIDAndTileType(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(0, 0, 0)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")

	// quadindex for 5,5,5 is 0b110011000000000000000000000000000000000000000000000000000101
	tile, err = qm.CreateTileAtSlippyCoords(5, 5, 5, 2)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")
	assert.Equal(t, QuadKey(0b110011000000000000000000000000000000000000000000000000000101), tile.QuadKey, "QuadKey incorrect")

	haveTile, err := qm.HaveTileForSlippyGroupIDAndTileType(5, 5, 5, 2, 3)
	assert.NoError(t, err, "Should not have error when checking tile/group")
	assert.True(t, haveTile, "Should have tile")

	_, err = qm.HaveTileForSlippyGroupIDAndTileType(5, 5, 5, 2, 4)
	assert.Error(t, err, "Should have error when checking tile/group")
}

// TestGetSlippyBoundsForGroupIDTileTypeAndZoom
func TestGetSlippyBoundsForGroupIDTileTypeAndZoom(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(0, 0, 0)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")

	tile, err = qm.CreateTileAtSlippyCoords(3, 2, 2, 2)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")

	////tile.SetFullForGroupIDAndTileType(2, 3, true)
	////minX, minY, maxX, maxY, err := qm.GetSlippyBoundsForGroupIDTileTypeAndZoom(2, 3, 3)
	////assert.NoError(t, err, "Should not have error when getting bounds")
	//
	//// check min/max are top left and bottom right
	//assert.Equal(t, uint32(6), minX, "MinX incorrect")
	//assert.Equal(t, uint32(4), minY, "MinY incorrect")
	//assert.Equal(t, uint32(7), maxX, "MaxX incorrect")
	//assert.Equal(t, uint32(5), maxY, "MaxY incorrect")
}
