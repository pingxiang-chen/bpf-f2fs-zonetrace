package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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

type fibmap struct {
	file_pos  int
	start_blk int
	end_blk   int
	blks      int
}

func parseFibmap(output_lines []string) []fibmap {
	var fibmaps []fibmap
	for i := 0; i < len(output_lines); i++ {
		file_pos, start_blk, end_blk, blks := 0, 0, 0, 0
		fmt.Sscanf(output_lines[i], "%d %d %d %d", &file_pos, &start_blk, &end_blk, &blks)
		if blks != 0 {
			fibmaps = append(fibmaps, fibmap{
				file_pos:  file_pos,
				start_blk: start_blk,
				end_blk:   end_blk,
				blks:      blks,
			})
		}
	}
	return fibmaps
}

/**
 * path: file path to highlight
 * regular_device: f2fs regular device. ex) nvme0n1p1
 * zns_device: f2fs zns device. ex) nvme1n1
 */
func getFileInfo(path string, regular_device string, zns_device string) (znsmemory.FileInfo, error) {
	info, err := get_zns_info(regular_device, zns_device)
	if err != nil {
		fmt.Printf("%s\n", err)
		return znsmemory.FileInfo{}, err
	}
	cmd := exec.Command("fibmap.f2fs", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("fibmap err: %s\n", err)
		return znsmemory.FileInfo{}, err
	}
	fileInfo := znsmemory.FileInfo{}
	output := string(out)
	// fmt.Println(output)
	zoneSize := info.zone_blocks * 4 / 1024 // MiB
	segPerZone := zoneSize / 2              // MiB
	output_lines := strings.Split(output, "\n")
	fibmaps := parseFibmap(output_lines)
	sitMap := make(map[int][]byte)
	for _, fibmap := range fibmaps {
		segNo := fibmap.start_blk / 512 // 512block per segment (4K * 512 = 2M segment)
		// endSegNo := fibmap.end_blk / 512
		sentryStartOffset := fibmap.start_blk % 512
		sentryEndOffset := sentryStartOffset + fibmap.blks
		for offset := sentryStartOffset; offset < sentryEndOffset; offset++ {
			// get sit and update data
			byteOffset := offset / 8
			curSegNo := segNo + (byteOffset / 64)
			byteOffset %= 64
			bitOffset := offset % 8
			if byteOffset >= 64 {

			}
			sit, ok := sitMap[curSegNo]
			if !ok {
				sitMap[curSegNo] = make([]byte, 64)
				sit = sitMap[curSegNo]
			}
			sit[byteOffset] |= 1 << (7 - bitOffset)
		}
	}
	for segNo, sit := range sitMap {
		curZone := segNo / int(segPerZone)
		fileInfo.FileSegments = append(fileInfo.FileSegments, znsmemory.FileSegment{
			ZoneIndex:    curZone,
			SegmentIndex: segNo,
			ValidMap:     sit,
		})
	}
	return fileInfo, nil
}

func main() {
	// info, err := get_zns_info("nvme4n1p1", "nvme3n1")
	// if err != nil {
	// 	fmt.Printf("zns info:%s\n", err)
	// 	return
	// }
	// fmt.Printf("%#v\n", info)
	// extents, err := get_extents("/mnt/f2fs/normal.0.0")
	// if err != nil {
	// 	fmt.Printf("extents:%s\n", err)
	// 	return
	// }
	// fmt.Printf("%#v\n", extents[len(extents)-1])
	// // fmt.Printf("%#v\n", extents)
	// files, err := dir_list("/mnt/f2fs/")
	// if err != nil {
	// 	fmt.Printf("%s\n", err)
	// 	return
	// }
	fileInfo, err := getFileInfo("/mnt/f2fs/target_file.png", "nvme4n1p1", "nvme3n1")
	if err != nil {
		fmt.Printf("fileinfo%s\n", err)
		return
	}
	fmt.Printf("%#v\n", fileInfo)
	// fmt.Println(files)
}
