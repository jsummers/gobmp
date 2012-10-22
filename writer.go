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

	writePaletted bool
	paletted      *image.Paletted
	nColors       int // Number of colors in palette; 0 if no palette
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
	setDWORD(h[32:36], uint32(e.nColors))
}

func (e *encoder) writeHeaders() error {
	var h [54]byte

	e.generateFileHeader(h[0:14])
	e.generateInfoHeader(h[14:54])

	_, err := e.w.Write(h[:])
	return err
}

func (e *encoder) writePalette() error {
	if !e.writePaletted {
		return nil
	}

	pal := make([]uint8, 4*e.nColors)
	for i := 0; i < e.nColors; i++ {
		r, g, b, _ := e.paletted.Palette[i].RGBA()
		pal[4*i+0] = uint8(b >> 8)
		pal[4*i+1] = uint8(g >> 8)
		pal[4*i+2] = uint8(r >> 8)
	}

	_, err := e.w.Write(pal)
	return err
}

func generateRow_1(e *encoder, j int, rowBuf []byte) {
	for i := range rowBuf {
		rowBuf[i] = 0
	}
	for i := 0; i < e.width; i++ {
		if e.paletted.Pix[j*e.paletted.Stride+i] != 0 {
			rowBuf[i/8] |= uint8(1 << uint(7-i%8))
		}
	}
}

func generateRow_4(e *encoder, j int, rowBuf []byte) {
	for i := range rowBuf {
		rowBuf[i] = 0
	}
	for i := 0; i < e.width; i++ {
		v := e.paletted.Pix[j*e.paletted.Stride+i]
		if i%2 == 0 {
			v <<= 4
		}
		rowBuf[i/2] |= v
	}
}

// Read a row from the source image, and store it in rowBuf in 8-bit BMP format
func generateRow_8(e *encoder, j int, rowBuf []byte) {
	copy(rowBuf[0:e.width], e.paletted.Pix[j*e.paletted.Stride:])
}

// Read a row from the source image, and store it in rowBuf in 24-bit BMP format
func generateRow_24(e *encoder, j int, rowBuf []byte) {
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
	var genRowFunc func(e *encoder, j int, rowBuf []byte)

	if e.writePaletted {
		switch e.dstBitCount {
		case 1:
			genRowFunc = generateRow_1
		case 4:
			genRowFunc = generateRow_4
		default:
			genRowFunc = generateRow_8
		}
	} else {
		genRowFunc = generateRow_24
	}

	rowBuf := make([]byte, e.dstStride)

	for j := 0; j < e.height; j++ {
		genRowFunc(e, e.height-j-1, rowBuf)
		_, err = e.w.Write(rowBuf)
		if err != nil {
			return err
		}
	}
	return nil
}

// If the image can be written as a paletted image, sets e.writePaletted
// to true (and sets e.paletted, e.nColors).
func (e *encoder) checkPaletted() {
	e.paletted, _ = e.m.(*image.Paletted)
	if e.paletted == nil {
		return
	}

	e.nColors = len(e.paletted.Palette)
	if e.nColors < 1 || e.nColors > 256 {
		e.nColors = 0
		return
	}

	e.writePaletted = true
}

// Plot out the structure of the file that we're going to write.
func (e *encoder) strategize() error {
	e.srcBounds = e.m.Bounds()
	e.width = e.srcBounds.Dx()
	e.height = e.srcBounds.Dy()
	e.checkPaletted()
	if e.writePaletted {
		if e.nColors <= 2 {
			e.dstBitCount = 1
		} else if e.nColors <= 16 {
			e.dstBitCount = 4
		} else {
			e.dstBitCount = 8
		}
	} else {
		e.dstBitCount = 24
	}
	e.dstStride = ((e.width*e.dstBitCount + 31) / 32) * 4
	e.dstBitsOffset = 14 + 40 + 4*e.nColors
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

	err = e.writePalette()
	if err != nil {
		return err
	}

	err = e.writeBits()
	if err != nil {
		return err
	}

	return nil
}
