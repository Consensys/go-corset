package mmap

import (
	"errors"
	"io"
	"runtime/debug"
	"syscall"

	pkgErrors "github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// BlockDevice represents a mmap block device holding a reference to a file descriptor.
type BlockDevice struct {
	FileDescriptor int
	Data           []byte
}

// NewBlockDevice creates a BlockDevice from a file
// descriptor referring either to a regular file or UNIX device node. To
// speed up reads, a memory map is used.
func NewBlockDevice(fileDescriptor, sizeBytes int) (*BlockDevice, error) {
	data, err := unix.Mmap(fileDescriptor, 0, sizeBytes, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "failed to memory map block device")
	}

	return &BlockDevice{
		FileDescriptor: fileDescriptor,
		Data:           data,
	}, nil
}

// ReadAt reads through the memory map at a given offset.
func (bd *BlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	// Let read actions go through the memory map to prevent system
	// call overhead for commonly requested objects.
	if off < 0 {
		return 0, syscall.EINVAL
	}

	if off > int64(len(bd.Data)) {
		return 0, io.EOF
	}
	// Install a page fault handler, so that I/O errors against the
	// memory map (e.g., due to disk failure) don't cause us to
	// crash.
	old := debug.SetPanicOnFault(true)
	defer func() {
		debug.SetPanicOnFault(old)

		if recover() != nil {
			err = errors.New("page fault occurred while reading from memory map")
		}
	}()

	n = copy(p, bd.Data[off:])
	if n < len(p) {
		err = io.EOF
	}

	return
}

// WriteAt writes at a given offset.
func (bd *BlockDevice) WriteAt(p []byte, off int64) (int, error) {
	// Let write actions go through the file descriptor. Doing so
	// yields better performance, as writes through a memory map
	// would trigger a page fault that causes data to be read.
	//
	// The pwrite() system call cannot return a size and error at
	// the same time. If an error occurs after one or more bytes are
	// written, it returns the size without an error (a "short
	// write"). As WriteAt() must return an error in those cases, we
	// must invoke pwrite() repeatedly.
	//
	// TODO: Maybe it makes sense to let unaligned writes that would
	// trigger reads anyway to go through the memory map?
	nTotal := 0

	for len(p) > 0 {
		n, err := unix.Pwrite(bd.FileDescriptor, p, off)
		nTotal += n

		if err != nil {
			return nTotal, err
		}

		p = p[n:]
		off += int64(n)
	}

	return nTotal, nil
}

// Sync synchronizes a file's in-core state with storage device.
func (bd *BlockDevice) Sync() error {
	return unix.Fsync(bd.FileDescriptor)
}
