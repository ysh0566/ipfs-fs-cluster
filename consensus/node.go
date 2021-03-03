package consensus

import (
	"context"
	"errors"
	"github.com/hashicorp/raft"
	"github.com/ipfs/go-mfs"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/ysh0566/ipfs-fs-cluster/consensus/pb"
	"github.com/ysh0566/ipfs-fs-cluster/network"
	"google.golang.org/grpc"
	"sync"
)

type Node struct {
	raft       *raft.Raft
	fsm        *Fsm
	retryTimes int
	ID         string
	operator   Operator
	mtx        sync.Mutex
	network    *network.Network
	ctx        context.Context
}

func (n *Node) Raft() *raft.Raft {
	return n.raft
}

func (n *Node) Op(ctx context.Context, code pb.Operation_Code, params ...string) error {
	if n.fsm.Inconsistent() {
		return errors.New("inconsistent state")
	}
	before := n.fsm.State.MustGetRoot()
	fork, err := n.fsm.State.Copy(ctx)
	if err != nil {
		return err
	}
	if err := fork.RpcOp(ctx, code, params); err != nil {
		return err
	}
	after := fork.MustGetRoot()
	err = n.TrySwitchOperator()
	if err != nil {
		return err
	}
	switch code {
	case pb.Operation_CP:
		return n.operator.Cp(ctx, before, after, params[0], params[1])
	case pb.Operation_MV:
		return n.operator.Mv(ctx, before, after, params[0], params[1])
	case pb.Operation_RM:
		return n.operator.Rm(ctx, before, after, params[0])
	case pb.Operation_MKDIR:
		return n.operator.MkDir(ctx, before, after, params[0])
	default:
		return errors.New("no matched operator")
	}
}

func (n *Node) Ls(ctx context.Context, path string) ([]mfs.NodeListing, error) {
	return n.fsm.State.Ls(ctx, path)
}

func (n *Node) Leader() string {
	return string(n.raft.Leader())
}

func (n *Node) Operator() string {
	return n.operator.Address()
}

func (n *Node) TrySwitchOperator() error {
	if n.Leader() != n.Operator() {
		if err := n.SwitchOperator(); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) SwitchOperator() error {
	if n.ID == n.Leader() {
		n.operator = NewLocalOperator(n.raft, n.ID)
	} else {
		conn, err := n.network.Connect(n.ctx, n.Leader())
		if err != nil {
			return err
		}
		n.operator = NewRemoteOperator(conn, n.Leader())
	}
	return nil
}

func NewNode(ctx context.Context, r *raft.Raft, fsm *Fsm, id string, net *network.Network) (*Node, error) {
	node := &Node{
		raft:       r,
		fsm:        fsm,
		retryTimes: 3,
		ID:         id,
		operator:   nil,
		mtx:        sync.Mutex{},
		network:    net,
		ctx:        ctx,
	}
	err := node.SwitchOperator()
	listener, err := gostream.Listen(net.Host(), network.Protocol)
	if err != nil {
		return nil, err
	}
	s1 := grpc.NewServer()

	pb.RegisterFileOpServer(s1, FsOpServer{operator: NewLocalOperator(r, id)})
	go s1.Serve(listener)
	return node, err
}
