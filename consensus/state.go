package consensus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/ysh0566/ipfs-fs-cluster/datastore"
	"io"
	"io/ioutil"
	"os"
	gopath "path"
	"sync"
)

var ErrParamsNum = errors.New("params num error")

type FileTreeState struct {
	dag   format.DAGService
	root  *mfs.Root
	store *datastore.BadgerStore
	ctx   context.Context
	copy  *FileTreeState
	once  sync.Once
	index uint64
}

func (fs *FileTreeState) Ls(ctx context.Context, path string) ([]mfs.NodeListing, error) {
	fsn, err := mfs.Lookup(fs.root, path)
	if err != nil {
		return nil, err
	}
	switch fsn := fsn.(type) {
	case *mfs.Directory:
		return fsn.List(ctx)
	case *mfs.File:
		_, name := gopath.Split(path)
		node := make([]mfs.NodeListing, 1)
		node[0] = mfs.NodeListing{
			Name: name,
		}
		if size, err := fsn.Size(); err == nil {
			node[0].Size = size
		}
		if nd, err := fsn.GetNode(); err == nil {
			node[0].Hash = nd.Cid().String()
		}
		return node, nil
	default:
		return nil, errors.New("unrecognized type")
	}
}

func (fs *FileTreeState) resolvePath(ctx context.Context, path string) (format.Node, error) {
	if len(path) > 0 && path[0] == '/' {
		fsNode, err := mfs.Lookup(fs.root, path)
		if err != nil {
			return nil, err
		}
		return fsNode.GetNode()
	}
	id, err := cid.Decode(path)
	if err != nil {
		return nil, err
	}
	return fs.dag.Get(ctx, id)
}

func (fs *FileTreeState) Cp(ctx context.Context, dir string, path string) error {
	node, err := fs.resolvePath(ctx, path)
	if err != nil {
		return err
	}
	err = mfs.PutNode(fs.root, dir, node)
	if err != nil {
		return err
	}
	_, err = mfs.FlushPath(ctx, fs.root, "/")
	return err
}

func (fs *FileTreeState) Mv(ctx context.Context, src string, dst string) error {
	src, err := checkPath(src)
	if err != nil {
		return err
	}
	dst, err = checkPath(dst)
	if err != nil {
		return err
	}
	err = mfs.Mv(fs.root, src, dst)
	if err != nil {
		return err
	}
	_, err = mfs.FlushPath(ctx, fs.root, "/")
	return err
}

func (fs *FileTreeState) Mkdir(src string) error {
	src, err := checkPath(src)
	if err != nil {
		return err
	}
	return mfs.Mkdir(fs.root, src, mfs.MkdirOpts{
		Mkparents:  true,
		Flush:      true,
		CidBuilder: fs.root.GetDirectory().GetCidBuilder(),
	})
}

func (fs *FileTreeState) Rm(path string) error {
	dir, name := gopath.Split(path)

	pdir, err := getParentDir(fs.root, dir)
	if err != nil {
		if err == os.ErrNotExist {
			return nil
		}
		return fmt.Errorf("parent lookup: %s", err)
	}
	err = pdir.Unlink(name)
	if err != nil {
		if err == os.ErrNotExist {
			return nil
		}
		return err
	}
	return pdir.Flush()
}

func (fs *FileTreeState) Flush() error {
	return fs.store.StoreState(fs.String())
}

func (fs *FileTreeState) Root() (string, error) {
	n, err := fs.root.GetDirectory().GetNode()
	if err != nil {
		return "", err
	}
	return n.Cid().String(), nil
}

func (fs *FileTreeState) MustGetRoot() string {
	n, err := fs.root.GetDirectory().GetNode()
	if err != nil {
		panic(err)
	}
	return n.Cid().String()
}

