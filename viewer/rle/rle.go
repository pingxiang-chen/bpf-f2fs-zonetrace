package rle

import (
	"bytes"
)

// Compress performs Run-Length Encoding (RLE) on the bits of a given byte slice.
// It iterates through the bits of the input data, and for each bit, it writes
// the bit value followed by a count of how many times the bit is
// repeated consecutively to a buffer. It returns the compressed data
// as a byte slice.
func Compress(data []byte) []byte {
	var compressed bytes.Buffer
	bitLength := len(data) * 8

	var currentBit byte
	var count byte = 0

	for i := 0; i < bitLength; i++ {
		bit := (data[i/8] >> (7 - (i % 8))) & 1
		if i == 0 {
			currentBit = bit
		}

		if bit == currentBit && count < 255 {
			count++
		} else {
			compressed.WriteByte(currentBit)
			compressed.WriteByte(count)
			currentBit = bit
			count = 1
		}
	}

	// Write the last run
	compressed.WriteByte(currentBit)
	compressed.WriteByte(count)

	return compressed.Bytes()
}

// Decompress reverses the Run-Length Encoding (RLE) compression on a
// given byte slice. It iterates through the input data by incrementing
// by 2 since the data consists of bit-value and count pairs. For each
// pair, it writes the bit value to a buffer for the number of times
// specified by the count. It returns the decompressed data as a byte slice.
func Decompress(data []byte) []byte {
	var decompressed bytes.Buffer
	var byteValue byte = 0
	var bitPosition byte = 7

	for i := 0; i < len(data); i += 2 {
		bitValue := data[i]
		count := data[i+1]

		for j := byte(0); j < count; j++ {
			byteValue |= (bitValue << bitPosition)
			bitPosition--
			if bitPosition > 7 {
				decompressed.WriteByte(byteValue)
				byteValue = 0
				bitPosition = 7
			}
		}
	}

	if bitPosition < 7 {
		decompressed.WriteByte(byteValue)
	}

	return decompressed.Bytes()
}
