package images

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // Register WebP decoder
)

const (
	// MaxUploadSize is the maximum file size for uploads (5 MB)
	MaxUploadSize = 5 * 1024 * 1024
	// MaxProcessedSize is the maximum size after processing (5 MB)
	MaxProcessedSize = 5 * 1024 * 1024
	// MaxDimension is the maximum width or height for processed images
	MaxDimension = 1920
	// JPEGQuality is the quality setting for JPEG encoding (0-100)
	JPEGQuality = 85
	// MimeTypeJPEG is the MIME type for JPEG images
	MimeTypeJPEG = "image/jpeg"
	// MimeTypePNG is the MIME type for PNG images
	MimeTypePNG = "image/png"
	// MimeTypeWebP is the MIME type for WebP images
	MimeTypeWebP = "image/webp"
)

// ValidateImageFormat checks if the image format is supported by examining magic bytes
func ValidateImageFormat(data []byte) (string, error) {
	if len(data) < 12 {
		return "", ErrInvalidFormat
	}

	// Check JPEG magic bytes (FF D8 FF)
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return MimeTypeJPEG, nil
	}

	// Check PNG magic bytes (89 50 4E 47 0D 0A 1A 0A)
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return MimeTypePNG, nil
	}

	// Check WebP magic bytes (RIFF....WEBP)
	if len(data) >= 12 &&
		data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P' {
		return MimeTypeWebP, nil
	}

	return "", ErrInvalidFormat
}

// DecodeImage decodes an image from bytes
func DecodeImage(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCorrupted, err)
	}
	return img, nil
}

// ResizeImage resizes an image if it exceeds MaxDimension, maintaining aspect ratio
func ResizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is within limits, return as-is
	if width <= MaxDimension && height <= MaxDimension {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight int
	if width > height {
		newWidth = MaxDimension
		newHeight = (height * MaxDimension) / width
		if newHeight < 1 {
			newHeight = 1
		}
	} else {
		newHeight = MaxDimension
		newWidth = (width * MaxDimension) / height
		if newWidth < 1 {
			newWidth = 1
		}
	}

	// Create new image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Use high-quality scaling
	draw.CatmullRom.Scale(dst, dst.Rect, img, bounds, draw.Over, nil)

	return dst
}

// EncodeToJPEG encodes an image to JPEG format with specified quality
// This also effectively strips EXIF data as we're re-encoding from scratch
func EncodeToJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer

	opts := &jpeg.Options{Quality: JPEGQuality}
	err := jpeg.Encode(&buf, img, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

// ProcessImage validates and processes an uploaded image
// Returns the processed image data and metadata
func ProcessImage(data []byte) (*ProcessedImage, error) {
	// Validate upload size
	if len(data) > MaxUploadSize {
		return nil, fmt.Errorf("%w: upload size %d exceeds %d bytes",
			ErrTooLarge, len(data), MaxUploadSize)
	}

	// Validate format
	_, err := ValidateImageFormat(data)
	if err != nil {
		return nil, err
	}

	// Decode image
	img, err := DecodeImage(data)
	if err != nil {
		return nil, err
	}

	// Resize if necessary
	img = ResizeImage(img)

	// Convert to JPEG (this also strips EXIF data)
	processedData, err := EncodeToJPEG(img)
	if err != nil {
		return nil, err
	}

	// Validate processed size
	if len(processedData) > MaxProcessedSize {
		return nil, fmt.Errorf("%w: processed size %d exceeds %d bytes",
			ErrTooLarge, len(processedData), MaxProcessedSize)
	}

	// Get dimensions
	bounds := img.Bounds()

	return &ProcessedImage{
		Data: processedData,
		Metadata: ImageMetadata{
			MimeType: MimeTypeJPEG,
			FileSize: len(processedData),
			Width:    bounds.Dx(),
			Height:   bounds.Dy(),
		},
	}, nil
}

// ProcessImageFromReader processes an image from an io.Reader
func ProcessImageFromReader(reader io.Reader) (*ProcessedImage, error) {
	// Read all data
	data, err := io.ReadAll(io.LimitReader(reader, MaxUploadSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return ProcessImage(data)
}
