package protos

import _ "embed"

// SegmentProtoFile reads the `zns.proto` file and stores its bytes.
//
//go:embed zns.proto
var SegmentProtoFile []byte
