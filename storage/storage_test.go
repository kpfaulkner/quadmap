package storage

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/kpfaulkner/quadmap/quadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewStorage(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func mustQK(t *testing.T, x, y uint32, z byte) quadmap.QuadKey {
	t.Helper()
	qk, err := quadmap.GenerateQuadKeyIndexFromSlippy(x, y, z)
	require.NoError(t, err)
	return qk
}

// TestNewStorage_CreatesSchema verifies that NewStorage creates the details
// and processed tables and that the WAL pragma sticks.
func TestNewStorage_CreatesSchema(t *testing.T) {
	s := newTestStorage(t)

	var names []string
	err := s.db.Select(&names, `SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`)
	require.NoError(t, err)
	assert.Contains(t, names, "details")
	assert.Contains(t, names, "processed")

	var journalMode string
	err = s.db.Get(&journalMode, `PRAGMA journal_mode`)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode)
}

// TestGenerateTableName covers the partitioning rules.
func TestGenerateTableName(t *testing.T) {
	s := &Storage{}

	t.Run("zoom below partition zoom yields quadmap_high", func(t *testing.T) {
		assert.Equal(t, "quadmap_high", s.GenerateTableName(mustQK(t, 1, 1, 5)))
	})

	t.Run("zoom at partition zoom uses own value", func(t *testing.T) {
		qk := mustQK(t, 123, 456, TablePartitionZoomLevel)
		assert.Equal(t, fmt.Sprintf("quadmap_%d", qk), s.GenerateTableName(qk))
	})

	t.Run("zoom above partition zoom uses ancestor", func(t *testing.T) {
		qk := mustQK(t, 500, 500, 12)
		ancestor := qk
		for ancestor.Zoom() > TablePartitionZoomLevel {
			ancestor = ancestor.ParentUnchecked()
		}
		assert.Equal(t, fmt.Sprintf("quadmap_%d", ancestor), s.GenerateTableName(qk))
	})

	t.Run("siblings under same ancestor share a table", func(t *testing.T) {
		ancestor := mustQK(t, 50, 50, TablePartitionZoomLevel)
		c1 := ancestor.Children()[0]
		c2 := ancestor.Children()[3]
		assert.Equal(t, s.GenerateTableName(c1), s.GenerateTableName(c2))
	})
}

// TestGenerateTileTypesQuery covers the unexported helper. For each tile type
// the helper emits two values: the presence bit (tile<<TileTypeOffset) and
// presence-or-full (presence | tile). These are the values stored in
// details_mask, so the IN-list matches both full and non-full tiles.
func TestGenerateTileTypesQuery(t *testing.T) {
	t.Run("empty list yields empty string", func(t *testing.T) {
		assert.Empty(t, generateTileTypesQuery(nil))
	})

	t.Run("single type emits presence-only and presence|self", func(t *testing.T) {
		// TileTypeVert (1) with TileTypeOffset (10): 1<<10=1024, |1=1025.
		assert.Equal(t, "1024 , 1025",
			generateTileTypesQuery([]quadmap.TileType{quadmap.TileTypeVert}))
	})

	t.Run("multiple types emit pairs joined by commas", func(t *testing.T) {
		// Vert=1 → 1024,1025; North=4 → 4096,4100.
		assert.Equal(t, "1024 , 1025 , 4096 , 4100",
			generateTileTypesQuery([]quadmap.TileType{quadmap.TileTypeVert, quadmap.TileTypeNorth}))
	})
}

// TestBeginAndCommitTxx exercises the transaction wrappers and verifies that
// committed writes are visible afterwards.
func TestBeginAndCommitTxx(t *testing.T) {
	s := newTestStorage(t)
	tx, err := s.BeginTxx()
	require.NoError(t, err)
	require.NotNil(t, tx)
	// Trivial write inside the tx — visible after commit.
	tx.MustExec(`INSERT INTO processed (identifier, tiletype) VALUES ($1, $2)`, "tx-test", 1)
	require.NoError(t, s.CommitTxx(tx))

	var ids []string
	require.NoError(t, s.db.Select(&ids, `SELECT identifier FROM processed`))
	assert.Equal(t, []string{"tx-test"}, ids)
}

