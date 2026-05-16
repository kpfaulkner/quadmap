package quadmap

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddTile create quadmap and adds tile
func TestAddTile(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(1, 1, 1)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")
}

// TestNumberOfTiles create quadmap and adds tile
func TestNumberOfTiles(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(1, 1, 1)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")

	numTiles := qm.NumberOfTiles()
	assert.EqualValues(t, 1, numTiles, "Should have 1 tile")
}

// TestCreateTileAtSlippyCoords create quadmap and adds tile
func TestCreateTileAtSlippyCoords(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(0, 0, 0)
	assert.Error(t, err, "Should have error when creating tile")

	// quadindex for 1,1,1 is 0b1100000000000000000000000000000000000000000000000000000000000001
	tile, err = qm.CreateTileAtSlippyCoords(1, 1, 1, TileTypeVert, true)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")
	assert.Equal(t, QuadKey(0b1100000000000000000000000000000000000000000000000000000000000001), tile.QuadKey, "QuadKey incorrect")
}

// TestGetTileForSlippyAndTileType create quadmap and adds tile and check if exists
func TestGetTileForSlippyAndTileType(t *testing.T) {
	qm := NewQuadMap(10)
	tile, err := NewTile(1, 1, 1)
	assert.NoError(t, err, "Should not have error when creating tile")
	err = qm.AddTile(tile)
	assert.NoError(t, err, "Should not have error when adding tile")

	// quadindex for 5,5,5 is 0b110011000000000000000000000000000000000000000000000000000101
	tile, err = qm.CreateTileAtSlippyCoords(5, 5, 5, TileTypeNorth, true)
	assert.NoError(t, err, "Should not have error when adding tile")
	assert.NotNil(t, tile, "Should have tile")
	assert.Equal(t, QuadKey(0b110011000000000000000000000000000000000000000000000000000101), tile.QuadKey, "QuadKey incorrect")

	tile, err = qm.GetTileForSlippyAndTileType(5, 5, 5, TileTypeNorth)
	assert.NoError(t, err, "Should not have error when getting tile")

	x, y, z := tile.QuadKey.SlippyCoords()
	assert.EqualValues(t, x, 5)
	assert.EqualValues(t, y, 5)
	assert.EqualValues(t, z, 5)

}

// TestGetAllTiles verifies GetAllTiles returns every tile in the quadmap,
// optionally sorted by QuadKey, and that AddTile on a duplicate QuadKey
// replaces (rather than adds) the existing entry.
func TestGetAllTiles(t *testing.T) {
	t.Run("empty quadmap", func(t *testing.T) {
		qm := NewQuadMap(0)
		tiles, err := qm.GetAllTiles(false)
		require.NoError(t, err)
		assert.Empty(t, tiles)

		tiles, err = qm.GetAllTiles(true)
		require.NoError(t, err)
		assert.Empty(t, tiles)
	})

	coords := []struct {
		x, y uint32
		z    byte
	}{
		{60292, 39326, 16},
		{1, 1, 1},
		{123, 456, 10},
		{5, 5, 5},
	}

	t.Run("returns all added tiles (unsorted)", func(t *testing.T) {
		qm := NewQuadMap(len(coords))
		wantKeys := map[QuadKey]struct{}{}
		for _, c := range coords {
			tile, err := NewTile(c.x, c.y, c.z)
			require.NoError(t, err)
			require.NoError(t, qm.AddTile(tile))
			wantKeys[tile.QuadKey] = struct{}{}
		}

		tiles, err := qm.GetAllTiles(false)
		require.NoError(t, err)
		require.Len(t, tiles, len(coords))

		gotKeys := map[QuadKey]struct{}{}
		for _, tile := range tiles {
			require.NotNil(t, tile)
			gotKeys[tile.QuadKey] = struct{}{}
		}
		assert.Equal(t, wantKeys, gotKeys, "set of returned QuadKeys should match the set added")
	})

	t.Run("sorted ascending by QuadKey", func(t *testing.T) {
		qm := NewQuadMap(len(coords))
		for _, c := range coords {
			tile, err := NewTile(c.x, c.y, c.z)
			require.NoError(t, err)
			require.NoError(t, qm.AddTile(tile))
		}

		tiles, err := qm.GetAllTiles(true)
		require.NoError(t, err)
		require.Len(t, tiles, len(coords))
		for i := 1; i < len(tiles); i++ {
			assert.LessOrEqualf(t, uint64(tiles[i-1].QuadKey), uint64(tiles[i].QuadKey),
				"tiles[%d].QuadKey should be <= tiles[%d].QuadKey", i-1, i)
		}
	})

	t.Run("duplicate AddTile keeps single entry", func(t *testing.T) {
		qm := NewQuadMap(2)
		first, err := NewTile(1, 1, 1)
		require.NoError(t, err)
		second, err := NewTile(1, 1, 1)
		require.NoError(t, err)
		require.NoError(t, qm.AddTile(first))
		require.NoError(t, qm.AddTile(second))

		tiles, err := qm.GetAllTiles(false)
		require.NoError(t, err)
		require.Len(t, tiles, 1, "AddTile on same QuadKey should replace, not duplicate")
		assert.Same(t, second, tiles[0], "the later AddTile call should win")
	})
}