func (fs *FileTreeState) rpcOp(ctx context.Context, code Operation_Code, params []string) error {
	fmt.Println("code", code)
	switch code {
	case Operation_CP:
		if len(params) != 2 {
			return ErrParamsNum
		}
		return fs.Cp(ctx, params[0], params[1])
	case Operation_MV:
		if len(params) != 2 {
			return ErrParamsNum
		}
		return fs.Mv(ctx, params[0], params[1])
	case Operation_RM:
		if len(params) != 1 {
			return ErrParamsNum
		}
		return fs.Rm(params[0])
	case Operation_MKDIR:
		if len(params) != 1 {
			return ErrParamsNum
		}
		return fs.Mkdir(params[0])
	default:
		return errors.New("unrecognized operation")
	}
}

func (fts *FileTreeState) Marshal(writer io.Writer) error {
	_, err := writer.Write([]byte(fts.String()))
	return err
}

func (fts *FileTreeState) String() string {
	d := struct {
		Index uint64 `json:"index"`
		Root  string `json:"root"`
	}{
		Index: fts.index,
		Root:  fts.MustGetRoot(),
	}
	data, _ := json.Marshal(d)
	return string(data)
}

func (fts *FileTreeState) Copy(ctx context.Context) (*FileTreeState, error) {
	var err error
	fts.once.Do(func() {
		var raw format.Node
		var root *mfs.Root
		raw, err = fts.root.GetDirectory().GetNode()
		if err != nil {
			return
		}
		node, ok := raw.(*merkledag.ProtoNode)
		if !ok {
			panic(err)
		}
		root, err = mfs.NewRoot(ctx, fts.dag, node, func(ctx context.Context, cid cid.Cid) error {
			return nil
		})
		fts.copy = &FileTreeState{
			dag:   fts.dag,
			root:  root,
			store: fts.store,
			ctx:   ctx,
		}
	})
	if err != nil {
		return nil, err
	}
	return fts.copy, nil
}

func WalkDirectory(ctx context.Context, dir *mfs.Directory, visited map[string]bool) error {
	ls, err := dir.List(ctx)
	if err != nil {
		return err
	}
	for _, node := range ls {
		node, err := dir.Child(node.Name)
		if err != nil {
			return err
		}
		if fnode, err := node.GetNode(); err != nil {
			return err
		} else {
			if _, ok := visited[fnode.Cid().String()]; ok {
				return nil
			} else {
				visited[fnode.Cid().String()] = true
			}
		}
		if node.Type() == mfs.TDir {
			err = WalkDirectory(ctx, node.(*mfs.Directory), visited)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (fts *FileTreeState) Unmarshal(reader io.Reader) error {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	state := struct {
		Index uint64 `json:"index"`
		Root  string `json:"root"`
	}{}
	if err = json.Unmarshal(bs, &state); err != nil {
		return err
	}
	c, err := cid.Decode(state.Root)
	if err != nil {
		return err
	}
	raw, err := fts.dag.Get(fts.ctx, c)
	if err != nil {
		return err
	}

	rootNode, ok := raw.(*merkledag.ProtoNode)
	if !ok {
		return errors.New("invalid root node")
	}
	r, err := mfs.NewRoot(fts.ctx, fts.dag, rootNode, func(ctx context.Context, cid cid.Cid) error {
		return nil
	})
	if err != nil {
		return err
	}
	visited := make(map[string]bool)
	err = WalkDirectory(fts.ctx, r.GetDirectory(), visited)
	if err != nil {
		return err
	}
	fts.root = r
	fts.index = state.Index
	return nil
}

func checkPath(p string) (string, error) {
	if len(p) == 0 {
		return "", fmt.Errorf("paths must not be empty")
	}

	if p[0] != '/' {
		return "", fmt.Errorf("paths must start with a leading slash")
	}

	cleaned := gopath.Clean(p)
	if p[len(p)-1] == '/' && p != "/" {
		cleaned += "/"
	}
	return cleaned, nil
}

func getParentDir(root *mfs.Root, dir string) (*mfs.Directory, error) {
	parent, err := mfs.Lookup(root, dir)
	if err != nil {
		return nil, err
	}

	pdir, ok := parent.(*mfs.Directory)
	if !ok {
		return nil, errors.New("expected *mfs.Directory, didn't get it. This is likely a race condition")
	}
	return pdir, nil
}