// TestCreatePartitionTableIfNotExist creates a partition table & index, then
// re-creates them (idempotent path).
func TestCreatePartitionTableIfNotExist(t *testing.T) {
	s := newTestStorage(t)
	const tableName = "quadmap_42"

	tx, err := s.BeginTxx()
	require.NoError(t, err)
	s.CreatePartitionTableIfNotExist(tx, tableName)
	require.NoError(t, s.CommitTxx(tx))

	var tables []string
	require.NoError(t, s.db.Select(&tables,
		`SELECT name FROM sqlite_master WHERE type='table' AND name=$1`, tableName))
	require.Len(t, tables, 1)

	var idx []string
	require.NoError(t, s.db.Select(&idx,
		`SELECT name FROM sqlite_master WHERE type='index' AND name=$1`, tableName+"_index"))
	require.Len(t, idx, 1)

	// Idempotent: a second call must not error.
	tx, err = s.BeginTxx()
	require.NoError(t, err)
	s.CreatePartitionTableIfNotExist(tx, tableName)
	require.NoError(t, s.CommitTxx(tx))
}

// TestInsertTileWithTableName inserts a row into a manually-named partition
// table and reads it back via raw SQL.
func TestInsertTileWithTableName(t *testing.T) {
	s := newTestStorage(t)
	qk := mustQK(t, 50, 50, TablePartitionZoomLevel)
	tableName := s.GenerateTableName(qk)

	tx, err := s.BeginTxx()
	require.NoError(t, err)
	s.CreatePartitionTableIfNotExist(tx, tableName)
	require.NoError(t, s.InsertTileWithTableName(tx, tableName, TileEntity{
		QuadKey: qk, DetailsMask: 0xCAFE, DetailsID: 42,
	}))
	require.NoError(t, s.CommitTxx(tx))

	var rows []TileEntity
	require.NoError(t, s.db.Select(&rows,
		fmt.Sprintf(`SELECT quadkey, details_mask, details_id FROM %s`, tableName)))
	require.Len(t, rows, 1)
	assert.Equal(t, qk, rows[0].QuadKey)
	assert.Equal(t, uint64(0xCAFE), rows[0].DetailsMask)
	assert.Equal(t, int64(42), rows[0].DetailsID)
}

// TestInsertTileWith picks the table name automatically.
func TestInsertTileWith(t *testing.T) {
	s := newTestStorage(t)
	qk := mustQK(t, 50, 50, TablePartitionZoomLevel)
	tableName := s.GenerateTableName(qk)

	tx, err := s.BeginTxx()
	require.NoError(t, err)
	s.CreatePartitionTableIfNotExist(tx, tableName)
	require.NoError(t, s.InsertTileWith(tx, TileEntity{
		QuadKey: qk, DetailsMask: 0xBEEF, DetailsID: 7,
	}))
	require.NoError(t, s.CommitTxx(tx))

	var got TileEntity
	require.NoError(t, s.db.Get(&got,
		fmt.Sprintf(`SELECT quadkey, details_mask, details_id FROM %s WHERE quadkey = $1`, tableName),
		int64(qk)))
	assert.Equal(t, uint64(0xBEEF), got.DetailsMask)
	assert.Equal(t, int64(7), got.DetailsID)
}

