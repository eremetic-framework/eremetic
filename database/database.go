package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/spf13/viper"
)

var boltdb *bolt.DB

// NewDB is used to load the database handler into memory.
// It will create a new database file if it doesn't already exist.
func NewDB(file string) error {
	if !filepath.IsAbs(file) {
		dir, _ := os.Getwd()
		file = fmt.Sprintf("%s/../%s", dir, file)
	}
	os.MkdirAll(filepath.Dir(file), 0755)

	db, err := bolt.Open(file, 0600, nil)
	boltdb = db
	return err
}

// Clean is used to delete the tasks bucket
func Clean() error {
	return boltdb.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte("tasks")); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("tasks")); err != nil {
			return err
		}
		return nil
	})
}

// Close is used to Close the database
func Close() {
	if boltdb != nil {
		boltdb.Close()
	}
}

func ensureDB() error {
	if boltdb == nil {
		err := NewDB(viper.GetString("database"))
		return err
	}
	return nil
}
