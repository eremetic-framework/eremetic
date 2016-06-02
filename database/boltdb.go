package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/klarna/eremetic/types"
)

type boltDriver struct {
	database *bolt.DB
}

func boltDB(file string) (TaskDB, error) {
	if file == "" {
		return nil, errors.New("Missing BoltDB database loctation.")
	}

	if !filepath.IsAbs(file) {
		dir, _ := os.Getwd()
		file = fmt.Sprintf("%s/../%s", dir, file)
	}
	os.MkdirAll(filepath.Dir(file), 0755)

	db, err := bolt.Open(file, 0600, nil)

	wrapped := wrap(db)

	return wrapped, err
}

func wrap(db *bolt.DB) TaskDB {
	return boltDriver{
		database: db,
	}
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

		encoded, err := json.Marshal(task)
		if err != nil {
			return err
		}

		return b.Put([]byte(task.ID), []byte(encoded))
	})
}

// ReadTask fetches a task from the database and applies a mask to the
// MaskedEnvironment field
func (db boltDriver) ReadTask(id string) (types.EremeticTask, error) {
	task, err := db.ReadUnmaskedTask(id)

	for k := range task.MaskedEnvironment {
		task.MaskedEnvironment[k] = "*******"
	}

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