// TestInsertDetailsAndGetAllDetails confirms InsertDetails returns the new
// rowid and that GetAllDetails surfaces enabled rows.
func TestInsertDetailsAndGetAllDetails(t *testing.T) {
	s := newTestStorage(t)
	id, err := s.InsertDetails(DetailsEntity{
		Border:          "border",
		SimpleBorder:    "simple",
		TileType:        1,
		DateTime:        100,
		Scale:           5,
		Identifier:      "id-1",
		SimpleBorderWKB: []byte{0x01, 0x02, 0x03},
	})
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	all, err := s.GetAllDetails()
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, uint64(id), all[0].Id)
	assert.Equal(t, "id-1", all[0].Identifier)
	assert.Equal(t, uint16(5), all[0].Scale)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, all[0].SimpleBorderWKB)
}

// TestUpdateDetails updates simple_border_wkb on an existing row.
func TestUpdateDetails(t *testing.T) {
	s := newTestStorage(t)
	id, err := s.InsertDetails(DetailsEntity{
		Border: "b", Identifier: "x", SimpleBorderWKB: []byte{0xFF},
	})
	require.NoError(t, err)

	require.NoError(t, s.UpdateDetails(DetailsEntity{
		Id: uint64(id), SimpleBorderWKB: []byte{0xAA, 0xBB},
	}))

	all, err := s.GetAllDetails()
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, []byte{0xAA, 0xBB}, all[0].SimpleBorderWKB)
}

// TestGetDetails_BugReturnsEmpty documents a current bug: GetDetails uses
// sqlx.Select with a struct destination (Select requires a slice), so the
// query errors silently and the function returns an empty entity.
//
// TODO(storage): switch GetDetails to db.Get and propagate the error.
func TestGetDetails_BugReturnsEmpty(t *testing.T) {
	s := newTestStorage(t)
	id, err := s.InsertDetails(DetailsEntity{Border: "b", Identifier: "x"})
	require.NoError(t, err)

	got, err := s.GetDetails(int(id))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, DetailsEntity{}, *got,
		"Select-on-struct misuse leaves entity zero-valued; locks in current bug")
}

// TestGetTile_BugReturnsEmpty documents the same Select-vs-Get bug for the
// tile path. The SQL also references the non-existent `quadmap` table.
//
// TODO(storage): switch GetTile to db.Get and either create the `quadmap`
// table or use the partition table from GenerateTableName.
func TestGetTile_BugReturnsEmpty(t *testing.T) {
	s := newTestStorage(t)
	got, err := s.GetTile(mustQK(t, 50, 50, TablePartitionZoomLevel))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, TileEntity{}, *got)
}

// TestSearchDetailsBetweenQuadKeys verifies the range scan: only details
// reachable through tiles in [qk1, qk2) with a matching details_mask are
// returned.
func TestSearchDetailsBetweenQuadKeys(t *testing.T) {
	s := newTestStorage(t)
	qk1 := mustQK(t, 50, 50, TablePartitionZoomLevel)
	qk2 := mustQK(t, 51, 50, TablePartitionZoomLevel)
	tableName := s.GenerateTableName(qk1)

	tx, err := s.BeginTxx()
	require.NoError(t, err)
	s.CreatePartitionTableIfNotExist(tx, tableName)
	require.NoError(t, s.CommitTxx(tx))

	id, err := s.InsertDetails(DetailsEntity{
		Border: "b", Identifier: "id1", TileType: 1, Scale: 1,
		SimpleBorderWKB: []byte("wkb"),
	})
	require.NoError(t, err)

	tx, err = s.BeginTxx()
	require.NoError(t, err)
	mask := uint64(quadmap.TileTypeVert) << quadmap.TileTypeOffset
	require.NoError(t, s.InsertTileWithTableName(tx, tableName, TileEntity{
		QuadKey: qk1, DetailsMask: mask, DetailsID: id,
	}))
	require.NoError(t, s.CommitTxx(tx))

	t.Run("with simple border", func(t *testing.T) {
		rows, err := s.SearchDetailsBetweenQuadKeys(
			qk1, qk2, []quadmap.TileType{quadmap.TileTypeVert}, true, 10)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, uint64(id), rows[0].Id)
		assert.Equal(t, "id1", rows[0].Identifier)
		assert.Equal(t, []byte("wkb"), rows[0].SimpleBorderWKB)
	})

	t.Run("without simple border", func(t *testing.T) {
		rows, err := s.SearchDetailsBetweenQuadKeys(
			qk1, qk2, []quadmap.TileType{quadmap.TileTypeVert}, false, 10)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "id1", rows[0].Identifier)
		assert.Empty(t, rows[0].SimpleBorderWKB,
			"includeSimpleBorder=false should omit the WKB column from SELECT")
	})

	t.Run("non-matching tile type returns empty", func(t *testing.T) {
		rows, err := s.SearchDetailsBetweenQuadKeys(
			qk1, qk2, []quadmap.TileType{quadmap.TileTypeNorth}, true, 10)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})
}

