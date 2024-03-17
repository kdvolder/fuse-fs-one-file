package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"bazil.org/fuse"
)

type test_mount struct {
	mntPoint string
	storage  string
}

func (m *test_mount) Start() {
	mount(m.storage, m.mntPoint, 26, 9)
}

func (m *test_mount) Stop() {
	fuse.Unmount(m.mntPoint)
	// os.RemoveAll(m.mntPoint)
	// os.RemoveAll(m.storage)
}

func withTestMount(do_stuff func(*test_mount)) {
	mountDir, err := os.MkdirTemp("/tmp", "mnt")
	if err != nil {
		panic(err)
	}
	storeDir, err := os.MkdirTemp("/tmp", "store")
	if err != nil {
		panic(err)
	}
	mnt := test_mount{
		mntPoint: mountDir,
		storage:  storeDir,
	}
	go mnt.Start()
	defer mnt.Stop()
	info, err := os.Stat(path.Join(mnt.mntPoint, "disk.img"))
	for err != nil {
		info, err = os.Stat(path.Join(mnt.mntPoint, "disk.img"))
	}
	fmt.Printf("%v", info)
	do_stuff(&mnt)
}

func (m *test_mount) file(p string) string {
	return path.Join(m.mntPoint, p)
}

func Test_readFile(t *testing.T) {
	withTestMount(func(mnt *test_mount) {
		data, err := os.ReadFile(path.Join(mnt.mntPoint, "disk.img"))
		if err != nil {
			t.Fatalf("failed to read %v", err)
		}
		t.Logf("read = %s", string(data))
	})
}

func Test_writeFile(t *testing.T) {
	withTestMount(func(m *test_mount) {
		f, err := os.OpenFile(m.file("disk.img"), os.O_RDWR, 0o600)
		if f != nil {
			defer f.Close()
		}
		if err != nil {
			t.Fatalf("failed to open %v", err)
		}

		info, err := f.Stat()
		if err != nil {
			t.Fatalf("Failed to write %v", err)
		}
		fileSize := info.Size()
		zees := make([]byte, fileSize-2)
		for i := range zees {
			zees[i] = 'Z'
		}
		count, err := f.WriteAt(zees, 0)
		if err != nil {
			t.Fatalf("Failed to write %v", err)
		}
		if count != int(len(zees)) {
			t.Fatalf("Wrong number written %d", count)
		}

		bees := "BBBB"
		bytes := []byte(bees)
		count, err = f.WriteAt(bytes, 7)
		if err != nil {
			t.Fatalf("Failed to write %v", err)
		}
		if count != len(bytes) {
			t.Fatalf("Wrong number of written bytes %d", count)
		}

		contents, err := os.ReadFile(m.file("disk.img"))
		if err != nil {
			t.Fatalf("Failed to read %v", err)
		}
		result := string(contents[0 : len(contents)-2])
		if result != "ZZZZZZZBBBBZZZZZZZZZZZZZ" {
			t.Fatalf("The file contents is wrong: %s", result)
		}
		if contents[len(contents)-1] != 0 {
			t.Fatalf("The final 0 is wrong: %s", string(contents))
		}
		if contents[len(contents)-2] != 0 {
			t.Fatalf("The final 0 is wrong: %s", string(contents))
		}
	})
}

type shadow_file struct {
	testFile *os.File
	realFile *os.File
}
