package consensus

type OpType string

const (
	OpRm OpType = "OpRm"
	OpMv OpType = "OpMv"
	OpCp OpType = "OpCp"
)

type FsOperation struct {
	Op     OpType
	Params []string
}
