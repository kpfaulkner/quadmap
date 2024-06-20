package quadtree

import (
	"sync"
)

type TileType uint16
type GroupID uint32

const (
	TileTypeVert  TileType = 0b000000000001
	TileTypeEast  TileType = 0b000000000010
	TileTypeNorth TileType = 0b000000000100
	TileTypeSouth TileType = 0b00000001000
	TileTypeWest  TileType = 0b0000010000
)

// GroupTileTypeDetails has bare information about the "group" that created the tile this is associated with.
// The first 32 bits is a unique identifier for the group
// The next 32 bits are a combination of tiletype (can be maximum of 16) and a full flag.
// eg.
//
//		|63-----------------32|31-----------------------------0|
//		| GroupID             | TileType + Full                |
//		|                     |00000000000000010000000000000001|
//	    In this case, bit 16 is set (TileTypeVert) and bit 0 is set (Full). So GroupID has a tiletype of Vert and is full.
//      where-as
//		|                     |00000000000000100000000000000000|
//      Has Type of East, but is not full.

type GroupTileTypeDetails uint64

func (gd GroupTileTypeDetails) GroupID() GroupID {
	return GroupID(gd >> 32)
}

// HasTileTypeAndFull returns if tiletype is set for the GroupTileTypeDetails and if
// the tile is full.
func (gd GroupTileTypeDetails) HasTileTypeAndFull(tileType TileType) (bool, bool) {
	tt := uint32(gd >> 16)
	hasTileType := uint16(tt)&uint16(tileType) == 1
	isFull := false
	if hasTileType {
		if uint16(gd)&uint16(tileType) == 1 {
			isFull = true
		}
	}
	return hasTileType, isFull
}

func (gd GroupTileTypeDetails) ReturnTileTypes() []TileType {
	var tileTypes []TileType
	if gd.HasTileType(TileTypeVert) {
		tileTypes = append(tileTypes, TileTypeVert)
	}
	if gd.HasTileType(TileTypeEast) {
		tileTypes = append(tileTypes, TileTypeEast)

	}
	if gd.HasTileType(TileTypeNorth) {
		tileTypes = append(tileTypes, TileTypeNorth)
	}
	if gd.HasTileType(TileTypeSouth) {
		tileTypes = append(tileTypes, TileTypeSouth)
	}
	if gd.HasTileType(TileTypeWest) {
		tileTypes = append(tileTypes, TileTypeWest)
	}
	return tileTypes
}

func (gd GroupTileTypeDetails) HasTileType(tileType TileType) bool {
	tt := uint32(gd)
	hasTileType := uint16(tt>>16)&uint16(tileType) == 1
	return hasTileType
}

func (gd GroupTileTypeDetails) SetTileTypeAndFull(tileType TileType, full bool) GroupTileTypeDetails {

	gd = gd | GroupTileTypeDetails(uint32(tileType)<<16)
	if full {
		gd = gd | GroupTileTypeDetails(tileType)
	} else {
		gd = gd & ^GroupTileTypeDetails(tileType)
	}

	return gd
}

func NewGroupTileTypeDetails(gid GroupID, tt TileType, full bool) GroupTileTypeDetails {
	gd := GroupTileTypeDetails(uint64(gid) << 32)
	gd = gd.SetTileTypeAndFull(tt, full)
	return gd
}

type RawData struct {
	Data        *[]byte
	IsWatermark bool
}

// GroupDetails has the GroupTileTypeDetails (which has the groupID and tiletype/full flag)
// As well as a pointer to the raw data used for on-demand population
type GroupDetails struct {
	Details GroupTileTypeDetails

	Data map[TileType]RawData
	// Data is the raw data used to store the Quadmap data.
	//Data *[]byte

	// IsWatermark indicates that this group is the last that is populated within the Quadmap.
	// There may be further entries available at lower scale levels, but they have NOT been read
	// from the Data and populated into the Quadmap.
	//IsWatermark bool
}

func NewGroupDetails(gid GroupID, tt TileType, full bool) GroupDetails {
	gd := GroupDetails{}
	gd.Details = NewGroupTileTypeDetails(gid, tt, full)
	gd.Data = make(map[TileType]RawData)
	return gd
}

// Tile is a node within a quadtree.
// Although a Tile instance will only be in the quadmap once (for a given quadkey) it may
// be the case that the same tile+quadkey is used by multiple "groups" and also for
// multiple tiletypes within a group.
//
// For example, if we populate the quadmap with one set of data (called group1) and group1
// in turn has information about tiletype1 and tiletype2, this means that we'll need to track at a per quadkey
// level if the files are full (and for which tiletypes).
type Tile struct {
	QuadKey QuadKey

	// groups that have information for this tile. The IDs listed here can be used elsewhere to look up data.
	groups []GroupDetails

	lock sync.RWMutex
}

