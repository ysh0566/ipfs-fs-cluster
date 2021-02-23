package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/ipfs/go-cid"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/go-unixfs"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ysh0566/ipfs-fs-cluster/datastore"
	"io"
)

type Fsm struct {
	client       *httpapi.HttpApi
	Mfs          FileStore
	index        uint64
	ctx          context.Context
	inconsistent bool
}

func NewFsm(store *datastore.BadgerStore) (*Fsm, error) {
	addr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/5001")
	if err != nil {
		return nil, err
	}
	api, err := httpapi.NewApi(addr)
	if err != nil {
		return nil, err
	}
	l, err := store.LastIndex()
	if err != nil {
		return nil, err
	}
	r, err := mfs.NewRoot(context.Background(), api.Dag(), unixfs.EmptyDirNode(), func(ctx context.Context, cid cid.Cid) error {
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Fsm{
		client: api,
		Mfs: FileStore{
			client: api,
			dag:    api.Dag(),
			root:   r,
		},
		index:        l,
		ctx:          context.Background(),
		inconsistent: false,
	}, nil
}

func (f *Fsm) Apply(log *raft.Log) interface{} {
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	op := FsOperation{}
	if err = json.Unmarshal(log.Data, &op); err != nil {
		return err
	}
	ctx, c := context.WithCancel(f.ctx)
	defer c()
	err = f.Mfs.Op(ctx, op)
	if err != nil {
		return err
	}
	rh, err := f.Mfs.Root()
	if err != nil {
		return err
	}
	if rh != op.Root {
		f.inconsistent = true
		fmt.Println("error inconsistent")
	} else {
		f.inconsistent = false
	}
	return rh
}

func (f Fsm) Snapshot() (raft.FSMSnapshot, error) {
	panic("implement me")
}

func (f Fsm) Restore(closer io.ReadCloser) error {
	panic("implement me")
}
