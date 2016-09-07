package boltdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"

	"github.com/klarna/eremetic"
)

// Connection defines the functions needed to interact with a bolt database
type Connection interface {
	Close() error
	Update(func(*bolt.Tx) error) error
	View(func(*bolt.Tx) error) error
	Path() string
}

// ConnectorInterface assists in opening a boltdb connection
type ConnectorInterface interface {
	Open(path string) (Connection, error)
}

type driver struct {
	database Connection
}

type connector struct{}

func (b connector) Open(file string) (Connection, error) {
	if !filepath.IsAbs(file) {
		dir, _ := os.Getwd()
		file = fmt.Sprintf("%s/../%s", dir, file)
	}
	os.MkdirAll(filepath.Dir(file), 0755)

	return bolt.Open(file, 0600, nil)
}

func newConnector() ConnectorInterface {
	return ConnectorInterface(connector{})
}

func NewTaskDB(file string) (eremetic.TaskDB, error) {
	return newDriver(newConnector(), file)
}

func newDriver(connector ConnectorInterface, file string) (eremetic.TaskDB, error) {
	if file == "" {
		return nil, errors.New("Missing BoltDB database loctation.")
	}

	db, err := connector.Open(file)

	return driver{database: db}, err
}

// Close is used to Close the database
func (db driver) Close() {
	if db.database != nil {
		db.database.Close()
	}
}

// Clean is used to delete the tasks bucket
func (db driver) Clean() error {
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
func (db driver) PutTask(task *eremetic.Task) error {
	return db.database.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("tasks"))
		if err != nil {
			return err
		}

		encoded, err := eremetic.Encode(task)
		if err != nil {
			logrus.WithError(err).Error("Unable to encode task to byte-array.")
			return err
		}

		return b.Put([]byte(task.ID), encoded)
	})
}

// ReadTask fetches a task from the database and applies a mask to the
// MaskedEnvironment field
func (db driver) ReadTask(id string) (eremetic.Task, error) {
	task, err := db.ReadUnmaskedTask(id)

	eremetic.ApplyMask(&task)

	return task, err
}

// ReadUnmaskedTask fetches a task from the database and does not mask the
// MaskedEnvironment field.
// This function should be considered internal to Eremetic, and is used where
// we need to fetch a task and then re-save it to the database. It should not
// be returned to the API.
func (db driver) ReadUnmaskedTask(id string) (eremetic.Task, error) {
	var task eremetic.Task

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
func (db driver) ListNonTerminalTasks() ([]*eremetic.Task, error) {
	tasks := []*eremetic.Task{}

	err := db.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tasks"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		b.ForEach(func(_, v []byte) error {
			var task eremetic.Task
			json.Unmarshal(v, &task)
			if !task.IsTerminated() {
				eremetic.ApplyMask(&task)
				tasks = append(tasks, &task)
			}
			return nil
		})
		return nil
	})

	return tasks, err
}
