package quadmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestChildAtPos exercises ChildAtPos error paths and the slippy-coord
// relationship between a tile and each of its four children. Pos mapping:
// 0=top-left (2x,2y), 1=top-right (2x+1,2y), 2=bottom-left (2x,2y+1),
// 3=bottom-right (2x+1,2y+1).
func TestChildAtPos(t *testing.T) {
	t.Run("rejects subdivision at MaxZoom", func(t *testing.T) {
		qk, err := GenerateQuadKeyIndexFromSlippy(0, 0, MaxZoom)
		require.NoError(t, err)
		for pos := 0; pos < 4; pos++ {
			_, err := qk.ChildAtPos(pos)
			assert.Errorf(t, err, "pos=%d should error at MaxZoom", pos)
		}
	})

	t.Run("rejects invalid pos", func(t *testing.T) {
		parent, err := GenerateQuadKeyIndexFromSlippy(0, 0, 5)
		require.NoError(t, err)
		for _, badPos := range []int{-1, 4, 99} {
			_, err := parent.ChildAtPos(badPos)
			assert.Errorf(t, err, "pos=%d should be rejected", badPos)
		}
	})

	for _, tc := range []struct {
		name string
		x, y uint32
		z    byte
	}{
		{"shallow", 0, 0, 1},
		{"midmap", 60292, 39326, 16},
		{"near MaxZoom edge", (1 << 23) - 1, (1 << 23) - 1, 23},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent, err := GenerateQuadKeyIndexFromSlippy(tc.x, tc.y, tc.z)
			require.NoError(t, err)

			for _, e := range []struct {
				pos    int
				dx, dy uint32
			}{
				{0, 0, 0},
				{1, 1, 0},
				{2, 0, 1},
				{3, 1, 1},
			} {
				child, err := parent.ChildAtPos(e.pos)
				require.NoErrorf(t, err, "pos=%d", e.pos)
				assert.Equalf(t, tc.z+1, child.Zoom(), "child zoom for pos=%d", e.pos)

				cx, cy, cz := child.SlippyCoords()
				assert.Equalf(t, tc.z+1, cz, "child slippy zoom for pos=%d", e.pos)
				assert.Equalf(t, 2*tc.x+e.dx, cx, "child x for pos=%d", e.pos)
				assert.Equalf(t, 2*tc.y+e.dy, cy, "child y for pos=%d", e.pos)

				back, err := child.Parent()
				require.NoErrorf(t, err, "parent of pos=%d child", e.pos)
				assert.Equalf(t, parent, back, "parent round-trip for pos=%d", e.pos)
			}
		})
	}
}

// TestChildren verifies that Children() returns the same four children as
// ChildAtPos(0..3) in order. Children() is the inlinable hot-path variant of
// ChildAtPos and the two must stay in sync.
func TestChildren(t *testing.T) {
	for _, tc := range []struct {
		name string
		x, y uint32
		z    byte
	}{
		{"shallow", 0, 0, 1},
		{"midmap", 60292, 39326, 16},
		{"just below MaxZoom", (1 << 23) - 1, (1 << 23) - 1, MaxZoom - 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent, err := GenerateQuadKeyIndexFromSlippy(tc.x, tc.y, tc.z)
			require.NoError(t, err)

			kids := parent.Children()
			for pos := 0; pos < 4; pos++ {
				want, err := parent.ChildAtPos(pos)
				require.NoErrorf(t, err, "ChildAtPos(%d)", pos)
				assert.Equalf(t, want, kids[pos],
					"Children()[%d] should match ChildAtPos(%d)", pos, pos)
				assert.Equalf(t, tc.z+1, kids[pos].Zoom(),
					"child zoom for pos=%d", pos)
			}

			seen := make(map[QuadKey]struct{}, 4)
			for _, c := range kids {
				seen[c] = struct{}{}
			}
			assert.Equalf(t, 4, len(seen), "Children() returned duplicates: %v", kids)
		})
	}
}

