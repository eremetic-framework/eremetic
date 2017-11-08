package zk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"

	"github.com/cybricio/eremetic"
)

// connection wraps a zk.Conn struct for testability
type connection interface {
	Close()
	Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error)
	Delete(path string, n int32) error
	Exists(path string) (bool, *zk.Stat, error)
	Get(path string) ([]byte, *zk.Stat, error)
	Set(path string, data []byte, version int32) (*zk.Stat, error)
	Children(path string) ([]string, *zk.Stat, error)
}

// connector helps create a zookeeper connection
type connector interface {
	Connect(path string) (connection, error)
}

// TaskDB is a Zookeeper implementation of the task database.
type TaskDB struct {
	conn connection
	path string
}

type defaultConnector struct{}

func (z defaultConnector) Connect(zksStr string) (connection, error) {
	zks := strings.Split(zksStr, ",")
	conn, _, err := zk.Connect(zks, time.Second)

	return conn, err
}

func parsePath(zkpath string) (string, string, error) {
	u, err := url.Parse(zkpath)
	if err != nil {
		return "", "", err
	}

	path := strings.TrimRight(u.Path, "/")
	return u.Host, path, nil
}

// NewTaskDB returns a new instance of a Zookeeper TaskDB.
func NewTaskDB(zk string) (*TaskDB, error) {
	return newCustomTaskDB(defaultConnector{}, zk)
}

func newCustomTaskDB(c connector, path string) (*TaskDB, error) {
	if path == "" {
		return nil, errors.New("Missing ZK path")
	}

	servers, path, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	conn, err := c.Connect(servers)
	if err != nil {
		return nil, err
	}

	exists, _, err := conn.Exists(path)
	if err != nil {
		return nil, err
	}

	if !exists {
		flags := int32(0)
		acl := zk.WorldACL(zk.PermAll)

		_, err = conn.Create(path, nil, flags, acl)
		if err != nil {
			logrus.WithError(err).Error("Unable to create node.")
			return nil, err
		}
	}

	return &TaskDB{
		conn: conn,
		path: path,
	}, nil
}

// Close closes the connection to the database.
func (z *TaskDB) Close() {
	z.conn.Close()
}

// Clean removes all tasks from the database.
func (z *TaskDB) Clean() error {
	path := fmt.Sprintf("%s/", z.path)
	return z.conn.Delete(path, -1)
}

// PutTask adds a new task to the database.
func (z *TaskDB) PutTask(task *eremetic.Task) error {
	path := fmt.Sprintf("%s/%s", z.path, task.ID)

	encode, err := eremetic.Encode(task)
	if err != nil {
		logrus.WithError(err).Error("Unable to encode task to byte-array.")
		return err
	}

	exists, stat, err := z.conn.Exists(path)
	if err != nil {
		logrus.WithError(err).Error("Unable to check existance of database.")
		return err
	}

	if exists {
		_, err = z.conn.Set(path, encode, stat.Version)
		return err
	}

	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	_, err = z.conn.Create(path, encode, flags, acl)
	return err
}

// ReadTask returns a task with a given id, or an error if not found.
func (z *TaskDB) ReadTask(id string) (eremetic.Task, error) {
	task, err := z.ReadUnmaskedTask(id)

	eremetic.ApplyMask(&task)

	return task, err
}

// ReadUnmaskedTask returns a task with all its environment variables unmasked.
func (z *TaskDB) ReadUnmaskedTask(id string) (eremetic.Task, error) {
	var task eremetic.Task
	path := fmt.Sprintf("%s/%s", z.path, id)

	bytes, _, err := z.conn.Get(path)
	json.Unmarshal(bytes, &task)

	return task, err

}

func (z *TaskDB) DeleteTask(id string) error {
	path := fmt.Sprintf("%s/%s", z.path, id)
	_, stat, err := z.conn.Exists(path)
	if err != nil {
		logrus.WithError(err).Error("Unable to check existance of database.")
		return err
	}
	err = z.conn.Delete(path, stat.Version)
	return err
}

// ListNonTerminalTasks returns all non-terminal tasks.
func (z *TaskDB) ListNonTerminalTasks() ([]*eremetic.Task, error) {
	tasks := []*eremetic.Task{}
	paths, _, _ := z.conn.Children(z.path)
	for _, p := range paths {
		t, err := z.ReadTask(p)
		if err != nil {
			logrus.WithError(err).Error("Unable to read task from database, skipping")
			continue
		}
		if !t.IsTerminated() {
			eremetic.ApplyMask(&t)
			tasks = append(tasks, &t)
		}
	}

	return tasks, nil
}
