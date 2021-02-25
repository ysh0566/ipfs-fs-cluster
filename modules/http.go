package modules

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ysh0566/ipfs-fs-cluster/consensus"
)

type Op struct {
	Op     string   `json:"op"`
	Params []string `json:"params"`
	Root   string   `json:"root"`
}

func Server2(node *consensus.Node, config Config) {
	router := gin.Default()

	// Query string parameters are parsed using the existing underlying request object.
	// The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe
	router.POST("/fs", func(c *gin.Context) {
		d, err := c.GetRawData()
		if err != nil {
			return
		}
		op := &Op{}
		err = json.Unmarshal(d, op)
		if err != nil {
			c.JSON(200, err.Error())
		}
		switch op.Op {
		case "ls":
			n, _ := node.Ls(c, op.Params[0])
			c.JSON(200, n)
		case "cp":
			err := node.Op(c, consensus.Operation_CP, op.Params[0], op.Params[1])
			if err != nil {
				c.JSON(200, err.Error())
			} else {
				c.JSON(200, "success")
			}
		case "mv":
			err := node.Op(c, consensus.Operation_MV, op.Params[0], op.Params[1])
			if err != nil {
				c.JSON(200, err.Error())
			} else {
				c.JSON(200, "success")
			}
		case "rm":
			err := node.Op(c, consensus.Operation_RM, op.Params[0])
			if err != nil {
				c.JSON(200, err.Error())
			} else {
				c.JSON(200, "success")
			}
		case "mkdir":
			err := node.Op(c, consensus.Operation_MKDIR, op.Params[0])
			if err != nil {
				c.JSON(200, err.Error())
			} else {
				c.JSON(200, "success")
			}
		default:
			c.JSON(200, "???")
		}
	})
	go router.Run(fmt.Sprintf(":%d", config.Port))
}
