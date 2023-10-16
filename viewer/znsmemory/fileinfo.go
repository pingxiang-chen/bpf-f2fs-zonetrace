package znsmemory

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func getBeginBlock(znsInfo *ZoneInfo) (int64, error) {
	out, err := exec.Command("dump.f2fs", fmt.Sprintf("/dev/%s", znsInfo.RegularDeviceName)).Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run `dump.f2fs`: %w", err)
	}
	output := string(out)
	znsBlkAddrPattern, err := regexp.Compile(fmt.Sprintf(`/dev/%s blkaddr = (\w+)`, znsInfo.ZNSDeviceName))
	if err != nil {
		return 0, fmt.Errorf("failed to compile zns blk addr pattern: %w", err)
	}
	match := znsBlkAddrPattern.FindStringSubmatch(output)
	if len(match) < 2 {
		return 0, errors.New("cannot find zns blk addr")
	}
	znsBlkAddr, err := strconv.ParseInt(match[1], 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse zns blk addr: %w", err)
	}
	return znsBlkAddr, nil
}

func GetFileInfo(znsInfo *ZoneInfo, path string) (*FileInfo, error) {
	cmd := exec.Command("fibmap.f2fs", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run `fibmap.f2fs`: %w", err)
	}
	fileInfo := FileInfo{
		FilePath:     path,
		FileSegments: []FileSegment{},
		Fibmaps:      []Fibmap{},
	}
	output := string(out)
	zoneSize := znsInfo.TotalBlockPerZone * 4 / 1024 // MiB
	segPerZone := zoneSize / 2                       // MiB
	outputLines := strings.Split(output, "\n")
	beginBlock, err := getBeginBlock(znsInfo)
	if err != nil {
		return nil, fmt.Errorf("GetBeginBlock: %w", err)
	}
	fibmaps, err := parseFibmap(outputLines, beginBlock)
	fileInfo.Fibmaps = make([]Fibmap, len(fibmaps))
	copy(fileInfo.Fibmaps, fibmaps)
	if err != nil {
		return nil, fmt.Errorf("ParseFibmap: %w", err)
	}
	sitMap := make(map[int][]byte)
	for _, fbm := range fibmaps {
		segNo := fbm.StartBlk / 512 // 512block per segment (4K * 512 = 2M segment)
		// endSegNo := Fibmap.end_blk / 512
		sentryStartOffset := fbm.StartBlk % 512
		sentryEndOffset := sentryStartOffset + int64(fbm.Blks)
		for offset := sentryStartOffset; offset < sentryEndOffset; offset++ {
			// get sit and update data
			byteOffset := offset / 8
			curSegNo := segNo + (byteOffset / 64)
			byteOffset %= 64
			bitOffset := offset % 8
			sit, ok := sitMap[int(curSegNo)]
			if !ok {
				sitMap[int(curSegNo)] = make([]byte, 64)
				sit = sitMap[int(curSegNo)]
			}
			sit[byteOffset] |= 1 << (7 - bitOffset)
		}
	}
	for segNo, sit := range sitMap {
		curZone := segNo / segPerZone
		fileInfo.FileSegments = append(fileInfo.FileSegments, FileSegment{
			ZoneIndex:            curZone,
			SegmentIndex:         segNo,
			RelativeSegmentIndex: segNo % segPerZone,
			ValidMap:             sit,
		})
	}
	return &fileInfo, nil
}

func parseFibmap(outputLines []string, beginBlock int64) ([]Fibmap, error) {
	var fibmaps []Fibmap
	for _, line := range outputLines {
		var filePos, startBlk, endBlk int64
		var blks int
		_, err := fmt.Sscanf(line, "%d %d %d %d", &filePos, &startBlk, &endBlk, &blks)
		if err != nil {
			continue
		}
		if blks != 0 {
			fibmaps = append(fibmaps, Fibmap{
				FilePos:  filePos,
				StartBlk: startBlk - beginBlock,
				EndBlk:   endBlk - beginBlock,
				Blks:     blks,
			})
		}
	}
	return fibmaps, nil
}
