package fstool

type FileListType int

const (
	ListTypeParent FileListType = iota
	ListTypeFile
	ListTypeDirectory
)

type MountInfo struct {
	MountPath []string
	Device    string
}

type FileListItem struct {
	Name    string
	SizeStr string
	Type    FileListType
}
