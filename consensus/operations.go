package consensus

type OpType int

const (
	OpRm OpType = iota
	OpMv
	OpCp
	OpMkdir
)

type FsOperation struct {
	Op     OpType   `json:"op"`
	Params []string `json:"params"`
	Root   string   `json:"root"`
}
