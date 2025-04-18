package quadtree

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/peterstace/simplefeatures/geom"
)

// QuadKey is a key representing a Slippy tile.
//		|63-----------------16|15----------------------------5|4----------0|
//		| Identify Tile       |         Unused                | Zoom level |
//		|                     |                               |   00001    |
//
// Need to handle 24 levels of zoom... so 48 bits. Leaving 16 bits.
// Need 5 bits for zoom level..

type QuadKey uint64

const (
	MaxZoom = 24
	MinZoom = 1

	// Zoom level is the bottom 5 bits
	zoomMask = 0b11111
)

// Parent get parents quadkey for passed quadkey
func (q QuadKey) Parent() (QuadKey, error) {
	zoomLevel := q.Zoom()
	if zoomLevel <= 0 {
		return 0, errors.New("no parent")
	}
	parentZoomLevel := zoomLevel - 1

	shift := 64 - (parentZoomLevel * 2)
	parent := q >> shift
	parent = parent << shift
	parent |= QuadKey(parentZoomLevel)

	return parent, nil
}

// ChildAtPos where pos is 0-3
// based off https://learn.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system?redirectedfrom=MSDN
func (q QuadKey) ChildAtPos(pos int) (QuadKey, error) {
	zoomLevel := q.Zoom()
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

	q |= QuadKey(zoomLevel) + 1
	return q, nil
}

// IsAncestorOf checks whether a QuadKey is an ancestor of (or equal to)
// another QuadKey.
func (q QuadKey) IsAncestorOf(desc QuadKey) bool {
	return q.Zoom() <= desc.Zoom() && q.Range().Contains(desc)
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
// If zoom level is < MinZoomLevel or > MaxZoomLevel return error.
// Only generate/care about bits 63 -> 32..
// Although inefficient, we generate entire quadkey (zoom 1 -> zoomLevel) but then
// shift left to trim off the first 9 zoom levels. This is because I might want
// to repurpose the generic quad key generation for later.
func GenerateQuadKeyIndexFromSlippy(x uint32, y uint32, zoomLevel byte) (QuadKey, error) {

	if zoomLevel < MinZoom || zoomLevel > MaxZoom {
		return 0, errors.New("invalid zoom level")
	}
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
	return binaryQuadkey, nil
}

// SlippyCoords generates the slippy coords from quadkey index
// Needs to take into account that quadkey starts at zoom level 10.
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
// Zoom is stored in lower 5 bits of quadkey
func (q QuadKey) Zoom() byte {
	zoomLevel := byte(q & zoomMask)
	return zoomLevel
}

// Envelope returns the lat/lon bounds of the slippy tile represented by a QuadKey.
func (q QuadKey) Envelope() (geom.Envelope, error) {
	x, y, z := q.SlippyCoords()
	return geom.NewEnvelope([]geom.XY{
		SlippyTopLeftToLonLat(x, y, z),
		SlippyTopLeftToLonLat(x+1, y+1, z),
	}...), nil
}

// From https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Tile_numbers_to_lon./lat.
func SlippyTopLeftToLonLat(x, y uint32, z byte) geom.XY {
	n := float64(uint64(1) << z)
	lonDeg := float64(x)/n*360 - 180
	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
	latDeg := latRad * 180 / math.Pi
	return geom.XY{X: lonDeg, Y: latDeg}
}

// GetMinMaxEquivForZoomLevel given a quadkey and a desired zoom level, keep converting
// quadkey to desired zoom level and get min/max quadkeys (top left, bottom right)
// Practically this will only be valid if the tile associated with the quadKey is "full", but
// it's up the caller to check this.
// This name utterly sucks, please suggest a better one.
func (q QuadKey) GetMinMaxEquivForZoomLevel(zoom byte) (QuadKey, QuadKey, error) {
	currentZoom := q.Zoom()
	if currentZoom > zoom {
		return 0, 0, errors.New("unable to generate min/max zooms")
	}

	minChild := q
	maxChild := q
	for z := byte(0); z < zoom-currentZoom; z++ {
		minChild, _ = minChild.ChildAtPos(0)
		maxChild, _ = maxChild.ChildAtPos(3)
	}
	return minChild, maxChild, nil
}

// GetAllAncestorsAndSelf returns all ancestors of given QuadKey
// including the QuadKey itself
func (q QuadKey) GetAllAncestorsAndSelf() []QuadKey {
	var ancestors = []QuadKey{q}
	for {
		parent, err := q.Parent()
		if err != nil {
			break
		}
		ancestors = append(ancestors, parent)
		q = parent
	}

	// reverse list so that it's in order of zoom level.
	slices.Reverse(ancestors)
	return ancestors
}

// GetAllChildrenAndSelf returns all children of given QuadKey at given zoom level
func (q QuadKey) GetAllChildrenAtZoom(maxZoom byte) []QuadKey {
	var allQuadKeys []QuadKey
	if q.Zoom() >= maxZoom {
		return []QuadKey{q}
	}

	children := q.Children()
	for _, child := range children {
		quadKeys := child.GetAllChildrenAtZoom(maxZoom)
		allQuadKeys = append(allQuadKeys, quadKeys...)
	}

	return allQuadKeys
}
