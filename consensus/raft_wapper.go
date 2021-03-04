package consensus

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/ysh0566/ipfs-fs-cluster/consensus/pb"
)

type raftWapper struct {
	*raft.Raft
}

func (r raftWapper) Call(instructions []*pb.Instruction) []error {

	inss := pb.Instructions{
		Instruction:          instructions,
		Ctx:                  nil,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}
	bs, err := proto.Marshal(&inss)
	if err != nil {
		return CopyError(err, len(instructions))
	}
	future := r.Apply(bs, defaultTimeout)
	if future.Error() != nil {
		return CopyError(err, len(instructions))
	}
	switch res := future.Response().(type) {
	case error:
		return CopyError(res, len(instructions))
	case []error:
		return res
	default:
		panic("should not return other type")
	}
}

func CopyError(err error, num int) []error {
	errs := make([]error, num)
	for i := 0; i < num; i++ {
		errs[i] = err
	}
	return errs
}
