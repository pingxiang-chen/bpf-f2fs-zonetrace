package znsmemory

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetFileInfo(znsInfo *ZoneInfo, path string) (*FileInfo, error) {
	cmd := exec.Command("fibmap.f2fs", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run `fibmap.f2fs`: %w", err)
	}
	fileInfo := FileInfo{}
	output := string(out)
	fmt.Printf("====== fibmap.f2fs ======\n%s\n=====================", output)
	zoneSize := znsInfo.TotalBlockPerZone * 4 / 1024 // MiB
	segPerZone := zoneSize / 2                       // MiB
	outputLines := strings.Split(output, "\n")
	fibmaps, err := parseFibmap(outputLines)
	if err != nil {
		return nil, fmt.Errorf("ParseFibmap: %w", err)
	}
	sitMap := make(map[int][]byte)
	for _, fbm := range fibmaps {
		segNo := fbm.StartBlk / 512 // 512block per segment (4K * 512 = 2M segment)
		// endSegNo := Fibmap.end_blk / 512
		sentryStartOffset := fbm.StartBlk % 512
		sentryEndOffset := sentryStartOffset + fbm.Blks
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
		curZone := segNo / segPerZone
		fileInfo.FileSegments = append(fileInfo.FileSegments, FileSegment{
			ZoneIndex:    curZone,
			SegmentIndex: segNo,
			ValidMap:     sit,
		})
	}
	return &fileInfo, nil
}

func parseFibmap(output_lines []string) ([]Fibmap, error) {
	var fibmaps []Fibmap
	for _, line := range output_lines {
		filePos, startBlk, endBlk, blks := 0, 0, 0, 0
		_, err := fmt.Sscanf(line, "%d %d %d %d", &filePos, &startBlk, &endBlk, &blks)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Fibmap: %w (%s)", err, line)
		}
		if blks != 0 {
			fibmaps = append(fibmaps, Fibmap{
				FilePos:  filePos,
				StartBlk: startBlk,
				EndBlk:   endBlk,
				Blks:     blks,
			})
		}
	}
	return fibmaps, nil
}
