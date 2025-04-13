package quadtree

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupSuite(t testing.T) func(t testing.T) {
	log.Println("setup suite")

	return func(t testing.T) {
	}
}

// TestSingleSetTileType tests setting of single groupid/tiletype
func TestSingleSetTileType(t *testing.T) {

	tile, err := NewTile(1, 1, 1)
	assert.NoError(t, err, "Should not have error")

	// check tileType (none set)
	tileTypeExists, tileTypeFull := tile.HasTileTypeAndFull(TileTypeVert)
	assert.Equal(t, false, tileTypeExists, "Should not have tileType")
	assert.Equal(t, false, tileTypeFull, "Should not have tileType")
	tile.AddTileType(TileTypeVert, true)

	tileTypeExists, tileTypeFull = tile.HasTileTypeAndFull(TileTypeVert)
	assert.Equal(t, true, tileTypeExists, "Should not have tileType")
	assert.Equal(t, true, tileTypeFull, "Should not have tileType")
}

// TestTwoSetTileTypes tests setting of two unrelated groupid/tiletype
func TestTwoSetTileTypes(t *testing.T) {

	tile, err := NewTile(1, 1, 1)
	assert.NoError(t, err, "Should not have error")

	// check tileType (none set)
	tileTypeExists, tileTypeFull := tile.HasTileTypeAndFull(TileTypeVert)
	assert.Equal(t, false, tileTypeExists, "Should not have tileType")
	assert.Equal(t, false, tileTypeFull, "Should not have tileType")

	tileTypeExists, tileTypeFull = tile.HasTileTypeAndFull(TileTypeNorth)
	assert.Equal(t, false, tileTypeExists, "Should not have tileType")
	assert.Equal(t, false, tileTypeFull, "Should not have tileType")

	tile.AddTileType(TileTypeVert, true)
	tile.AddTileType(TileTypeVert, true)

	tileTypeExists, tileTypeFull = tile.HasTileTypeAndFull(TileTypeVert)
	assert.Equal(t, true, tileTypeExists, "Should not have tileType")
	assert.Equal(t, true, tileTypeFull, "Should not have tileType")

	tileTypeExists, tileTypeFull = tile.HasTileTypeAndFull(TileTypeNorth)
	assert.Equal(t, true, tileTypeExists, "Should not have tileType")
	assert.Equal(t, true, tileTypeFull, "Should not have tileType")
}
