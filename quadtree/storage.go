package quadtree

//
//import (
//	"encoding/binary"
//	"encoding/json"
//	errs "errors"
//	"fmt"
//
//	"github.com/cockroachdb/pebble"
//	"github.com/pingcap/errors"
//	log "github.com/sirupsen/logrus"
//)
//
//const (
//	storageDetailPrefex = "detail"
//	storageTilePrefex   = "tile"
//)
//
//type Storage interface {
//
//	// GetTileDetails returns the details (groupID, tileType,  full status etc) from storage.
//	GetTile(quadKey QuadKey) (Tile, error)
//
//	// SetTileDetails sets the details (groupID, tileType,  full status etc) in storage.
//	// If groupID+scale+tileType already exists, then it will be overwritten.
//	SetTile(quadKey QuadKey, detail TileDetail) error
//
//	// GetTileDetailsGroupByTileType returns the details for a given tileType at location.
//	GetTileByTileTypeAndGroupID(quadKey QuadKey, tileType TileType, groupID uint32) (TileDetails, error)
//
//	// SetDetail sets the details into storage. This will be reused across multiple tiles
//	SetDetail(quadKey QuadKey, detail TileDetail) (TileDetailID, error)
//
//	// GetDetail returns the detail from storage
//	GetDetail(tileDetailID TileDetailID) (TileDetail, error)
//}
//
//type PebbleStorage struct {
//	pdb *pebble.DB
//}
//
//func NewPebbleStorage() *PebbleStorage {
//	db := connectToPebbleDB()
//	return &PebbleStorage{
//		pdb: db,
//	}
//}
//
//func connectToPebbleDB() *pebble.DB {
//	// Pebble
//	pdb, err := pebble.Open("quadmapstorage", &pebble.Options{})
//	if err != nil {
//		log.Fatal(err)
//	}
//	return pdb
//}
//
//func (ps *PebbleStorage) GetTileDetails(quadKey QuadKey) (Tile, error) {
//
//	key := make([]byte, 8)
//	binary.LittleEndian.PutUint64(key, uint64(quadKey))
//	bytes, closer, err := ps.pdb.Get(key)
//	if err != nil {
//		// not a not-found... so return error.
//		return TileDetails{}, err
//
//	}
//	if closer != nil {
//		defer closer.Close()
//	}
//
//	var existingDetails TileDetails
//	err = json.Unmarshal(bytes, &existingDetails)
//	if err != nil {
//		return TileDetails{}, errors.Trace(err)
//	}
//
//	return existingDetails, nil
//}
//
//// SetTileDetail
//// 1) try and get existing details.
//// 2) update if exists, or create new
//// 3) store
//func (ps *PebbleStorage) SetTileDetail(quadKey QuadKey, details TileDetail) error {
//
//	// get details from storage (if they exist)
//
//	key := make([]byte, 8)
//	binary.LittleEndian.PutUint64(key, uint64(quadKey))
//	existingDetails, err := ps.GetTileDetails(quadKey)
//	if err != nil {
//		if !errs.Is(err, pebble.ErrNotFound) {
//
//			// not a not-found... so return error.
//			return err
//		}
//	}
//
//	found := false
//	// check if existing detail (TileType, groupID etc) are the same as existing
//	for _, ed := range existingDetails.Details {
//		if ed.GroupID == details.GroupID && ed.TileType == details.TileType {
//			// already exists, so just update.
//			ed.Full = details.Full
//			ed.Scale = details.Scale
//			found = true
//			break
//		}
//	}
//
//	if !found {
//		existingDetails.Details = append(existingDetails.Details, details)
//	}
//
//	bytes, err := json.Marshal(existingDetails)
//	if err != nil {
//		return errors.Trace(err)
//	}
//
//	// no sync...  titanic :P
//	err = ps.pdb.Set(key, bytes, pebble.NoSync)
//	if err != nil {
//		return errors.Trace(err)
//	}
//
//	return nil
//}
//
//// GetTileDetailByTileTypeAndGroupID returns the details for a given tileType/groupID at location.
//// Involves scanning details...  will do for now. Measure and optimise later. TODO(kpfaulkner)
//func (ps *PebbleStorage) GetTileDetailByTileTypeAndGroupID(quadKey QuadKey, tileType TileType, groupID uint32) (TileDetail, error) {
//	details, err := ps.GetTileDetails(quadKey)
//	if err != nil {
//		return TileDetail{}, err
//	}
//	for _, detail := range details.Details {
//		if detail.GroupID == groupID && detail.TileType == tileType {
//			return detail, nil
//		}
//	}
//	return TileDetail{}, errors.New("not found")
//}
//
//// SetDetail stores the TileDetail in storage. This will be reused across multiple tiles.
//func (ps *PebbleStorage) SetDetail(detail TileDetail) (TileDetailID, error) {
//	id := detail.Hash()
//
//	key := []byte(fmt.Sprintf("%s_%d", storageDetailPrefex, id))
//
//	bytes, err := json.Marshal(detail)
//	if err != nil {
//		return 0, errors.Trace(err)
//	}
//
//	err = ps.pdb.Set(key, bytes, pebble.NoSync)
//	if err != nil {
//		return 0, errors.Trace(err)
//	}
//
//	return TileDetailID(id), nil
//}
//
//func (ps *PebbleStorage) GetDetail(tileDetailID TileDetailID) (TileDetail, error) {
//	key := []byte(fmt.Sprintf("%s_%d", storageDetailPrefex, tileDetailID))
//
//	bytes, closer, err := ps.pdb.Get(key)
//	if err != nil {
//		// not a not-found... so return error.
//		return TileDetail{}, err
//	}
//
//	if closer != nil {
//		defer closer.Close()
//	}
//
//	var detail TileDetail
//	err = json.Unmarshal(bytes, &detail)
//	if err != nil {
//		return TileDetail{}, errors.Trace(err)
//	}
//
//	return detail, nil
//}
