package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

func executeLs() ([]byte, error) {
	cmd := exec.Command("ls", "-l")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

type fiemap_extent struct {
	fe_logical    uint64
	fe_physical   uint64
	fe_length     uint64
	fe_reserved64 [2]uint64
	fe_flags      uint32
	fe_reserved   [3]uint64
} // 56 bytes

type fiemap struct {
	fm_start          uint64
	fm_length         uint64
	fm_flags          uint32
	fm_mapped_extents uint32
	fm_extent_count   uint32
	fm_reserved       uint32
	fm_extents        [1 << 10]fiemap_extent
} // 32 bytes

func ioctl() {
	file, err := os.Open("/mnt/f2fs/normal.0.0.jpg")
	if err != nil {
		fmt.Printf("Error: open: %s\n", err)
		return
	}
	if err != nil {
		fmt.Printf("Error: stat: %s\n", err)
		return
	}
	fiemap_result := fiemap{fm_start: 0, fm_length: ^uint64(0), fm_flags: 0x1, fm_mapped_extents: 0, fm_extent_count: 0, fm_reserved: 0}
	ptr := uintptr(unsafe.Pointer(&fiemap_result))

	r1, r2, err := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), 3223348747, ptr)
	if errors.Is(err, os.ErrInvalid) {
		fmt.Printf("Error: fiemap: %s\n", err)
		return
	}
	fmt.Printf("syscall %d %d\n", r1, r2)
	fmt.Printf("mapped_extents: %d\n", fiemap_result.fm_mapped_extents)
	fiemap_extents_temp := make([]byte, 32+(56*fiemap_result.fm_mapped_extents))
	fiemap_extents_ptr := unsafe.Pointer(&fiemap_extents_temp)
	fiemap_extents := (*fiemap)(fiemap_extents_ptr)
	fiemap_extents.fm_start = 0
	fiemap_extents.fm_length = ^uint64(0)
	fiemap_extents.fm_flags = 0x1
	fiemap_extents.fm_extent_count = fiemap_result.fm_mapped_extents
	fiemap_extents.fm_mapped_extents = 0
	r1, r2, err = syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), 3223348747, uintptr(unsafe.Pointer(&fiemap_extents_temp)))
	if errors.Is(err, os.ErrInvalid) {
		fmt.Printf("Error: fiemap: %s\n", err)
		return
	}
	// fmt.Printf("%v\n", fiemap_extents)
	// fiemap_extents_array :=
	for i := 0; i < int(fiemap_extents.fm_extent_count); i++ {
		extent := fiemap_extents.fm_extents[i]
		fmt.Printf("extent: %v\n", extent)
		// fmt.Printf("extent: %d %d %d %d\n", extent.fe_logical, extent.fe_physical, extent.fe_length, extent.fe_flags)
	}
	// for i := 0; i < len(extents); i++ {
	// 	extent := extents[i]
	// 	fmt.Printf("extent: %v\n", extent)
	// }
}

func main() {
	// out, err := executeLs()
	// if err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// 	return
	// }
	ioctl()
	// fmt.Printf("Output: %s\n", out)
}
