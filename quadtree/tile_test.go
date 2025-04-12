package quadtree

import (
	"log"
	"testing"
)

func setupSuite(t testing.T) func(t testing.T) {
	log.Println("setup suite")

	return func(t testing.T) {
	}
}

//
//// TestSingleSetTileType tests setting of single groupid/tiletype
//func TestSingleSetTileType(t *testing.T) {
//
//	tile, err := NewTile(0, 0, 0)
//	assert.NoError(t, err, "Should not have error")
//	//err = tile.SetTileType(1, 2)
//	//assert.Nil(t, err, "Should not have error when setting tile type")
//
//	// check groupID
//	assert.EqualValues(t, 1, tile.groups[0].GroupID, "GroupID should be 1")
//
//	// check tiletype
//	assert.EqualValues(t, 2, tile.groups[0].Type, "TileType should be 2")
//}
//
//// TestTwoSetTileTypes tests setting of two unrelated groupid/tiletype
//func TestTwoSetTileTypes(t *testing.T) {
//
//	tile := NewTile(0, 0, 0)
//	err := tile.SetTileType(1, 2)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	err = tile.SetTileType(3, 4)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	assert.Equal(t, 2, len(tile.groups), "Should have 2 groups")
//
//	// check groupID 0
//	assert.EqualValues(t, 1, tile.groups[0].GroupID, "GroupID should be 1")
//
//	// check tiletype 0
//	assert.EqualValues(t, 2, tile.groups[0].Type, "TileType should be 2")
//
//	// check groupID 1
//	assert.EqualValues(t, 3, tile.groups[1].GroupID, "GroupID should be 3")
//
//	// check tiletype 1
//	assert.EqualValues(t, 4, tile.groups[1].Type, "TileType should be 4")
//}
//
//// TestDuplicateSetTileType tests setting of same groupid/tiletype twice
//func TestDuplicateSetTileType(t *testing.T) {
//
//	tile := NewTile(0, 0, 0)
//	err := tile.SetTileType(1, 2)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	err = tile.SetTileType(1, 2)
//	assert.Error(t, err, "should have thrown error due to duplicate groupid/tiletype")
//}
//
//// TestHasTileTypeSuccess add groupid/tiletype and check if it exists
//func TestHasTileTypeSuccess(t *testing.T) {
//
//	tile := NewTile(0, 0, 0)
//	err := tile.SetTileType(1, 2)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	hasTT := tile.HasTileType(1, 2)
//	assert.True(t, hasTT, "Should have found groupid/tiletype")
//}
//
//// TestHasTileTypeFail add groupid/tiletype and check for combination that does NOT exist
//func TestHasTileTypeFail(t *testing.T) {
//
//	tile := NewTile(0, 0, 0)
//	err := tile.SetTileType(1, 2)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	hasTT := tile.HasTileType(2, 3)
//	assert.False(t, hasTT, "Should NOT have found groupid/tiletype")
//}
//
//// TestHasTileTypeSuccessMultiple add multiple groupid/tiletype combinations and check the last one entered
//func TestHasTileTypeSuccessMultiple(t *testing.T) {
//
//	tile := NewTile(0, 0, 0)
//	err := tile.SetTileType(1, 2)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	err = tile.SetTileType(3, 4)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	hasTT := tile.HasTileType(3, 4)
//	assert.True(t, hasTT, "Should have found groupid/tiletype")
//}
//
//// TestGetTileZoomLevel extracts zoom level from quadkey
//// Picks a few zoom levels and checks
//func TestGetTileZoomLevel(t *testing.T) {
//
//	tile := NewTile(0, 0, 14)
//	zoom := tile.GetTileZoomLevel()
//	assert.Equal(t, byte(14), zoom, "Should have zoom level of 14")
//
//	tile = NewTile(0, 0, 16)
//	zoom = tile.GetTileZoomLevel()
//	assert.Equal(t, byte(16), zoom, "Should have zoom level of 16")
//
//	tile = NewTile(0, 0, 21)
//	zoom = tile.GetTileZoomLevel()
//	assert.Equal(t, byte(21), zoom, "Should have zoom level of 21")
//}
//
//// TestSetFullForGroupIDAndTileType sets the full flag for a given groupid/tiletype combo
//func TestSetFullForGroupIDAndTileType(t *testing.T) {
//
//	tile := NewTile(0, 0, 14)
//	err := tile.SetTileType(1, 2)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	// set existing groupid/tiletype
//	err = tile.SetFullForGroupIDAndTileType(1, 2, true)
//	assert.Nil(t, err, "Should not have error when setting full ")
//
//	err = tile.SetFullForGroupIDAndTileType(3, 4, true)
//	assert.Nil(t, err, "Should not have error when setting full ")
//}
//
//// TestGetFullForGroupIDAndTileType sets the full flag for a given groupid/tiletype combo
//// then retrieves it!
//func TestGetFullForGroupIDAndTileType(t *testing.T) {
//
//	tile := NewTile(0, 0, 14)
//	err := tile.SetTileType(1, 2)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	// set existing groupid/tiletype
//	err = tile.SetFullForGroupIDAndTileType(1, 2, true)
//	assert.Nil(t, err, "Should not have error when setting full ")
//
//	isFull := tile.GetFullForGroupIDAndTileType(1, 2)
//	assert.True(t, isFull, "Should have found full flag")
//
//	isFull = tile.GetFullForGroupIDAndTileType(3, 4)
//	assert.False(t, isFull, "Should NOT have found full flag")
//}
