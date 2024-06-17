package mmap

import (
	pkgErrors "github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// File represents a memory-mapped file.
type File struct {
	BlockDevice     *BlockDevice
	SectorSizeBytes int
	SectorCount     int64
}

// NewFile constructs a new instance of File.
func NewFile(path string, minimumSizeBytes int) (*File, error) {
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_APPEND, 0666)
	if err != nil {
		return nil, pkgErrors.Wrapf(err, "failed to open file %#v", path)
	}

	// Use the block size returned by fstat() to determine the
	// sector size and the number of sectors needed to store the
	// desired amount of space.
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return nil, pkgErrors.Wrapf(err, "failed to obtain size of file %#v", path)
	}

	sectorSizeBytes := int(stat.Blksize)
	sectorCount := int64((uint64(minimumSizeBytes) + uint64(stat.Blksize) - 1) / uint64(stat.Blksize))
	sizeBytes := int64(sectorSizeBytes) * sectorCount

	if err := unix.Ftruncate(fd, sizeBytes); err != nil {
		return nil, pkgErrors.Wrapf(err, "failed to truncate file %#v to %d bytes", path, sizeBytes)
	}

	bd, err := NewBlockDevice(fd, int(sizeBytes))

	if err != nil {
		return nil, err
	} else if err := unix.Close(fd); err != nil {
		return nil, err
	}

	return &File{
		BlockDevice:     bd,
		SectorSizeBytes: sectorSizeBytes,
		SectorCount:     sectorCount,
	}, nil
}
