package server

// http://oldblog.antirez.com/post/redis-persistence-demystified.html

import (
	"github.com/drone/mq/stomp"

	"github.com/syndtr/goleveldb/leveldb"
)

type store interface {
	put(*stomp.Message) error
	delete(*stomp.Message) error
	close() error
}

type datastore struct {
	db *leveldb.DB
}

func (d *datastore) put(m *stomp.Message) error {
	return d.db.Put(m.ID, m.Bytes(), nil)
}

func (d *datastore) delete(m *stomp.Message) error {
	return d.db.Delete(m.ID, nil)
}

func (d *datastore) close() error {
	return d.db.Close()
}

// loadDatastore reads the datastore from disk and restores
// persisted message to the appropriate queues.
func loadDatastore(path string, b *router) (store, error) {
	db, err := leveldb.RecoverFile(path, nil)
	if err != nil {
		return nil, err
	}

	// iterate through the persisted messages
	// and send to the broker.
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		m := stomp.NewMessage()
		m.Parse(iter.Value())
		b.publish(m)
	}
	iter.Release()

	return &datastore{db: db}, nil
}
