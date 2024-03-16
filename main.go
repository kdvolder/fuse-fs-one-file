package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/kdvolder/fuse-fs-one-file/pkg"
)

const THE_FILE = "disk.img"

var progName = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "  %s storagePath MOUNTPOINT\n", progName)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}
	path := flag.Arg(0)
	mountpoint := flag.Arg(1)
	if err := mount(path, mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(storagePath, mountpoint string) error {
	c, err := fuse.Mount(mountpoint)
	if c != nil {
		defer c.Close()
		fuse.Unmount(mountpoint)
	}
	if err != nil {
		return err
	}

	filesys := &FS{
		storage: pkg.NewStorage(storagePath, 26, 9),
	}
	filesys.root = &Dir{filesys}
	return fs.Serve(c, filesys)
}

////////// Filesystem ////////////////////////////////////

type FS struct {
	storage pkg.Storage
	root    *Dir
}

var _ fs.FS = (*FS)(nil)

func (f *FS) Root() (fs.Node, error) {
	return f.root, nil
}

/////////// Dir ///////////////////////////////////////////

// Our filesysten only has a single file in the root directory, which means
// it only has a single directory, which is the root directory. We don't need any
// info to determine the directory path since there is only one.
type Dir struct {
	fs *FS
}

var _ fs.Node = (*Dir)(nil)

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0o555
	return nil
}

var dirDirs = []fuse.Dirent{
	{Inode: 2, Name: THE_FILE, Type: fuse.DT_File},
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return dirDirs, nil
}

func (dir Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == THE_FILE {
		return File{storage: &dir.fs.storage}, nil
	}
	return nil, syscall.ENOENT
}

///////// File /////////////////////////////////////////////////////////

type File struct {
	storage *pkg.Storage
}

var _ fs.Node = File{}
var _ fs.HandleReader = File{}
var _ fs.HandleWriter = File{}
var _ fs.NodeSetattrer = File{}

func (f File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0o600
	a.Size = f.storage.Size()
	return nil
}

// implements fs.HandleReader
func (f File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	return f.storage.Read(ctx, req, resp)
}

// implements fs.HandleWriter.
func (f File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	return f.storage.Write(ctx, req, resp)
}

// implements fs.NodeSetattrer.
func (f File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	return nil
}
