package modules

import (
	"context"
	"github.com/hashicorp/raft"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	gostream "github.com/libp2p/go-libp2p-gostream"
	p2praft "github.com/libp2p/go-libp2p-raft"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ysh0566/ipfs-fs-cluster/consensus"
	"github.com/ysh0566/ipfs-fs-cluster/datastore"
	"github.com/ysh0566/ipfs-fs-cluster/http"
	"github.com/ysh0566/ipfs-fs-cluster/network"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"time"
)

func Network(lc fx.Lifecycle, cfg *network.NetConfig) (*network.Network, error) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: nil,
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})
	return network.NewNetwork(ctx, *cfg)
}

func RaftConfig(js Config) *raft.Config {
	cfg := raft.DefaultConfig()
	cfg.LogLevel = js.Raft.LogLevel
	cfg.LocalID = raft.ServerID(js.P2P.Identity.PeerID)
	return cfg
}

func DataStore(lc fx.Lifecycle, js Config) (*datastore.BadgerStore, error) {
	d, err := datastore.NewBadgerStore(js.DBPath)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: nil,
		OnStop: func(ctx context.Context) error {
			return d.Close()
		},
	})
	return d, err
}

func SnapshotStore() (raft.SnapshotStore, error) {
	return raft.NewFileSnapshotStore("snapshot", 5, nil)
}

func Fsm(store *datastore.BadgerStore, api *httpapi.HttpApi) (*consensus.Fsm, error) {
	return consensus.NewFsm(store, api)
}

func IpfsClient(js Config) (*httpapi.HttpApi, error) {
	addr, err := ma.NewMultiaddr(js.Ipfs)
	if err != nil {
		return nil, err
	}
	return httpapi.NewApi(addr)
}

func Transport(n *network.Network) (raft.Transport, error) {
	return p2praft.NewLibp2pTransport(n.Host(), time.Minute*2)
}

func Raft(conf *raft.Config, fsm *consensus.Fsm, snaps raft.SnapshotStore, trans raft.Transport, badger *datastore.BadgerStore, js Config) (*raft.Raft, error) {
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
	http.RegisterGreeterServer(s1, &http.Server{})
	go s1.Serve(listener)
	return s1
}

func Node(lc fx.Lifecycle, r *raft.Raft, fsm *consensus.Fsm, js Config, net *network.Network) (*consensus.Node, error) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: nil,
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})
	return consensus.NewNode(ctx, r, fsm, js.P2P.Identity.PeerID, net)
}

//type Clients struct {
//	c map[string]http.GreeterClient
//}
//
//func (client Clients) Client(id string) http.GreeterClient {
//	if c, ok := client.c[id]; !ok {
//		return nil
//	} else {
//		return c
//	}
//}
//
//func RpcClients(n *network.Network, js Config) (*Clients, error) {
//	c := make(map[string]http.GreeterClient)
//	for i := 0; i < len(js.Raft.Peers); i++ {
//		if js.Raft.Peers[i] == js.P2P.Identity.PeerID {
//			continue
//		}
//		conn, err := n.Connect(n.Context(), js.Raft.Peers[i])
//		if err != nil {
//			return nil, err
//		}
//		c[js.Raft.Peers[i]] = http.NewGreeterClient(conn)
//	}
//	return &Clients{c: c}, nil
//}
