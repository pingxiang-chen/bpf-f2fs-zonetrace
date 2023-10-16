package main

import (
	"fmt"

	"k8s.io/mount-utils"
)

func getMountPoints(device string) ([]string, error) {
	mounter := mount.New("")
	allMounts, err := mounter.List()
	if err != nil {
		return nil, err
	}

	var mountPoints []string
	for _, mountInfo := range allMounts {
		if mountInfo.Device == device {
			mountPoints = append(mountPoints, mountInfo.Path)
		}
	}

	return mountPoints, nil
}

func main() {
	device := "/dev/nvme4n1p1"
	mountPoints, err := getMountPoints(device)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, mountPoint := range mountPoints {
		fmt.Println("Mount Point:", mountPoint)
	}
}
