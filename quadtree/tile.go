package quadtree

// TileType is the type of a given tile
// Given we'll have very few tile types we could just use a byte, but
// may eventually use this as a bitmask to help filtering. So will
// keep it as a uint16 for now. Can change later if space becomes an issue
type TileType byte

// TileTypeMask indicates TileType and full indication.
// Something like: Vert|VertFull|North|NorthFull|South|SouthFull........
type TileTypeMask uint32

type TileDetailID uint32

// TileDetail contains 2 pieces of information.
// First 32 bits is groupID
// Second 32 bits is a mask of tiletype and indicator to say if full or not. (TileTypeMask )
type TileDetail uint64

func NewTileDetail(groupID uint32, tileType TileType, full bool) TileDetail {
	td := TileDetail(groupID) << 32
	td.SetTileTypeAndFull(tileType, full)
	return td
}

// Hash returns a hash of the TileDetail. Used for making
// unique index. Currently stupidly simple implementation but will
// refactor if it turns out this is a bottleneck. FIXME(kpfaulkner)
//func (td TileDetail) Hash() uint32 {
//	return hash(fmt.Sprintf("%d:%d:%d", td.GroupID, td.TileType, td.Scale))
//}

// Tile information about a tile, groups its associated with,
// tiletypes etc etc.
// A lot of this may be stored out of memory on storage, so is NOT required
// for querying the quadmap itself, but more for when you know you are interested
// in a specific tile and want more details
type Tile struct {
	Details []TileDetail
}

// SetTileType will add tileType to tile if it does not exist.
func (t *Tile) SetTileTypeForGroupID(groupID uint32, tileType TileType, full bool) error {
	found := false
	for i, detail := range t.Details {
		if detail.GetGroupID() == groupID {
			found = true
			err := detail.SetTileTypeAndFull(tileType, full)
			if err != nil {
				return err
			}
			t.Details[i] = detail
			break
		}
	}

	if !found {
		td := NewTileDetail(groupID, tileType, full)
		t.Details = append(t.Details, td)
	}

	return nil
}

func (t TileDetail) GetGroupID() uint32 {
	return uint32(t >> 32)
}

func (t TileDetail) GetTileTypeMask() uint32 {
	return uint32(t & 0xFFFFFFFF)
}

func (t *TileDetail) SetTileTypeMask(mask uint32) {
	cleared := *t & 0xFFFFFFFF00000000
	cleared = cleared | TileDetail(mask)
	*t = cleared
}

// GetTileTypes returns all the TileTypes for this tile.
func (t TileDetail) GetTileTypes() []TileType {

	var tileTypes []TileType

	tileTypeMask := t.GetTileTypeMask()
	for tt, typeBit := range TileLUT {
		if tileTypeMask&typeBit == 1 {
			tileTypes = append(tileTypes, tt)
		}
	}

	return tileTypes
}

// HasTileType indicates if TileDetail has type and if its full
func (t TileDetail) HasTileType(tt TileType) (bool, bool) {
	tileTypeMask := t.GetTileTypeMask()
	typeBit := TileLUT[tt]
	return tileTypeMask&typeBit == 1, tileTypeMask&(typeBit+1) == 1
}

func setBit(n uint64, pos uint64) uint64 {
	n |= uint64(1 << pos)
	return n
}

func clearBit(n uint64, pos uint64) uint64 {
	mask := uint64(^(1 << pos))
	n &= mask
	return n
}

func (t *TileDetail) SetTileTypeAndFull(tt TileType, full bool) error {

	typeBit := TileLUT[tt]
	*t = TileDetail(setBit(uint64(*t), uint64(typeBit)))

	if full {
		*t = TileDetail(setBit(uint64(*t), uint64(typeBit+1)))
	} else {
		*t = TileDetail(clearBit(uint64(*t), uint64(typeBit+1)))
	}

	return nil
}
