package infra

import (
	"fmt"

	"github.com/boltdb/bolt"
	"taiyigo.com/common"
)

var blotDb *bolt.DB
var (
	CONF_TABLE        = "conf"
	STF_HISTORY_TABLE = "stf_history"
)

type dbCloseGuid struct {
	common.TItemLife
}

func (dbg *dbCloseGuid) Close() {
	if blotDb != nil {
		blotDb.Close()
	}
}

func StartDb() error {
	dbName := fmt.Sprintf("%s/blot/blot.db", common.Conf.Infra.FsDir)
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		common.Logger.Infof("Open db:%s, failed:%s", dbName, err)
		return err
	}
	blotDb = db
	common.TaddLife(&dbCloseGuid{})
	return nil
}

func CheckAndSet(table string, key string, value string) (bool, error) {
	err := blotDb.Update(func(tx *bolt.Tx) error {
		buck, err := tx.CreateBucketIfNotExists([]byte(table))
		if err != nil {
			common.Logger.Infof("create bucket %s failed:%s", table, err)
			return err
		}
		bKey := []byte(key)
		v := buck.Get(bKey)
		if v != nil {
			return gIsBExist
		}
		err = buck.Put(bKey, []byte(value))
		if err != nil {
			common.Logger.Infof("put key:%s, failed:%s", key, err)
			return err
		}
		return nil
	})
	if err != nil {
		if isTargetError(err, gIsBExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
