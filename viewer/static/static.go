package static

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

// fileTracerJsFile reads the `js/fileTracer.js` file and stores its bytes.
//
//go:embed js/fileTracer.js
var fileTracerJsFile []byte

// explorerJsFile reads the `js/explorer.js` file and stores its bytes.
//
//go:embed js/explorer.js
var explorerJsFile []byte

type ServingFile struct {
	File        []byte
	ContentType string
}

var ServingFileMap = map[string]ServingFile{
	"zns.proto": {
		File:        protos.SegmentProtoFile,
		ContentType: "text/plain",
	},
	"js/tracer.js": {
		File:        tracerJsFile,
		ContentType: "text/javascript",
	},
	"js/fileTracer.js": {
		File:        fileTracerJsFile,
		ContentType: "text/javascript",
	},
	"js/explorer.js": {
		File:        explorerJsFile,
		ContentType: "text/javascript",
	},
}
