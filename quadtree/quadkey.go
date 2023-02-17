package quadtree

import (
	"errors"
	"fmt"
)

// Misc functions for generating/calculating quadkeys

type QuadKey uint64

// Parent get parents quadkey for passed quadkey
func (q QuadKey) Parent() (QuadKey, error) {
	zoomLevel := q & 0xFF
	parentZoomLevel := zoomLevel - 1

	if parentZoomLevel <= 0 {
		return 0, errors.New("no parent")
	}

	shift := 64 - (parentZoomLevel * 2)
	parent := q >> shift
	parent = parent << shift
	parent |= parentZoomLevel

	return parent, nil
}

// ChildAtPos where pos is 0-3
// based off https://learn.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system?redirectedfrom=MSDN
func (q QuadKey) ChildAtPos(pos int) (QuadKey, error) {
	zoomLevel := q & 0xFF

	rightShift := 63 - (zoomLevel * 2) + 1
	q = q >> rightShift

	switch pos {
	case 0:
		q = q << 2
	case 1:
		q = (q << 2) | 0b01
	case 2:
		q = (q << 2) | 0b10
	case 3:
		q = (q << 2) | 0b11
	default:
		return 0, errors.New(fmt.Sprintf("invalid pos %d", pos))
	}

	q = q << (64 - (zoomLevel * 2) - 2)

	q |= zoomLevel + 1
	return q, nil
}

// Children get all the quadkeys for the 4 children of the passed quadkey
func (q QuadKey) Children() []QuadKey {
	var children []QuadKey
	for i := 0; i < 4; i++ {
		child, _ := q.ChildAtPos(i)
		children = append(children, child)
	}
	return children
}

// GenerateQuadKeyIndexFromSlippy generates the quadkey index from slippy coords
func GenerateQuadKeyIndexFromSlippy(x uint32, y uint32, zoomLevel byte) QuadKey {
	var binaryQuadkey QuadKey
	for i := zoomLevel; i > 0; i-- {
		var mask uint32 = 1 << (i - 1)
		var bitLocation QuadKey = 64 - (QuadKey(zoomLevel-i+1) * 2) + 1
		if x&mask != 0 {
			binaryQuadkey |= 0b1 << (bitLocation - 1)
		}
		if y&mask != 0 {
			binaryQuadkey |= 0b1 << bitLocation
		}
	}
	binaryQuadkey |= QuadKey(zoomLevel)
	return binaryQuadkey
}

// SlippyCoords generates the slippy coords from quadkey index
func (q QuadKey) SlippyCoords() (int32, int32, byte) {
	var x int32
	var y int32

	zoomLevel := q.Zoom()

	minPos := 64 - (int(zoomLevel) * 2)
	for i := 63; i > minPos; i -= 2 {

		firstBit := (q >> i) & 1
		secondBit := (q >> (i - 1)) & 1
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

// Zoom get the zoom level of the quadkey
func (q QuadKey) Zoom() byte {
	zoomLevel := byte(q & 0xFF)
	return zoomLevel
}

// GenerateMinMaxQuadKeysForZoom given a quadkey and a desired zoom level, keep converting
// quadkey to desired zoom level and get min/max quadkeys (top left, bottom right)
// Practically this will only be valid if the tile associated with the quadKey is "full", but
// it's up the caller to check this.
func GenerateMinMaxQuadKeysForZoom(quadKey QuadKey, zoom byte) (QuadKey, QuadKey, error) {
	currentZoom := quadKey.Zoom()
	if currentZoom > zoom {
		return 0, 0, errors.New("unable to generate min/max zooms")
	}

	minChild := quadKey
	maxChild := quadKey
	for z := byte(0); z < zoom-currentZoom; z++ {
		minChild, _ = minChild.ChildAtPos(0)
		maxChild, _ = maxChild.ChildAtPos(3)
	}

	if minChild == 0 {
		fmt.Printf("snoop\n")
	}
	return minChild, maxChild, nil
}
