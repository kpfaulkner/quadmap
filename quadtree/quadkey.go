package quadtree

import (
	"errors"
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/geom"
)

// QuadKey is a key representing a Slippy tile.
type QuadKey uint64

const MaxZoom = 29

// Zoom level is the bottom 5 bits
const zoomMask = 0b11111

// Parent get parents quadkey for passed quadkey
func (q QuadKey) Parent() (QuadKey, error) {
	z := q.Zoom()
	if z <= 0 {
		return 0, errors.New("no parent")
	}
	parentZoomLevel := z - 1

	shift := 64 - (parentZoomLevel * 2)
	parent := q >> shift
	parent = parent << shift
	parent |= QuadKey(parentZoomLevel)

	return parent, nil
}

// ChildAtPos where pos is 0-3
// based off https://learn.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system?redirectedfrom=MSDN
func (q QuadKey) ChildAtPos(pos int) (QuadKey, error) {
	zoomLevel := q & zoomMask
	if zoomLevel >= MaxZoom {
		return 0, fmt.Errorf("maximum zoom is %d", MaxZoom)
	}

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
		return 0, fmt.Errorf("invalid pos %d", pos)
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
	// var qk QuadKey
	// for zoomLevel > 0 {
	// 	qk |= QuadKey(x&1) << (63 - 2*(zoomLevel-1))
	// 	qk |= QuadKey(y&1) << (63 - 2*(zoomLevel-1) + 1)
	// 	x >>= 1
	// 	y >>= 1
	// 	zoomLevel--
	// }
	// qk |= QuadKey(zoomLevel)
	// return qk
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
func (q QuadKey) SlippyCoords() (uint32, uint32, byte) {
	var x uint32
	var y uint32

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
	zoomLevel := byte(q & zoomMask)
	return zoomLevel
}

// Envelope returns the lat/lon bounds of the slippy tile represented by a QuadKey.
func (q QuadKey) Envelope() (geom.Envelope, error) {
	x, y, z := q.SlippyCoords()
	return geom.NewEnvelope([]geom.XY{
		slippyTopLeftToLonLat(x, y, z),
		slippyTopLeftToLonLat(x+1, y+1, z),
	})
}

// From https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Tile_numbers_to_lon./lat.
func slippyTopLeftToLonLat(x, y uint32, z byte) geom.XY {
	n := float64(uint64(1) << z)
	lonDeg := float64(x)/n*360 - 180
	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
	latDeg := latRad * 180 / math.Pi
	return geom.XY{X: lonDeg, Y: latDeg}
}
