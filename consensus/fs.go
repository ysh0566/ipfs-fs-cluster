package consensus

import (
	"context"
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
)

var ErrParamsNum = errors.New("params num error")

type FileTreeState struct {
	dag   format.DAGService
	root  *mfs.Root
	store *datastore.BadgerStore
	ctx   context.Context
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

		size, err := fsn.Size()
		if err != nil {
			return nil, err
		}

		nd, err := fsn.GetNode()
		if err != nil {
			return nil, err
		}
		node[0] = mfs.NodeListing{
			Name: name,
			Type: int(fsn.Type()),
			Size: size,
			Hash: nd.Cid().Hash().B58String(),
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

func (fs *FileTreeState) Commit() error {
	root, err := fs.Root()
	if err != nil {
		return err
	}
	return fs.store.StoreState(root)
}

func (fs *FileTreeState) Root() (string, error) {

	n, err := fs.root.GetDirectory().GetNode()
	if err != nil {
		return "", err
	}
	return n.Cid().Hash().B58String(), nil
}

func (fs *FileTreeState) Op(ctx context.Context, ops FsOperation) error {
	switch ops.Op {
	case OpCp:
		if len(ops.Params) != 2 {
			return ErrParamsNum
		}
		return fs.Cp(ctx, ops.Params[0], ops.Params[1])
	case OpMv:
		if len(ops.Params) != 2 {
			return ErrParamsNum
		}
		return fs.Mv(ctx, ops.Params[0], ops.Params[1])
	case OpRm:
		if len(ops.Params) != 1 {
			return ErrParamsNum
		}
		return fs.Rm(ops.Params[0])
	case OpMkdir:
		if len(ops.Params) != 1 {
			return ErrParamsNum
		}
		return fs.Mkdir(ops.Params[0])
	default:
		return errors.New("unrecognized operation")
	}
}

func (fts *FileTreeState) Marshal(writer io.Writer) error {
	r, err := fts.Root()
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(r))
	return err
}

func WalkDirectory(ctx context.Context, dir *mfs.Directory) error {
	return dir.ForEachEntry(ctx, func(listing mfs.NodeListing) error {
		node, err := dir.Child(listing.Name)
		if err != nil {
			return err
		}
		if node.Type() == mfs.TDir {
			err = WalkDirectory(ctx, node.(*mfs.Directory))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (fts *FileTreeState) Unmarshal(reader io.Reader) error {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	c, err := cid.Decode(string(bs))
	if err != nil {
		return err
	}
	raw, err := fts.dag.Get(fts.ctx, c)
	if err != nil {
		return err
	}
	dir, err := mfs.NewDirectory(fts.ctx, "tmp", raw, nil, fts.dag)
	if err != nil {
		return err
	}
	err = WalkDirectory(fts.ctx, dir)
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
	fts.root = r
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
