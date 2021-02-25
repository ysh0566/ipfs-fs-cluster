package consensus

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"time"
)

const defaultTimeout = time.Second * 5

type Operator interface {
	Cp(ctx context.Context, preHash, hash, dir, path string) error
	Mv(ctx context.Context, preHash, hash, dir, path string) error
	Rm(ctx context.Context, preHash, hash, path string) error
	MkDir(ctx context.Context, preHash, hash, path string) error
	Address() string
}

type LocalOperator struct {
	raft *raft.Raft
	addr string
}

func (l *LocalOperator) Cp(ctx context.Context, preHash, hash, dir, path string) error {
	return l.operation(Operation_CP, preHash, hash, dir, path)
}

func (l *LocalOperator) Mv(ctx context.Context, preHash, hash, dir, path string) error {
	return l.operation(Operation_MV, preHash, hash, dir, path)
}

func (l *LocalOperator) Rm(ctx context.Context, preHash, hash, path string) error {
	return l.operation(Operation_RM, preHash, hash, path)
}

func (l *LocalOperator) MkDir(ctx context.Context, preHash, hash, path string) error {
	return l.operation(Operation_MKDIR, preHash, hash, path)
}

func (l *LocalOperator) Address() string {
	return l.addr
}

func NewLocalOperator(r *raft.Raft, address string) *LocalOperator {
	return &LocalOperator{
		raft: r,
		addr: address,
	}
}

func (l *LocalOperator) operation(code Operation_Code, preHash, hash string, params ...string) error {
	op := &Operation{
		Code:   code,
		Params: params,
		Ctx: &Ctx{
			Pre:  preHash,
			Next: hash,
		},
	}
	cmd, err := proto.Marshal(op)
	if err != nil {
		return err
	}
	future := l.raft.Apply(cmd, defaultTimeout)
	return future.Error()
}

type RemoteOperator struct {
	client FileOpClient
	addr   string
}

func (r *RemoteOperator) Cp(ctx context.Context, preHash, hash, dir, path string) error {
	_, err := r.client.Cp(ctx, &DirPath{
		Dir:  dir,
		Path: path,
		Ctx: &Ctx{
			Pre:  preHash,
			Next: hash,
		},
	})
	return err
}

func (r *RemoteOperator) Mv(ctx context.Context, preHash, hash, dir, path string) error {
	_, err := r.client.Mv(ctx, &DirPath{
		Dir:  dir,
		Path: path,
		Ctx: &Ctx{
			Pre:  preHash,
			Next: hash,
		},
	})
	return err
}

func (r *RemoteOperator) Rm(ctx context.Context, preHash, hash, path string) error {
	_, err := r.client.Rm(ctx, &Path{
		Path: path,
		Ctx: &Ctx{
			Pre:  preHash,
			Next: hash,
		},
	})
	return err
}

func (r *RemoteOperator) MkDir(ctx context.Context, preHash, hash, path string) error {
	_, err := r.client.MkDir(ctx, &Path{
		Path: path,
		Ctx: &Ctx{
			Pre:  preHash,
			Next: hash,
		},
	})
	return err
}

func (r *RemoteOperator) Address() string {
	return r.addr
}

func NewRemoteOperator(conn grpc.ClientConnInterface, addr string) *RemoteOperator {
	return &RemoteOperator{
		client: NewFileOpClient(conn),
		addr:   addr,
	}
}

type FsOpServer struct {
	operator *LocalOperator
}

func (f FsOpServer) mustEmbedUnimplementedFileOpServer() {

}

func (f FsOpServer) Cp(ctx context.Context, in *DirPath) (*Empty, error) {
	err := f.operator.Cp(ctx, in.Ctx.Pre, in.Ctx.Next, in.Dir, in.Path)
	return &Empty{}, err
}

func (f FsOpServer) Mv(ctx context.Context, in *DirPath) (*Empty, error) {
	err := f.operator.Mv(ctx, in.Ctx.Pre, in.Ctx.Next, in.Dir, in.Path)
	return &Empty{}, err
}

func (f FsOpServer) Rm(ctx context.Context, in *Path) (*Empty, error) {
	err := f.operator.Rm(ctx, in.Ctx.Pre, in.Ctx.Next, in.Path)
	return &Empty{}, err
}

func (f FsOpServer) MkDir(ctx context.Context, in *Path) (*Empty, error) {
	err := f.operator.MkDir(ctx, in.Ctx.Pre, in.Ctx.Next, in.Path)
	return &Empty{}, err
}