// TestSearchDetailsWithinQuadKey routes through the convenience wrapper that
// computes the eastward sibling as qk2.
func TestSearchDetailsWithinQuadKey(t *testing.T) {
	s := newTestStorage(t)
	qk := mustQK(t, 50, 50, TablePartitionZoomLevel)
	tableName := s.GenerateTableName(qk)

	tx, err := s.BeginTxx()
	require.NoError(t, err)
	s.CreatePartitionTableIfNotExist(tx, tableName)
	require.NoError(t, s.CommitTxx(tx))

	id, err := s.InsertDetails(DetailsEntity{Border: "b", Identifier: "id1", TileType: 1})
	require.NoError(t, err)

	tx, err = s.BeginTxx()
	require.NoError(t, err)
	mask := uint64(quadmap.TileTypeVert) << quadmap.TileTypeOffset
	require.NoError(t, s.InsertTileWithTableName(tx, tableName, TileEntity{
		QuadKey: qk, DetailsMask: mask, DetailsID: id,
	}))
	require.NoError(t, s.CommitTxx(tx))

	rows, err := s.SearchDetailsWithinQuadKey(qk, []quadmap.TileType{quadmap.TileTypeVert}, false, 10)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "id1", rows[0].Identifier)
}

// TestSearchQuadKeysWithinQuadKey checks the DISTINCT-on-details_id behaviour
// for the slippy-bbox range query.
func TestSearchQuadKeysWithinQuadKey(t *testing.T) {
	s := newTestStorage(t)
	qk := mustQK(t, 50, 50, TablePartitionZoomLevel)
	tableName := s.GenerateTableName(qk)

	tx, err := s.BeginTxx()
	require.NoError(t, err)
	s.CreatePartitionTableIfNotExist(tx, tableName)
	// Same QuadKey + same details_id inserted twice — DISTINCT collapses them.
	require.NoError(t, s.InsertTileWithTableName(tx, tableName, TileEntity{
		QuadKey: qk, DetailsID: 7,
	}))
	require.NoError(t, s.InsertTileWithTableName(tx, tableName, TileEntity{
		QuadKey: qk, DetailsID: 7,
	}))
	require.NoError(t, s.CommitTxx(tx))

	ids, err := s.SearchQuadKeysWithinQuadKey(qk)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, int64(7), ids[0])
}

// TestIdentifierLifecycle verifies the (identifier, tileType) presence
// checks: missing before insert, present after, doesn't match on differing
// identifier or tileType.
func TestIdentifierLifecycle(t *testing.T) {
	s := newTestStorage(t)
	assert.False(t, s.HasIdentifier("abc", quadmap.TileTypeVert),
		"should not exist before insertion")

	require.NoError(t, s.InsertIdentifier("abc", quadmap.TileTypeVert))
	assert.True(t, s.HasIdentifier("abc", quadmap.TileTypeVert))

	assert.False(t, s.HasIdentifier("abc", quadmap.TileTypeNorth),
		"different tile type must not match")
	assert.False(t, s.HasIdentifier("xyz", quadmap.TileTypeVert),
		"different identifier must not match")
}
