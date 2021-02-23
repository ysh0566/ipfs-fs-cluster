package modules

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"time"
)

type Op struct {
	Op     int      `json:"op"`
	Params []string `json:"params"`
	Root   string   `json:"root"`
}

func Server2(raft *raft.Raft, fsm raft.FSM, config Config) {
	router := gin.Default()

	// Query string parameters are parsed using the existing underlying request object.
	// The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe
	router.POST("/fs", func(c *gin.Context) {
		d, err := c.GetRawData()
		if err != nil {
			return
		}
		if string(raft.Leader()) == config.P2P.Identity.PeerID {
			raft.Apply(d, time.Second*60)
		}

	})
	go router.Run(fmt.Sprintf(":%d", config.Port))
}
