package extentparser

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/znsmemory"
)

const (
	FIEMAP_EXTENT_SIZE = 56
	FIEMAP_SIZE        = 32
	FS_IOC_FIEMAP      = 3223348747
	FIEMAP_MAX_OFFSET  = ^uint64(0)
	FIEMAP_FLAG_SYNC   = 0x0001
	FIEMAP_EXTENT_LAST = 0x0001
)

type zns_info struct {
	zns_start_blkaddr uint64
	zone_blocks       uint64
}

type fiemap_extent struct {
	fe_logical    uint64
	fe_physical   uint64
	fe_length     uint64
	fe_reserved64 [2]uint64
	fe_flags      uint32
	fe_reserved   [3]uint32
} // 56 bytes

type fiemap struct {
	fm_start          uint64
	fm_length         uint64
	fm_flags          uint32
	fm_mapped_extents uint32
	fm_extent_count   uint32
	fm_reserved       uint32
} // 32 bytes

type extent struct {
	logical  uint64
	physical uint64
	length   uint64
	flags    uint32
}

func dir_list(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil, err
	}
	ret := []string{}
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			ret = append(ret, entry.Name())
		}
	}
	return ret, nil
}

func get_zns_info(regular_device string, zns_device string) (zns_info, error) {
	out, err := exec.Command("dump.f2fs", fmt.Sprintf("/dev/%s", regular_device)).Output()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return zns_info{}, err
	}
	output := string(out)
	zns_blkaddr_pattern := regexp.MustCompile(fmt.Sprintf(`/dev/%s blkaddr = (\w+)`, zns_device))
	match := zns_blkaddr_pattern.FindStringSubmatch(output)
	if len(match) < 2 {
		return zns_info{}, errors.New("Cannot find zns blk addr")
	}
	zns_blkaddr, err := strconv.ParseInt(match[1], 16, 64)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return zns_info{}, err
	}
	zone_blocks_pattern := regexp.MustCompile(`(\d+) blocks per zone`)
	match = zone_blocks_pattern.FindStringSubmatch(output)
	if len(match) < 2 {
		return zns_info{}, errors.New("Cannot find zone blocks")
	}
	zone_blocks, err := strconv.Atoi(match[1])
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return zns_info{}, err
	}
	// fmt.Println(output)
	return zns_info{
		zns_start_blkaddr: uint64(zns_blkaddr),
		zone_blocks:       uint64(zone_blocks),
	}, nil
}

func get_extents(path string) ([]extent, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error: open: %s\n", err)
		return nil, err
	}
	if err != nil {
		fmt.Printf("Error: stat: %s\n", err)
		return nil, err
	}
	// check fm_mappd_extents count
	fiemap_result := fiemap{fm_start: 0, fm_length: FIEMAP_MAX_OFFSET, fm_flags: FIEMAP_FLAG_SYNC, fm_mapped_extents: 0, fm_extent_count: 0, fm_reserved: 0}
	ptr := uintptr(unsafe.Pointer(&fiemap_result))
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), FS_IOC_FIEMAP, ptr)
	if errors.Is(err, os.ErrInvalid) {
		fmt.Printf("Error: ioctl: %s\n", err)
		return nil, err
	}
	// allocate for actual extents count
	fiemap_extents := make([]fiemap_extent, fiemap_result.fm_mapped_extents+1)
	// index 0 element as fiemap
	fiemap_ptr := unsafe.Pointer(uintptr(unsafe.Pointer(&fiemap_extents[1])) - FIEMAP_SIZE)
	fiemap_struct := (*fiemap)(fiemap_ptr)
	fiemap_struct.fm_start = 0
	fiemap_struct.fm_length = FIEMAP_MAX_OFFSET
	fiemap_struct.fm_flags = FIEMAP_FLAG_SYNC
	fiemap_struct.fm_extent_count = fiemap_result.fm_mapped_extents
	fiemap_struct.fm_mapped_extents = 0
	// get extents
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), FS_IOC_FIEMAP, uintptr(fiemap_ptr))
	if errors.Is(err, os.ErrInvalid) {
		fmt.Printf("Error: ioctl: %s\n", err)
		return nil, err
	}
	// convert
	extents := make([]extent, fiemap_struct.fm_extent_count)
	for i := 1; i <= int(fiemap_struct.fm_extent_count); i++ {
		extents[i-1] = extent{
			logical:  fiemap_extents[i].fe_logical,
			physical: fiemap_extents[i].fe_physical,
			length:   fiemap_extents[i].fe_length,
			flags:    fiemap_extents[i].fe_flags,
		}
	}
	if extents[len(extents)-1].flags&FIEMAP_EXTENT_LAST == 0 {
		fmt.Printf("WARN: incomplete extents list.")
	}
	return extents, nil
}

func getFileInfo(path string, regular_device string, zns_device string) (znsmemory.FileInfo, error) {
	info, err := get_zns_info(regular_device, zns_device)
	if err != nil {
		fmt.Printf("%s\n", err)
		return znsmemory.FileInfo{}, err
	}
	extents, err := get_extents(path)
	if err != nil {
		return znsmemory.FileInfo{}, err
	}
	fileInfo := znsmemory.FileInfo{}
	for i := 0; i < len(extents); i++ {
		zoneAddress := (extents[i].physical - info.zns_start_blkaddr)
		zoneNo := zoneAddress / info.zone_blocks
		zoneOffset := zoneAddress - (zoneNo * info.zone_blocks)
		segmentNo := zoneOffset / (2 * 1024 * 1024)
		// segmentOffset := zoneOffset - (segmentNo * (2*1024*1024))
		fileSegment := znsmemory.FileSegment{
			ZoneIndex:    int(zoneNo),
			SegmentIndex: int(segmentNo),
			ValidMap:     znsmemory.ValidMap{},
		}
		fileInfo.FileSegments = append(fileInfo.FileSegments, fileSegment)
	}
	return fileInfo, nil
}

func main() {
	info, err := get_zns_info("nvme0n1p3", "nvme1n1")
	if err != nil {
		fmt.Printf("zns info:%s\n", err)
		return
	}
	fmt.Printf("%#v\n", info)
	extents, err := get_extents("/mnt/f2fs/normal.0.0.jpg")
	if err != nil {
		fmt.Printf("extents:%s\n", err)
		return
	}
	fmt.Printf("%#v\n", extents[len(extents)-1])
	// fmt.Printf("%#v\n", extents)
	files, err := dir_list("/mnt/f2fs/")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fileInfo, err := getFileInfo("nvme0n1p3", "nvme1n1", "/mnt/f2fs/normal.0.0.jpg")
	if err != nil {
		fmt.Printf("fileinfo%s\n", err)
		return
	}
	fmt.Printf("%#v\n", fileInfo)
	fmt.Println(files)
}
