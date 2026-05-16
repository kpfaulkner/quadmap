package quadmap

import (
	"errors"
	"fmt"
	"math"

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

// ParentUnchecked returns q's parent. Caller MUST ensure q.Zoom() > 0; no
// validation is performed. Provided as an inlinable alternative to Parent()
// for hot loops where the error return blocks inlining and the caller already
// knows the input is non-root.
func (q QuadKey) ParentUnchecked() QuadKey {
	parentZoomLevel := q.Zoom() - 1
	shift := 64 - (parentZoomLevel * 2)
	return ((q >> shift) << shift) | QuadKey(parentZoomLevel)
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
	// Direct prefix comparison instead of building a QuadKeyRange and calling
	// Contains. q is an ancestor of desc iff q's zoom is no deeper than
	// desc's AND the top 2*q.Zoom() bits match. The zoom-guard also rules
	// out the false-positive case noted in QuadKey.Range().
	qz := q.Zoom()
	if qz > desc.Zoom() {
		return false
	}
	shift := 64 - 2*qz
	return uint64(q)>>shift == uint64(desc)>>shift
}

// Children get all the quadkeys for the 4 children of the passed quadkey.
// Callers must ensure q.Zoom() < MaxZoom; no validation is performed.
func (q QuadKey) Children() [4]QuadKey {
	// The parent's next two bits (one level deeper) are guaranteed zero,
	// and the low 5 bits hold zoom Z with no risk of carry for Z < 31,
	// so q+1 both sets the child's zoom bits and leaves the new position
	// bits cleared, ready to be OR-ed in below.
	shift := 62 - 2*(q&zoomMask)
	base := q + 1
	return [4]QuadKey{
		base,
		base | (0b01 << shift),
		base | (0b10 << shift),
		base | (0b11 << shift),
	}
}

// GenerateQuadKeyIndexFromSlippy generates the quadkey index from slippy coords.
// Returns an error if zoomLevel is outside [MinZoom, MaxZoom].
// The tile portion is a Morton (Z-order) interleave of y (odd bit positions)
// and x (even bit positions), left-aligned into bits 63..(64-2z).
func GenerateQuadKeyIndexFromSlippy(x uint32, y uint32, zoomLevel byte) (QuadKey, error) {
	if zoomLevel < MinZoom || zoomLevel > MaxZoom {
		return 0, errors.New("invalid zoom level")
	}
	morton := spreadBits(x) | (spreadBits(y) << 1)
	return QuadKey(morton<<(64-2*zoomLevel)) | QuadKey(zoomLevel), nil
}

// SlippyCoords returns the (x, y, z) slippy coords for the quadkey.
func (q QuadKey) SlippyCoords() (uint32, uint32, byte) {
	zoomLevel := q.Zoom()
	morton := uint64(q) >> (64 - 2*zoomLevel)
	return compactBits(morton), compactBits(morton >> 1), zoomLevel
}

// spreadBits spreads the low 32 bits of v into the even bit positions
// (0, 2, 4, ..., 62) of the result. Inverse of compactBits.
func spreadBits(v uint32) uint64 {
	x := uint64(v)
	x = (x | (x << 16)) & 0x0000FFFF0000FFFF
	x = (x | (x << 8)) & 0x00FF00FF00FF00FF
	x = (x | (x << 4)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x << 2)) & 0x3333333333333333
	x = (x | (x << 1)) & 0x5555555555555555
	return x
}

// compactBits gathers bits at even positions (0, 2, 4, ..., 62) into the low
// 32 bits of the result, discarding bits at odd positions.
func compactBits(x uint64) uint32 {
	x &= 0x5555555555555555
	x = (x | (x >> 1)) & 0x3333333333333333
	x = (x | (x >> 2)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x >> 4)) & 0x00FF00FF00FF00FF
	x = (x | (x >> 8)) & 0x0000FFFF0000FFFF
	x = (x | (x >> 16)) & 0x00000000FFFFFFFF
	return uint32(x)
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
	// Compute n once and call NewEnvelope with explicit args. The previous
	// `[]geom.XY{...}...` form forced the literal slice to escape; passing
	// two args lets the compiler stack-allocate the variadic slice.
	n := float64(uint64(1) << z)
	tl := slippyToLonLatN(x, y, n)
	br := slippyToLonLatN(x+1, y+1, n)
	return geom.NewEnvelope(tl, br), nil
}

