package storage

import "github.com/kpfaulkner/quadmap/quadmap"

// Entities:
// TileEntity has quadkey and details mask. The details mask will indicate if there are
// QuadKey is NOT the primary key... but will be indexed and will be main column we search on.
type TileEntity struct {
	QuadKey     quadmap.QuadKey `db:"quadkey"`
	TileType    uint16          `db:"tiletype"`
	Full        bool            `db:"full"`
	DetailsMask uint64          `db:"details_mask"`
	DetailsID   int64           `db:"details_id"`
}

type DetailsEntity struct {
	Id              uint64 `db:"id"`
	Border          string `db:"border"`
	SimpleBorder    string `db:"simple_border"`
	SimpleBorderWKB []byte `db:"simple_border_wkb"`
	TileType        uint16 `db:"tiletype"`
	DateTime        int64  `db:"datetime"`
	Enabled         bool   `db:"enabled"`
	Identifier      string `db:"identifier"`
	Scale           uint16 `db:"scale"`
}
