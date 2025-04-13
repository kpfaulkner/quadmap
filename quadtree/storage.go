package quadtree

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

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

	_, err = conn.ExecContext(context.Background(), `create table if not exists tile_type (id int primary keyquadkey varchar(20) primary key, quadkey varchar(20), type int, full bool)`)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetTileByQuadkey(quadkey QuadKey) (*Tile, error) {

	return nil, nil
}
