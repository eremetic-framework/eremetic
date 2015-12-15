package database

import (
	"encoding/json"

	"github.com/klarna/eremetic/types"
	"github.com/boltdb/bolt"
)

// PutTask stores a requested task in the database
func PutTask(task *types.EremeticTask) error {
	err := ensureDB()
	if err != nil {
		return err
	}

	return boltdb.Update(func(tx *bolt.Tx) error {
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

// ReadTask fetches a task from the database
func ReadTask(id string) (types.EremeticTask, error) {
	var task types.EremeticTask

	err := ensureDB()
	if err != nil {
		return task, err
	}

	err = boltdb.View(func(tx *bolt.Tx) error {
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

func ListNonTerminalTasks() ([]*types.EremeticTask, error) {
	var tasks []*types.EremeticTask

	err := ensureDB()
	if err != nil {
		return tasks, err
	}

	err = boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tasks"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		b.ForEach(func(_, v []byte) error {
			var task types.EremeticTask
			json.Unmarshal(v, &task)
			if !task.IsTerminated() {
				tasks = append(tasks, &task)
			}
			return nil
		})
		return nil
	})

	return tasks, err
}
