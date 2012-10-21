// ◄◄◄ gobmp/rle.go ►►►
// Copyright (c) 2012 Jason Summers
// Use of this code is governed by an MIT-style license that can
// be found in the readme.md file.
//
// BMP RLE decoder
//

package gobmp

import "io"

type rleState struct {
	xpos, ypos   int // Position in the target image
	bufPos       int
	bufUsed      int // Number of valid bytes in .buf
	eofFlag      bool
	badColorFlag bool
	buf          []byte // RLE data buffer
}

// Read until rle.buf is full, or EOF or an error is encountered.
// On error other than EOF, returns the error.
// On EOF, sets rle.eofFlag.
func (d *decoder) rleReadMore(rle *rleState) error {
	var n int
	var err error

	rle.bufPos = 0
	rle.bufUsed = 0

	for {
		if rle.eofFlag {
			break
		}
		if rle.bufUsed >= len(rle.buf) {
			break
		}

		n, err = d.r.Read(rle.buf[rle.bufUsed:])
		if err != nil {
			if err == io.EOF {
				rle.eofFlag = true
			} else {
				return err
			}
		}
		rle.bufUsed += n
	}
	return nil
}

func (d *decoder) rlePutPixel(rle *rleState, v byte) {
	// Make sure the position is valid.
	if rle.xpos < 0 || rle.xpos >= d.width ||
		rle.ypos < 0 || rle.ypos >= d.height {
		return
	}
	// Make sure the palette index is valid.
	if int(v) >= d.dstPalNumEntries {
		rle.badColorFlag = true
		return
	}

	// Set the pixel, and advance the current position.
	d.img_Paletted.Pix[rle.ypos*d.img_Paletted.Stride+rle.xpos] = v
	rle.xpos++
}

func (d *decoder) readBitsRLE() error {
	var err error
	var b1, b2 byte
	var uncPixelsLeft int
	var deltaFlag bool
	var k int

	rle := new(rleState)
	rle.xpos = 0
	rle.ypos = d.height - 1 // RLE images are not allowed to be top-down.
	rle.buf = make([]byte, 4096)

	for {
		if rle.badColorFlag {
			return FormatError("palette index out of range")
		}

		if rle.ypos < 0 || (rle.ypos == 0 && rle.xpos >= d.width) {
			break // Reached the end of the target image; may as well stop
		}

		// If there aren't at least 2 bytes available, read more data
		// from the file.
		if rle.bufUsed-rle.bufPos < 2 {
			err = d.rleReadMore(rle)
			if err != nil {
				return err
			}
			if rle.bufUsed < 2 {
				break // End of file, presumably
			}
		}

		// Look at the next two bytes
		b1 = rle.buf[rle.bufPos]
		b2 = rle.buf[rle.bufPos+1]
		rle.bufPos += 2

		if uncPixelsLeft > 0 {
			if d.biCompression == bI_RLE4 {
				// The two bytes we're processing store up to 4 uncompressed pixels.
				d.rlePutPixel(rle, b1>>4)
				uncPixelsLeft--
				if uncPixelsLeft > 0 {
					d.rlePutPixel(rle, b1&0x0f)
					uncPixelsLeft--
				}
				if uncPixelsLeft > 0 {
					d.rlePutPixel(rle, b2>>4)
					uncPixelsLeft--
				}
				if uncPixelsLeft > 0 {
					d.rlePutPixel(rle, b2&0x0f)
					uncPixelsLeft--
				}
			} else { // RLE8
				// The two bytes we're processing store up to 2 uncompressed pixels.
				d.rlePutPixel(rle, b1)
				uncPixelsLeft--
				if uncPixelsLeft > 0 {
					d.rlePutPixel(rle, b2)
					uncPixelsLeft--
				}
			}
		} else if deltaFlag {
			rle.xpos += int(b1)
			rle.ypos -= int(b2)
			deltaFlag = false
		} else if b1 == 0 {
			// An uncompressed run, or a special code.
			//
			// Any pixels skipped by special codes will be left at whatever
			// image.NewPaletted() initialized them to, which we assume is 0,
			// for palette entry 0.
			if b2 == 0 { // End of row
				rle.ypos--
				rle.xpos = 0
			} else if b2 == 1 { // End of bitmap
				break
			} else if b2 == 2 { // Delta
				deltaFlag = true
			} else {
				// An upcoming uncompressed run of b2 pixels
				uncPixelsLeft = int(b2)
			}
		} else { // A compressed run of pixels
			if d.biCompression == bI_RLE4 {
				// b1 pixels, alternating between two colors
				for k = 0; k < int(b1); k++ {
					if k%2 == 0 {
						d.rlePutPixel(rle, b2>>4)
					} else {
						d.rlePutPixel(rle, b2&0x0f)
					}
				}
			} else { // RLE8
				// b1 pixels of color b2
				for k = 0; k < int(b1); k++ {
					d.rlePutPixel(rle, b2)
				}
			}
		}
	}

	return nil
}
