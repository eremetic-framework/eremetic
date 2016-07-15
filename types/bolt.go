package types

import "github.com/boltdb/bolt"

// BoltConnection defines the functions needed to interact with a bolt database
type BoltConnection interface {
	Close() error
	Update(func(*bolt.Tx) error) error
	View(func(*bolt.Tx) error) error
	Path() string
}

// BoltConnectorInterface assists in opening a boltdb connection
type BoltConnectorInterface interface {
	Open(path string) (BoltConnection, error)
}
