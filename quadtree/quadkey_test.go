package quadtree

import (
	"fmt"
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

func TestEnv(t *testing.T) {
	for _, tc := range []struct {
		qk             QuadKey
		minLon, minLat float64
		maxLon, maxLat float64
	}{
		{
			qk:     GenerateQuadKeyIndexFromSlippy(60292, 39326, 16),
			minLon: 151.19384765625,
			minLat: -33.86585445407186,
			maxLon: 151.1993408203125,
			maxLat: -33.861293113515515,
		},
	} {
		// TODO: QuadKey.String()
		t.Run(fmt.Sprint(tc.qk), func(t *testing.T) {
			env, err := tc.qk.Envelope()
			assert.NoError(t, err)
			min, max, ok := env.MinMaxXYs()
			assert.True(t, ok)
			assert.InDelta(t, tc.minLon, min.X, 1e-9)
			assert.InDelta(t, tc.minLat, min.Y, 1e-9)
			assert.InDelta(t, tc.maxLon, max.X, 1e-9)
			assert.InDelta(t, tc.maxLat, max.Y, 1e-9)
		})
	}
}
