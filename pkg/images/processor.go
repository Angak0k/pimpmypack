package images

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"

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
	// MaxProfilePicDimension is the maximum width or height for profile pictures
	MaxProfilePicDimension = 512
	// MaxInventoryItemDimension is the maximum width or height for inventory item pictures
	MaxInventoryItemDimension = 400
	// JPEGQuality is the quality setting for JPEG encoding (0-100)
	JPEGQuality = 85
	// InventoryItemJPEGQuality is the quality setting for inventory item images (0-100)
	InventoryItemJPEGQuality = 70
	// BannerJPEGQuality is the quality setting for banner images (0-100)
	BannerJPEGQuality = 95
	// MaxBannerWidth is the maximum width for banner images
	MaxBannerWidth = 1920
	// MaxBannerHeight is the maximum height for banner images
	MaxBannerHeight = 600
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
	// Apply EXIF orientation correction (no-op for non-JPEG or orientation=1)
	orientation := readEXIFOrientation(data)
	img = applyOrientation(img, orientation)
	return img, nil
}

// ResizeImageWithMax resizes an image if it exceeds maxDim, maintaining aspect ratio
func ResizeImageWithMax(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is within limits, return as-is
	if width <= maxDim && height <= maxDim {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight int
	if width > height {
		newWidth = maxDim
		newHeight = (height * maxDim) / width
		if newHeight < 1 {
			newHeight = 1
		}
	} else {
		newHeight = maxDim
		newWidth = (width * maxDim) / height
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

// ResizeImage resizes an image if it exceeds MaxDimension, maintaining aspect ratio
func ResizeImage(img image.Image) image.Image {
	return ResizeImageWithMax(img, MaxDimension)
}

// EncodeToJPEGWithQuality encodes an image to JPEG format with the given quality (1-100)
// Values outside the range are clamped. This also effectively strips EXIF data.
func EncodeToJPEGWithQuality(img image.Image, quality int) ([]byte, error) {
	quality = max(1, min(100, quality))

	var buf bytes.Buffer

	opts := &jpeg.Options{Quality: quality}
	err := jpeg.Encode(&buf, img, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

// EncodeToJPEG encodes an image to JPEG format with default quality (JPEGQuality)
func EncodeToJPEG(img image.Image) ([]byte, error) {
	return EncodeToJPEGWithQuality(img, JPEGQuality)
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

// ProcessProfileImage validates and processes an uploaded profile picture
// Uses MaxProfilePicDimension (512px) instead of MaxDimension (1920px)
func ProcessProfileImage(data []byte) (*ProcessedImage, error) {
	if len(data) > MaxUploadSize {
		return nil, fmt.Errorf("%w: upload size %d exceeds %d bytes",
			ErrTooLarge, len(data), MaxUploadSize)
	}

	_, err := ValidateImageFormat(data)
	if err != nil {
		return nil, err
	}

	img, err := DecodeImage(data)
	if err != nil {
		return nil, err
	}

	img = ResizeImageWithMax(img, MaxProfilePicDimension)

	processedData, err := EncodeToJPEG(img)
	if err != nil {
		return nil, err
	}

	if len(processedData) > MaxProcessedSize {
		return nil, fmt.Errorf("%w: processed size %d exceeds %d bytes",
			ErrTooLarge, len(processedData), MaxProcessedSize)
	}

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

// ProcessProfileImageFromReader processes a profile image from an io.Reader
func ProcessProfileImageFromReader(reader io.Reader) (*ProcessedImage, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxUploadSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return ProcessProfileImage(data)
}

// ProcessInventoryItemImage validates and processes an uploaded inventory item picture
// Uses MaxInventoryItemDimension (400px) and InventoryItemJPEGQuality (70%)
func ProcessInventoryItemImage(data []byte) (*ProcessedImage, error) {
	if len(data) > MaxUploadSize {
		return nil, fmt.Errorf("%w: upload size %d exceeds %d bytes",
			ErrTooLarge, len(data), MaxUploadSize)
	}

	_, err := ValidateImageFormat(data)
	if err != nil {
		return nil, err
	}

	img, err := DecodeImage(data)
	if err != nil {
		return nil, err
	}

	img = ResizeImageWithMax(img, MaxInventoryItemDimension)

	processedData, err := EncodeToJPEGWithQuality(img, InventoryItemJPEGQuality)
	if err != nil {
		return nil, err
	}

	if len(processedData) > MaxProcessedSize {
		return nil, fmt.Errorf("%w: processed size %d exceeds %d bytes",
			ErrTooLarge, len(processedData), MaxProcessedSize)
	}

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

// ProcessInventoryItemImageFromReader processes an inventory item image from an io.Reader
func ProcessInventoryItemImageFromReader(reader io.Reader) (*ProcessedImage, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxUploadSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return ProcessInventoryItemImage(data)
}

// ResizeBannerImage scales and center-crops an image to exactly MaxBannerWidth x MaxBannerHeight
func ResizeBannerImage(img image.Image) image.Image {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	if origW == MaxBannerWidth && origH == MaxBannerHeight {
		return img
	}

	// Compute the source crop rectangle matching the target aspect ratio,
	// then scale directly to the target size. This avoids allocating a
	// potentially huge intermediate image for extreme aspect ratios.
	targetAspect := float64(MaxBannerWidth) / float64(MaxBannerHeight)
	sourceAspect := float64(origW) / float64(origH)

	var cropRect image.Rectangle
	if sourceAspect > targetAspect {
		// Source is wider than target — crop sides
		cropH := origH
		cropW := int(math.Round(float64(cropH) * targetAspect))
		cropX := (origW - cropW) / 2
		cropRect = image.Rect(bounds.Min.X+cropX, bounds.Min.Y, bounds.Min.X+cropX+cropW, bounds.Min.Y+cropH)
	} else {
		// Source is taller than target — crop top/bottom
		cropW := origW
		cropH := int(math.Round(float64(cropW) / targetAspect))
		cropY := (origH - cropH) / 2
		cropRect = image.Rect(bounds.Min.X, bounds.Min.Y+cropY, bounds.Min.X+cropW, bounds.Min.Y+cropY+cropH)
	}

	dst := image.NewRGBA(image.Rect(0, 0, MaxBannerWidth, MaxBannerHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, cropRect, draw.Over, nil)

	return dst
}

// ProcessBannerImage validates and processes an uploaded banner image
func ProcessBannerImage(data []byte) (*ProcessedImage, error) {
	if len(data) > MaxUploadSize {
		return nil, fmt.Errorf("%w: upload size %d exceeds %d bytes",
			ErrTooLarge, len(data), MaxUploadSize)
	}

	_, err := ValidateImageFormat(data)
	if err != nil {
		return nil, err
	}

	img, err := DecodeImage(data)
	if err != nil {
		return nil, err
	}

	img = ResizeBannerImage(img)

	processedData, err := EncodeToJPEGWithQuality(img, BannerJPEGQuality)
	if err != nil {
		return nil, err
	}

	if len(processedData) > MaxProcessedSize {
		return nil, fmt.Errorf("%w: processed size %d exceeds %d bytes",
			ErrTooLarge, len(processedData), MaxProcessedSize)
	}

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

// ProcessBannerImageFromReader processes a banner image from an io.Reader
func ProcessBannerImageFromReader(reader io.Reader) (*ProcessedImage, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxUploadSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return ProcessBannerImage(data)
}
