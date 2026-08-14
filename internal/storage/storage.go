package storage

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const FallbackDir = "picture/screnshoot"

func init() {
	_ = godotenv.Load()
}

// GetDefaultDir returns configured directory from FASTSS_OUTPUT_DIR or default fallback
func GetDefaultDir() string {
	_ = godotenv.Load()
	dir := os.Getenv("FASTSS_OUTPUT_DIR")
	if strings.TrimSpace(dir) != "" {
		return strings.TrimSpace(dir)
	}
	return FallbackDir
}

// SaveImage saves an image to a destination file. If outPath is empty or a directory,
// a timestamped PNG file will be generated in that directory.
func SaveImage(img image.Image, outPath string) (string, error) {
	if strings.TrimSpace(outPath) == "" {
		outPath = GetDefaultDir()
	}

	// Check if outPath has an image extension
	ext := strings.ToLower(filepath.Ext(outPath))
	var targetFile string

	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		targetFile = outPath
	} else {
		// Treat as a directory and append timestamped filename
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("screenshot_%s.png", timestamp)
		targetFile = filepath.Join(outPath, filename)
		ext = ".png"
	}

	// Ensure target directory exists
	dir := filepath.Dir(targetFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.Create(targetFile)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", targetFile, err)
	}
	defer f.Close()

	if ext == ".jpg" || ext == ".jpeg" {
		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 95}); err != nil {
			return "", fmt.Errorf("failed to encode JPEG image: %w", err)
		}
	} else {
		if err := png.Encode(f, img); err != nil {
			return "", fmt.Errorf("failed to encode PNG image: %w", err)
		}
	}

	absPath, err := filepath.Abs(targetFile)
	if err != nil {
		return targetFile, nil
	}

	return absPath, nil
}
