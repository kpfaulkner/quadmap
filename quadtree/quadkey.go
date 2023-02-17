package quadtree

import (
	"errors"
	"fmt"
)

// Misc functions for generating/calculating quadkeys

// GetParentQuadKey get parents quadkey for passed quadkey
func GetParentQuadKey(quadKey uint64) (uint64, error) {
	zoomLevel := quadKey & 0xFF
	parentZoomLevel := zoomLevel - 1

	if parentZoomLevel <= 0 {
		return 0, errors.New("no parent")
	}

	shift := 64 - (parentZoomLevel * 2)
	parentQuadKey := quadKey >> shift
	parentQuadKey = parentQuadKey << shift
	parentQuadKey |= uint64(parentZoomLevel)

	return parentQuadKey, nil
}

// GetChildQuadKeyForPos where pos is 0-3
// based off https://learn.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system?redirectedfrom=MSDN
func GetChildQuadKeyForPos(quadKey uint64, pos int) (uint64, error) {
	zoomLevel := quadKey & 0xFF

	rightShift := 63 - (zoomLevel * 2) + 1
	quadKey = quadKey >> rightShift

	switch pos {
	case 0:
		quadKey = quadKey << 2
	case 1:
		quadKey = (quadKey << 2) | 0b01
	case 2:
		quadKey = (quadKey << 2) | 0b10
	case 3:
		quadKey = (quadKey << 2) | 0b11
	default:
		return 0, errors.New(fmt.Sprintf("invalid pos %d", pos))
	}

	quadKey = quadKey << (64 - (zoomLevel * 2) - 2)

	quadKey |= uint64(zoomLevel + 1)
	return quadKey, nil
}

// GetChildrenQuadKeys get all the quadkeys for the 4 children of the passed quadkey
func GetChildrenQuadKeys(quadKey uint64) []uint64 {
	var quadKeys []uint64
	for i := 0; i < 4; i++ {
		quadKey, _ := GetChildQuadKeyForPos(quadKey, i)
		quadKeys = append(quadKeys, quadKey)
	}
	return quadKeys
}

// GenerateQuadKeyIndexFromSlippy generates the quadkey index from slippy coords
func GenerateQuadKeyIndexFromSlippy(x uint32, y uint32, zoomLevel byte) uint64 {
	var binaryQuadkey uint64
	for i := zoomLevel; i > 0; i-- {
		var mask uint32 = 1 << (i - 1)
		var bitLocation uint64 = 64 - (uint64(zoomLevel-i+1) * 2) + 1
		if x&mask != 0 {
			binaryQuadkey |= uint64(0b1) << (bitLocation - 1)
		}
		if y&mask != 0 {
			binaryQuadkey |= uint64(0b1) << bitLocation
		}
	}
	binaryQuadkey |= uint64(zoomLevel)
	return binaryQuadkey
}

// GenerateSlippyCoordsFromQuadKeyIndex generates the slippy coords from quadkey index
func GenerateSlippyCoordsFromQuadKey(quadKey uint64) (int32, int32, byte) {
	var x int32
	var y int32

	zoomLevel := GetTileZoomLevel(quadKey)

	minPos := 64 - (int(zoomLevel) * 2)
	for i := 63; i > minPos; i -= 2 {

		firstBit := (quadKey >> i) & 1
		secondBit := (quadKey >> (i - 1)) & 1
		twoBits := (firstBit << 1) | secondBit
		switch twoBits {

		case 0b01:
			x += 1
		case 0b10:
			y += 1
		case 0b11:
			x += 1
			y += 1
		}
		x = x << 1
		y = y << 1
	}

	// undo last shift.
	x = x >> 1
	y = y >> 1
	return x, y, zoomLevel
}

// GetTileZoomLevel get the zoom level of the quadkey
func GetTileZoomLevel(quadKey uint64) byte {
	zoomLevel := byte(quadKey & 0xFF)
	return zoomLevel
}

// GenerateMinMaxQuadKeysForZoom given a quadkey and a desired zoom level, keep converting
// quadkey to desired zoom level and get min/max quadkeys (top left, bottom right)
// Practically this will only be valid if the tile associated with the quadKey is "full", but
// it's up the caller to check this.
func GenerateMinMaxQuadKeysForZoom(quadKey uint64, zoom byte) (uint64, uint64, error) {
	currentZoom := GetTileZoomLevel(quadKey)
	if currentZoom > zoom {
		return 0, 0, errors.New("unable to generate min/max zooms")
	}

	minChild := quadKey
	maxChild := quadKey
	for z := byte(0); z < zoom-currentZoom; z++ {
		minChild, _ = GetChildQuadKeyForPos(minChild, 0)
		maxChild, _ = GetChildQuadKeyForPos(maxChild, 3)
	}

	return minChild, maxChild, nil
}