// TestGetAllAncestorsAndSelf verifies that the returned slice is the chain
// of ancestors from the root (zoom 0) up to q inclusive, in ascending zoom
// order. anc[i] must be the unique ancestor at zoom i.
func TestGetAllAncestorsAndSelf(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		root := QuadKey(0)
		assert.Equal(t, []QuadKey{root}, root.GetAllAncestorsAndSelf())
	})

	t.Run("known fixture", func(t *testing.T) {
		anc := quadKey.GetAllAncestorsAndSelf()
		require.Len(t, anc, int(quadKey.Zoom())+1)
		assert.Equal(t, QuadKey(0), anc[0], "anc[0] must be the root")
		assert.Equal(t, parent, anc[5], "anc[5] must equal the fixture parent")
		assert.Equal(t, quadKey, anc[6], "anc[zoom] must equal q itself")
	})

	for _, tc := range []struct {
		name string
		x, y uint32
		z    byte
	}{
		{"shallow", 0, 0, 1},
		{"midmap", 60292, 39326, 16},
		{"MaxZoom", 12345, 67890, MaxZoom},
	} {
		t.Run(tc.name, func(t *testing.T) {
			qk, err := GenerateQuadKeyIndexFromSlippy(tc.x, tc.y, tc.z)
			require.NoError(t, err)

			anc := qk.GetAllAncestorsAndSelf()
			require.Len(t, anc, int(tc.z)+1, "length should be zoom+1")
			require.Equal(t, QuadKey(0), anc[0], "anc[0] must be the root")
			require.Equal(t, qk, anc[tc.z], "anc[zoom] must be q itself")

			for i, a := range anc {
				assert.Equalf(t, byte(i), a.Zoom(), "anc[%d] zoom", i)
				assert.Truef(t, a.IsAncestorOf(qk),
					"anc[%d] should be an ancestor of q", i)
				if i > 0 {
					back, err := a.Parent()
					require.NoErrorf(t, err, "Parent(anc[%d])", i)
					assert.Equalf(t, anc[i-1], back,
						"Parent(anc[%d]) should equal anc[%d]", i, i-1)
				}
			}
		})
	}
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

// TestEnvelope verifies that Envelope returns the correct lat/lon bounds for
// a QuadKey's slippy tile. Covers the Web Mercator pole limit (~85.0511°) at
// zoom 1 and a known mid-zoom reference (Sydney area at z=16).
func TestEnvelope(t *testing.T) {
	const mercatorLat = 85.05112877980659
	for _, tc := range []struct {
		name           string
		x, y           uint32
		z              byte
		minLon, minLat float64
		maxLon, maxLat float64
	}{
		{"z=1 NW quadrant", 0, 0, 1, -180, 0, 0, mercatorLat},
		{"z=1 NE quadrant", 1, 0, 1, 0, 0, 180, mercatorLat},
		{"z=1 SW quadrant", 0, 1, 1, -180, -mercatorLat, 0, 0},
		{"z=1 SE quadrant", 1, 1, 1, 0, -mercatorLat, 180, 0},
		{
			name: "Sydney z=16",
			x:    60292, y: 39326, z: 16,
			minLon: 151.19384765625, minLat: -33.86585445407186,
			maxLon: 151.1993408203125, maxLat: -33.861293113515515,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			qk, err := GenerateQuadKeyIndexFromSlippy(tc.x, tc.y, tc.z)
			require.NoError(t, err)

			env, err := qk.Envelope()
			require.NoError(t, err)

			min, max, ok := env.MinMaxXYs()
			require.True(t, ok, "envelope should not be empty")
			assert.InDelta(t, tc.minLon, min.X, 1e-9, "min lon")
			assert.InDelta(t, tc.minLat, min.Y, 1e-9, "min lat")
			assert.InDelta(t, tc.maxLon, max.X, 1e-9, "max lon")
			assert.InDelta(t, tc.maxLat, max.Y, 1e-9, "max lat")
		})
	}
}

// TestRange verifies that QuadKey.Range produces a numeric interval covering
// q and every descendant in [Start, End]. Start and End themselves aren't
// valid QuadKeys (the low bits — including the zoom field — are zeroed/set);
// they're bounds for numeric filtering, not lookup keys.
func TestRange(t *testing.T) {
	t.Run("root covers full key space", func(t *testing.T) {
		r := QuadKey(0).Range()
		assert.Equal(t, uint64(0), r.Start)
		assert.Equal(t, uint64(0xFFFFFFFFFFFFFFFF), r.End)
	})

	t.Run("contains q itself", func(t *testing.T) {
		qk := mustQK(t, 123, 456, 10)
		assert.True(t, qk.Range().Contains(qk))
	})

	t.Run("contains all 4 direct children", func(t *testing.T) {
		qk := mustQK(t, 5, 5, 5)
		r := qk.Range()
		for i, child := range qk.Children() {
			assert.Truef(t, r.Contains(child),
				"child pos %d (%#x) should be inside range [%#x..%#x]",
				i, uint64(child), r.Start, r.End)
		}
	})

	t.Run("contains all descendants at deeper zoom", func(t *testing.T) {
		qk := mustQK(t, 5, 5, 5)
		r := qk.Range()
		for _, d := range qk.GetAllPossibleChildrenAtZoom(8) {
			assert.Truef(t, r.Contains(d), "descendant %#x should be in range", uint64(d))
		}
	})

	t.Run("does not contain a sibling tile", func(t *testing.T) {
		qk := mustQK(t, 5, 5, 5)
		sibling := mustQK(t, 6, 5, 5)
		assert.False(t, qk.Range().Contains(sibling))
	})

	t.Run("width matches the 64-2z low bits", func(t *testing.T) {
		// End - Start should equal (1<<(64-2z)) - 1 — the low-bit mask.
		for _, z := range []byte{1, 5, 10, 16, 23} {
			qk := mustQK(t, 1, 1, z)
			r := qk.Range()
			want := (^uint64(0)) >> (uint(z) * 2)
			assert.Equalf(t, want, r.End-r.Start, "zoom %d range width", z)
		}
	})

	t.Run("Start and End align on the 2z-bit boundary", func(t *testing.T) {
		for _, z := range []byte{1, 5, 16} {
			qk := mustQK(t, 1, 1, z)
			r := qk.Range()
			mask := (^uint64(0)) >> (uint(z) * 2)
			assert.Equalf(t, uint64(0), r.Start&mask, "zoom %d: Start low bits should be 0", z)
			assert.Equalf(t, mask, r.End&mask, "zoom %d: End low bits should be all set", z)
			assert.Equalf(t, r.Start&^mask, r.End&^mask, "zoom %d: Start/End share top 2z bits", z)
		}
	})
}

