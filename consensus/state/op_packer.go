package state

import (
	"context"
	"github.com/ysh0566/ipfs-fs-cluster/consensus/pb"
	"sync"
)

type Instruction interface {
	GetParams() []string
	GetNode() []byte
	GetCode() pb.Instruction_Code
}

type Call struct {
	done chan *Call
	ins  Instruction
	err  error
}

type Packer struct {
	chanPool sync.Pool
}

func (fs *Packer) Execute(ctx context.Context, ins Instruction) error {
	//switch ins.GetCode() {
	//case pb.:
	//
	//}
	return nil
}
