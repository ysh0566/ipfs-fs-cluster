package modules

import (
	"context"
	"github.com/hashicorp/raft"
	gostream "github.com/libp2p/go-libp2p-gostream"
	p2praft "github.com/libp2p/go-libp2p-raft"
	"github.com/ysh0566/ipfs-fs-cluster/consensus"
	"github.com/ysh0566/ipfs-fs-cluster/datastore"
	"github.com/ysh0566/ipfs-fs-cluster/network"
	"github.com/ysh0566/ipfs-fs-cluster/rpc"
	"google.golang.org/grpc"
	"time"
)

const protocol = "/grpc/0.0.1"

func Network(ctx context.Context, cfg *network.NetConfig) (*network.Network, error) {
	return network.NewNetwork(ctx, *cfg)
}

func RaftConfig(js Config) *raft.Config {
	cfg := raft.DefaultConfig()
	cfg.LogLevel = js.Raft.LogLevel
	cfg.LocalID = raft.ServerID(js.P2P.Identity.PeerID)
	return cfg
}

func DataStore(js Config) (*datastore.BadgerStore, error) {
	return datastore.NewBadgerStore(js.DBPath)
}

func SnapshotStore() raft.SnapshotStore {
	return raft.NewInmemSnapshotStore()
}

func Fsm() raft.FSM {
	return &consensus.Fsm{}
}

func Transport(n *network.Network) (raft.Transport, error) {
	return p2praft.NewLibp2pTransport(n.Host(), time.Minute*2)
}

func Raft(conf *raft.Config, fsm raft.FSM, snaps raft.SnapshotStore, trans raft.Transport, badger *datastore.BadgerStore, js Config) (*raft.Raft, error) {
	servers := make([]raft.Server, len(js.Raft.Peers))
	for i := 0; i < len(js.Raft.Peers); i++ {
		servers[i] = raft.Server{
			Suffrage: 0,
			ID:       raft.ServerID(js.Raft.Peers[i]),
			Address:  raft.ServerAddress(js.Raft.Peers[i]),
		}
	}
	r, err := raft.NewRaft(conf, fsm, badger, badger, snaps, trans)
	if err != nil {
		return nil, err
	}
	r.BootstrapCluster(raft.Configuration{Servers: servers})
	return r, nil
}

func RpcServer(n *network.Network) *grpc.Server {
	listener, err := gostream.Listen(n.Host(), network.Protocol)
	if err != nil {
		return nil
	}
	s1 := grpc.NewServer()
	rpc.RegisterGreeterServer(s1, &rpc.Server{})
	go s1.Serve(listener)
	return s1
}

type Clients struct {
	c map[string]rpc.GreeterClient
}

func (client Clients) Client(id string) rpc.GreeterClient {
	if c, ok := client.c[id]; !ok {
		return nil
	} else {
		return c
	}
}

func RpcClients(n *network.Network, js Config) (*Clients, error) {
	c := make(map[string]rpc.GreeterClient)
	for i := 0; i < len(js.Raft.Peers); i++ {
		if js.Raft.Peers[i] == js.P2P.Identity.PeerID {
			continue
		}
		conn, err := n.Connect(n.Context(), js.Raft.Peers[i])
		if err != nil {
			return nil, err
		}
		c[js.Raft.Peers[i]] = rpc.NewGreeterClient(conn)
	}
	return &Clients{c: c}, nil
}
