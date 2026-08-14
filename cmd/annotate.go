package cmd

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"fastss/internal/draw"
	"fastss/internal/ocr"
	"fastss/internal/storage"

	"github.com/spf13/cobra"
)

var annotateCmd = &cobra.Command{
	Use:   "annotate [image-file]",
	Short: "Annotate an existing screenshot or image file using OCR",
	Args:  cobra.ExactArgs(1),
	RunE:  runAnnotate,
}

func init() {
	annotateCmd.Flags().StringSliceVar(&boxQueries, "box", nil, "Text to draw a box around")
	annotateCmd.Flags().StringSliceVar(&arrowQueries, "arrow", nil, "Text to point an arrow at")
	annotateCmd.Flags().StringVar(&arrowFrom, "arrow-from", "top-left", "Direction arrow originates from ('top', 'bottom', 'left', 'right', 'top-left', etc.)")
	annotateCmd.Flags().StringSliceVar(&highlightList, "highlight", nil, "Text to highlight")
	annotateCmd.Flags().StringSliceVar(&blurList, "blur", nil, "Text to blur/censor")
	annotateCmd.Flags().StringSliceVar(&badgeList, "badge", nil, "Draw badge on text, format 'TEXT:LABEL'")
	annotateCmd.Flags().StringVar(&colorName, "color", "red", "Annotation color (red, green, blue, yellow, orange, magenta, cyan, hex)")
	annotateCmd.Flags().Float64Var(&strokeWidth, "stroke", 4.0, "Stroke width for boxes and arrows")
	annotateCmd.Flags().StringVarP(&outputPath, "output", "o", storage.GetDefaultDir(), "Destination output path or folder")
	annotateCmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case-sensitive text matching for OCR")
	annotateCmd.Flags().BoolVar(&ocrDump, "ocr-dump", false, "Print all detected OCR text lines and words")

	rootCmd.AddCommand(annotateCmd)
}

func runAnnotate(cmd *cobra.Command, args []string) error {
	imagePath := args[0]
	f, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image '%s': %w", imagePath, err)
	}
	defer f.Close()

	rawImg, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	fmt.Printf("🔍 Running OCR engine on '%s'...\n", imagePath)
	ocrRes, err := ocr.RecognizeFile(imagePath)
	if err != nil {
		return fmt.Errorf("OCR engine failed: %w", err)
	}

	if ocrDump {
		fmt.Println("\n--- OCR Text Detection Dump ---")
		for idx, line := range ocrRes.Lines {
			fmt.Printf("[%d] '%s' (Bounds: %v)\n", idx+1, line.Text, line.Bounds.Rect())
		}
		fmt.Println("-------------------------------")
	}

	annotator := draw.NewAnnotator(rawImg)
	annotatedCount := 0

	// 1. Process Boxes
	for _, target := range boxQueries {
		matches, err := ocr.FindText(ocrRes, target, caseSensitive)
		if err != nil || len(matches) == 0 {
			fmt.Printf("⚠️ Box: Text '%s' not found in image\n", target)
			continue
		}
		for _, m := range matches {
			fmt.Printf("📦 Box added around '%s' at %v\n", m.MatchedText, m.Bounds)
			annotator.DrawBox(m.Bounds, colorName, strokeWidth, 4.0, 4.0)
			annotatedCount++
		}
	}

	// 2. Process Arrows
	for _, target := range arrowQueries {
		matches, err := ocr.FindText(ocrRes, target, caseSensitive)
		if err != nil || len(matches) == 0 {
			fmt.Printf("⚠️ Arrow: Text '%s' not found in image\n", target)
			continue
		}
		for _, m := range matches {
			fmt.Printf("🏹 Arrow pointing to '%s' at %v\n", m.MatchedText, m.Center)
			annotator.DrawArrow(m.Center, arrowFrom, colorName, strokeWidth, 55.0)
			annotatedCount++
		}
	}

	// 3. Process Highlights
	for _, target := range highlightList {
		matches, err := ocr.FindText(ocrRes, target, caseSensitive)
		if err != nil || len(matches) == 0 {
			fmt.Printf("⚠️ Highlight: Text '%s' not found in image\n", target)
			continue
		}
		for _, m := range matches {
			fmt.Printf("✨ Highlight added over '%s' at %v\n", m.MatchedText, m.Bounds)
			annotator.DrawHighlight(m.Bounds, colorName, 2.0)
			annotatedCount++
		}
	}

	// 4. Process Blurs
	for _, target := range blurList {
		matches, err := ocr.FindText(ocrRes, target, caseSensitive)
		if err != nil || len(matches) == 0 {
			fmt.Printf("⚠️ Blur: Text '%s' not found in image\n", target)
			continue
		}
		for _, m := range matches {
			fmt.Printf("🔒 Blur applied to '%s' at %v\n", m.MatchedText, m.Bounds)
			annotator.DrawBlur(m.Bounds, 8)
			annotatedCount++
		}
	}

	// 5. Process Badges
	for _, badgeItem := range badgeList {
		parts := strings.SplitN(badgeItem, ":", 2)
		target := parts[0]
		badgeText := "1"
		if len(parts) > 1 {
			badgeText = parts[1]
		}

		matches, err := ocr.FindText(ocrRes, target, caseSensitive)
		if err != nil || len(matches) == 0 {
			fmt.Printf("⚠️ Badge: Text '%s' not found in image\n", target)
			continue
		}
		for _, m := range matches {
			fmt.Printf("🏷️ Badge '%s' added to '%s' at %v\n", badgeText, m.MatchedText, m.Bounds.Min)
			annotator.DrawBadge(m.Bounds.Min, badgeText, colorName, "white")
			annotatedCount++
		}
	}

	finalImg := annotator.Result()
	fmt.Printf("🎨 Successfully rendered %d annotation(s)\n", annotatedCount)

	savedFile, err := storage.SaveImage(finalImg, outputPath)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	fmt.Printf("✅ Annotated image saved successfully:\n   📁 %s\n", savedFile)
	return nil
}
