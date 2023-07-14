package statics

import (
	_ "embed"
	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/protos"
)

//go:embed index.html
var IndexHtmlFile []byte

type StaticFile struct {
	File        []byte
	ContentType string
}

var StaticFileMap = map[string]StaticFile{
	"zns.proto": {
		File:        protos.SegmentProtoFile,
		ContentType: "text/plain",
	},
}
