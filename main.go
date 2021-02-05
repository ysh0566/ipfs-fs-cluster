package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipns"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	relay "github.com/libp2p/go-libp2p-circuit"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/libp2p/go-libp2p-core/routing"
	gostream "github.com/libp2p/go-libp2p-gostream"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	noise "github.com/libp2p/go-libp2p-noise"
	p2praft "github.com/libp2p/go-libp2p-raft"
	record "github.com/libp2p/go-libp2p-record"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/ysh0566/ipfs-fs-cluster/consensus"
	"google.golang.org/grpc"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	DefaultConfigCrypto    = crypto.ECDSA
	DefaultConfigKeyLength = -1
)

var logger = logging.Logger("cluster")

var cfg = struct {
	ConnMgr struct {
		LowWater    int
		HighWater   int
		GracePeriod time.Duration
	}
	EnableRelayHop bool
	Secret         pnet.PSK
}{}

func init() {
	cfg.ConnMgr.LowWater = 100
	cfg.ConnMgr.LowWater = 400
	cfg.ConnMgr.GracePeriod = 2 * time.Minute
	cfg.Secret, _ = hex.DecodeString("2fa6063a945f2ae5fc9c40648601e8fe0de73466463743ea198d74eee67a97fd")
}

var keys = []string{
	"08031279307702010104207e37a50b2570d10a49b519a6f55f1c8e48724cc597120231f545c93d3710d383a00a06082a8648ce3d030107a144034200045e3822bb139fca42876aae1f408b43f5ff76331f9efcf337fabc979c01752c56b36c626f14c9b8c1d66d7dd9172ac8ed11a87a840ecff5726fa5c91f9aecf072",
	"08031279307702010104203ab4b28a95e66e78926f9ee20b72246f8de87b70769b0b95f9a1ac3744ad5fb9a00a06082a8648ce3d030107a14403420004dc127498a2d58c4cbf1d5274bd88deccdeab8db512c991d10e7bb87e506fdbec396907cc1a6897621a36c911fcf1b9ac8bd099e6abf9fc2fe6d90e9e22d5bc1c",
	"0803127930770201010420a9fdfd7c48a6bfd16d4c6cf5c03ecc212d549da2191fd93a9c3818eb21d62aeca00a06082a8648ce3d030107a14403420004f2bcce13879aafe4c603caaf516b3c173605db8da72a8187c335822a7f48183712a96129dc5a711f69be29d67e5d50a6e053b663cccaf4e906f9812ac53625fb",
}

var addrs = []string{
	"localhost:10086",
	"localhost:10087",
	"localhost:10088",
	"localhost:10089",
}

func newDHT(ctx context.Context, h host.Host, store datastore.Datastore, extraopts ...dual.Option) (*dual.DHT, error) {
	opts := []dual.Option{
		dual.DHTOption(dht.NamespacedValidator("pk", record.PublicKeyValidator{})),
		dual.DHTOption(dht.NamespacedValidator("ipns", ipns.Validator{KeyBook: h.Peerstore()})),
		dual.DHTOption(dht.Concurrency(10)),
	}

	opts = append(opts, extraopts...)

	//if batchingDs, ok := store.(datastore.Batching); ok {
	//	dhtDatastore := namespace.Wrap(batchingDs, datastore.NewKey("dht"))
	//	opts = append(opts, dual.DHTOption(dht.Datastore(dhtDatastore)))
	//	logger.Debug("enabling DHT record persistence to datastore")
	//}

	return dual.New(ctx, h, opts...)
}

func baseOpts(psk pnet.PSK) []libp2p.Option {
	return []libp2p.Option{
		libp2p.PrivateNetwork(psk),
		libp2p.EnableNATService(),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// TODO: quic does not support private networks
		// libp2p.Transport(libp2pquic.NewTransport),
		libp2p.DefaultTransports,
	}
}

func NewHost(ctx context.Context, addr []ma.Multiaddr, priv crypto.PrivKey) (host.Host, error) {
	relayOpts := []relay.RelayOpt{}
	if cfg.EnableRelayHop {
		relayOpts = append(relayOpts, relay.OptHop)
	}
	connman := connmgr.NewConnManager(cfg.ConnMgr.LowWater, cfg.ConnMgr.HighWater, cfg.ConnMgr.GracePeriod)
	opts := []libp2p.Option{
		libp2p.ListenAddrs(addr...),
		libp2p.NATPortMap(),
		libp2p.ConnectionManager(connman),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err := newDHT(ctx, h, nil)
			return idht, err
		}),
		libp2p.EnableRelay(relayOpts...),
		libp2p.EnableAutoRelay(),
	}

	finalOpts := []libp2p.Option{
		libp2p.Identity(priv),
	}
	finalOpts = append(finalOpts, baseOpts(cfg.Secret)...)
	finalOpts = append(finalOpts, opts...)

	h, err := libp2p.New(
		ctx,
		finalOpts...,
	)
	return h, err
}

