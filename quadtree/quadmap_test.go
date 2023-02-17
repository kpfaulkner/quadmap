package quadtree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAddTile create quadmap and adds tile
func TestAddTile(t *testing.T) {
	qm := NewQuadMap(10)
	tile := NewTile(0, 0, 0)
	err := qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")
}

// TestNumberOfTiles create quadmap and adds tile
func TestNumberOfTiles(t *testing.T) {
	qm := NewQuadMap(10)
	tile := NewTile(0, 0, 0)
	err := qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")

	numTiles := qm.NumberOfTiles()
	assert.EqualValues(t, 1, numTiles, "Should have 1 tile")
}

// TestCreateTileAtSlippyCoords create quadmap and adds tile
func TestCreateTileAtSlippyCoords(t *testing.T) {
	qm := NewQuadMap(10)
	tile := NewTile(0, 0, 0)
	err := qm.AddTile(tile)
	assert.Nil(t, err, "Should not have error when adding tile")

	// quadindex for 1,1,1 is 0xc000000000000001
	tile, err = qm.CreateTileAtSlippyCoords(1, 1, 1, 2, 3)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")
	assert.Equal(t, QuadKey(0xc000000000000001), tile.QuadKey, "QuadKey incorrect")
}

// TestHaveTileForSlippyGroupIDAndTileType create quadmap and adds tile and check if exists
func TestHaveTileForSlippyGroupIDAndTileType(t *testing.T) {
	qm := NewQuadMap(10)
	tile := NewTile(0, 0, 0)
	err := qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")

	// quadindex for 5,5,5 is 0xcc0000000000005
	tile, err = qm.CreateTileAtSlippyCoords(5, 5, 5, 2, 3)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")
	assert.Equal(t, QuadKey(0xcc0000000000005), tile.QuadKey, "QuadKey incorrect")

	haveTile, err := qm.HaveTileForSlippyGroupIDAndTileType(5, 5, 5, 2, 3)
	assert.NoError(t, err, "Should not have error when checking tile/group")
	assert.True(t, haveTile, "Should have tile")

	_, err = qm.HaveTileForSlippyGroupIDAndTileType(5, 5, 5, 2, 4)
	assert.Error(t, err, "Should have error when checking tile/group")
}
