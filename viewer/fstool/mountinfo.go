package fstool

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

func GetMountInfo(device string) (*MountInfo, error) {
	mountPaths, err := getMountPoints(device)
	if err != nil {
		return nil, fmt.Errorf("failed to get mount points: %w", err)
	}
	return &MountInfo{
		MountPath: mountPaths,
		Device:    device,
	}, nil
}
