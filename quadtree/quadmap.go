package quadtree

import (
	"errors"
	"fmt"
	"math"
	"sync"

	log "github.com/sirupsen/logrus"
)

// TileDetailsGroup is same as TileDetails but we also want
// the quadkey that gave us the match.
// TileDetailsGroup is used when returning query results and NOT
// actually part of the quadmap itself.
type TileDetailsGroup struct {
	GroupTileTypeDetails
	QuadKey QuadKey
}

// TileDetails information about a tile, groups its associated with,
// tiletypes etc etc.
// TileDetails is used when returning query results and NOT
// actually part of the quadmap itself.
type TileDetails struct {
	Groups []TileDetailsGroup
	lock   sync.RWMutex
}

func (td *TileDetails) AddTileDetailsGroup(tdg TileDetailsGroup) {
	td.lock.Lock()
	defer td.lock.Unlock()
	td.Groups = append(td.Groups, tdg)
}

func (td *TileDetails) GetTileDetailsGroups() []TileDetailsGroup {
	td.lock.RLock()
	defer td.lock.RUnlock()
	return td.Groups
}

// DataReader function is provided by the consumer of the Quadmap.
// This will read a byte slice and scale and populate the Quadmap with
// the appropriate deserialised data.
// Will read as far as expandToLevel value (exclusive)
type DataReader func(qm *QuadMap, data *[]byte, groupID GroupID, tileType TileType, expandToLevel byte) error

// QuadMap is a quadtree in disguise...
type QuadMap struct {

	// map of quadkey to tile
	quadKeyMap map[QuadKey]*Tile

	// function able to take byte slices and populate Quadmap.
	dataReader DataReader

	lock sync.RWMutex
}

// NewQuadMap create a new quadmap
// Should provide a large initialCapacity when dealing with large quadtree structures
func NewQuadMap(initialCapacity int) *QuadMap {
	return &QuadMap{
		quadKeyMap: make(map[QuadKey]*Tile, initialCapacity),
	}
}

// SetDataReader sets the data reader for the quadmap
// Ideally would like it to be as part of NewQuadMap but need to pass in the quadmap...  so catch 22
func (qm *QuadMap) SetDataReader(dr DataReader) {
	qm.dataReader = dr
}

// GetParentTile returns parent tile of passed in tile t
func (qm *QuadMap) GetParentTile(t *Tile) (*Tile, error) {
	parentKey, err := t.QuadKey.Parent()
	if err != nil {
		return nil, err
	}

	qm.lock.RLock()
	parentTile, ok := qm.quadKeyMap[parentKey]
	qm.lock.RUnlock()

	if !ok {
		return nil, errors.New("parent tile not found")
	}
	return parentTile, nil
}

// GetChildInPos returns child tile of passed in tile t which is in position pos
// pos is a number between 0 and 3, where 0 is top left, 1 is top right, 2 is bottom left and 3 is bottom right
func (qm *QuadMap) GetChildInPos(t *Tile, pos int) (*Tile, error) {
	childKey, err := t.QuadKey.ChildAtPos(pos)
	if err != nil {
		return nil, err
	}
	qm.lock.RLock()
	childTile, ok := qm.quadKeyMap[childKey]
	qm.lock.RUnlock()

	if !ok {
		return nil, errors.New(fmt.Sprintf("child tile in pos %d not found", pos))
	}
	return childTile, nil
}

// GetExactTileForSlippy returns tile for slippy co-ord match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForSlippy(x uint32, y uint32, z byte) (*Tile, error) {
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return nil, err
	}
	return qm.GetExactTileForQuadKey(quadKey)
}

// GetExactTileForQuadKey returns tile for quadkey match. Does NOT traverse up the ancestry
func (qm *QuadMap) GetExactTileForQuadKey(quadKey QuadKey) (*Tile, error) {

	qm.lock.RLock()
	defer qm.lock.RUnlock()

	// if actual quadkey exists, return tile.
	if t, ok := qm.quadKeyMap[quadKey]; ok {
		return t, nil
	}
	return nil, errors.New("no tile found")
}

// NumberOfTilesForZoom returns number of tiles for a given zoom level.
// It will NOT include parents that may be used when querying (and the parents
// are marked as full)
// Given we don't keep track of zoom levels separately, we need to traverse the
// entire quadmap. If this is a common operation we'll need to track/cache this
// information somewhere. Although for the limited test cases so far it's pretty much instant
func (qm *QuadMap) NumberOfTilesForZoom(zoom byte) int {
	count := 0
	for _, t := range qm.quadKeyMap {
		if t.QuadKey.Zoom() == zoom {
			count++
		}
	}
	return count
}

