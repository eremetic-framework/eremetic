package types

import "github.com/samuel/go-zookeeper/zk"

// ZkConnection wraps a zk.Conn struct for testability
type ZkConnection interface {
	Close()
	Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error)
	Delete(path string, n int32) error
	Exists(path string) (bool, *zk.Stat, error)
	Get(path string) ([]byte, *zk.Stat, error)
	Set(path string, data []byte, version int32) (*zk.Stat, error)
	Children(path string) ([]string, *zk.Stat, error)
}

// ZkConnectorInterface helps create a zookeeper connection
type ZkConnectorInterface interface {
	Connect(path string) (ZkConnection, error)
}
