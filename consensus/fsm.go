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
	"github.com/ysh0566/ipfs-fs-cluster/datastore"
	"io"
)

type Fsm struct {
	client       *httpapi.HttpApi
	State        *FileTreeState
	ctx          context.Context
	inconsistent bool
}

func NewFsm(store *datastore.BadgerStore, api *httpapi.HttpApi) (*Fsm, error) {
	r, err := mfs.NewRoot(context.Background(), api.Dag(), unixfs.EmptyDirNode(), func(ctx context.Context, cid cid.Cid) error {
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Fsm{
		client: api,
		State: &FileTreeState{
			dag:   api.Dag(),
			root:  r,
			store: store,
		},
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
	err = f.State.Op(ctx, op)
	if err != nil {
		return err
	}
	rh, err := f.State.Root()
	if err != nil {
		return err
	}
	if rh != op.Root {
		f.inconsistent = true
	} else {
		f.inconsistent = false
	}
	return rh
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot{state: f.State}, nil
}

func (f *Fsm) Restore(closer io.ReadCloser) error {
	defer closer.Close()
	return f.State.Unmarshal(closer)
}

type Snapshot struct {
	state *FileTreeState
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	return s.state.Marshal(sink)
}

func (s Snapshot) Release() {

}
