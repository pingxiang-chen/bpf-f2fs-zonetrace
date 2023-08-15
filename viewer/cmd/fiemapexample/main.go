package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/fibmap"
)

func main() {
	file, err := os.Open("/mnt/f2fs/normal.0.0.jpg")
	if err != nil {
		fmt.Printf("Error: open: %s\n", err)
		return
	}

	// extents size 를 구한다.
	f := fibmap.NewFibmapFile(file)
	extents, err := f.FibmapExtents()
	if err != nil {
		if !errors.Is(err, syscall.Errno(0)) {
			fmt.Printf("Error: fibmap: %s\n", err)
			return
		}
	}
	extentsLength := len(extents)

	// fieamp 호출
	e, err := f.Fiemap(uint32(extentsLength))
	if !errors.Is(err, syscall.Errno(0)) {
		fmt.Printf("Error: fiemap: %s\n", err)
		return
	}

	// 보기 좋게 json 으로 변환
	jsonBytes, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes)) // json 으로 변환된 것을 출력
}
