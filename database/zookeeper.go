package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/klarna/eremetic/types"
	"github.com/samuel/go-zookeeper/zk"
)

type zkDriver struct {
	connection types.ZkConnection
	path       string
}

type zkConnector struct{}

func (z zkConnector) Connect(zksStr string) (types.ZkConnection, error) {
	zks := strings.Split(zksStr, ",")
	conn, _, err := zk.Connect(zks, time.Second)

	return conn, err
}

func createZKConnector() types.ZkConnectorInterface {
	return types.ZkConnectorInterface(zkConnector{})
}

func parsePath(zkpath string) (string, string, error) {
	u, err := url.Parse(zkpath)
	if err != nil {
		return "", "", err
	}

	path := strings.TrimRight(u.Path, "/")
	return u.Host, path, nil
}

func createZKDriver(connector types.ZkConnectorInterface, zkPath string) (TaskDB, error) {
	if zkPath == "" {
		return nil, errors.New("Missing ZK path")
	}

	servers, path, err := parsePath(zkPath)
	if err != nil {
		return nil, err
	}

	conn, err := connector.Connect(servers)
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

	driver := zkDriver{connection: types.ZkConnection(conn), path: path}

	return driver, nil
}

func (z zkDriver) Close() {
	z.connection.Close()
}

func (z zkDriver) Clean() error {
	path := fmt.Sprintf("%s/", z.path)
	return z.connection.Delete(path, -1)
}

func (z zkDriver) PutTask(task *types.EremeticTask) error {
	path := fmt.Sprintf("%s/%s", z.path, task.ID)

	encode, err := encode(task)
	if err != nil {
		logrus.WithError(err).Error("Unable to encode task to byte-array.")
		return err
	}

	exists, stat, err := z.connection.Exists(path)
	if err != nil {
		logrus.WithError(err).Error("Unable to check existance of database.")
		return err
	}

	if exists {
		_, err = z.connection.Set(path, encode, stat.Version)
		return err
	}

	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	_, err = z.connection.Create(path, encode, flags, acl)
	return err
}

func (z zkDriver) ReadTask(id string) (types.EremeticTask, error) {
	task, err := z.ReadUnmaskedTask(id)

	applyMask(&task)

	return task, err
}

func (z zkDriver) ReadUnmaskedTask(id string) (types.EremeticTask, error) {
	var task types.EremeticTask
	path := fmt.Sprintf("%s/%s", z.path, id)

	bytes, _, err := z.connection.Get(path)
	json.Unmarshal(bytes, &task)

	return task, err

}

func (z zkDriver) ListNonTerminalTasks() ([]*types.EremeticTask, error) {
	tasks := []*types.EremeticTask{}
	paths, _, _ := z.connection.Children(z.path)
	for _, p := range paths {
		t, err := z.ReadTask(p)
		if err != nil {
			logrus.WithError(err).Error("Unable to read task from database, skipping")
			continue
		}
		if !t.IsTerminated() {
			applyMask(&t)
			tasks = append(tasks, &t)
		}
	}

	return tasks, nil
}
