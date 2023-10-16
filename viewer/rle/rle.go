package rle

import (
	"bytes"
)

// Compress performs Run-Length Encoding (RLE) on a given byte slice.
// It iterates through the input data, and for each byte, it writes
// the byte value followed by a count of how many times the byte is
// repeated consecutively to a buffer. It returns the compressed data
// as a byte slice.
func Compress(data []byte) []byte {
	var compressed bytes.Buffer
	length := len(data)

	for i := 0; i < length; i++ {
		count := 1
		for i+1 < length && data[i] == data[i+1] {
			count++
			i++
		}
		compressed.WriteByte(data[i])
		compressed.WriteByte(byte(count))
	}

	return compressed.Bytes()
}

// Decompress reverses the Run-Length Encoding (RLE) compression on a
// given byte slice. It iterates through the input data by incrementing
// by 2 since the data consists of byte-value and count pairs. For each
// pair, it writes the byte value to a buffer for the number of times
// specified by the count. It returns the decompressed data as a byte slice.
func Decompress(data []byte) []byte {
	var decompressed bytes.Buffer

	for i := 0; i < len(data); i += 2 {
		value := data[i]
		count := int(data[i+1])

		for j := 0; j < count; j++ {
			decompressed.WriteByte(value)
		}
	}

	return decompressed.Bytes()
}
