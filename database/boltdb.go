package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/klarna/eremetic/types"
)

type boltDriver struct {
	database types.BoltConnection
}

type boltConnector struct{}

func (b boltConnector) Open(file string) (types.BoltConnection, error) {
	if !filepath.IsAbs(file) {
		dir, _ := os.Getwd()
		file = fmt.Sprintf("%s/../%s", dir, file)
	}
	os.MkdirAll(filepath.Dir(file), 0755)

	return bolt.Open(file, 0600, nil)
}

func createBoltConnector() types.BoltConnectorInterface {
	return types.BoltConnectorInterface(boltConnector{})
}

func createBoltDriver(connector types.BoltConnectorInterface, file string) (TaskDB, error) {
	if file == "" {
		return nil, errors.New("Missing BoltDB database loctation.")
	}

	db, err := connector.Open(file)

	return boltDriver{database: db}, err
}

// Close is used to Close the database
func (db boltDriver) Close() {
	if db.database != nil {
		db.database.Close()
	}
}

// Clean is used to delete the tasks bucket
func (db boltDriver) Clean() error {
	return db.database.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte("tasks")); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("tasks")); err != nil {
			return err
		}
		return nil
	})
}

// PutTask stores a requested task in the database
func (db boltDriver) PutTask(task *types.EremeticTask) error {
	return db.database.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("tasks"))
		if err != nil {
			return err
		}

		encoded, err := encode(task)
		if err != nil {
			logrus.WithError(err).Error("Unable to encode task to byte-array.")
			return err
		}

		return b.Put([]byte(task.ID), encoded)
	})
}

// ReadTask fetches a task from the database and applies a mask to the
// MaskedEnvironment field
func (db boltDriver) ReadTask(id string) (types.EremeticTask, error) {
	task, err := db.ReadUnmaskedTask(id)

	applyMask(&task)

	return task, err
}

// ReadUnmaskedTask fetches a task from the database and does not mask the
// MaskedEnvironment field.
// This function should be considered internal to Eremetic, and is used where
// we need to fetch a task and then re-save it to the database. It should not
// be returned to the API.
func (db boltDriver) ReadUnmaskedTask(id string) (types.EremeticTask, error) {
	var task types.EremeticTask

	err := db.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tasks"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		v := b.Get([]byte(id))
		json.Unmarshal(v, &task)
		return nil
	})

	return task, err
}

// ListNonTerminalTasks returns a list of tasks that are not yet finished in one
// way or another.
func (db boltDriver) ListNonTerminalTasks() ([]*types.EremeticTask, error) {
	var tasks []*types.EremeticTask

	err := db.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tasks"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		b.ForEach(func(_, v []byte) error {
			var task types.EremeticTask
			json.Unmarshal(v, &task)
			if !task.IsTerminated() {
				applyMask(&task)
				tasks = append(tasks, &task)
			}
			return nil
		})
		return nil
	})

	return tasks, err
}

func (db boltDriver) Count() int {
	var tasks []*types.EremeticTask

	db.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tasks"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		b.ForEach(func(_, v []byte) error {
			var task types.EremeticTask
			json.Unmarshal(v, &task)
			tasks = append(tasks, &task)
			return nil
		})
		return nil
	})

	return len(tasks)
}
