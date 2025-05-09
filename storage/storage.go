package storage

import (
	"context"
	"fmt"
	"sync"
	"strings"
	_ "modernc.org/sqlite"

	"github.com/jmoiron/sqlx"
	"github.com/kpfaulkner/quadmap/quadtree"
	log "github.com/sirupsen/logrus"
)

const (
	TablePartitionZoomLevel = 10
)

type Storage struct {
	db     *sqlx.DB
	dbLock sync.Mutex
}

func NewStorage(dbName string) (*Storage, error) {

	db, err := sqlx.Connect("sqlite", dbName)

	if err != nil {
		return nil, err
	}

	// table
	//db.MustExec(`create table if not exists quadmap (id integer primary key, quadkey integer , details_mask integer, details_id integer)`)
	db.MustExec(`create table if not exists details (id integer primary key, border varchar(500000),simple_border varchar(500000), simple_border_wkb blob, tiletype integer, datetime integer, scale integer, identifier varchar(50), enabled bool)`)
	db.MustExec(`create table if not exists processed (id integer primary key, identifier varchar(50))`)
	//db.MustExec(`create index if not exists quadmap_index on quadmap(quadkey)`)
	db.MustExec(`create index if not exists details_index on details(id)`)

	_, err = db.Exec(`PRAGMA cache_size = -1000000`)
	if err != nil {
		log.Errorf("error setting cache size %s", err)
		return nil, err
	}

	_, err = db.Exec(`PRAGMA temp_store = MEMORY`)
	if err != nil {
		log.Errorf("error setting cache size %s", err)
		return nil, err
	}

	s := &Storage{
		db: db,
	}
	return s, nil
}

func (s *Storage) CreatePartitionTableIfNotExist(txx *sqlx.Tx, tableName string) {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	statement := fmt.Sprintf("create table if not exists %s (id integer primary key, quadkey integer , details_mask integer, details_id integer)", tableName)
	txx.MustExec(statement)

	indexName := fmt.Sprintf("%s_index", tableName)
	statement = fmt.Sprintf("create index if not exists %s on %s(quadkey)", indexName, tableName)
	//s.db.MustExec(statement)
	txx.MustExec(statement)
}

// GenerateTableName generates the table name that should be associated with the provided quadkey.
// Currently the assumption is that the table name will be associated with an ancestor of the quadkey at level
// 10. This is an attempt to find a sweet spot between performance and the number of tables.
// If the provided quadkey is already smaller than 10, then the table name will be "quadmap_high".
func (s *Storage) GenerateTableName(key quadtree.QuadKey) string {

	targetKey := key
	zoom := targetKey.Zoom()
	if zoom < TablePartitionZoomLevel {
		return "quadmap_high"
	}

	for targetKey.Zoom() > TablePartitionZoomLevel {
		targetKey, _ = targetKey.Parent()
	}

	newTableName := fmt.Sprintf("quadmap_%d", targetKey)
	return newTableName
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) BeginTxx() (*sqlx.Tx, error) {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	tx, err := s.db.BeginTxx(context.Background(), nil)
	return tx, err
}

func (s *Storage) CommitTxx(txx *sqlx.Tx) error {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	return txx.Commit()
}

func (s *Storage) InsertTileWithTableName(txx *sqlx.Tx, tableName string, tile TileEntity) error {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	statement := fmt.Sprintf("INSERT INTO %s (quadkey, details_mask, details_id ) VALUES ($1,$2,$3)", tableName)
	txx.MustExec(statement, int64(tile.QuadKey), int64(tile.DetailsMask), tile.DetailsID)
	return nil
}

func (s *Storage) InsertTileWith(txx *sqlx.Tx, tile TileEntity) error {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	tableName := s.GenerateTableName(tile.QuadKey)
	statement := fmt.Sprintf("INSERT INTO %s (quadkey, details_mask, details_id ) VALUES ($1,$2,$3)", tableName)
	txx.MustExec(statement, int64(tile.QuadKey), int64(tile.DetailsMask), tile.DetailsID)
	return nil
}

func (s *Storage) InsertDetails(details DetailsEntity) (int64, error) {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	res := s.db.MustExec(`INSERT INTO details ( border, simple_border, tiletype, datetime, enabled, scale, identifier, simple_border_wkb) VALUES ($1,$2,$3,$4,$5,$6, $7, $8);`, details.Border, details.SimpleBorder, details.TileType, details.DateTime, true, details.Scale, details.Identifier, details.SimpleBorderWKB)

	lastInsertedID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertedID, nil
}

func (s *Storage) UpdateDetails(details DetailsEntity) error {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	s.db.MustExec(`UPDATE details set simple_border_wkb = $1 WHERE id=$2;`, details.SimpleBorderWKB, details.Id)
	return nil
}

func (s *Storage) GetDetails(id int) (*DetailsEntity, error) {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	var entity DetailsEntity
	s.db.Select(&entity, `SELECT id, border, simple_border, tiletype, datetime, scale, identifier, simple_border_wkb FROM details WHERE enabled = true AND id = $1`, fmt.Sprintf("%d", id))
	return &entity, nil
}

