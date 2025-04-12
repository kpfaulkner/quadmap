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
	MinChildZoom21 QuadKey = 0b1101110110110000000000000000000000000000000000000000000000010101

	// MaxChild (bottom right) of QuadKey (above) at level 21
	MaxChildZoom21 QuadKey = 0b1101110110111111111111111111111111111111110000000000000000010101

	// Parent is same as Quadkey but bits 10-11 are zeroed and length (at end of binary) now reads 5
	parent QuadKey = 0b1101110110000000000000000000000000000000000000000000000000000101
)

func TestGenerateQuadKeyIndexFromSlippy(t *testing.T) {
	for _, tc := range []struct {
		x, y      uint32
		z         byte
		qk        QuadKey
		expectErr bool
	}{
		{
			x: 0, y: 0, z: 0,
			qk:        0b0000000000000000000000000000000000000000000000000000000000000000,
			expectErr: true,
		},
		{
			x: 0, y: 0, z: 1,
			qk: 0b0000000000000000000000000000000000000000000000000000000000000001,
		},
		{
			x: 1, y: 0, z: 1,
			qk: 0b0100000000000000000000000000000000000000000000000000000000000001,
		},
		{
			x: 0, y: 1, z: 1,
			qk: 0b1000000000000000000000000000000000000000000000000000000000000001,
		},
		{
			x: 1, y: 1, z: 1,
			qk: 0b1100000000000000000000000000000000000000000000000000000000000001,
		},
		{
			x: 2, y: 2, z: 2,
			qk: 0b1100000000000000000000000000000000000000000000000000000000000010,
		},
		{
			x: 3, y: 2, z: 2,
			qk: 0b1101000000000000000000000000000000000000000000000000000000000010,
		},
		{
			x: 6, y: 4, z: 3,
			qk: 0b1101000000000000000000000000000000000000000000000000000000000011,
		},
		{
			x: 6, y: 5, z: 3,
			qk: 0b1101100000000000000000000000000000000000000000000000000000000011,
		},
		{
			x: 13, y: 11, z: 4,
			qk: 0b1101101100000000000000000000000000000000000000000000000000000100,
		},
	} {
		qk, err := GenerateQuadKeyIndexFromSlippy(tc.x, tc.y, tc.z)
		if tc.expectErr && err == nil {
			t.Error("expected error, got none")
			return
		}

		if !tc.expectErr && err != nil {
			t.Errorf("unexpected error %s", err.Error())
			return
		}

		assert.Equal(t, tc.qk, qk)
		x, y, z := qk.SlippyCoords()
		assert.Equal(t, tc.x, x)
		assert.Equal(t, tc.y, y)
		assert.Equal(t, tc.z, z)
	}
}

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

// TestGetMinMaxEquivForZoomLevel confirms that min/max (top left, bottom right) quadkeys are generated
// based off an original quadkey and zoom target
func TestGetMinMaxEquivForZoomLevel(t *testing.T) {

	minChild, maxChild, err := quadKey.GetMinMaxEquivForZoomLevel(7)
	assert.NoErrorf(t, err, "no error expected")
	assert.Equal(t, Child0, minChild, "min child incorrect")
	assert.Equal(t, Child3, maxChild, "max child incorrect")

	minChild, maxChild, err = quadKey.GetMinMaxEquivForZoomLevel(21)
	assert.NoErrorf(t, err, "no error expected")
	assert.Equal(t, MinChildZoom21, minChild, "min child incorrect")
	assert.Equal(t, MaxChildZoom21, maxChild, "max child incorrect")

}

//func TestEnv(t *testing.T) {
//	for _, tc := range []struct {
//		qk             QuadKey
//		minLon, minLat float64
//		maxLon, maxLat float64
//	}{
//		{
//			qk:     GenerateQuadKeyIndexFromSlippy(60292, 39326, 16),
//			minLon: 151.19384765625,
//			minLat: -33.86585445407186,
//			maxLon: 151.1993408203125,
//			maxLat: -33.861293113515515,
//		},
//	} {
//		// TODO: QuadKey.String()
//		t.Run(fmt.Sprint(tc.qk), func(t *testing.T) {
//			env, err := tc.qk.Envelope()
//			assert.NoError(t, err)
//			min, max, ok := env.MinMaxXYs()
//			assert.True(t, ok)
//			assert.InDelta(t, tc.minLon, min.X, 1e-9)
//			assert.InDelta(t, tc.minLat, min.Y, 1e-9)
//			assert.InDelta(t, tc.maxLon, max.X, 1e-9)
//			assert.InDelta(t, tc.maxLat, max.Y, 1e-9)
//		})
//	}
//}
