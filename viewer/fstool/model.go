package fstool

type PathType int

const (
	UnknownPath PathType = iota
	RootPath
	ParentPath
	FilePath
	DirectoryPath
)

type MountInfo struct {
	MountPath []string
	Device    string
}

type FileInfo struct {
	FilePath string   `json:"file_path"`
	Name     string   `json:"name"`
	Type     PathType `json:"type"`
	SizeStr  string   `json:"size_str"`
}