func (s *Storage) GetAllDetails() ([]DetailsEntity, error) {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	var entities []DetailsEntity
	s.db.Select(&entities, `SELECT id, border, simple_border, tiletype, datetime, scale, identifier, simple_border_wkb FROM details WHERE enabled = true`)
	return entities, nil
}

func (s *Storage) GetTile(qk quadtree.QuadKey) (*TileEntity, error) {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	var entity TileEntity
	s.db.Select(&entity, `SELECT quadkey,detail_mask  FROM quadmap WHERE quadkey = $1`, fmt.Sprintf("%d", qk))
	return &entity, nil
}

func (s *Storage) SearchDetailsWithinQuadKey(qk quadtree.QuadKey, includeSimpleBorder bool, limit int) ([]DetailsEntity, error) {
	// get range of quadkeys to cover the entire tile. So convert to slippy... then get next tile
	// to the right and down...
	x1, y1, z1 := qk.SlippyCoords()
	x2 := x1 + 1
	y2 := y1
	qk2, err := quadtree.GenerateQuadKeyIndexFromSlippy(x2, y2, z1)
	if err != nil {
		return nil, err
	}
	return s.SearchDetailsBetweenQuadKeys(qk, qk2, includeSimpleBorder, limit)
}

// SearchDetailsWithinQuadKey returns details for any hits within a particular QuadKey
func (s *Storage) SearchDetailsBetweenQuadKeys(qk1 quadtree.QuadKey, qk2 quadtree.QuadKey, includeSimpleBorder bool, limit int) ([]DetailsEntity, error) {

	qkint64 := int64(qk1)
	qk2int64 := int64(qk2)
	//var allKeys []int64
	//s.db.Select(&allKeys, `select distinct details_id from quadmap qm where  qm.quadkey >= $1 AND qm.quadkey <= $2 limit 48;`, qkint64, qk2int64)

	fmt.Printf("qk1 %d\n", qkint64)
	fmt.Printf("qk2 %d\n", qk2int64)
	fmt.Printf("Diff between qks %d\n", qk2-qk1)
	var entities []DetailsEntity

	tableName := s.GenerateTableName(qk1)
	fmt.Printf("Searching table %s\n", tableName)
	var statement string
	if includeSimpleBorder {
		statement = fmt.Sprintf("select d.id,d.scale,d.identifier, d.simple_border_wkb from details d where d.id in (select distinct details_id from %s qm where  qm.quadkey >= $1 AND qm.quadkey < $2 ) limit $3;", tableName)
	} else {
		statement = fmt.Sprintf("select d.id,d.scale,d.identifier from details d where d.id in (select distinct details_id from %s qm where  qm.quadkey >= $1 AND qm.quadkey < $2 ) limit $3;", tableName)
	}

	fmt.Printf("statement %s\n", statement)
	printStatement := strings.Replace(statement, "$1", fmt.Sprintf("%d", qkint64), -1)
	printStatement = strings.Replace(printStatement, "$2", fmt.Sprintf("%d", qk2int64), -1)
	printStatement = strings.Replace(printStatement, "$3", fmt.Sprintf("%d", limit), -1)

	fmt.Printf("printStatement %s\n", printStatement)
	//s.dbLock.Lock()
	//defer s.dbLock.Unlock()
	err := s.db.Select(&entities, statement, qkint64, qk2int64, limit)
	if err != nil {
		fmt.Printf("XXX err %+v\n", err)
		return nil, err
	}

	return entities, nil
}

func (s *Storage) SearchQuadKeysWithinQuadKey(qk quadtree.QuadKey) ([]int64, error) {

	// get range of quadkeys to cover the entire tile. So convert to slippy... then get next tile
	// to the right and down...
	x1, y1, z1 := qk.SlippyCoords()
	x2 := x1 + 1
	y2 := y1 + 1
	qk2, err := quadtree.GenerateQuadKeyIndexFromSlippy(x2, y2, z1)
	if err != nil {
		return nil, err
	}

	var entities []int64

	qkint64 := int64(qk)
	qk2int64 := int64(qk2)
	fmt.Printf("qk1 %d\n", qkint64)
	fmt.Printf("qk2 %d\n", qk2int64)
	tableName := s.GenerateTableName(qk)
	fmt.Printf("Searching table %s\n", tableName)
	statement := fmt.Sprintf("select distinct qm.details_id from %s qm where qm.quadkey >= $1 AND qm.quadkey <= $2;", tableName)

	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	s.db.Select(&entities, statement, qkint64, qk2int64)
	return entities, nil
}

func (s *Storage) InsertIdentifier(identifier string) error {
	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	s.db.MustExec(`INSERT INTO processed ( identifier ) VALUES ($1);`, identifier)

	return nil
}

func (s *Storage) HasIdentifier(identifier string) bool {

	s.dbLock.Lock()
	defer s.dbLock.Unlock()
	var existingIdentifier []string
	err := s.db.Select(&existingIdentifier, `SELECT identifier  FROM processed WHERE identifier = $1`, identifier)
	if err != nil {
		log.Errorf("error checking for identifier", err)
		return false
	}
	return len(existingIdentifier) > 0
}