// TestGetAllPossibleChildrenAtZoom verifies the enumeration of all descendant
// QuadKeys at a target zoom level. The bit-enumeration algorithm should
// produce the same sequence as recursively walking Children(), so the test
// cross-checks against Children()/ChildAtPos for shallow deltas and against
// the descendant slippy-coord range for deeper deltas.
func TestGetAllPossibleChildrenAtZoom(t *testing.T) {
	sydney, err := GenerateQuadKeyIndexFromSlippy(60292, 39326, 16)
	require.NoError(t, err)

	t.Run("at maxZoom returns self only", func(t *testing.T) {
		assert.Equal(t, []QuadKey{sydney}, sydney.GetAllPossibleChildrenAtZoom(16))
	})

	t.Run("below q.Zoom returns self only", func(t *testing.T) {
		assert.Equal(t, []QuadKey{sydney}, sydney.GetAllPossibleChildrenAtZoom(10))
	})

	t.Run("delta 1 matches Children()", func(t *testing.T) {
		kids := sydney.Children()
		all := sydney.GetAllPossibleChildrenAtZoom(17)
		require.Len(t, all, 4)
		assert.Equal(t, kids[:], all)
	})

	t.Run("delta 2 grouped by ChildAtPos", func(t *testing.T) {
		all := sydney.GetAllPossibleChildrenAtZoom(18)
		require.Len(t, all, 16)
		for pos := 0; pos < 4; pos++ {
			child, err := sydney.ChildAtPos(pos)
			require.NoErrorf(t, err, "ChildAtPos(%d)", pos)
			sub := child.GetAllPossibleChildrenAtZoom(18)
			require.Lenf(t, sub, 4, "pos %d sub-block length", pos)
			assert.Equalf(t, sub, all[pos*4:(pos+1)*4],
				"block %d should equal recursive ChildAtPos(%d) enumeration", pos, pos)
		}
	})

	for _, tc := range []struct {
		name       string
		x, y       uint32
		z, maxZoom byte
	}{
		{"z 1 -> 4", 0, 0, 1, 4},
		{"z 16 -> 18", 60292, 39326, 16, 18},
	} {
		t.Run(tc.name, func(t *testing.T) {
			qk, err := GenerateQuadKeyIndexFromSlippy(tc.x, tc.y, tc.z)
			require.NoError(t, err)

			delta := tc.maxZoom - tc.z
			wantCount := 1 << (2 * delta)

			all := qk.GetAllPossibleChildrenAtZoom(tc.maxZoom)
			require.Lenf(t, all, wantCount, "count should be 4^delta = %d", wantCount)

			seen := make(map[QuadKey]struct{}, wantCount)
			xLow, xHigh := tc.x<<delta, (tc.x+1)<<delta
			yLow, yHigh := tc.y<<delta, (tc.y+1)<<delta
			for i, c := range all {
				assert.Equalf(t, tc.maxZoom, c.Zoom(), "result[%d] zoom", i)
				assert.Truef(t, qk.IsAncestorOf(c),
					"result[%d]=%#x should be a descendant of q", i, uint64(c))
				seen[c] = struct{}{}

				cx, cy, _ := c.SlippyCoords()
				assert.Truef(t, xLow <= cx && cx < xHigh,
					"result[%d] x=%d outside [%d,%d)", i, cx, xLow, xHigh)
				assert.Truef(t, yLow <= cy && cy < yHigh,
					"result[%d] y=%d outside [%d,%d)", i, cy, yLow, yHigh)
			}
			assert.Equalf(t, wantCount, len(seen), "all results should be unique")
		})
	}
}
