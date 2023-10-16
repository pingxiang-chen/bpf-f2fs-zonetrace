package fstool

import (
	"fmt"
	"os"
	"path/filepath"
)

func ListFiles(dirPath string) ([]FileInfo, error) {
	// 응답 객체 초기화
	var result []FileInfo

	// 디렉터리 열기
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error opening directory: %w", err)
	}
	defer dir.Close()

	// 디렉터리 내의 파일 및 디렉터리 목록 가져오기
	entries, err := dir.Readdir(-1) // 모든 항목 가져오기
	if err != nil {
		return nil, fmt.Errorf("error reading directory contents: %w", err)
	}

	// 각 항목에 대해 정보 수집
	for _, entry := range entries {
		sizeStr := ""
		if !entry.IsDir() {
			sizeStr = fmt.Sprintf("%d", entry.Size()) // 파일 크기를 문자열로 변환
		}

		fileInfo := FileInfo{
			FilePath: filepath.Join(dirPath, entry.Name()),
			Name:     entry.Name(),
			Type:     getType(entry),
			SizeStr:  sizeStr,
		}
		result = append(result, fileInfo)
	}
	return result, nil
}

func getType(entry os.FileInfo) PathType {
	// 실제 시스템에서는 파일 유형을 확인하고 znsmemory.PathType 값 중 하나를 반환해야 합니다.
	// 여기서는 단순화를 위해 모든 파일을 "File"로 가정합니다.
	if entry.IsDir() {
		return DirectoryPath
	}
	return FilePath
}
