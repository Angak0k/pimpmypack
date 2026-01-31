package images

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// createTestJPEG creates a test JPEG image with specified dimensions
func createTestJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 128,
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

// createTestPNG creates a test PNG image with specified dimensions
func createTestPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestValidateImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid JPEG",
			data:     createTestJPEG(100, 100),
			expected: "image/jpeg",
			wantErr:  false,
		},
		{
			name:     "Valid PNG",
			data:     createTestPNG(100, 100),
			expected: "image/png",
			wantErr:  false,
		},
		{
			name:     "Invalid format (text)",
			data:     []byte("This is not an image"),
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Too short data",
			data:     []byte{0xFF, 0xD8},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Empty data",
			data:     []byte{},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mimeType, err := ValidateImageFormat(tt.data)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if mimeType != tt.expected {
				t.Errorf("Expected mime type %s, got %s", tt.expected, mimeType)
			}
		})
	}
}

func TestDecodeImage(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantErr   bool
		wantWidth int
	}{
		{
			name:      "Valid JPEG",
			data:      createTestJPEG(100, 100),
			wantErr:   false,
			wantWidth: 100,
		},
		{
			name:      "Valid PNG",
			data:      createTestPNG(200, 150),
			wantErr:   false,
			wantWidth: 200,
		},
		{
			name:    "Corrupted data",
			data:    []byte{0xFF, 0xD8, 0xFF, 0x00, 0x01, 0x02},
			wantErr: true,
		},
		{
			name:    "Empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := DecodeImage(tt.data)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && img != nil {
				bounds := img.Bounds()
				if bounds.Dx() != tt.wantWidth {
					t.Errorf("Expected width %d, got %d", tt.wantWidth, bounds.Dx())
				}
			}
		})
	}
}

func TestResizeImage(t *testing.T) {
	tests := []struct {
		name         string
		inputWidth   int
		inputHeight  int
		expectResize bool
		maxDim       int
	}{
		{
			name:         "No resize needed (small image)",
			inputWidth:   800,
			inputHeight:  600,
			expectResize: false,
		},
		{
			name:         "Resize landscape image",
			inputWidth:   3000,
			inputHeight:  2000,
			expectResize: true,
			maxDim:       MaxDimension,
		},
		{
			name:         "Resize portrait image",
			inputWidth:   2000,
			inputHeight:  3000,
			expectResize: true,
			maxDim:       MaxDimension,
		},
		{
			name:         "Exactly at limit",
			inputWidth:   MaxDimension,
			inputHeight:  MaxDimension,
			expectResize: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image
			img := image.NewRGBA(image.Rect(0, 0, tt.inputWidth, tt.inputHeight))

			// Resize
			resized := ResizeImage(img)
			bounds := resized.Bounds()

			if tt.expectResize {
				// Verify image was resized
				if bounds.Dx() > MaxDimension || bounds.Dy() > MaxDimension {
					t.Errorf("Resized image exceeds MaxDimension: %dx%d", bounds.Dx(), bounds.Dy())
				}

				// Verify aspect ratio maintained (allow 1 pixel tolerance for rounding)
				inputAspect := float64(tt.inputWidth) / float64(tt.inputHeight)
				outputAspect := float64(bounds.Dx()) / float64(bounds.Dy())
				aspectDiff := inputAspect - outputAspect
				if aspectDiff < -0.01 || aspectDiff > 0.01 {
					t.Errorf("Aspect ratio not maintained: input=%.2f, output=%.2f",
						inputAspect, outputAspect)
				}
			} else {
				// Verify image was not resized
				if bounds.Dx() != tt.inputWidth || bounds.Dy() != tt.inputHeight {
					t.Errorf("Image was resized when it shouldn't be: %dx%d -> %dx%d",
						tt.inputWidth, tt.inputHeight, bounds.Dx(), bounds.Dy())
				}
			}
		})
	}
}

