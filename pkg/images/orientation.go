package images

import (
	"encoding/binary"
	"image"
	"image/color"
)

// readEXIFOrientation parses JPEG EXIF data to extract the orientation tag.
// Returns 1-8 for valid orientations, defaults to 1 (normal) on any error,
// non-JPEG input, or missing EXIF data.
func readEXIFOrientation(data []byte) int {
	// Need at least enough bytes for JPEG SOI + APP1 marker
	if len(data) < 14 {
		return 1
	}

	// Verify JPEG SOI marker
	if data[0] != 0xFF || data[1] != 0xD8 {
		return 1
	}

	// Walk through JPEG markers to find APP1 (EXIF)
	offset := 2
	for offset+4 <= len(data) {
		if data[offset] != 0xFF {
			return 1
		}

		marker := data[offset+1]
		// Skip padding bytes
		if marker == 0xFF {
			offset++
			continue
		}

		// SOS marker — no more metadata segments
		if marker == 0xDA {
			return 1
		}

		// Read segment length
		if offset+4 > len(data) {
			return 1
		}
		segLen := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))
		if segLen < 2 {
			return 1
		}

		// APP1 marker (0xE1) — EXIF data
		if marker == 0xE1 {
			return parseEXIFOrientation(data[offset+4 : offset+2+segLen])
		}

		// Skip to next marker
		offset += 2 + segLen
	}

	return 1
}

// parseEXIFOrientation extracts the orientation value from an APP1 EXIF segment body.
func parseEXIFOrientation(exif []byte) int {
	// Verify "Exif\0\0" header
	if len(exif) < 14 || string(exif[0:6]) != "Exif\x00\x00" {
		return 1
	}

	tiff := exif[6:]

	// Determine byte order
	var bo binary.ByteOrder
	switch string(tiff[0:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 1
	}

	// Verify TIFF magic number (42)
	if bo.Uint16(tiff[2:4]) != 0x002A {
		return 1
	}

	// Get offset to IFD0
	ifdOffset := int(bo.Uint32(tiff[4:8]))
	if ifdOffset+2 > len(tiff) {
		return 1
	}

	// Read number of IFD entries
	numEntries := int(bo.Uint16(tiff[ifdOffset : ifdOffset+2]))
	entryStart := ifdOffset + 2

	// Walk IFD0 entries looking for orientation tag (0x0112)
	for i := 0; i < numEntries; i++ {
		entryOffset := entryStart + i*12
		if entryOffset+12 > len(tiff) {
			return 1
		}

		tag := bo.Uint16(tiff[entryOffset : entryOffset+2])
		if tag == 0x0112 {
			// Orientation tag found — value is a SHORT (uint16)
			val := int(bo.Uint16(tiff[entryOffset+8 : entryOffset+10]))
			if val >= 1 && val <= 8 {
				return val
			}
			return 1
		}
	}

	return 1
}

// applyOrientation transforms an image according to the EXIF orientation value.
// Returns the original image unchanged for orientation 1 (normal) or invalid values.
//
// EXIF orientation values:
//
//	1 = Normal (no transform)
//	2 = Flip horizontal
//	3 = Rotate 180°
//	4 = Flip vertical
//	5 = Transpose (flip horizontal + rotate 270° CW)
//	6 = Rotate 90° CW
//	7 = Transverse (flip horizontal + rotate 90° CW)
//	8 = Rotate 270° CW
func applyOrientation(img image.Image, orientation int) image.Image {
	if orientation <= 1 || orientation > 8 {
		return img
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// For orientations 5-8, width and height are swapped
	var dst *image.NRGBA
	if orientation >= 5 {
		dst = image.NewNRGBA(image.Rect(0, 0, h, w))
	} else {
		dst = image.NewNRGBA(image.Rect(0, 0, w, h))
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			var dx, dy int
			switch orientation {
			case 2: // Flip horizontal
				dx, dy = w-1-x, y
			case 3: // Rotate 180
				dx, dy = w-1-x, h-1-y
			case 4: // Flip vertical
				dx, dy = x, h-1-y
			case 5: // Transpose
				dx, dy = y, x
			case 6: // Rotate 90 CW
				dx, dy = h-1-y, x
			case 7: // Transverse
				dx, dy = h-1-y, w-1-x
			case 8: // Rotate 270 CW
				dx, dy = y, w-1-x
			}
			dst.Set(dx, dy, color.NRGBAModel.Convert(c))
		}
	}

	return dst
}
