package images

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// buildEXIFJPEG creates a JPEG with an EXIF APP1 segment containing the given orientation tag.
// The image is created with dimensions w×h (raw pixel dimensions before EXIF transform).
func buildEXIFJPEG(t *testing.T, w, h, orientation int, bigEndian bool) []byte {
	t.Helper()

	// Create a test image with an asymmetric color pattern for verification
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / max(w-1, 1)),
				G: uint8((y * 255) / max(h-1, 1)),
				B: 128,
				A: 255,
			})
		}
	}

	// Encode to JPEG
	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("failed to encode JPEG: %v", err)
	}
	jpegData := jpegBuf.Bytes()

	// Build EXIF APP1 segment with orientation tag
	exifSegment := buildEXIFSegment(orientation, bigEndian)

	// Insert APP1 segment right after SOI marker (first 2 bytes)
	var result bytes.Buffer
	result.Write(jpegData[:2]) // SOI: FF D8
	result.Write(exifSegment)
	result.Write(jpegData[2:]) // Rest of JPEG

	return result.Bytes()
}

// buildEXIFSegment creates a minimal EXIF APP1 segment with just the orientation tag.
func buildEXIFSegment(orientation int, bigEndian bool) []byte {
	var buf bytes.Buffer

	var bo binary.ByteOrder
	var boMarker []byte
	if bigEndian {
		bo = binary.BigEndian
		boMarker = []byte("MM")
	} else {
		bo = binary.LittleEndian
		boMarker = []byte("II")
	}

	// Build TIFF data
	var tiff bytes.Buffer

	// Byte order marker
	tiff.Write(boMarker)

	// TIFF magic number (42)
	magic := make([]byte, 2)
	bo.PutUint16(magic, 0x002A)
	tiff.Write(magic)

	// Offset to IFD0 (immediately after header = 8)
	ifdOffset := make([]byte, 4)
	bo.PutUint32(ifdOffset, 8)
	tiff.Write(ifdOffset)

	// IFD0: 1 entry
	numEntries := make([]byte, 2)
	bo.PutUint16(numEntries, 1)
	tiff.Write(numEntries)

	// IFD entry: orientation tag (0x0112), type SHORT (3), count 1, value
	tag := make([]byte, 2)
	bo.PutUint16(tag, 0x0112)
	tiff.Write(tag)

	typ := make([]byte, 2)
	bo.PutUint16(typ, 3) // SHORT
	tiff.Write(typ)

	count := make([]byte, 4)
	bo.PutUint32(count, 1)
	tiff.Write(count)

	// Value (4 bytes, value in first 2)
	val := make([]byte, 4)
	bo.PutUint16(val, uint16(orientation))
	tiff.Write(val)

	// Next IFD offset (0 = no more IFDs)
	nextIFD := make([]byte, 4)
	tiff.Write(nextIFD)

	// Build APP1 segment
	exifHeader := []byte("Exif\x00\x00")
	exifHeader = append(exifHeader, tiff.Bytes()...)
	segBody := exifHeader
	segLen := len(segBody) + 2 // +2 for the length field itself

	// APP1 marker + length + body
	buf.Write([]byte{0xFF, 0xE1})
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(segLen))
	buf.Write(lenBytes)
	buf.Write(segBody)

	return buf.Bytes()
}

// buildXMPThenEXIFJPEG creates a JPEG with a non-EXIF APP1 segment (XMP) followed
// by a real EXIF APP1 segment with the given orientation. This tests that the parser
// correctly skips non-EXIF APP1 segments.
func buildXMPThenEXIFJPEG(t *testing.T, w, h, orientation int) []byte {
	t.Helper()

	base := buildEXIFJPEG(t, w, h, orientation, false)

	// Build a fake XMP APP1 segment (starts with "http://ns.adobe.com/xap/", not "Exif\0\0")
	xmpBody := []byte("http://ns.adobe.com/xap/1.0/\x00<x:xmpmeta/>")
	xmpSegLen := len(xmpBody) + 2
	var xmpSeg bytes.Buffer
	xmpSeg.Write([]byte{0xFF, 0xE1})
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(xmpSegLen))
	xmpSeg.Write(lenBytes)
	xmpSeg.Write(xmpBody)

	// Insert XMP segment after SOI, before the existing EXIF APP1
	var result bytes.Buffer
	result.Write(base[:2]) // SOI
	result.Write(xmpSeg.Bytes())
	result.Write(base[2:]) // EXIF APP1 + rest of JPEG

	return result.Bytes()
}

