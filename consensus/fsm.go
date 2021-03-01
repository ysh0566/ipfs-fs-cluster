package consensus

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/ipfs/go-cid"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/go-unixfs"
	"github.com/ysh0566/ipfs-fs-cluster/datastore"
	"io"
	"strings"
)

type Fsm struct {
	client       *httpapi.HttpApi
	State        *FileTreeState
	ctx          context.Context
	inconsistent bool
}

func NewFsm(store *datastore.BadgerDB, api *httpapi.HttpApi) (*Fsm, error) {
	s, err := store.LoadState()
	state := &FileTreeState{
		dag:   api.Dag(),
		store: store,
		ctx:   context.Background(),
	}
	if err != nil {
		if err != datastore.ErrKeyNotFound {
			return nil, err
		} else {
			r, _ := mfs.NewRoot(context.Background(), api.Dag(), unixfs.EmptyDirNode(), func(ctx context.Context, cid cid.Cid) error {
				return nil
			})
			state.root = r
		}
	} else {
		err := state.Unmarshal(strings.NewReader(s))
		if err != nil {
			return nil, err
		}
	}
	return &Fsm{
		client:       api,
		State:        state,
		ctx:          context.Background(),
		inconsistent: false,
	}, nil
}

func (f *Fsm) Apply(log *raft.Log) interface{} {
	var err error
	if log.Index < f.State.index {
		f.inconsistent = true
		return f.State
	} else if log.Index == f.State.index {
		f.inconsistent = false
		return f.State
	}
	op := &Operation{}
	if err = proto.Unmarshal(log.Data, op); err != nil {
		return err
	}
	err = f.State.rpcOp(f.ctx, op.Code, op.Params)
	if err != nil {
		return err
	}
	if newRoot, err := f.State.Root(); err != nil && op.Ctx.Next == newRoot {
		f.inconsistent = false
	}
	f.State.index = log.Index
	_ = f.State.Flush()
	return f.State
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot{state: f.State}, nil
}

func (f *Fsm) Restore(closer io.ReadCloser) error {
	defer closer.Close()
	return f.State.Unmarshal(closer)
}

func (f *Fsm) Inconsistent() bool {
	return f.inconsistent
}

type Snapshot struct {
	state *FileTreeState
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	return s.state.Marshal(sink)
}

func (s Snapshot) Release() {

}
