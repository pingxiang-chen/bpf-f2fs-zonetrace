package main

type FileListType int

const (
	ListTypeParent FileListType = iota
	ListTypeFile
	ListTypeDirectory
)

type FileListItem struct {
	Name    string
	SizeStr string
	Type    FileListType
}

func main() {

}
