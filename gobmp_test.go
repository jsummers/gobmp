// ◄◄◄ gobmp/gobmp_test.go ►►►
// 
// Copyright (c) 2012 Jason Summers

package gobmp

import "testing"
import "image"
import "image/color"
import "image/png"
import "os"
import "io/ioutil"
import "bytes"
import "fmt"

func readImageFromFile(t *testing.T, srcFilename string) image.Image {
	var err error
	var srcImg image.Image

	file, err := os.Open(srcFilename)
	if err != nil {
		t.Logf("%s\n", err.Error())
		t.FailNow()
		return nil
	}
	defer file.Close()

	srcImg, _, err = image.Decode(file)
	if err != nil {
		t.Logf("%s: %s\n", srcFilename, err.Error())
		t.FailNow()
		return nil
	}
	if srcImg == nil {
		t.Logf("%s: Decode failed\n", srcFilename)
		t.FailNow()
		return nil
	}

	return srcImg
}

func writeImageToFile(t *testing.T, img image.Image, dstFilename string, fileFmt string) {
	var err error

	file, err := os.Create(dstFilename)
	if err != nil {
		t.Logf("%s\n", err.Error())
		t.FailNow()
		return
	}
	defer file.Close()

	if fileFmt == "png" {
		err = png.Encode(file, img)
	} else {
		err = Encode(file, img)
	}
	if err != nil {
		t.Logf("%s\n", err.Error())
		t.FailNow()
		return
	}
}

func compareFiles(t *testing.T, expectedFN string, actualFN string) {
	var expectedBytes []byte
	var actualBytes []byte
	var err error

	expectedBytes, err = ioutil.ReadFile(expectedFN)
	if err != nil {
		t.Logf("Failed to open for compare: %s\n", err.Error())
		t.Fail()
		return
	}

	actualBytes, err = ioutil.ReadFile(actualFN)
	if err != nil {
		t.Logf("Failed to open for compare: %s\n", err.Error())
		t.FailNow()
		return
	}

	if len(expectedBytes) != len(actualBytes) {
		t.Logf("%s and %s differ in size\n", expectedFN, actualFN)
		t.Fail()
		return
	}

	if 0 != bytes.Compare(actualBytes, expectedBytes) {
		t.Logf("%s and %s differ\n", expectedFN, actualFN)
		t.Fail()
		return
	}
}

type encodeTestType struct {
	srcFN, dstFN, expectedFN string
}

var encodeTests = []encodeTestType{
	{"rgb8a.png", "rgb8a.bmp", "rgb8a.bmp"},
	{"p8.png", "p8.bmp", "p8.bmp"},
}

func TestEncode(t *testing.T) {
	var m image.Image

	for i := range encodeTests {
		srcFN := fmt.Sprintf("testdata%csrcimg%c%s", os.PathSeparator, os.PathSeparator, encodeTests[i].srcFN)
		dstFN := fmt.Sprintf("testdata%cactual%c%s", os.PathSeparator, os.PathSeparator, encodeTests[i].dstFN)
		expectedFN := fmt.Sprintf("testdata%cexpected%c%s", os.PathSeparator, os.PathSeparator, encodeTests[i].expectedFN)
		m = readImageFromFile(t, srcFN)
		writeImageToFile(t, m, dstFN, "bmp")
		compareFiles(t, expectedFN, dstFN)
	}
}

func decodeConfig(t *testing.T, shortFN string, hasPalette bool, pal_len int) {
	var err error
	var ok bool
	var fmtName string
	var cfg image.Config

	fn := fmt.Sprintf("testdata%csrcimg%c%s", os.PathSeparator, os.PathSeparator, shortFN)
	file, err := os.Open(fn)
	if err != nil {
		t.Logf("%s\n", err.Error())
		t.FailNow()
		return
	}
	defer file.Close()

	cfg, fmtName, err = image.DecodeConfig(file)
	if err != nil {
		t.Logf("%s\n", err.Error())
		t.Fail()
		return
	}

	if fmtName != "bmp" || cfg.Width != 31 || cfg.Height != 32 {
		t.Logf("%s: Wrong size or format name\n", shortFN)
		t.Fail()
		return
	}

	var pal color.Palette
	pal, ok = cfg.ColorModel.(color.Palette)
	if hasPalette {
		if !ok {
			t.Logf("DecodeConfig %s, Expected palette\n", fn)
			t.Fail()
		} else {
			if len(pal) != pal_len {
				t.Logf("DecodeConfig %s, Palette length expected %v, got %v\n", fn, pal_len, len(pal))
				t.Fail()
			}
		}
	} else {
		if ok {
			t.Logf("DecodeConfig %s, Unexpected palette\n", fn)
			t.Fail()
		}
	}
}

func TestDecodeConfig(t *testing.T) {
	decodeConfig(t, "rgb24.bmp", false, 0)
	decodeConfig(t, "pal8.bmp", true, 0)
}

type decodeTestType struct {
	srcFN, dstFN, expectedFN string
}

var decodeTests = []decodeTestType{
	{"rgb24.bmp", "rgb24.png", "rgb24.png"},
	{"pal8.bmp", "pal8.png", "pal8.png"},
	{"pal4.bmp", "pal4.png", "pal4.png"},
	{"pal1bg.bmp", "pal1bg.png", "pal1bg.png"},
	{"pal8offs.bmp", "pal8offs.png", "pal8.png"},
	{"pal8os2.bmp", "pal8os2.png", "pal8os2.png"},
	{"pal8v4.bmp", "pal8v4.png", "pal8.png"},
	{"pal8v5.bmp", "pal8v5.png", "pal8.png"},
	{"rgb16-565pal.bmp", "rgb16-565.png", "rgb16-565.png"},
}

func TestDecode(t *testing.T) {
	var m image.Image

	for i := range decodeTests {
		srcFN := fmt.Sprintf("testdata%csrcimg%c%s", os.PathSeparator, os.PathSeparator, decodeTests[i].srcFN)
		dstFN := fmt.Sprintf("testdata%cactual%c%s", os.PathSeparator, os.PathSeparator, decodeTests[i].dstFN)
		expectedFN := fmt.Sprintf("testdata%cexpected%c%s", os.PathSeparator, os.PathSeparator, decodeTests[i].expectedFN)
		m = readImageFromFile(t, srcFN)
		writeImageToFile(t, m, dstFN, "png")
		compareFiles(t, expectedFN, dstFN)
	}
}