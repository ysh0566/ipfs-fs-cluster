package consensus

import (
	"fmt"
	"github.com/hashicorp/raft"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"io"
)

type Fsm struct {
	client *httpapi.HttpApi
}

func (f Fsm) Apply(log *raft.Log) interface{} {
	fmt.Println(string(log.Data))
	return "????"
}

func (f Fsm) Snapshot() (raft.FSMSnapshot, error) {
	panic("implement me")
}

func (f Fsm) Restore(closer io.ReadCloser) error {
	panic("implement me")
}
