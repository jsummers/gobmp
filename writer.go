// ◄◄◄ gobmp/writer.go ►►►
// Copyright (c) 2012 Jason Summers
// Use of this code is governed by an MIT-style license that can
// be found in the readme.md file.
//
// BMP file encoder
//

package gobmp

import "image"
import "io"

type encoder struct {
	w io.Writer
	m image.Image

	srcBounds     image.Rectangle
	width         int
	height        int
	dstStride     int
	dstBitsSize   int
	dstBitCount   int
	dstBitsOffset int
	dstFileSize   int
}

func setWORD(b []byte, n uint16) {
	b[0] = byte(n)
	b[1] = byte(n >> 8)
}

func setDWORD(b []byte, n uint32) {
	b[0] = byte(n)
	b[1] = byte(n >> 8)
	b[2] = byte(n >> 16)
	b[3] = byte(n >> 24)
}

// Write the BITMAPFILEHEADER structure to a slice[14].
func (e *encoder) generateFileHeader(h []byte) {
	h[0] = 0x42 // 'B'
	h[1] = 0x4d // 'M'
	setDWORD(h[2:6], uint32(e.dstFileSize))
	setDWORD(h[10:14], uint32(e.dstBitsOffset))
}

// Write the BITMAPINFOHEADER structure to a slice[40].
func (e *encoder) generateInfoHeader(h []byte) {
	setDWORD(h[0:4], 40)
	setDWORD(h[4:8], uint32(e.width))
	setDWORD(h[8:12], uint32(e.height))
	setWORD(h[12:14], 1) // biPlanes
	setWORD(h[14:16], uint16(e.dstBitCount))
	setDWORD(h[20:24], uint32(e.dstBitsSize))
	setDWORD(h[24:28], 2835) // biXPelsPerMeter
	setDWORD(h[28:32], 2835) // biYPelsPerMeter
}

func (e *encoder) writeHeaders() error {
	var h [54]byte

	e.generateFileHeader(h[0:14])
	e.generateInfoHeader(h[14:54])

	_, err := e.w.Write(h[:])
	return err
}

// Read a row from the source image, and store it in rowBuf in BMP format
func (e *encoder) generateRow(j int, rowBuf []byte) {
	for i := 0; i < e.width; i++ {
		srcclr := e.m.At(e.srcBounds.Min.X+i, e.srcBounds.Min.Y+j)
		r, g, b, _ := srcclr.RGBA()
		rowBuf[i*3+0] = uint8(b >> 8)
		rowBuf[i*3+1] = uint8(g >> 8)
		rowBuf[i*3+2] = uint8(r >> 8)
	}
}

func (e *encoder) writeBits() error {
	var err error

	rowBuf := make([]byte, e.dstStride)

	for j := 0; j < e.height; j++ {
		e.generateRow(e.height-j-1, rowBuf)
		_, err = e.w.Write(rowBuf)
		if err != nil {
			return err
		}
	}
	return nil
}

// Figure out the vital statistics of the target image.
func (e *encoder) strategize() error {
	e.srcBounds = e.m.Bounds()
	e.width = e.srcBounds.Max.X - e.srcBounds.Min.X
	e.height = e.srcBounds.Max.Y - e.srcBounds.Min.Y
	e.dstBitCount = 24
	e.dstStride = ((e.width*e.dstBitCount + 31) / 32) * 4
	e.dstBitsOffset = 14 + 40
	e.dstBitsSize = e.height * e.dstStride
	e.dstFileSize = e.dstBitsOffset + e.dstBitsSize
	return nil
}

// Encode writes the Image m to w in BMP format.
func Encode(w io.Writer, m image.Image) error {
	var err error

	e := new(encoder)
	e.w = w
	e.m = m

	err = e.strategize()
	if err != nil {
		return err
	}

	err = e.writeHeaders()
	if err != nil {
		return err
	}

	err = e.writeBits()
	if err != nil {
		return err
	}

	return nil
}