// NewTile creates a new tile at slippy co-ords x,y,z
// Will probably only be used for root tile
func NewTile(x uint32, y uint32, z byte) *Tile {
	t := &Tile{}
	quadKey := GenerateQuadKeyIndexFromSlippy(x, y, z)
	t.QuadKey = quadKey
	return t
}

func NewTileWithQuadKey(quadKey QuadKey) *Tile {
	t := &Tile{}
	t.QuadKey = quadKey
	return t
}

func (t *Tile) AddGroupDetails(gd GroupDetails) error {
	t.groups = append(t.groups, gd)
	return nil
}

func (t *Tile) GetGroupDetails() []GroupDetails {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.groups
}

func (t *Tile) SetWatermarkForGroupIDAndTileType(groupID GroupID, tt TileType) error {

	// coarse grain locking for now
	t.lock.Lock()
	defer t.lock.Unlock()
	for i, g := range t.groups {
		if g.Details.GroupID() == groupID && g.Details.HasTileType(tt) {
			data := t.groups[i].Data[tt]
			data.IsWatermark = true
			t.groups[i].Data[tt] = data
		}
	}
	return nil
}

func (t *Tile) ClearWatermarkForGroupIDAndTileType(groupID GroupID, tt TileType) error {
	// coarse grain locking for now
	t.lock.Lock()
	defer t.lock.Unlock()
	for i, g := range t.groups {
		if g.Details.GroupID() == groupID && g.Details.HasTileType(tt) {
			data := t.groups[i].Data[tt]
			data.IsWatermark = false
			t.groups[i].Data[tt] = data
		}
	}
	return nil
}

func (t *Tile) GetWatermarkForGroupIDAndTileType(groupID GroupID, tt TileType) bool {
	// coarse grain locking for now
	t.lock.RLock()
	defer t.lock.RUnlock()

	for i, g := range t.groups {
		if g.Details.GroupID() == groupID && g.Details.HasTileType(tt) {
			return t.groups[i].Data[tt].IsWatermark
		}
	}
	return false
}

// UpdateTileTypeFullRawDataByGroupID for a given GroupID, either update the tiletype, full and rawdata
// if they exist, otherwise create them.
func (t *Tile) UpdateTileTypeFullRawDataWatermarkByGroupID(groupID GroupID, tt TileType, full bool, isWaterMark bool, rawData *[]byte) error {

	t.lock.Lock()
	defer t.lock.Unlock()

	found := false
	for groupIdx, g := range t.groups {
		if g.Details.GroupID() == groupID {
			newGroupDetails := g.Details.SetTileTypeAndFull(tt, full)
			t.groups[groupIdx].Details = newGroupDetails
			data := t.groups[groupIdx].Data[tt]
			data.Data = rawData
			data.IsWatermark = isWaterMark
			t.groups[groupIdx].Data[tt] = data
			found = true
		}
	}

	if !found {
		gd := NewGroupDetails(groupID, tt, full)
		data := RawData{Data: rawData, IsWatermark: isWaterMark}
		gd.Data[tt] = data
		t.AddGroupDetails(gd)
	}

	return nil
}

// HasTileType... needs to loop through array. Hopefully wouldn't be too many
// FIXME(kpfaulkner) measure and confirm
func (t *Tile) HasTileType(groupID GroupID, tt TileType) bool {

	// coarse grain locking for now
	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, g := range t.groups {
		if g.Details.GroupID() == groupID && g.Details.HasTileType(tt) {
			return true
		}
	}
	return false
}

// GetTileZoomLevel returns zoom level of tile
func (t *Tile) GetTileZoomLevel() byte {
	return t.QuadKey.Zoom()
}

// IsTileFullByGroupIDAndTileType gets the full flag for a given tile type. Defaults to false if no flag found
func (t *Tile) IsTileFullByGroupIDAndTileType(groupID GroupID, tileType TileType) bool {

	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, g := range t.groups {
		if g.Details.GroupID() == groupID {
			hasTileType, isFull := g.Details.HasTileTypeAndFull(tileType)
			if hasTileType {
				return isFull
			}

		}
	}
	return false
}

func (t *Tile) GetGroupDetailsByGroupIDAndTileType(groupID GroupID, tileType TileType) *GroupDetails {

	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, g := range t.groups {
		if g.Details.GroupID() == groupID {
			hasTileType, _ := g.Details.HasTileTypeAndFull(tileType)
			if hasTileType {
				return &g
			}
		}
	}
	return nil
}
