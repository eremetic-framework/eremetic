package boltdb

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"

	"github.com/klarna/eremetic"
)

// connection defines the functions needed to interact with a bolt database
type connection interface {
	Close() error
	Update(func(*bolt.Tx) error) error
	View(func(*bolt.Tx) error) error
	Path() string
}

// connector assists in opening a boltdb connection
type connector interface {
	Open(path string) (connection, error)
}

type defaultConnector struct{}

func (b defaultConnector) Open(file string) (connection, error) {
	os.MkdirAll(filepath.Dir(file), 0755)

	return bolt.Open(file, 0600, nil)
}

// TaskDB is a boltdb implementation of the task database.
type TaskDB struct {
	conn connection
}

// NewTaskDB returns a new instance of TaskDB.
func NewTaskDB(file string) (*TaskDB, error) {
	return newCustomTaskDB(defaultConnector{}, file)
}

func newCustomTaskDB(c connector, file string) (*TaskDB, error) {
	if file == "" {
		return nil, errors.New("missing boltdb database location")
	}

	conn, err := c.Open(file)
	if err != nil {
		return nil, err
	}

	err = conn.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("tasks"))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &TaskDB{conn: conn}, nil
}

// Close is used to Close the database
func (db *TaskDB) Close() {
	if db.conn != nil {
		db.conn.Close()
	}
}

// Clean is used to delete the tasks bucket
func (db *TaskDB) Clean() error {
	return db.conn.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte("tasks")); err != nil {
			return err
		}

		return nil
	})
}

// PutTask stores a requested task in the database
func (db *TaskDB) PutTask(task *eremetic.Task) error {
	return db.conn.Update(func(tx *bolt.Tx) error {
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
func (db *TaskDB) ReadTask(id string) (eremetic.Task, error) {
	task, err := db.ReadUnmaskedTask(id)

	eremetic.ApplyMask(&task)

	return task, err
}

// ReadUnmaskedTask fetches a task from the database and does not mask the
// MaskedEnvironment field.
// This function should be considered internal to Eremetic, and is used where
// we need to fetch a task and then re-save it to the database. It should not
// be returned to the API.
func (db *TaskDB) ReadUnmaskedTask(id string) (eremetic.Task, error) {
	var task eremetic.Task

	err := db.conn.View(func(tx *bolt.Tx) error {
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

func (db *TaskDB) DeleteTask(id string) error {
	return db.conn.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("tasks"))
		if err != nil {
			return err
		}
		return b.Delete([]byte(id))
	})
}

// ListNonTerminalTasks returns a list of tasks that are not yet finished in one
// way or another.
func (db *TaskDB) ListNonTerminalTasks() ([]*eremetic.Task, error) {
	tasks := []*eremetic.Task{}

	err := db.conn.View(func(tx *bolt.Tx) error {
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
