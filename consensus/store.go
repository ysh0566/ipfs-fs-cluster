package consensus

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"sync"
)

var (
	dbLogsPrefix     = []byte("logs")
	dbLogsFirstIndex = []byte("logs_first")
	dbLogsLastIndex  = []byte("logs_last")
	dbConfPrefix     = []byte("conf")
	ErrNotFound      = errors.New("not found")
)

type BadgerStore struct {
	db       *badger.DB
	encoders *sync.Pool
	decoders *sync.Pool
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	var err error
	store := &BadgerStore{}
	if store.db, err = badger.Open(badger.DefaultOptions(path)); err != nil {
		return nil, err
	}
	//store.encoders = &sync.Pool{New: func() interface{}{
	//	return NewEncoder()
	//}}
	//store.decoders = &sync.Pool{New: func() interface{}{
	//	return NewDecoder()
	//}}
	return store, nil
}

func (s *BadgerStore) set(key []byte, val []byte) error {
	fmt.Println(string(key), string(val))
	tx := s.db.NewTransaction(true)
	if err := tx.Set(key, val); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *BadgerStore) get(key []byte) ([]byte, error) {
	tx := s.db.NewTransaction(false)
	defer tx.Discard()
	item, err := tx.Get(key)
	if err != nil {
		return nil, err
	}
	return item.ValueCopy(nil)
}

func (s *BadgerStore) Set(key []byte, val []byte) error {
	k := append(dbConfPrefix, key...)
	return s.set(k, val)
}

func (s *BadgerStore) Get(key []byte) ([]byte, error) {

	k := append(dbConfPrefix, key...)
	if bs, err := s.get(k); err != nil && err == badger.ErrKeyNotFound {
		return nil, ErrNotFound
	} else {
		return bs, err
	}
}

func (s *BadgerStore) SetUint64(key []byte, val uint64) error {
	tmp := make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(tmp, val)
	return s.Set(key, tmp)
}

func (s *BadgerStore) GetUint64(key []byte) (uint64, error) {
	if tmp, err := s.Get(key); err != nil {
		return 0, err
	} else {
		return binary.LittleEndian.Uint64(tmp), nil
	}
}

func (s *BadgerStore) FirstIndex() (uint64, error) {
	if v, err := s.get(dbLogsFirstIndex); err == nil {
		return binary.LittleEndian.Uint64(v), nil

	} else {
		if err == badger.ErrKeyNotFound {
			return 0, nil
		}
		return 0, err
	}
}

func (s *BadgerStore) LastIndex() (uint64, error) {
	if v, err := s.get(dbLogsLastIndex); err == nil {
		return binary.LittleEndian.Uint64(v), nil

	} else {
		if err == badger.ErrKeyNotFound {
			return 0, nil
		}
		return 0, err
	}
}

func (s *BadgerStore) GetLog(index uint64, log *raft.Log) error {
	tmp := make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(tmp, index)
	tmp = append(dbLogsPrefix, tmp...)
	if bs, err := s.get(tmp); err == nil {
		return json.Unmarshal(bs, log)
	} else {
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}
		return err
	}
}

func (s *BadgerStore) StoreLog(log *raft.Log) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()
	tmp := make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(tmp, log.Index)
	tmp = append(dbLogsPrefix, tmp...)
	bs, err := json.Marshal(log)
	if err != nil {
		return err
	}
	if err := tx.Set(tmp, bs); err != nil {
		return err
	}
	tmp = make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(tmp, log.Index)
	if err := tx.Set(dbLogsLastIndex, tmp); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *BadgerStore) StoreLogs(logs []*raft.Log) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()
	tmp := make([]byte, 8, 8)
	var max uint64
	for _, log := range logs {
		binary.LittleEndian.PutUint64(tmp, log.Index)
		key := append(dbLogsPrefix, tmp...)

		bs, err := json.Marshal(log)
		if err != nil {
			return err
		}
		if err := tx.Set(key, bs); err != nil {
			return err
		}
		if log.Index > max {
			max = log.Index
		}
	}
	tmp = make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(tmp, max)
	if err := tx.Set(dbLogsLastIndex, tmp); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *BadgerStore) DeleteRange(min, max uint64) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()
	tmp := make([]byte, 8, 8)
	lowIndex, err := s.FirstIndex()
	if err != nil {
		return err
	}
	highIndex, err := s.LastIndex()
	if err != nil {
		return err
	}
	for i := min; i <= max; i++ {
		binary.LittleEndian.PutUint64(tmp, i)
		key := append(dbLogsPrefix, tmp...)
		if err := tx.Delete(key); err != nil && err != badger.ErrKeyNotFound {
			return err
		}
	}
	if min <= lowIndex {
		lowIndex = max + 1
	}
	if max >= highIndex {
		highIndex = min - 1
	}
	if lowIndex > highIndex {
		highIndex = 0
		lowIndex = 0
	}
	tmp = make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(tmp, lowIndex)
	if err := tx.Set(dbLogsFirstIndex, tmp); err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(tmp, highIndex)
	if err := tx.Set(dbLogsLastIndex, tmp); err != nil {
		return err
	}
	return tx.Commit()
}

//type Encoder struct {
//	writer *bytes.Buffer
//	internal gob.Encoder
//}
//
//func(e *Encoder) Reset() {
//	e.writer.Reset()
//}
//
//func(e *Encoder) Encode(obj interface{}) ([]byte, error) {
//	err := e.internal.Encode(obj)
//	if err != nil {
//		return nil, err
//	}
//	return e.writer.Bytes(), nil
//}
//
//func NewEncoder() *Encoder{
//	return &Encoder{
//		writer:   bytes.NewBufferString(""),
//		internal: gob.Encoder{},
//	}
//}
//
//type Decoder struct {
//	reader *bytes.Buffer
//	internal gob.Encoder
//}
//
//func(e *Decoder) Reset() {
//	e.reader.Reset()
//}
//
//func(e *Decoder) Decode(bs []byte, obj interface{}) error {
//	if _, err := e.reader.Write(bs); err != nil {
//		return err
//	}
//	return e.internal.Encode(obj)
//}
//
//func NewDecoder() *Decoder{
//	return &Decoder{
//		reader:   bytes.NewBufferString("?"),
//		internal: gob.Encoder{},
//	}
//}
