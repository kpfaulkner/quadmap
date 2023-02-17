package quadtree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (

	// levels 1-6 are populated (first 12 bits) and can see level (6) indicated at end of binary
	quadKey QuadKey = 0b1101110110110000000000000000000000000000000000000000000000000110

	// children of QuadKey. Note bits 12 and 13 as well as zoom at end.
	Child0 QuadKey = 0b1101110110110000000000000000000000000000000000000000000000000111
	Child1 QuadKey = 0b1101110110110100000000000000000000000000000000000000000000000111
	Child2 QuadKey = 0b1101110110111000000000000000000000000000000000000000000000000111
	Child3 QuadKey = 0b1101110110111100000000000000000000000000000000000000000000000111

	// MinChild (top left) of QuadKey (above) at level 21
	MinChildZoom21 uint64 = 0b1101110110110000000000000000000000000000000000000000000000010101

	// MaxChild (bottom right) of QuadKey (above) at level 21
	MaxChildZoom21 uint64 = 0b1101110110111111111111111111111111111111110000000000000000010101

	// Parent is same as Quadkey but bits 10-11 are zeroed and length (at end of binary) now reads 5
	parent QuadKey = 0b1101110110000000000000000000000000000000000000000000000000000101
)

// TestGetParentQuadKey checks parent calculation is correct
func TestGetParentQuadKey(t *testing.T) {

	// confirm we've got zoom 6
	assert.Equal(t, uint8(6), quadKey.Zoom(), "Zoom level should be 6")

	parentQuadKey, err := quadKey.Parent()
	assert.Nil(t, err, "Should not have error when getting parent quadkey")
	assert.Equal(t, parent, parentQuadKey, "Parent quadkey incorrect")
	assert.Equal(t, uint8(5), parentQuadKey.Zoom(), "Parent zoom level should be 5")

}

// TestGetQuadKeyFromSlippyCoords checks child quadkey calculation is correct.
// see https://learn.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system?redirectedfrom=MSDN
// for details
func TestGetChildQuadKeyForPos(t *testing.T) {

	// confirm we've got zoom 6
	assert.Equal(t, uint8(6), quadKey.Zoom(), "Zoom level should be 6")

	childPos0, err := quadKey.ChildAtPos(0)
	assert.Nil(t, err, "Should not have error when getting child quadkey")
	assert.Equal(t, Child0, childPos0, "Child quadkey incorrect")
	assert.Equal(t, uint8(7), childPos0.Zoom(), "Child zoom level should be 7")

	childPos1, err := quadKey.ChildAtPos(1)
	assert.Nil(t, err, "Should not have error when getting child quadkey")
	assert.Equal(t, Child1, childPos1, "Child quadkey incorrect")
	assert.Equal(t, uint8(7), childPos1.Zoom(), "Child zoom level should be 7")

	childPos2, err := quadKey.ChildAtPos(2)
	assert.Nil(t, err, "Should not have error when getting child quadkey")
	assert.Equal(t, Child2, childPos2, "Child quadkey incorrect")
	assert.Equal(t, uint8(7), childPos2.Zoom(), "Child zoom level should be 7")

	childPos3, err := quadKey.ChildAtPos(3)
	assert.Nil(t, err, "Should not have error when getting child quadkey")
	assert.Equal(t, Child3, childPos3, "Child quadkey incorrect")
	assert.Equal(t, uint8(7), childPos3.Zoom(), "Child zoom level should be 7")

}

// TestGenerateMinMaxQuadKeysForZoom confirms that min/max (top left, bottom right) quadkeys are generated
// based off an original quadkey and zoom target
func TestGenerateMinMaxQuadKeysForZoom(t *testing.T) {

	minChild, maxChild, err := GenerateMinMaxQuadKeysForZoom(QuadKey, 7)
	assert.NoErrorf(t, err, "no error expected")
	assert.Equal(t, Child0, minChild, "min child incorrect")
	assert.Equal(t, Child3, maxChild, "max child incorrect")

	minChild, maxChild, err = GenerateMinMaxQuadKeysForZoom(QuadKey, 21)
	assert.NoErrorf(t, err, "no error expected")
	assert.Equal(t, MinChildZoom21, minChild, "min child incorrect")
	assert.Equal(t, MaxChildZoom21, maxChild, "max child incorrect")

}
