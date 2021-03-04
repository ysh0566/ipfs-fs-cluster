package consensus

import (
	"context"
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/raft"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ysh0566/ipfs-fs-cluster/consensus/pb"
	"github.com/ysh0566/ipfs-fs-cluster/consensus/state"
	"github.com/ysh0566/ipfs-fs-cluster/datastore"
	"io"
)

var ErrInconsistent = errors.New("inconsistent")

type Fsm struct {
	client       *httpapi.HttpApi
	State        *state.FileTreeState
	ctx          context.Context
	inconsistent bool
}

func NewFsm(store *datastore.BadgerDB, api *httpapi.HttpApi) (*Fsm, error) {
	_state, err := state.NewFileTreeState(store, api.Dag())
	if err != nil {
		return nil, err
	}
	return &Fsm{
		client:       api,
		State:        _state,
		ctx:          context.Background(),
		inconsistent: false,
	}, nil
}

func (f *Fsm) Apply(log *raft.Log) interface{} {
	var err error
	index := f.State.Index()
	if log.Index < index {
		f.inconsistent = true
		return ErrInconsistent
	} else if log.Index == index {
		f.inconsistent = false
		return ErrInconsistent
	}
	inss := &pb.Instructions{}
	if err = proto.Unmarshal(log.Data, inss); err != nil {
		return err
	}
	errs := make([]error, len(inss.Instruction))
	for index, ins := range inss.Instruction {
		errs[index] = f.State.Execute(ins)
	}
	f.State.SetIndex(log.Index)
	_ = f.State.Flush()
	return errs
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
	state *state.FileTreeState
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	return s.state.Marshal(sink)
}

func (s Snapshot) Release() {

}