// SlippyTopLeftToLonLat converts a slippy (x, y, z) tile coord to its top-left
// lon/lat in degrees.
// From https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Tile_numbers_to_lon./lat.
func SlippyTopLeftToLonLat(x, y uint32, z byte) geom.XY {
	return slippyToLonLatN(x, y, float64(uint64(1)<<z))
}

func slippyToLonLatN(x, y uint32, n float64) geom.XY {
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
	if zoom > MaxZoom {
		return 0, 0, errors.New("invalid zoom level")
	}

	// The tile bits of q occupy positions [64-2*currentZoom, 63]. Descendants
	// add 2 position bits per level below currentZoom. ChildAtPos(0) appends
	// 0b00 each level, so the min descendant keeps q's tile bits unchanged and
	// only swaps in the new zoom. ChildAtPos(3) appends 0b11 each level, so
	// the max descendant additionally sets all 2*delta position bits just
	// below q's tile bits.
	stripZoom := uint64(q) &^ uint64(zoomMask)
	delta := zoom - currentZoom
	var addMask uint64
	if delta > 0 {
		addMask = ((uint64(1) << (2 * delta)) - 1) << (64 - 2*zoom)
	}
	minChild := QuadKey(stripZoom | uint64(zoom))
	maxChild := QuadKey(stripZoom | addMask | uint64(zoom))
	return minChild, maxChild, nil
}

// GetAllAncestorsAndSelf returns all ancestors of given QuadKey
// including the QuadKey itself, ordered by ascending zoom level
// (root first, q last).
func (q QuadKey) GetAllAncestorsAndSelf() []QuadKey {
	zoom := q.Zoom()
	ancestors := make([]QuadKey, zoom+1)
	// Ancestor at zoom z keeps the top 2*z tile bits of q. Build directly
	// from q's bits instead of stepping through Parent() — each Parent()
	// call re-reads zoom and does its own shift pair, so the loop is O(z)
	// avoidable work per level. ancestors[0] is the root (zoom 0, no tile
	// bits), already zero-initialised.
	tileBits := uint64(q) &^ uint64(zoomMask)
	for z := byte(1); z <= zoom; z++ {
		mask := (^uint64(0)) << (64 - 2*z)
		ancestors[z] = QuadKey(tileBits&mask) | QuadKey(z)
	}
	return ancestors
}

// GetAllPossibleChildrenAtZoom returns all children QuadKeys of given QuadKey at given zoom level
// This returns all children QuadKeys even if the child itself doesn't exist. (ie doesn't check full flag)
// Results are in Morton (recursive child-order) traversal order.
func (q QuadKey) GetAllPossibleChildrenAtZoom(maxZoom byte) []QuadKey {
	qz := q.Zoom()
	if qz >= maxZoom {
		return []QuadKey{q}
	}

	// A descendant of q at maxZoom shares q's tile bits and adds 2*delta
	// position bits in positions [64-2*maxZoom, 64-2*qz - 1]. Enumerating
	// those bits from 0..4^delta-1 yields the same order as the recursive
	// Children()-based walk, so we can fill the slice in one pass.
	delta := maxZoom - qz
	count := uint64(1) << (2 * delta)
	stripZoom := uint64(q) &^ uint64(zoomMask)
	shift := 64 - 2*uint64(maxZoom)
	tz := uint64(maxZoom)

	result := make([]QuadKey, count)
	for i := uint64(0); i < count; i++ {
		result[i] = QuadKey(stripZoom | (i << shift) | tz)
	}
	return result
}
