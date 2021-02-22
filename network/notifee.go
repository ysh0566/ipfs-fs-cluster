package network

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerHandler struct {
	host host.Host
	ctx  context.Context
}

func (p *PeerHandler) HandlePeerFound(info peer.AddrInfo) {
	_ = p.host.Connect(p.ctx, info)
}
