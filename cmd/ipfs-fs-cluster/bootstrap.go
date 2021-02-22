package main

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/ysh0566/ipfs-fs-cluster/modules"
	"go.uber.org/fx"
)

type App struct {
	*fx.App
	Clients modules.Clients
	Raft    *raft.Raft
}

func New(opts ...fx.Option) *App {

	app := fx.New(opts...)
	if err := app.Err(); err != nil {
		fmt.Printf("fx.New failed: %v", err)
		panic(err)
	}
	return &App{
		App: app,
	}
}