// TestGetParentTile verifies that GetParentTile resolves a tile's parent via
// the quadmap, errors when the parent isn't present, and errors when the
// input tile is at the root (no parent exists).
func TestGetParentTile(t *testing.T) {
	t.Run("returns parent when present", func(t *testing.T) {
		qm := NewQuadMap(2)
		parent, err := NewTile(0, 0, 1)
		require.NoError(t, err)
		require.NoError(t, qm.AddTile(parent))

		child, err := NewTile(0, 0, 2)
		require.NoError(t, err)
		require.NoError(t, qm.AddTile(child))

		got, err := qm.GetParentTile(child)
		require.NoError(t, err)
		assert.Same(t, parent, got, "should return the exact parent pointer stored in the map")
	})

	t.Run("walks up multiple levels via repeated calls", func(t *testing.T) {
		qm := NewQuadMap(3)
		ancestors := []struct {
			x, y uint32
			z    byte
		}{
			{0, 0, 1},
			{0, 0, 2},
			{0, 0, 3},
		}
		tiles := make([]*Tile, len(ancestors))
		for i, a := range ancestors {
			tile, err := NewTile(a.x, a.y, a.z)
			require.NoError(t, err)
			require.NoError(t, qm.AddTile(tile))
			tiles[i] = tile
		}

		// tiles[2] (z=3) → tiles[1] (z=2) → tiles[0] (z=1).
		mid, err := qm.GetParentTile(tiles[2])
		require.NoError(t, err)
		assert.Same(t, tiles[1], mid)

		top, err := qm.GetParentTile(mid)
		require.NoError(t, err)
		assert.Same(t, tiles[0], top)
	})

	t.Run("parent not in map returns error", func(t *testing.T) {
		qm := NewQuadMap(1)
		child, err := NewTile(0, 0, 2)
		require.NoError(t, err)
		require.NoError(t, qm.AddTile(child))

		got, err := qm.GetParentTile(child)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("root tile has no parent", func(t *testing.T) {
		qm := NewQuadMap(1)
		root := NewTileWithQuadKey(QuadKey(0))
		got, err := qm.GetParentTile(root)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

// TestGetSlippyBoundsForTileTypeAndZoom verifies the bounding-box scan:
//   - tiles at the exact target zoom contribute their own slippy coord;
//   - tiles at a shallower zoom that are "full" for the requested tileType
//     contribute their full descendant box at the target zoom;
//   - shallower tiles that are NOT full are skipped;
//   - tiles deeper than the target zoom are skipped;
//   - tiles with a different tileType are skipped.
func TestGetSlippyBoundsForTileTypeAndZoom(t *testing.T) {
	t.Run("empty map yields sentinel bounds", func(t *testing.T) {
		qm := NewQuadMap(0)
		minX, minY, maxX, maxY, err := qm.GetSlippyBoundsForTileTypeAndZoom(TileTypeVert, 5)
		require.NoError(t, err)
		// With no contributors, the initial sentinels (MaxUint32, MaxUint32, 0, 0)
		// are returned unchanged. Documenting current behavior so callers can
		// detect "no matches".
		assert.Equal(t, uint32(math.MaxUint32), minX)
		assert.Equal(t, uint32(math.MaxUint32), minY)
		assert.Equal(t, uint32(0), maxX)
		assert.Equal(t, uint32(0), maxY)
	})

	t.Run("single tile at target zoom", func(t *testing.T) {
		qm := NewQuadMap(1)
		_, err := qm.CreateTileAtSlippyCoords(5, 7, 10, TileTypeVert, true)
		require.NoError(t, err)

		minX, minY, maxX, maxY, err := qm.GetSlippyBoundsForTileTypeAndZoom(TileTypeVert, 10)
		require.NoError(t, err)
		assert.Equal(t, uint32(5), minX)
		assert.Equal(t, uint32(7), minY)
		assert.Equal(t, uint32(5), maxX)
		assert.Equal(t, uint32(7), maxY)
	})

	t.Run("multiple tiles at target zoom span their bbox", func(t *testing.T) {
		qm := NewQuadMap(4)
		for _, c := range []struct{ x, y uint32 }{{3, 4}, {10, 4}, {5, 12}, {7, 2}} {
			_, err := qm.CreateTileAtSlippyCoords(c.x, c.y, 8, TileTypeVert, true)
			require.NoError(t, err)
		}
		minX, minY, maxX, maxY, err := qm.GetSlippyBoundsForTileTypeAndZoom(TileTypeVert, 8)
		require.NoError(t, err)
		assert.Equal(t, uint32(3), minX)
		assert.Equal(t, uint32(2), minY)
		assert.Equal(t, uint32(10), maxX)
		assert.Equal(t, uint32(12), maxY)
	})

	t.Run("full ancestor expands to descendant range", func(t *testing.T) {
		qm := NewQuadMap(1)
		// (1,1,1) full for TileTypeVert. Descendants at zoom 3 occupy
		// x,y in [1*2^2 .. 1*2^2 + 2^2 - 1] = [4, 7].
		_, err := qm.CreateTileAtSlippyCoords(1, 1, 1, TileTypeVert, true)
		require.NoError(t, err)
		minX, minY, maxX, maxY, err := qm.GetSlippyBoundsForTileTypeAndZoom(TileTypeVert, 3)
		require.NoError(t, err)
		assert.Equal(t, uint32(4), minX)
		assert.Equal(t, uint32(4), minY)
		assert.Equal(t, uint32(7), maxX)
		assert.Equal(t, uint32(7), maxY)
	})

	t.Run("non-full ancestor is ignored at deeper target zoom", func(t *testing.T) {
		qm := NewQuadMap(1)
		_, err := qm.CreateTileAtSlippyCoords(1, 1, 1, TileTypeVert, false)
		require.NoError(t, err)
		minX, minY, maxX, maxY, err := qm.GetSlippyBoundsForTileTypeAndZoom(TileTypeVert, 3)
		require.NoError(t, err)
		assert.Equal(t, uint32(math.MaxUint32), minX)
		assert.Equal(t, uint32(math.MaxUint32), minY)
		assert.Equal(t, uint32(0), maxX)
		assert.Equal(t, uint32(0), maxY)
	})

	t.Run("tiles deeper than target zoom are skipped", func(t *testing.T) {
		qm := NewQuadMap(2)
		_, err := qm.CreateTileAtSlippyCoords(5, 5, 5, TileTypeVert, true)
		require.NoError(t, err)
		// Zoom 12 > target zoom 8 → ignored.
		_, err = qm.CreateTileAtSlippyCoords(1000, 1000, 12, TileTypeVert, true)
		require.NoError(t, err)

		minX, minY, maxX, maxY, err := qm.GetSlippyBoundsForTileTypeAndZoom(TileTypeVert, 8)
		require.NoError(t, err)
		// Only the (5,5,5) tile contributes. delta=3 → descendant box
		// [5*8 .. 5*8+7] in both x and y = [40, 47].
		assert.Equal(t, uint32(40), minX)
		assert.Equal(t, uint32(40), minY)
		assert.Equal(t, uint32(47), maxX)
		assert.Equal(t, uint32(47), maxY)
	})

	t.Run("tiles with different tile type are skipped", func(t *testing.T) {
		qm := NewQuadMap(2)
		_, err := qm.CreateTileAtSlippyCoords(5, 5, 5, TileTypeVert, true)
		require.NoError(t, err)
		_, err = qm.CreateTileAtSlippyCoords(20, 20, 5, TileTypeNorth, true)
		require.NoError(t, err)

		minX, minY, maxX, maxY, err := qm.GetSlippyBoundsForTileTypeAndZoom(TileTypeVert, 5)
		require.NoError(t, err)
		assert.Equal(t, uint32(5), minX)
		assert.Equal(t, uint32(5), minY)
		assert.Equal(t, uint32(5), maxX)
		assert.Equal(t, uint32(5), maxY)
	})
}

// TestGetAllChildrenForQuadKeyAndZoomLocked exercises every branch of the
// recursive helper:
//   - base case (qk.Zoom() == zoom) returns [qk];
//   - missing children are skipped;
//   - children present + full for tileType expand via GetAllPossibleChildrenAtZoom;
//   - children present + not-full recurse;
//   - children present without tileType block descent (intermediate gate);
//   - children with a different tileType are skipped.
func TestGetAllChildrenForQuadKeyAndZoomLocked(t *testing.T) {
	t.Run("base case at target zoom returns self", func(t *testing.T) {
		qm := NewQuadMap(0)
		qk, err := GenerateQuadKeyIndexFromSlippy(0, 0, 5)
		require.NoError(t, err)
		got, err := qm.getAllChildrenForQuadKeyAndZoomLocked(qk, TileTypeVert, 5)
		require.NoError(t, err)
		assert.Equal(t, []QuadKey{qk}, got)
	})

	t.Run("no children in map yields empty", func(t *testing.T) {
		qm := NewQuadMap(0)
		qk, err := GenerateQuadKeyIndexFromSlippy(0, 0, 1)
		require.NoError(t, err)
		got, err := qm.getAllChildrenForQuadKeyAndZoomLocked(qk, TileTypeVert, 3)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("full child expands to all descendants at target zoom", func(t *testing.T) {
		qm := NewQuadMap(1)
		parent, err := GenerateQuadKeyIndexFromSlippy(0, 0, 1)
		require.NoError(t, err)

		// pos-0 child (0,0,2), full for TileTypeVert.
		_, err = qm.CreateTileAtSlippyCoords(0, 0, 2, TileTypeVert, true)
		require.NoError(t, err)

		got, err := qm.getAllChildrenForQuadKeyAndZoomLocked(parent, TileTypeVert, 4)
		require.NoError(t, err)
		require.Len(t, got, 16, "delta=2 from (0,0,2) full → 4^2 descendants")

		fullChild, err := GenerateQuadKeyIndexFromSlippy(0, 0, 2)
		require.NoError(t, err)
		for _, k := range got {
			assert.Equal(t, byte(4), k.Zoom())
			assert.True(t, fullChild.IsAncestorOf(k))
		}
	})

	t.Run("non-full child recurses into grandchildren", func(t *testing.T) {
		qm := NewQuadMap(2)
		parent, err := GenerateQuadKeyIndexFromSlippy(0, 0, 1)
		require.NoError(t, err)

		// pos-0 child has tileType but not full → recurse.
		_, err = qm.CreateTileAtSlippyCoords(0, 0, 2, TileTypeVert, false)
		require.NoError(t, err)
		// Grandchild at target zoom, full → expanded yields just itself.
		_, err = qm.CreateTileAtSlippyCoords(1, 1, 3, TileTypeVert, true)
		require.NoError(t, err)

		got, err := qm.getAllChildrenForQuadKeyAndZoomLocked(parent, TileTypeVert, 3)
		require.NoError(t, err)
		want, err := GenerateQuadKeyIndexFromSlippy(1, 1, 3)
		require.NoError(t, err)
		assert.Equal(t, []QuadKey{want}, got)
	})

	t.Run("intermediate child missing tileType blocks recursion", func(t *testing.T) {
		qm := NewQuadMap(2)
		parent, err := GenerateQuadKeyIndexFromSlippy(0, 0, 1)
		require.NoError(t, err)

		// Child has TileTypeNorth (not Vert) — the lookup gate fails so the
		// recursion never reaches descendants, even if they have Vert.
		_, err = qm.CreateTileAtSlippyCoords(0, 0, 2, TileTypeNorth, false)
		require.NoError(t, err)
		_, err = qm.CreateTileAtSlippyCoords(1, 1, 3, TileTypeVert, true)
		require.NoError(t, err)

		got, err := qm.getAllChildrenForQuadKeyAndZoomLocked(parent, TileTypeVert, 3)
		require.NoError(t, err)
		assert.Empty(t, got, "descendant unreachable because intermediate lacks tileType")
	})

	t.Run("mixed full and recursive children", func(t *testing.T) {
		qm := NewQuadMap(4)
		parent, err := GenerateQuadKeyIndexFromSlippy(0, 0, 1)
		require.NoError(t, err)

		// pos-0 child full → 4 descendants at zoom 3.
		_, err = qm.CreateTileAtSlippyCoords(0, 0, 2, TileTypeVert, true)
		require.NoError(t, err)
		// pos-1 child not full → recurse; one grandchild at zoom 3 is full.
		_, err = qm.CreateTileAtSlippyCoords(1, 0, 2, TileTypeVert, false)
		require.NoError(t, err)
		_, err = qm.CreateTileAtSlippyCoords(2, 0, 3, TileTypeVert, true)
		require.NoError(t, err)
		// pos-2 and pos-3 children absent → contribute nothing.

		got, err := qm.getAllChildrenForQuadKeyAndZoomLocked(parent, TileTypeVert, 3)
		require.NoError(t, err)
		assert.Len(t, got, 5, "4 from full pos-0 + 1 from recursed pos-1")

		// Confirm the contributing grandchild is in the result.
		grand, err := GenerateQuadKeyIndexFromSlippy(2, 0, 3)
		require.NoError(t, err)
		assert.Contains(t, got, grand)
	})
}

// TestIsTileCoveredForSlippyCoordsAndTileTypeTopDown verifies the ancestor
// walk: starting from the input slippy coord, walk up to root, returning the
// deepest ancestor that either (a) is the exact-zoom tile (regardless of
// tileType — current contract!) or (b) has the requested tileType AND is
// marked full.
func TestIsTileCoveredForSlippyCoordsAndTileTypeTopDown(t *testing.T) {
	t.Run("invalid zoom returns error", func(t *testing.T) {
		qm := NewQuadMap(0)
		covered, qk, err := qm.IsTileCoveredForSlippyCoordsAndTileTypeTopDown(0, 0, 0, TileTypeVert)
		assert.Error(t, err)
		assert.False(t, covered)
		assert.Equal(t, QuadKey(0), qk)
	})

	t.Run("empty map yields not covered", func(t *testing.T) {
		qm := NewQuadMap(0)
		covered, qk, err := qm.IsTileCoveredForSlippyCoordsAndTileTypeTopDown(5, 5, 5, TileTypeVert)
		require.NoError(t, err)
		assert.False(t, covered)
		assert.Equal(t, QuadKey(0), qk)
	})

	t.Run("exact tile present returns true regardless of tileType", func(t *testing.T) {
		// Locks in current behavior: the exact-zoom branch doesn't consult
		// tileType. A tile added with TileTypeNorth still counts as "covered"
		// when querying for TileTypeVert.
		qm := NewQuadMap(1)
		_, err := qm.CreateTileAtSlippyCoords(5, 5, 5, TileTypeNorth, true)
		require.NoError(t, err)

		want, err := GenerateQuadKeyIndexFromSlippy(5, 5, 5)
		require.NoError(t, err)
		covered, qk, err := qm.IsTileCoveredForSlippyCoordsAndTileTypeTopDown(5, 5, 5, TileTypeVert)
		require.NoError(t, err)
		assert.True(t, covered)
		assert.Equal(t, want, qk)
	})

	t.Run("full ancestor covers descendant", func(t *testing.T) {
		qm := NewQuadMap(1)
		// (0,0,2) full → covers any zoom-5 descendant in its quadrant.
		_, err := qm.CreateTileAtSlippyCoords(0, 0, 2, TileTypeVert, true)
		require.NoError(t, err)

		// (3,2,5) → ancestors at zoom 4 (1,1,4), zoom 3 (0,0,3), zoom 2 (0,0,2).
		want, err := GenerateQuadKeyIndexFromSlippy(0, 0, 2)
		require.NoError(t, err)
		covered, qk, err := qm.IsTileCoveredForSlippyCoordsAndTileTypeTopDown(3, 2, 5, TileTypeVert)
		require.NoError(t, err)
		assert.True(t, covered)
		assert.Equal(t, want, qk, "should surface the covering ancestor's QuadKey")
	})

	t.Run("non-full ancestor does not cover", func(t *testing.T) {
		qm := NewQuadMap(1)
		_, err := qm.CreateTileAtSlippyCoords(0, 0, 2, TileTypeVert, false)
		require.NoError(t, err)

		covered, qk, err := qm.IsTileCoveredForSlippyCoordsAndTileTypeTopDown(3, 2, 5, TileTypeVert)
		require.NoError(t, err)
		assert.False(t, covered)
		assert.Equal(t, QuadKey(0), qk)
	})

	t.Run("ancestor with wrong tileType does not cover", func(t *testing.T) {
		qm := NewQuadMap(1)
		_, err := qm.CreateTileAtSlippyCoords(0, 0, 2, TileTypeNorth, true)
		require.NoError(t, err)

		covered, qk, err := qm.IsTileCoveredForSlippyCoordsAndTileTypeTopDown(3, 2, 5, TileTypeVert)
		require.NoError(t, err)
		assert.False(t, covered)
		assert.Equal(t, QuadKey(0), qk)
	})

	t.Run("returns nearest covering ancestor", func(t *testing.T) {
		// Two full ancestors of (3,2,5): (0,0,1) at zoom 1 and (0,0,3) at
		// zoom 3. The bottom-up walk hits zoom 3 first.
		qm := NewQuadMap(2)
		_, err := qm.CreateTileAtSlippyCoords(0, 0, 1, TileTypeVert, true)
		require.NoError(t, err)
		_, err = qm.CreateTileAtSlippyCoords(0, 0, 3, TileTypeVert, true)
		require.NoError(t, err)

		want, err := GenerateQuadKeyIndexFromSlippy(0, 0, 3)
		require.NoError(t, err)
		covered, qk, err := qm.IsTileCoveredForSlippyCoordsAndTileTypeTopDown(3, 2, 5, TileTypeVert)
		require.NoError(t, err)
		assert.True(t, covered)
		assert.Equal(t, want, qk, "deeper ancestor should be returned")
	})
}

// TestGetTilesForTypeAndZoom verifies the filter: returns every tile at the
// exact target zoom whose Details carry the requested tileType (the full flag
// is not considered — HasTileType, not HasTileTypeAndFull).
func TestGetTilesForTypeAndZoom(t *testing.T) {
	t.Run("empty map yields empty slice", func(t *testing.T) {
		qm := NewQuadMap(0)
		assert.Empty(t, qm.GetTilesForTypeAndZoom(TileTypeVert, 5))
	})

	t.Run("no tiles at target zoom yields empty", func(t *testing.T) {
		qm := NewQuadMap(1)
		_, err := qm.CreateTileAtSlippyCoords(5, 5, 5, TileTypeVert, true)
		require.NoError(t, err)
		assert.Empty(t, qm.GetTilesForTypeAndZoom(TileTypeVert, 8))
	})

	t.Run("matches at zoom regardless of full flag", func(t *testing.T) {
		qm := NewQuadMap(2)
		// One full, one not full — both should be returned.
		_, err := qm.CreateTileAtSlippyCoords(1, 1, 5, TileTypeVert, true)
		require.NoError(t, err)
		_, err = qm.CreateTileAtSlippyCoords(2, 2, 5, TileTypeVert, false)
		require.NoError(t, err)

		tiles := qm.GetTilesForTypeAndZoom(TileTypeVert, 5)
		require.Len(t, tiles, 2)

		gotKeys := map[QuadKey]struct{}{}
		for _, tile := range tiles {
			gotKeys[tile.QuadKey] = struct{}{}
		}
		want1, err := GenerateQuadKeyIndexFromSlippy(1, 1, 5)
		require.NoError(t, err)
		want2, err := GenerateQuadKeyIndexFromSlippy(2, 2, 5)
		require.NoError(t, err)
		assert.Contains(t, gotKeys, want1)
		assert.Contains(t, gotKeys, want2)
	})

	t.Run("filters out wrong tileType and wrong zoom", func(t *testing.T) {
		qm := NewQuadMap(4)
		// Target: (TileTypeVert, zoom 5).
		_, err := qm.CreateTileAtSlippyCoords(1, 1, 5, TileTypeVert, true)
		require.NoError(t, err)
		// Wrong tileType, right zoom — excluded.
		_, err = qm.CreateTileAtSlippyCoords(2, 2, 5, TileTypeNorth, true)
		require.NoError(t, err)
		// Right tileType, wrong zoom — excluded.
		_, err = qm.CreateTileAtSlippyCoords(0, 0, 4, TileTypeVert, true)
		require.NoError(t, err)
		// Right tileType, wrong zoom (deeper) — excluded.
		_, err = qm.CreateTileAtSlippyCoords(10, 10, 6, TileTypeVert, true)
		require.NoError(t, err)

		tiles := qm.GetTilesForTypeAndZoom(TileTypeVert, 5)
		require.Len(t, tiles, 1)
		want, err := GenerateQuadKeyIndexFromSlippy(1, 1, 5)
		require.NoError(t, err)
		assert.Equal(t, want, tiles[0].QuadKey)
	})

	t.Run("tile with multiple tileTypes is returned for each matching type", func(t *testing.T) {
		qm := NewQuadMap(1)
		tile, err := qm.CreateTileAtSlippyCoords(1, 1, 5, TileTypeVert, true)
		require.NoError(t, err)
		tile.AddTileType(TileTypeNorth, false)

		vertTiles := qm.GetTilesForTypeAndZoom(TileTypeVert, 5)
		require.Len(t, vertTiles, 1)
		assert.Same(t, tile, vertTiles[0])

		northTiles := qm.GetTilesForTypeAndZoom(TileTypeNorth, 5)
		require.Len(t, northTiles, 1)
		assert.Same(t, tile, northTiles[0])
	})
}

// TestGetChildInPos verifies that GetChildInPos looks up the child of the
// passed-in tile via ChildAtPos and returns the entry stored in the map.
// Surfaces ChildAtPos errors (invalid pos, MaxZoom parent) and the "not
// found" error when the computed child isn't in the map.
func TestGetChildInPos(t *testing.T) {
	t.Run("returns each child when present", func(t *testing.T) {
		qm := NewQuadMap(5)
		parent, err := qm.CreateTileAtSlippyCoords(0, 0, 1, TileTypeVert, false)
		require.NoError(t, err)

		// pos 0..3 → slippy (0,0), (1,0), (0,1), (1,1) at zoom 2.
		coords := [4]struct{ x, y uint32 }{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
		children := [4]*Tile{}
		for pos, c := range coords {
			children[pos], err = qm.CreateTileAtSlippyCoords(c.x, c.y, 2, TileTypeVert, true)
			require.NoErrorf(t, err, "create child at pos %d", pos)
		}

		for pos := 0; pos < 4; pos++ {
			got, err := qm.GetChildInPos(parent, pos)
			require.NoErrorf(t, err, "GetChildInPos(%d)", pos)
			assert.Samef(t, children[pos], got, "pos %d should return the stored *Tile pointer", pos)
		}
	})

	t.Run("invalid pos propagates ChildAtPos error", func(t *testing.T) {
		qm := NewQuadMap(1)
		parent, err := qm.CreateTileAtSlippyCoords(0, 0, 1, TileTypeVert, false)
		require.NoError(t, err)
		for _, pos := range []int{-1, 4, 99} {
			got, err := qm.GetChildInPos(parent, pos)
			assert.Errorf(t, err, "pos=%d should error", pos)
			assert.Nil(t, got)
		}
	})

	t.Run("MaxZoom parent cannot subdivide", func(t *testing.T) {
		qm := NewQuadMap(1)
		parent, err := qm.CreateTileAtSlippyCoords(0, 0, MaxZoom, TileTypeVert, true)
		require.NoError(t, err)
		got, err := qm.GetChildInPos(parent, 0)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("child not in map returns error", func(t *testing.T) {
		qm := NewQuadMap(1)
		parent, err := qm.CreateTileAtSlippyCoords(0, 0, 1, TileTypeVert, false)
		require.NoError(t, err)
		// No child added.
		got, err := qm.GetChildInPos(parent, 0)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

// TestGetExactTileForSlippy verifies the slippy→tile lookup: an exact match
// returns the stored pointer; a missing tile returns TileNotFoundError; an
// invalid zoom returns the GenerateQuadKeyIndexFromSlippy error.
func TestGetExactTileForSlippy(t *testing.T) {
	t.Run("returns stored pointer for exact match", func(t *testing.T) {
		qm := NewQuadMap(1)
		want, err := qm.CreateTileAtSlippyCoords(60292, 39326, 16, TileTypeVert, true)
		require.NoError(t, err)

		got, err := qm.GetExactTileForSlippy(60292, 39326, 16)
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("missing tile returns TileNotFoundError", func(t *testing.T) {
		qm := NewQuadMap(1)
		_, err := qm.CreateTileAtSlippyCoords(5, 5, 5, TileTypeVert, true)
		require.NoError(t, err)

		got, err := qm.GetExactTileForSlippy(6, 6, 5)
		assert.ErrorIs(t, err, TileNotFoundError)
		assert.Nil(t, got)
	})

	t.Run("does not surface ancestors", func(t *testing.T) {
		// (0,0,2) being present should NOT satisfy a query for its descendant.
		qm := NewQuadMap(1)
		_, err := qm.CreateTileAtSlippyCoords(0, 0, 2, TileTypeVert, true)
		require.NoError(t, err)

		got, err := qm.GetExactTileForSlippy(0, 0, 5)
		assert.ErrorIs(t, err, TileNotFoundError)
		assert.Nil(t, got)
	})

	t.Run("invalid zoom returns error", func(t *testing.T) {
		qm := NewQuadMap(0)
		got, err := qm.GetExactTileForSlippy(0, 0, 0)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

// TestNumberOfTilesForZoom verifies the per-zoom counter only counts tiles at
// the exact target zoom, regardless of tileType or full state.
func TestNumberOfTilesForZoom(t *testing.T) {
	t.Run("empty map returns zero", func(t *testing.T) {
		qm := NewQuadMap(0)
		assert.Equal(t, 0, qm.NumberOfTilesForZoom(5))
	})

	t.Run("counts only the target zoom", func(t *testing.T) {
		qm := NewQuadMap(6)
		// Three at zoom 5, two at zoom 6, one at zoom 8.
		for _, c := range []struct{ x, y uint32 }{{1, 1}, {2, 2}, {3, 3}} {
			_, err := qm.CreateTileAtSlippyCoords(c.x, c.y, 5, TileTypeVert, true)
			require.NoError(t, err)
		}
		for _, c := range []struct{ x, y uint32 }{{4, 4}, {5, 5}} {
			_, err := qm.CreateTileAtSlippyCoords(c.x, c.y, 6, TileTypeVert, true)
			require.NoError(t, err)
		}
		_, err := qm.CreateTileAtSlippyCoords(100, 100, 8, TileTypeVert, true)
		require.NoError(t, err)

		assert.Equal(t, 0, qm.NumberOfTilesForZoom(4))
		assert.Equal(t, 3, qm.NumberOfTilesForZoom(5))
		assert.Equal(t, 2, qm.NumberOfTilesForZoom(6))
		assert.Equal(t, 0, qm.NumberOfTilesForZoom(7))
		assert.Equal(t, 1, qm.NumberOfTilesForZoom(8))
	})

	t.Run("ignores tileType and full flag", func(t *testing.T) {
		qm := NewQuadMap(3)
		_, err := qm.CreateTileAtSlippyCoords(1, 1, 5, TileTypeVert, true)
		require.NoError(t, err)
		_, err = qm.CreateTileAtSlippyCoords(2, 2, 5, TileTypeNorth, false)
		require.NoError(t, err)
		// Tile without a tile type via the lower-level API.
		bare := NewTileWithQuadKey(mustQK(t, 3, 3, 5))
		require.NoError(t, qm.AddTile(bare))

		assert.Equal(t, 3, qm.NumberOfTilesForZoom(5))
	})
}

// mustQK is a tiny test helper for inline QuadKey construction.
func mustQK(t *testing.T, x, y uint32, z byte) QuadKey {
	t.Helper()
	qk, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	require.NoError(t, err)
	return qk
}
