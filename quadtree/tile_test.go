package quadtree

//
//import (
//	"log"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//const (
//
//	// the positioning is type at bit X...  full at bit X-1
//	VertPos   = 1 << 31
//	VertFull  = 1 << 30
//	NorthPos  = 1 << 29
//	NorthFull = 1 << 28
//
//	VertType TileType = iota
//	NorthType
//)
//
//func setupSuite(t testing.T) func(t testing.T) {
//	log.Println("setup suite")
//
//	return func(t testing.T) {
//
//	}
//}
//
//// TestSingleSetTileTypeAndFull tests setting of single tiletype/full
//func TestSingleSetTileTypeAndFull(t *testing.T) {
//
//	m := map[TileType]int{
//		VertType:  VertPos,
//		NorthType: NorthPos,
//	}
//	SetupTileLUT(m)
//
//	tile := NewTile()
//	err := tile.SetTileType(VertType, false)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	isFull := tile.IsFullForTileType(VertType)
//	assert.EqualValues(t, false, isFull, "should not be full")
//
//	tile2 := NewTile()
//	err = tile2.SetTileType(VertType, true)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	isFull = tile2.IsFullForTileType(VertType)
//	assert.EqualValues(t, true, isFull, "should be full")
//}
//
//// TestTwoSetTileTypes tests setting of two unrelated groupid/tiletype
//func TestTwoSetTileTypes(t *testing.T) {
//
//	m := map[TileType]int{
//		VertType:  VertPos,
//		NorthType: NorthPos,
//	}
//	SetupTileLUT(m)
//
//	tile := NewTile()
//	err := tile.SetTileType(VertType, false)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	err = tile.SetTileType(NorthType, true)
//	assert.Nil(t, err, "Should not have error when setting tile type")
//
//	assert.EqualValues(t, true, tile.HasTileType(VertType), "should have tile type 1")
//	assert.EqualValues(t, true, tile.HasTileType(NorthType), "should have tile type 3")
//
//}