type PeerHandler struct {
	host host.Host
	ctx  context.Context
}

func (p *PeerHandler) HandlePeerFound(info peer.AddrInfo) {
	fmt.Println("id", info.ID)
	err := p.host.Connect(p.ctx, info)
	if err != nil {
		fmt.Println("???????", err.Error())
	}
}

const protocol = "/grpc/0.0.1"

func DialOption(hosts host.Host) grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		id, err := peer.Decode(s)
		if err != nil {
			return nil, errors.Wrapf(err, "parse peer.ID error %s", s)
		}
		return gostream.Dial(ctx, hosts, id, protocol)
	})
}

func ConnGrpc(ctx context.Context, id string, hosts host.Host) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, id, DialOption(hosts), grpc.WithInsecure())
}

func main() {

	s := os.Args[1]
	i, _ := strconv.Atoi(s)
	tmp, _ := hex.DecodeString(keys[i])
	priv, err := crypto.UnmarshalPrivateKey(tmp)
	if err != nil {
		panic(err)
	}
	//id, err := peer.IDFromPublicKey(priv.GetPublic())
	//if err != nil {
	//	panic(err)
	//}
	//t, err := raft.NewTCPTransport(ad, nil, 2, time.Second, nil)
	ctx := context.Background()
	mu, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/1004%d", i))
	if err != nil {
		panic(err)
	}
	addr := []ma.Multiaddr{mu}
	hosts, err := NewHost(ctx, addr, priv)
	if err != nil {
		panic(err)
	}

	tp, err := p2praft.NewLibp2pTransport(hosts, time.Minute*2)

	if err != nil {
		panic(err)
	}
	config := raft.DefaultConfig()
	config.ShutdownOnRemove = false

	//mock := &raft.MockFSMConfigStore{
	//	&consensus.Fsm{},
	//}

	lconfig := raft.Configuration{Servers: []raft.Server{
		raft.Server{
			Suffrage: 0,
			ID:       "Qme2D4VnKNZchsNBCqzsayPADzcQXi91NcNTfzXtSyyM6e",
			Address:  "Qme2D4VnKNZchsNBCqzsayPADzcQXi91NcNTfzXtSyyM6e",
		},
		raft.Server{
			Suffrage: 0,
			ID:       "QmcL3oFY7RWDN5ZcG9ChU2CmhBRaRZAbM5qv7K17PRSxnr",
			Address:  "QmcL3oFY7RWDN5ZcG9ChU2CmhBRaRZAbM5qv7K17PRSxnr",
		},
		raft.Server{
			Suffrage: 0,
			ID:       "QmdsWJRjmhdELnUqDMUhdTWTNCRj3eZzXuLP1KsuGenDMi",
			Address:  "QmdsWJRjmhdELnUqDMUhdTWTNCRj3eZzXuLP1KsuGenDMi",
		},
	}}

	config.LocalID = lconfig.Servers[i].ID
	//mock.StoreConfiguration(0, lconfig)
	snap := raft.NewInmemSnapshotStore()

	mdns, err := discovery.NewMdnsService(ctx, hosts, time.Second*20, "mdnsServiceTag")
	if err != nil {
		panic(err)
	}
	handle := PeerHandler{
		host: hosts,
		ctx:  ctx,
	}
	mdns.RegisterNotifee(&handle)

	store, err := consensus.NewBadgerStore(fmt.Sprintf("badger%d", i))
	if err != nil {
		panic(err)
	}
	raft.BootstrapCluster(config, store, store, snap, tp, lconfig)

	r, err := raft.NewRaft(config, &consensus.Fsm{}, store, store, snap, tp)
	if err != nil {
		panic(err)
	}

	r.BootstrapCluster(lconfig)

	ticker := time.NewTicker(time.Second * 3)
	for range ticker.C {
		fmt.Println(r.State())
		futuer := r.Apply([]byte("test test"), time.Second)
		if futuer.Error() == nil {
			fmt.Println("leader", futuer.Response())
		}
	}

	listener, err := gostream.Listen(hosts, "grpc")
	s1 := grpc.NewServer()
	s1.Serve(listener)

}
