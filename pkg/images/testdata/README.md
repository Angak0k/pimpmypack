# Test Images

This directory contains test images for the images package unit and integration tests.

## Test Files

| File | Size | Purpose |
|------|------|---------|
| `valid.jpg` | 8.0K | Valid JPEG image (800x600) - tests normal JPEG processing |
| `valid.png` | 1.9K | Valid PNG image (640x480) - tests PNG to JPEG conversion |
| `valid.webp` | 638B | Valid WebP image (500x500) - tests WebP to JPEG conversion |
| `large.jpg` | 139K | Large JPEG (3000x3000) - tests image resize functionality |
| `too_large.jpg` | 39M | Exceeds 5MB limit (6000x6000) - tests file size validation |
| `invalid.txt` | 43B | Plain text file - tests invalid format detection |
| `corrupted.jpg` | 33B | Corrupted JPEG header - tests corrupted file handling |

## Test Scenarios Covered

1. **Valid Format Processing**
   - JPEG → JPEG (minimal processing)
   - PNG → JPEG (format conversion)
   - WebP → JPEG (format conversion)

2. **Image Resize**
   - Large image (3000x3000) resized to max dimension (1920x1920)
   - Aspect ratio preservation during resize

3. **Validation**
   - File size limit enforcement (5MB max)
   - Format validation via magic bytes
   - Corrupted image detection

4. **Error Handling**
   - Invalid file format rejection
   - Corrupted image handling
   - Over-sized image rejection

## Generating Test Images

Test images are generated using:

- `generate_test_images.go` - Creates most test images
- `create_large_file.go` - Creates the 39MB file
- ImageMagick `convert` - Converts PNG to WebP

To regenerate:

```bash
cd pkg/images/testdata
go run generate_test_images.go
go run create_large_file.go
magick convert valid.webp.png valid.webp  # If you need to recreate WebP
```

## Usage in Tests

These files are used by:

- `processor_test.go` - Image processing unit tests
- Future integration tests for upload/download endpoints

Example:

```go
data, _ := os.ReadFile("testdata/valid.jpg")
processed, err := ProcessImage(data)
```