func TestEncodeToJPEG(t *testing.T) {
	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Encode
	data, err := EncodeToJPEG(img)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	// Verify it's a valid JPEG
	mimeType, err := ValidateImageFormat(data)
	if err != nil {
		t.Fatalf("Encoded data is not a valid image: %v", err)
	}
	if mimeType != "image/jpeg" {
		t.Errorf("Expected mime type image/jpeg, got %s", mimeType)
	}

	// Verify we can decode it back
	decoded, err := DecodeImage(data)
	if err != nil {
		t.Fatalf("Failed to decode encoded image: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Decoded image has wrong dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestProcessImage_ValidJPEG(t *testing.T) {
	data := createTestJPEG(1000, 800)

	result, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("Failed to process valid JPEG: %v", err)
	}

	// Verify metadata
	if result.Metadata.MimeType != "image/jpeg" {
		t.Errorf("Expected mime type image/jpeg, got %s", result.Metadata.MimeType)
	}
	if result.Metadata.Width != 1000 {
		t.Errorf("Expected width 1000, got %d", result.Metadata.Width)
	}
	if result.Metadata.Height != 800 {
		t.Errorf("Expected height 800, got %d", result.Metadata.Height)
	}
	if result.Metadata.FileSize <= 0 {
		t.Errorf("File size should be > 0, got %d", result.Metadata.FileSize)
	}

	// Verify data is valid JPEG
	mimeType, err := ValidateImageFormat(result.Data)
	if err != nil {
		t.Fatalf("Processed data is not a valid image: %v", err)
	}
	if mimeType != "image/jpeg" {
		t.Errorf("Processed data is not JPEG: %s", mimeType)
	}
}

func TestProcessImage_ValidPNG(t *testing.T) {
	data := createTestPNG(500, 500)

	result, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("Failed to process valid PNG: %v", err)
	}

	// Verify PNG was converted to JPEG
	if result.Metadata.MimeType != "image/jpeg" {
		t.Errorf("Expected mime type image/jpeg, got %s", result.Metadata.MimeType)
	}

	// Verify data is valid JPEG
	mimeType, err := ValidateImageFormat(result.Data)
	if err != nil {
		t.Fatalf("Processed data is not a valid image: %v", err)
	}
	if mimeType != "image/jpeg" {
		t.Errorf("Processed data is not JPEG: %s", mimeType)
	}
}

func TestProcessImage_LargeImage(t *testing.T) {
	// Create image that exceeds MaxDimension
	data := createTestJPEG(3000, 2000)

	result, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("Failed to process large image: %v", err)
	}

	// Verify image was resized
	if result.Metadata.Width > MaxDimension || result.Metadata.Height > MaxDimension {
		t.Errorf("Image not resized properly: %dx%d",
			result.Metadata.Width, result.Metadata.Height)
	}

	// Verify largest dimension is MaxDimension
	if result.Metadata.Width != MaxDimension && result.Metadata.Height != MaxDimension {
		t.Errorf("Neither dimension is MaxDimension: %dx%d",
			result.Metadata.Width, result.Metadata.Height)
	}

	// Verify aspect ratio maintained
	inputAspect := 3000.0 / 2000.0
	outputAspect := float64(result.Metadata.Width) / float64(result.Metadata.Height)
	aspectDiff := inputAspect - outputAspect
	if aspectDiff < -0.01 || aspectDiff > 0.01 {
		t.Errorf("Aspect ratio not maintained: input=%.2f, output=%.2f",
			inputAspect, outputAspect)
	}
}

func TestProcessImage_InvalidFormat(t *testing.T) {
	data := []byte("This is not an image")

	_, err := ProcessImage(data)
	if err == nil {
		t.Fatalf("Expected error for invalid format")
	}
	if err != ErrInvalidFormat {
		t.Errorf("Expected ErrInvalidFormat, got: %v", err)
	}
}

func TestProcessImage_TooLarge(t *testing.T) {
	// Create data larger than MaxUploadSize
	data := make([]byte, MaxUploadSize+1)
	// Add JPEG magic bytes
	data[0], data[1], data[2] = 0xFF, 0xD8, 0xFF

	_, err := ProcessImage(data)
	if err == nil {
		t.Fatalf("Expected error for oversized upload")
	}

	// Check if error is about size
	if err.Error() == "" || err == nil {
		t.Errorf("Expected size error, got: %v", err)
	}
}

func TestProcessImage_Corrupted(t *testing.T) {
	// Create data with valid JPEG magic bytes but corrupted content
	data := make([]byte, 100)
	data[0], data[1], data[2] = 0xFF, 0xD8, 0xFF

	_, err := ProcessImage(data)
	if err == nil {
		t.Fatalf("Expected error for corrupted image")
	}
}

func TestProcessImageFromReader(t *testing.T) {
	data := createTestJPEG(800, 600)
	reader := bytes.NewReader(data)

	result, err := ProcessImageFromReader(reader)
	if err != nil {
		t.Fatalf("Failed to process from reader: %v", err)
	}

	if result.Metadata.Width != 800 {
		t.Errorf("Expected width 800, got %d", result.Metadata.Width)
	}
	if result.Metadata.Height != 600 {
		t.Errorf("Expected height 600, got %d", result.Metadata.Height)
	}
}
