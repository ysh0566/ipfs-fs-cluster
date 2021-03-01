package datastore

import (
	"errors"
	"github.com/dgraph-io/badger/v3"
)

var (
	dbState        = []byte("state")
	ErrKeyNotFound = errors.New("not found")
)

type BadgerDB struct {
	db *badger.DB
}

func NewBadgerStore(path string) (*BadgerDB, error) {
	var err error
	store := &BadgerDB{}
	if store.db, err = badger.Open(badger.DefaultOptions(path)); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *BadgerDB) Close() error {
	return s.db.Close()
}

func (s *BadgerDB) set(key []byte, val []byte) error {
	tx := s.db.NewTransaction(true)
	if err := tx.Set(key, val); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *BadgerDB) get(key []byte) ([]byte, error) {
	tx := s.db.NewTransaction(false)
	defer tx.Discard()
	item, err := tx.Get(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return item.ValueCopy(nil)
}

func (s *BadgerDB) StoreState(hash string) error {
	return s.set(dbState, []byte(hash))
}

func (s *BadgerDB) LoadState() (string, error) {
	v, err := s.get(dbState)
	if err != nil {
		return "", err
	}
	return string(v), err
}