// GetTilesForTypeAndZoom gets tiles for a given tile type and zoom level
func (qm *QuadMap) GetTilesForTypeAndZoom(tt TileType, zoom byte) []*Tile {
	tiles := []*Tile{}

	qm.lock.RLock()
	defer qm.lock.RUnlock()
	for _, t := range qm.quadKeyMap {
		if t.QuadKey.Zoom() == zoom {
			for _, g := range t.groups {
				if g.Details.HasTileType(tt) {
					tiles = append(tiles, t)
				}
			}
		}
	}
	return tiles
}

// NumberOfTiles returns number of tiles in quadmap
func (qm *QuadMap) NumberOfTiles() int {
	return len(qm.quadKeyMap)
}

// AddTile adds a pre-generated tile (which has its quadkey already)
func (qm *QuadMap) AddTile(t *Tile) error {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	qm.quadKeyMap[t.QuadKey] = t
	return nil
}

// CreateTileAtSlippyCoords creates a tile to the quadmap at slippy coords
func (qm *QuadMap) CreateTileAtSlippyCoords(x uint32, y uint32, z uint32, tileType TileType) (*Tile, error) {

	// x,y,z are already child coords...  so no need to take pos into account
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, byte(z))
	if err != nil {
		return nil, err
	}

	qm.lock.Lock()
	defer qm.lock.Unlock()
	// check if child exists.
	if tile, ok := qm.quadKeyMap[quadKey]; ok {
		//err := tile.UpdateTileTypeFullRawDataWatermarkByGroupID(groupID, tileType, false, false, rawData)
		//if err != nil {
		//	return nil, err
		//}

		return tile, nil
	}

	t := NewTileWithQuadKey(quadKey)
	//err = t.UpdateTileTypeFullRawDataWatermarkByGroupID(groupID, tileType, false, false, rawData)
	//if err != nil {
	//	return nil, err
	//}

	qm.quadKeyMap[t.QuadKey] = t
	return t, nil
}

// createChildForPos creates child tile for tile t in appropriate position
// Populates tile type and full flags based off parent.
// FIXME(kpfaulkner) confirm can delete
func createChildForPos(childQuadKey QuadKey, pos int) (*Tile, error) {
	//child := &Tile{QuadKey: childQuadKey}
	child := NewTileWithQuadKey(childQuadKey)
	return child, nil
}

// HaveTileForSlippyGroupIDAndTileType returns bool indicating if we have details for a tile at the provided
// slippy co-ords but also matching the tiletype and groupID.
func (qm *QuadMap) HaveTileForSlippyGroupIDAndTileType(x uint32, y uint32, z byte, groupID GroupID, tileType TileType) (bool, error) {
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return false, err
	}
	return qm.HaveTileForGroupIDAndTileType(quadKey, groupID, tileType, true)
}

// HaveTileForGroupIDAndTileType returns bool indicating if we have details for a tile at the provided
// quadKey provided but also matching the tiletype and groupID.
// It will return true if tile requested is found OR an ancestor that is FULL
// The process is:
//
//   - if quadkey exists, groupID exists and tiletype exists, return true
//
//   - else
//
//   - get parent quadkey
//
//   - if parent quadkey exists and is full, return parent details
//
//   - loop until no parent.
//
//     What happens if we hit a parent that is NOT full? No tile therefore return error?
//     Returns tile (actual or parent), bool indicating if actual (true == actual, false == ancestor) and error
func (qm *QuadMap) HaveTileForGroupIDAndTileType(quadKey QuadKey, groupID GroupID, tileType TileType, actualTile bool) (bool, error) {

	qm.lock.RLock()

	// if actual quadkey exists, check tiletype and groupID
	if t, ok := qm.quadKeyMap[quadKey]; ok {
		for _, g := range t.groups {
			if g.Details.GroupID() == groupID {
				hasTileType, isFull := g.Details.HasTileTypeAndFull(tileType)
				if hasTileType {
					if isFull || actualTile {
						qm.lock.RUnlock()
						return true, nil
					}
				}
			}
		}
	}
	qm.lock.RUnlock()

	parentQuadKey, err := quadKey.Parent()
	if err != nil {
		return false, err
	}

	// check parents and upwards. actualTile is false since we're querying ancestors
	found, err := qm.HaveTileForGroupIDAndTileType(parentQuadKey, groupID, tileType, false)
	if err != nil {
		return false, err
	}

	// return whether found or not
	return found, nil
}

