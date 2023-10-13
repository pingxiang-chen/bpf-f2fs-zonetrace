package statics

import (
	_ "embed"

	"github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer/protos"
)

// IndexHtmlFile reads the `index.html` file and stores its bytes.
//
//go:embed index.html
var IndexHtmlFile []byte

// HighlightHtmlFile reads the 'highlight.html' file and stores its bytes.
//
//go:embed highlight.html
var HighlightHtmlFile []byte

// TracerJsFile reads the `js/tracer.js` file and stores its bytes.
//
//go:embed js/tracer.js
var tracerJsFile []byte

type StaticFile struct {
	File        []byte
	ContentType string
}

var StaticFileMap = map[string]StaticFile{
	"zns.proto": {
		File:        protos.SegmentProtoFile,
		ContentType: "text/plain",
	},
	"js/tracer.js": {
		File:        tracerJsFile,
		ContentType: "text/javascript",
	},
}
