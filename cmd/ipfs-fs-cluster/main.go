package main

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/ysh0566/ipfs-fs-cluster/consensus"
	"github.com/ysh0566/ipfs-fs-cluster/modules"
	"go.uber.org/fx"
	"time"
)

//func BootStrap() error {
//
//}
func T(fsm raft.FSM, clients *modules.Clients, config modules.Config) {
	fsm2 := fsm.(*consensus.Fsm)
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for range ticker.C {
			fmt.Println(fsm2.State.Root())
		}
	}()
}

func main() {
	var options = []fx.Option{
		fx.Provide(modules.InitConfig),
		fx.Provide(modules.NetConfig),
		fx.Provide(modules.Network),
		fx.Provide(modules.RaftConfig),
		fx.Provide(modules.DataStore),
		fx.Provide(modules.SnapshotStore),
		fx.Provide(modules.Fsm),
		fx.Provide(modules.IpfsClient),
		fx.Provide(modules.Transport),
		fx.Provide(modules.Raft),
		fx.Provide(modules.RpcClients),
		fx.StopTimeout(time.Minute),
		fx.Invoke(modules.RpcServer),
		fx.Invoke(modules.Server2),
		fx.Invoke(T),
	}
	app := New(options...)
	app.Run()

}