func TestReadEXIFOrientation(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
	}{
		{
			name:     "JPEG with orientation 6 (little-endian)",
			data:     buildEXIFJPEG(t, 100, 80, 6, false),
			expected: 6,
		},
		{
			name:     "JPEG with orientation 6 (big-endian)",
			data:     buildEXIFJPEG(t, 100, 80, 6, true),
			expected: 6,
		},
		{
			name:     "JPEG with orientation 1",
			data:     buildEXIFJPEG(t, 100, 80, 1, false),
			expected: 1,
		},
		{
			name:     "JPEG with orientation 3",
			data:     buildEXIFJPEG(t, 100, 80, 3, false),
			expected: 3,
		},
		{
			name:     "JPEG with orientation 8",
			data:     buildEXIFJPEG(t, 100, 80, 8, true),
			expected: 8,
		},
		{
			name:     "JPEG without EXIF (plain JPEG)",
			data:     createTestJPEG(100, 80),
			expected: 1,
		},
		{
			name:     "PNG data",
			data:     createTestPNG(100, 80),
			expected: 1,
		},
		{
			name:     "Empty data",
			data:     []byte{},
			expected: 1,
		},
		{
			name:     "Short data",
			data:     []byte{0xFF, 0xD8},
			expected: 1,
		},
		{
			name:     "Not an image",
			data:     []byte("hello world this is not an image"),
			expected: 1,
		},
		{
			name:     "XMP APP1 before EXIF APP1",
			data:     buildXMPThenEXIFJPEG(t, 100, 80, 6),
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readEXIFOrientation(tt.data)
			if got != tt.expected {
				t.Errorf("readEXIFOrientation() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestApplyOrientation(t *testing.T) {
	// Create a 4x2 test image with known pixel colors:
	//   (0,0)=Red    (1,0)=Green  (2,0)=Blue   (3,0)=White
	//   (0,1)=Yellow (1,1)=Cyan   (2,1)=Magenta (3,1)=Black
	img := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	pixels := []struct {
		x, y int
		c    color.NRGBA
	}{
		{0, 0, color.NRGBA{255, 0, 0, 255}},     // Red
		{1, 0, color.NRGBA{0, 255, 0, 255}},     // Green
		{2, 0, color.NRGBA{0, 0, 255, 255}},     // Blue
		{3, 0, color.NRGBA{255, 255, 255, 255}}, // White
		{0, 1, color.NRGBA{255, 255, 0, 255}},   // Yellow
		{1, 1, color.NRGBA{0, 255, 255, 255}},   // Cyan
		{2, 1, color.NRGBA{255, 0, 255, 255}},   // Magenta
		{3, 1, color.NRGBA{0, 0, 0, 255}},       // Black
	}
	for _, p := range pixels {
		img.SetNRGBA(p.x, p.y, p.c)
	}

	red := color.NRGBA{255, 0, 0, 255}
	white := color.NRGBA{255, 255, 255, 255}
	yellow := color.NRGBA{255, 255, 0, 255}
	black := color.NRGBA{0, 0, 0, 255}

	tests := []struct {
		name        string
		orientation int
		wantW       int
		wantH       int
		// Check a few key pixel positions
		checks []struct {
			x, y int
			c    color.NRGBA
		}
	}{
		{
			name:        "Orientation 1 (normal)",
			orientation: 1,
			wantW:       4,
			wantH:       2,
			checks: []struct {
				x, y int
				c    color.NRGBA
			}{
				{0, 0, red}, {3, 0, white}, {0, 1, yellow}, {3, 1, black},
			},
		},
		{
			name:        "Orientation 2 (flip horizontal)",
			orientation: 2,
			wantW:       4,
			wantH:       2,
			checks: []struct {
				x, y int
				c    color.NRGBA
			}{
				{0, 0, white}, {3, 0, red}, {0, 1, black}, {3, 1, yellow},
			},
		},
		{
			name:        "Orientation 3 (rotate 180)",
			orientation: 3,
			wantW:       4,
			wantH:       2,
			checks: []struct {
				x, y int
				c    color.NRGBA
			}{
				{0, 0, black}, {3, 0, yellow}, {0, 1, white}, {3, 1, red},
			},
		},
		{
			name:        "Orientation 4 (flip vertical)",
			orientation: 4,
			wantW:       4,
			wantH:       2,
			checks: []struct {
				x, y int
				c    color.NRGBA
			}{
				{0, 0, yellow}, {3, 0, black}, {0, 1, red}, {3, 1, white},
			},
		},
		{
			name:        "Orientation 5 (transpose)",
			orientation: 5,
			wantW:       2,
			wantH:       4,
			checks: []struct {
				x, y int
				c    color.NRGBA
			}{
				{0, 0, red}, {1, 0, yellow}, {0, 3, white}, {1, 3, black},
			},
		},
		{
			name:        "Orientation 6 (rotate 90 CW)",
			orientation: 6,
			wantW:       2,
			wantH:       4,
			checks: []struct {
				x, y int
				c    color.NRGBA
			}{
				{0, 0, yellow}, {1, 0, red}, {0, 3, black}, {1, 3, white},
			},
		},
		{
			name:        "Orientation 7 (transverse)",
			orientation: 7,
			wantW:       2,
			wantH:       4,
			checks: []struct {
				x, y int
				c    color.NRGBA
			}{
				{0, 0, black}, {1, 0, white}, {0, 3, yellow}, {1, 3, red},
			},
		},
		{
			name:        "Orientation 8 (rotate 270 CW)",
			orientation: 8,
			wantW:       2,
			wantH:       4,
			checks: []struct {
				x, y int
				c    color.NRGBA
			}{
				{0, 0, white}, {1, 0, black}, {0, 3, red}, {1, 3, yellow},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyOrientation(img, tt.orientation)
			bounds := result.Bounds()

			if bounds.Dx() != tt.wantW || bounds.Dy() != tt.wantH {
				t.Errorf("dimensions = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.wantW, tt.wantH)
			}

			for _, check := range tt.checks {
				converted := color.NRGBAModel.Convert(result.At(check.x, check.y))
				got, ok := converted.(color.NRGBA)
				if !ok {
					t.Fatalf("pixel(%d,%d) failed color conversion", check.x, check.y)
				}
				if got != check.c {
					t.Errorf("pixel(%d,%d) = %v, want %v", check.x, check.y, got, check.c)
				}
			}
		})
	}
}

// TestProcessImage_EXIFOrientation6 verifies the full pipeline correctly rotates
// a JPEG with EXIF orientation 6 (iPhone portrait).
func TestProcessImage_EXIFOrientation6(t *testing.T) {
	// Stored pixels: 200w × 300h, tagged with EXIF orientation 6 (rotate 90° CW).
	// After applying the EXIF orientation, the output image should be 300w × 200h.
	data := buildEXIFJPEG(t, 200, 300, 6, false)

	result, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("ProcessImage() error: %v", err)
	}

	// After rotating 90° CW, a 200×300 image becomes 300×200
	if result.Metadata.Width != 300 || result.Metadata.Height != 200 {
		t.Errorf("dimensions = %dx%d, want 300x200", result.Metadata.Width, result.Metadata.Height)
	}
}

// TestProcessImage_NoEXIF verifies that images without EXIF data are processed unchanged.
func TestProcessImage_NoEXIF(t *testing.T) {
	data := createTestJPEG(200, 300)

	result, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("ProcessImage() error: %v", err)
	}

	if result.Metadata.Width != 200 || result.Metadata.Height != 300 {
		t.Errorf("dimensions = %dx%d, want 200x300", result.Metadata.Width, result.Metadata.Height)
	}
}

// TestProcessInventoryItemImage_EXIFOrientation6 verifies inventory item processing
// correctly handles EXIF orientation.
func TestProcessInventoryItemImage_EXIFOrientation6(t *testing.T) {
	// 300x400 pixels with orientation 6 → after rotation: 400x300
	data := buildEXIFJPEG(t, 300, 400, 6, false)

	result, err := ProcessInventoryItemImage(data)
	if err != nil {
		t.Fatalf("ProcessInventoryItemImage() error: %v", err)
	}

	// After rotation: 400×300, then resize to fit 400px max → already fits
	if result.Metadata.Width != 400 || result.Metadata.Height != 300 {
		t.Errorf("dimensions = %dx%d, want 400x300", result.Metadata.Width, result.Metadata.Height)
	}
}
