package test

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"

	"github.com/consensys/go-corset/pkg/mmap"

	"github.com/stretchr/testify/require"
)

func Ignored_TestNewBlockDeviceFromFile(t *testing.T) {
	minSizeBytes := 123456
	blockDevicePath := filepath.Join(t.TempDir(), "test_blockdevice")

	println(blockDevicePath)

	mmapFile, err := mmap.NewFile(blockDevicePath, minSizeBytes)
	require.NoError(t, err)

	sectorSizeBytes := mmapFile.SectorSizeBytes
	sectorCount := mmapFile.SectorCount
	blockDevice := mmapFile.BlockDevice
	// The sector size should be a power of two, and the number of
	// sectors should be sufficient to hold the required space.
	require.LessOrEqual(t, 512, sectorSizeBytes)
	require.Equal(t, 0, sectorSizeBytes&(sectorSizeBytes-1))
	require.Equal(t, int64((minSizeBytes+sectorSizeBytes-1)/sectorSizeBytes), sectorCount)

	// The file on disk should have a size that corresponds to the
	// sector size and count.
	fileInfo, err := os.Stat(blockDevicePath)
	require.NoError(t, err)
	require.Equal(t, int64(sectorSizeBytes)*sectorCount, fileInfo.Size())

	// Test read, write and sync operations.
	n, err := blockDevice.WriteAt([]byte("Hello"), 12345)
	require.Equal(t, 5, n)
	require.NoError(t, err)

	var b [16]byte
	n, err = blockDevice.ReadAt(b[:], 12340)
	require.Equal(t, 16, n)
	require.NoError(t, err)
	require.Equal(t, []byte("\x00\x00\x00\x00\x00Hello\x00\x00\x00\x00\x00\x00"), b[:])

	require.NoError(t, mmapFile.BlockDevice.Sync())

	// Truncating the file will cause future read access to the
	// memory map underneath the BlockDevice to raise SIGBUS. This
	// may also occur in case of actual I/O errors. These page
	// faults should be caught properly.
	//
	// To be able to implement this, ReadAt() temporary enables the
	// debug.SetPanicOnFault() option. Test that the original value
	// of this option is restored upon completion.
	require.NoError(t, os.Truncate(blockDevicePath, 0))

	debug.SetPanicOnFault(false)

	n, err = blockDevice.ReadAt(b[:], 12340)
	require.NoError(t, err)

	require.False(t, debug.SetPanicOnFault(false))
	require.Equal(t, 0, n)
	debug.SetPanicOnFault(true)

	n, err = blockDevice.ReadAt(b[:], 12340)

	require.True(t, debug.SetPanicOnFault(false))
	require.Equal(t, 0, n)
	require.Error(t, err, "page fault occurred while reading from memory map")
}
