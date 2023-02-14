package quadtree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (

	// levels 1-6 are populated (first 12 bits) and can see level (6) indicated at end of binary
	QuadKey uint64 = 0b1101110110110000000000000000000000000000000000000000000000000110

	Child0 uint64 = 0b1101110110110000000000000000000000000000000000000000000000000111
	Child1 uint64 = 0b1101110110110100000000000000000000000000000000000000000000000111
	Child2 uint64 = 0b1101110110111000000000000000000000000000000000000000000000000111
	Child3 uint64 = 0b1101110110111100000000000000000000000000000000000000000000000111

	// Parent is same as Quadkey but bits 10-11 are zeroed and length (at end of binary) now reads 5
	ParentQuadKey uint64 = 0b1101110110000000000000000000000000000000000000000000000000000101
)

// TestGetParentQuadKey checks parent calculation is correct
func TestGetParentQuadKey(t *testing.T) {

	// confirm we've got zoom 6
	zoom := GetTileZoomLevel(QuadKey)
	assert.Equal(t, uint8(6), zoom, "Zoom level should be 6")

	parentQuadKey, err := GetParentQuadKey(QuadKey)
	assert.Nil(t, err, "Should not have error when getting parent quadkey")
	assert.Equal(t, ParentQuadKey, parentQuadKey, "Parent quadkey incorrect")
	assert.Equal(t, uint8(5), GetTileZoomLevel(parentQuadKey), "Parent zoom level should be 5")

}

// TestGetQuadKeyFromSlippyCoords checks child quadkey calculation is correct.
// see https://learn.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system?redirectedfrom=MSDN
// for details
func TestGetChildQuadKeyForPos(t *testing.T) {

	// confirm we've got zoom 6
	zoom := GetTileZoomLevel(QuadKey)
	assert.Equal(t, uint8(6), zoom, "Zoom level should be 6")

	childPos0, err := GetChildQuadKeyForPos(QuadKey, 0)
	assert.Nil(t, err, "Should not have error when getting child quadkey")
	assert.Equal(t, Child0, childPos0, "Child quadkey incorrect")
	assert.Equal(t, uint8(7), GetTileZoomLevel(childPos0), "Child zoom level should be 7")

	childPos1, err := GetChildQuadKeyForPos(QuadKey, 1)
	assert.Nil(t, err, "Should not have error when getting child quadkey")
	assert.Equal(t, Child1, childPos1, "Child quadkey incorrect")
	assert.Equal(t, uint8(7), GetTileZoomLevel(childPos1), "Child zoom level should be 7")

	childPos2, err := GetChildQuadKeyForPos(QuadKey, 2)
	assert.Nil(t, err, "Should not have error when getting child quadkey")
	assert.Equal(t, Child2, childPos2, "Child quadkey incorrect")
	assert.Equal(t, uint8(7), GetTileZoomLevel(childPos2), "Child zoom level should be 7")

	childPos3, err := GetChildQuadKeyForPos(QuadKey, 3)
	assert.Nil(t, err, "Should not have error when getting child quadkey")
	assert.Equal(t, Child3, childPos3, "Child quadkey incorrect")
	assert.Equal(t, uint8(7), GetTileZoomLevel(childPos3), "Child zoom level should be 7")

}
