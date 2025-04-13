package quadtree

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

type TileDetails struct {
	ID         int
	Quadkey    QuadKey
	WebTileUID string
	TileType   int
	Full       bool
}

// SQLite storage (probably)
type Storage struct {
	conn *sql.Conn
}

func NewStorage() (*Storage, error) {
	s := &Storage{}

	err := s.createDB()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) Close() error {
	return s.conn.Close()
}

func (s *Storage) createDB() error {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		return err
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	s.conn = conn
	_, err = conn.ExecContext(context.Background(), `create table if not exists tile (quadkey varchar(20) primary key, name varchar(50))`)
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(context.Background(), `create table if not exists tile_type (id int primary, quadkey varchar(20), webtile_uid varchar(40), tile_type int, full bool)`)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetTileDetailsByQuadkey(quadkey QuadKey) (*Tile, error) {

	return nil, nil
}