func (qm *QuadMap) GetTileDetailsForSlippyCoordsAndTileTypeTopDown(x uint32, y uint32, z byte, tileTypes []TileType, tileDetails *TileDetails) error {
	quadKey, err := GenerateQuadKeyIndexFromSlippy(x, y, z)
	if err != nil {
		return err
	}
	return qm.GetTileDetailsForQuadkeyAndTileTypeTopDown(quadKey, tileTypes, tileDetails)
}

// GetTileDetailsForQuadkeyAndTileTypeTopDown returns details for the tile for quadkey
// This starts at the highest quadkey possible and works its way down. This will allow dynamic populating of the quadmap
// if required.
// TODO(kpfaulkner) check if we should only be considering full or target tiles?
// The more I think about it I think we should just return if there's a QK match at all.
func (qm *QuadMap) GetTileDetailsForQuadkeyAndTileTypeTopDown(quadKey QuadKey, tileTypes []TileType, tileDetails *TileDetails) error {

	allAncestors := quadKey.GetAllAncestorsAndSelf()

	targetScale := quadKey.Zoom()
	for _, qk := range allAncestors {

		qm.lock.RLock()
		t, ok := qm.quadKeyMap[qk]
		qm.lock.RUnlock()
		if ok {

			// whatever groups are in tile t....  add the details to tileDetails but only if full (if we're processing parent)
			for _, g := range t.GetGroupDetails() {
				for _, tileType := range tileTypes {
					hasTileType, isFull := g.Details.HasTileTypeAndFull(tileType)
					if hasTileType {
						if isFull || qk.Zoom() == targetScale {
							tileDetails.AddTileDetailsGroup(TileDetailsGroup{GroupTileTypeDetails: g.Details, QuadKey: qk})
							continue
						}
					}

					// If at watermark, then need to populate the quadmap to further scale depths and
					// reset the IsWatermark to a different depth.
					if g.Data[tileType].IsWatermark {
						// reset IsWaterMark...

						t.ClearWatermarkForGroupIDAndTileType(g.Details.GroupID(), tileType)
						// populate the quadmap down to targetScale (so dont have to populate all scale/zoom levels)
						err := qm.dataReader(qm, g.Data[tileType].Data, g.Details.GroupID(), tileType, targetScale)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	// by this stage tileDetails should be populated with all tiles from top down that have matches (full or target scale)
	return nil
}

type GroupIDAndTileTypePair struct {
	GroupID  GroupID
	TileType TileType
}

// GetGroupIDsForQuadKeysTopDown searches the QuadMap for the passed in quadKeys.
func (qm *QuadMap) GetGroupIDsForQuadKeysTopDown(quadKeys []QuadKey, tileTypes []TileType, upperScale byte, lowerScale byte) ([]GroupIDAndTileTypePair, error) {

	groupIDs := make(map[GroupID]map[TileType]bool)
	for _, quadKey := range quadKeys {
		allAncestors := quadKey.GetAllAncestorsAndSelf()

		// skip to level 12
		// need to determine if this is realistic for ALL surveys... or just the collection I've tried?
		allAncestors = allAncestors[upperScale:]
		for _, qk := range allAncestors {
			qm.lock.RLock()
			t, ok := qm.quadKeyMap[qk]
			qm.lock.RUnlock()
			if ok {

				// whatever groups are in tile t....  add the details to tileDetails but only if full (if we're processing parent)
				for _, g := range t.GetGroupDetails() {
					for _, tileType := range tileTypes {
						hasTileType, isFull := g.Details.HasTileTypeAndFull(tileType)
						if hasTileType {
							if isFull || qk.Zoom() == lowerScale {
								if _, ok := groupIDs[g.Details.GroupID()]; !ok {
									groupIDs[g.Details.GroupID()] = make(map[TileType]bool)
								}
								groupIDs[g.Details.GroupID()][tileType] = true
								continue
							}
						}

						// If at watermark, then need to populate the quadmap to further scale depths and
						// reset the IsWatermark to a different depth.
						if g.Data[tileType].IsWatermark {
							// reset IsWaterMark...

							t.ClearWatermarkForGroupIDAndTileType(g.Details.GroupID(), tileType)
							// populate the quadmap down to targetScale (so dont have to populate all scale/zoom levels)
							err := qm.dataReader(qm, g.Data[tileType].Data, g.Details.GroupID(), tileType, lowerScale)
							if err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
	}

	groupIDList := []GroupIDAndTileTypePair{}

	for k, v := range groupIDs {
		for tt, _ := range v {
			groupIDList = append(groupIDList, GroupIDAndTileTypePair{
				GroupID:  k,
				TileType: tt,
			})
		}
	}

	// by this stage tileDetails should be populated with all tiles from top down that have matches (full or target scale)
	return groupIDList, nil
}

// GetSlippyBoundsForGroupIDTileTypeAndZoom returns the minx,miny,maxx,maxy slippy coords for a given zoom level
// extracted from the quadmap. Brute forcing it for now.
func (qm *QuadMap) GetSlippyBoundsForGroupIDTileTypeAndZoom(groupID GroupID, tileType TileType, zoom byte) (uint32, uint32, uint32, uint32, error) {

	var minX uint32 = math.MaxUint32
	var minY uint32 = math.MaxUint32
	var maxX uint32 = 0
	var maxY uint32 = 0

	for quadKey, v := range qm.quadKeyMap {

		if quadKey == 0 {
			continue // should this be in the quadMap at all?
		}

		z := quadKey.Zoom()
		if z > zoom {
			continue
		}

		// only get tiletype and groupID that we want. Also needs to be either equal zoom OR is full.
		for _, g := range v.groups {
			gID := g.Details.GroupID()
			if gID == groupID {
				hasTileType, isFull := g.Details.HasTileTypeAndFull(tileType)

				if !hasTileType {
					continue
				}

				// only continue if precise zoom level OR this tile is considered full.
				if z == zoom || isFull {
					minChild, maxChild, err := quadKey.GetMinMaxEquivForZoomLevel(zoom)
					if err != nil {
						log.Errorf("error while generating min/max for quadkey %s", err.Error())
						return 0, 0, 0, 0, err
					}

					x, y, _ := minChild.SlippyCoords()
					if x < minX {
						minX = x
					}
					if y < minY {
						minY = y
					}

					x, y, _ = maxChild.SlippyCoords()
					if x > maxX {
						maxX = x
					}

					if y > maxY {
						maxY = y
					}
				}
			}
		}

	}

	return minX, minY, maxX, maxY, nil
}

func (qm *QuadMap) PrintStats(tileType TileType) {
	fmt.Printf("Number of tiles %d\n", qm.NumberOfTiles())

	groupDetailSizes := make(map[int]int)
	quadKeyScaleDetails := make(map[byte]int)
	groupsWithoutData := make(map[GroupID]int)
	groupsWithData := make(map[GroupID]int)
	for k, v := range qm.quadKeyMap {

		for _, g := range v.groups {
			if g.Data[tileType].Data == nil {
				groupsWithoutData[g.Details.GroupID()]++
				x, y, z := k.SlippyCoords()
				fmt.Printf("coord %d %d %d has no data: groupid %d\n", x, y, z, g.Details.GroupID())

				qks := k.GetAllAncestorsAndSelf()
				for _, qk := range qks {
					if t, ok := qm.quadKeyMap[qk]; ok {
						gd := t.GetGroupDetailsByGroupIDAndTileType(g.Details.GroupID(), tileType)
						if gd != nil {
							x, y, z := qk.SlippyCoords()
							if gd.Data[tileType].Data == nil {
								fmt.Printf("coord %d %d %d has NO data: groupid %d\n", x, y, z, g.Details.GroupID())
							} else {
								fmt.Printf("coord %d %d %d has data: groupid %d\n", x, y, z, g.Details.GroupID())
							}
						}
					}
				}

			} else {
				groupsWithData[g.Details.GroupID()]++
			}
		}

		groupSize := len(v.groups)

		x, y, z := k.SlippyCoords()

		if _, ok := quadKeyScaleDetails[z]; !ok {
			quadKeyScaleDetails[z] = 1
		} else {
			quadKeyScaleDetails[z]++
		}

		if _, ok := groupDetailSizes[groupSize]; !ok {
			groupDetailSizes[groupSize] = 1
		} else {
			groupDetailSizes[groupSize] += groupSize
		}

		// only 1 group and only vert..
		if v.groups[0].Data[tileType].IsWatermark {
			fmt.Printf("X:%d Y:%d Z:%d is watermark\n", x, y, z)
		}

	}

	total := 0
	for k, v := range groupDetailSizes {
		fmt.Printf("Group size %d : count %d\n", k, v)
		total += v
	}

	for k, v := range quadKeyScaleDetails {
		fmt.Printf("qk scale %d : count %d\n", k, v)
	}

	fmt.Printf("total groups recorded is %d\n", total)

	fmt.Printf("total number of uint64s stored (QKs + group details) %d\n", total+qm.NumberOfTiles())

	count := 0
	for k, _ := range groupsWithoutData {
		if _, ok := groupsWithData[k]; ok {
			fmt.Printf("Group %d has data and no data\n", k)
			count++
		}
	}
	fmt.Printf("Groups with both data and no data %d\n", count)
	fmt.Printf("Groups without data %d\n", len(groupsWithoutData))
	fmt.Printf("Groups with data %d\n", len(groupsWithData))

}
