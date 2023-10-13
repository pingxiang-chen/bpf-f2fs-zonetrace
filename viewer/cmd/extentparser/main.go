package main

import (
	"encoding/json"
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
	FiemapExtentSize = 56
	FiemapSize       = 32
	FsIocFiemap      = 3223348747
	FiemapMaxOffset  = ^uint64(0)
	FiemapFlagSync   = 0x0001
	FiemapExtentLast = 0x0001
)

type znsInfo struct {
	znsStartBlkAddr uint64
	zoneBlocks      uint64
}

type fiemapExtent struct {
	feLogical    uint64
	fePhysical   uint64
	feLength     uint64
	feReserved64 [2]uint64
	feFlags      uint32
	feReserved   [3]uint32
} // 56 bytes

type fiemap struct {
	fmStart         uint64
	fmLength        uint64
	fmFlags         uint32
	fmMappedExtents uint32
	fmExtentCount   uint32
	fmReserved      uint32
} // 32 bytes

type extent struct {
	logical  uint64
	physical uint64
	length   uint64
	flags    uint32
}

func dirList(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil, err
	}
	var ret []string
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			ret = append(ret, entry.Name())
		}
	}
	return ret, nil
}

func getZnsInfo(regularDevice string, znsDevice string) (znsInfo, error) {
	out, err := exec.Command("dump.f2fs", fmt.Sprintf("/dev/%s", regularDevice)).Output()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return znsInfo{}, err
	}
	output := string(out)
	znsBlkAddrPattern := regexp.MustCompile(fmt.Sprintf(`/dev/%s blkaddr = (\w+)`, znsDevice))
	match := znsBlkAddrPattern.FindStringSubmatch(output)
	if len(match) < 2 {
		return znsInfo{}, errors.New("Cannot find zns blk addr")
	}
	znsBlkAddr, err := strconv.ParseInt(match[1], 16, 64)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return znsInfo{}, err
	}
	zoneBlocksPattern := regexp.MustCompile(`(\d+) blocks per zone`)
	match = zoneBlocksPattern.FindStringSubmatch(output)
	if len(match) < 2 {
		return znsInfo{}, errors.New("Cannot find zone blocks")
	}
	zoneBlocks, err := strconv.Atoi(match[1])
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return znsInfo{}, err
	}
	// fmt.Println(output)
	return znsInfo{
		znsStartBlkAddr: uint64(znsBlkAddr),
		zoneBlocks:      uint64(zoneBlocks),
	}, nil
}

func getExtents(path string) ([]extent, error) {
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
	fiemapResult := fiemap{fmStart: 0, fmLength: FiemapMaxOffset, fmFlags: FiemapFlagSync, fmMappedExtents: 0, fmExtentCount: 0, fmReserved: 0}
	ptr := uintptr(unsafe.Pointer(&fiemapResult))
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), FsIocFiemap, ptr)
	if errors.Is(err, os.ErrInvalid) {
		fmt.Printf("Error: ioctl: %s\n", err)
		return nil, err
	}
	// allocate for actual extents count
	fiemapExtents := make([]fiemapExtent, fiemapResult.fmMappedExtents+1)
	// index 0 element as fiemap
	fiemapPtr := unsafe.Pointer(uintptr(unsafe.Pointer(&fiemapExtents[1])) - FiemapSize)
	fiemapStruct := (*fiemap)(fiemapPtr)
	fiemapStruct.fmStart = 0
	fiemapStruct.fmLength = FiemapMaxOffset
	fiemapStruct.fmFlags = FiemapFlagSync
	fiemapStruct.fmExtentCount = fiemapResult.fmMappedExtents
	fiemapStruct.fmMappedExtents = 0
	// get extents
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), FsIocFiemap, uintptr(fiemapPtr))
	if errors.Is(err, os.ErrInvalid) {
		fmt.Printf("Error: ioctl: %s\n", err)
		return nil, err
	}
	// convert
	extents := make([]extent, fiemapStruct.fmExtentCount)
	for i := 1; i <= int(fiemapStruct.fmExtentCount); i++ {
		extents[i-1] = extent{
			logical:  fiemapExtents[i].feLogical,
			physical: fiemapExtents[i].fePhysical,
			length:   fiemapExtents[i].feLength,
			flags:    fiemapExtents[i].feFlags,
		}
	}
	if extents[len(extents)-1].flags&FiemapExtentLast == 0 {
		fmt.Printf("WARN: incomplete extents list.")
	}
	return extents, nil
}

type fibmap struct {
	filePos  int
	startBlk int
	endBlk   int
	blks     int
}

func parseFibmap(output_lines []string) []fibmap {
	var fibmaps []fibmap
	for i := 0; i < len(output_lines); i++ {
		filePos, startBlk, endBlk, blks := 0, 0, 0, 0
		fmt.Sscanf(output_lines[i], "%d %d %d %d", &filePos, &startBlk, &endBlk, &blks)
		if blks != 0 {
			fibmaps = append(fibmaps, fibmap{
				filePos:  filePos,
				startBlk: startBlk,
				endBlk:   endBlk,
				blks:     blks,
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
func getFileInfo(path string, regularDevice string, znsDevice string) (znsmemory.FileInfo, error) {
	info, err := getZnsInfo(regularDevice, znsDevice)
	if err != nil {
		fmt.Printf("%s\n", err)
		return znsmemory.FileInfo{}, err
	}
	b, _ := json.Marshal(info)
	fmt.Printf("info: %s\n", string(b))

	cmd := exec.Command("fibmap.f2fs", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("fibmap err: %s\n", err)
		return znsmemory.FileInfo{}, err
	}
	fileInfo := znsmemory.FileInfo{}
	output := string(out)
	// fmt.Println(output)
	zoneSize := info.zoneBlocks * 4 / 1024 // MiB
	segPerZone := zoneSize / 2             // MiB
	outputLines := strings.Split(output, "\n")
	fibmaps := parseFibmap(outputLines)
	sitMap := make(map[int][]byte)
	for _, fibmap := range fibmaps {
		segNo := fibmap.startBlk / 512 // 512block per segment (4K * 512 = 2M segment)
		// endSegNo := fibmap.end_blk / 512
		sentryStartOffset := fibmap.startBlk % 512
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
	// info, err := getZnsInfo("nvme4n1p1", "nvme3n1")
	// if err != nil {
	// 	fmt.Printf("zns info:%s\n", err)
	// 	return
	// }
	// fmt.Printf("%#v\n", info)
	// extents, err := getExtents("/mnt/f2fs/normal.0.0")
	// if err != nil {
	// 	fmt.Printf("extents:%s\n", err)
	// 	return
	// }
	// fmt.Printf("%#v\n", extents[len(extents)-1])
	// // fmt.Printf("%#v\n", extents)
	// files, err := dirList("/mnt/f2fs/")
	// if err != nil {
	// 	fmt.Printf("%s\n", err)
	// 	return
	// }
	fileInfo, err := getFileInfo("/mnt/f2fs/target_file.png", "nvme4n1p1", "nvme3n1")
	if err != nil {
		fmt.Printf("fileinfo%s\n", err)
		return
	}
	b, err := json.Marshal(fileInfo)
	fmt.Printf("fileInfo: %s\n", string(b))
	// fmt.Println(files)
}
