package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/ysh0566/ipfs-fs-cluster/modules"
	"github.com/ysh0566/ipfs-fs-cluster/rpc"
	"go.uber.org/fx"
	"time"
)

//func BootStrap() error {
//
//}
func T(raft *raft.Raft, clients *modules.Clients, config modules.Config) {
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for range ticker.C {
			if config.P2P.Identity.PeerID == string(raft.Leader()) {
				fmt.Println("im leader")
				continue
			}
			c := clients.Client(string(raft.Leader()))
			if c != nil {
				rep, err := c.SayHello(context.Background(), &rpc.HelloRequest{Name: config.P2P.Identity.PeerID})
				if err == nil {
					fmt.Println("reply:", rep.Message)
				} else {
					fmt.Println("error:", err.Error())
				}
			}
		}
	}()
}

func Context() context.Context {
	return context.Background()
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
		fx.Provide(modules.Transport),
		fx.Provide(modules.Raft),
		fx.Invoke(modules.RpcServer),
		fx.Provide(modules.RpcClients),
		fx.Provide(Context),
		fx.Invoke(T),
	}
	app := New(options...)
	app.Run()

}
