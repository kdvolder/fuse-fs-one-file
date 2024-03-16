package pkg

import (
	"context"
	"fmt"
	"os"
	"path"

	"bazil.org/fuse"
)

const MEGABYTE = 1024 * 1024
const TERRABYTE = 1024 * MEGABYTE
const DFLT_BLOCKSIZE = 4 * MEGABYTE

type Storage struct {
	path      string
	size      uint64
	numBlocks uint64
	blockSize uint
}

func NewStorage(path string, size uint64, blockSize uint) Storage {
	numBlocks := size / uint64(blockSize)
	if size < numBlocks*uint64(blockSize) {
		numBlocks = numBlocks + 1
	}
	return Storage{
		path:      path,
		size:      size,
		numBlocks: numBlocks,
		blockSize: blockSize,
	}
}

func (s *Storage) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data := resp.Data[0:cap(resp.Data)]
	offset := uint64(max(0, req.Offset))
	to_read := min(uint64(req.Size), s.size)
	total_read := 0
	for to_read > 0 {
		blk := int64(offset) / int64(s.blockSize)
		f := s.getExistingBlockFile(blk)
		offset_in_blk := uint(offset) - uint(blk)*uint(s.blockSize)
		var len_in_blk int = int(s.blockSize) - int(offset_in_blk)
		len_in_blk = min(len_in_blk, int(to_read))
		if f != nil {
			defer f.Close()
			readCount, err := f.ReadAt(data[total_read:total_read+len_in_blk], int64(offset_in_blk))
			total_read += readCount
			offset += uint64(readCount)
			to_read -= uint64(readCount)
			resp.Data = data[0:total_read]
			if err != nil {
				// expected when blocks are not full-length
				zeroFill := len_in_blk - readCount
				offset += uint64(zeroFill)
				to_read -= uint64(zeroFill)
				for zeroFill > 0 {
					resp.Data = append(resp.Data, 0)
					zeroFill -= 1
				}
			}
			f.Close()
		} else { // f == nil, blockfile not yet created. So we assume this data is all zeros
			total_read += len_in_blk
			to_read -= uint64(len_in_blk)
			offset += uint64(len_in_blk)
			resp.Data = data[0:total_read]
			toFill := resp.Data[(len(resp.Data) - len_in_blk):]
			for i := range toFill {
				toFill[i] = 0
			}
		}
	}
	return nil
}

func (s *Storage) Size() uint64 {
	return s.size
}

func (s *Storage) blkFileName(blk int64) string {
	return path.Join(s.path, fmt.Sprintf("block_%06d", blk))
}

func (s *Storage) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	data := req.Data
	offset := req.Offset
	for len(data) > 0 {
		blk := offset / int64(s.blockSize)
		blk_f, err := os.OpenFile(s.blkFileName(blk), os.O_RDWR, 0o600)
		if err != nil {
			blk_f, err = s.createBlockFile(blk)
		}
		if blk_f != nil {
			defer blk_f.Close()
		}
		if err != nil {
			return err
		}

		offset_in_blk := int(offset % int64(s.blockSize))
		var len_in_blk int = int(s.blockSize) - int(offset_in_blk)
		len_to_write := min(len_in_blk, len(data))
		written, err := blk_f.WriteAt(data[:len_to_write], int64(offset_in_blk))
		if err != nil {
			return err
		}
		data = data[written:]
		offset = offset + int64(written)
		resp.Size += written
		blk_f.Close()
	}

	return nil
}

func (s *Storage) createBlockFile(blk int64) (*os.File, error) {
	f, err := os.Create(s.blkFileName(blk))
	if err != nil {
		return f, err
	}
	return f, err
}

// Get an exiting block file and opens it in read_only mode.
func (s *Storage) getExistingBlockFile(blk int64) *os.File {
	f, _ := os.Open(s.blkFileName(blk))
	return f
}
