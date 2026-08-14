package cmd

import (
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	"fastss/internal/capture"
	"fastss/internal/draw"
	"fastss/internal/ocr"
	"fastss/internal/storage"

	"github.com/spf13/cobra"
)

var (
	windowQuery   string
	boxQueries    []string
	arrowQueries  []string
	arrowFrom     string
	highlightList []string
	blurList      []string
	badgeList     []string
	colorName     string
	strokeWidth   float64
	outputPath    string
	delaySeconds  int
	listWindows   bool
	caseSensitive bool
	ocrDump       bool
	markAll       bool
	globalNth     int
	boxWidth      float64
	boxHeight     float64
	boxPadding    float64
	boxRadius     float64
)

var rootCmd = &cobra.Command{
	Use:   "fastss",
	Short: "FastSS - CLI Tool for Smart Screenshot Capture and Auto-Annotation",
	Long: `FastSS is an automated CLI screenshot tool written in Go.
It captures your screen or a specific application window, automatically runs OCR
to locate text, and draws boxes, arrows, highlights, and blurs directly on targets.

Examples:
  # Capture full screen and draw red box on first 'Login':
  fastss --box "Login"

  # Custom Box Size (Width, Height, Padding, Radius):
  fastss --box "Login" --box-width 200 --box-height 60
  fastss --box "Submit|200x50"
  fastss --box "Save|w=180,h=50,pad=10,radius=8"
  fastss --box "Delete" --padding 12 --radius 8

  # Target specific occurrence (1st, 2nd, last, all):
  fastss --box "Submit[1]"
  fastss --box "Submit[2]"
  fastss --box "Submit[last]"
  fastss --box "Submit[all]"

  # Target text near / below a specific context anchor:
  fastss --box "Edit near Profile"
  fastss --arrow "Save below Settings"

  # Capture a specific window:
  fastss -w "Chrome" --arrow "Submit" --color red
`,
	RunE: runCapture,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&windowQuery, "window", "w", "fullscreen", "Window title to capture ('fullscreen', 'active', or partial title like 'Chrome')")
	rootCmd.Flags().StringSliceVar(&boxQueries, "box", nil, "Text to draw a box around (supports 'Text|200x50', 'Text|w=200,h=50,pad=10', 'Text[1]', 'Text near Anchor')")
	rootCmd.Flags().StringSliceVar(&arrowQueries, "arrow", nil, "Text to point an arrow at")
	rootCmd.Flags().StringVar(&arrowFrom, "arrow-from", "top-left", "Direction arrow originates from ('top', 'bottom', 'left', 'right', 'top-left', 'top-right', 'bottom-left', 'bottom-right')")
	rootCmd.Flags().StringSliceVar(&highlightList, "highlight", nil, "Text to highlight with semi-transparent color")
	rootCmd.Flags().StringSliceVar(&blurList, "blur", nil, "Text to blur/censor")
	rootCmd.Flags().StringSliceVar(&badgeList, "badge", nil, "Draw badge on text, format 'TEXT:LABEL' (e.g. 'Submit:Step 1')")
	rootCmd.Flags().StringVar(&colorName, "color", "red", "Primary annotation color (red, green, blue, yellow, orange, magenta, cyan, or hex code #FF0000)")
	rootCmd.Flags().Float64Var(&strokeWidth, "stroke", 4.0, "Stroke width for boxes and arrows")
	rootCmd.Flags().Float64VarP(&boxWidth, "box-width", "W", 0, "Custom fixed width for box in pixels (e.g. --box-width 200)")
	rootCmd.Flags().Float64VarP(&boxHeight, "box-height", "H", 0, "Custom fixed height for box in pixels (e.g. --box-height 60)")
	rootCmd.Flags().Float64VarP(&boxPadding, "padding", "p", 6.0, "Padding around text for bounding box (default: 6.0)")
	rootCmd.Flags().Float64VarP(&boxRadius, "radius", "r", 4.0, "Corner radius for rounded bounding box (default: 4.0)")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", storage.GetDefaultDir(), "Destination output path or folder (default: configured in .env or picture/screnshoot)")
	rootCmd.Flags().IntVarP(&delaySeconds, "delay", "d", 0, "Delay in seconds before capturing screenshot")
	rootCmd.Flags().BoolVarP(&listWindows, "list", "l", false, "List all visible open windows and exit")
	rootCmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case-sensitive text matching for OCR")
	rootCmd.Flags().BoolVar(&ocrDump, "ocr-dump", false, "Print all detected OCR text lines and words")
	rootCmd.Flags().BoolVarP(&markAll, "all", "a", false, "Mark ALL matching occurrences of text (default is 1st match unless specified)")
	rootCmd.Flags().IntVarP(&globalNth, "nth", "n", 0, "Select N-th occurrence (1 for 1st, 2 for 2nd, -1 for last)")
}

func applyTargetOverrides(target string) string {
	if globalNth != 0 && !strings.ContainsAny(target, "[]:#") {
		if globalNth == -1 {
			return target + "[last]"
		}
		return fmt.Sprintf("%s[%d]", target, globalNth)
	}
	return target
}

func runCapture(cmd *cobra.Command, args []string) error {
	if listWindows {
		return PrintWindowsList()
	}

	if len(args) > 0 && (windowQuery == "fullscreen" || windowQuery == "") {
		windowQuery = args[0]
	}

	if delaySeconds > 0 {
		fmt.Printf("⏳ Waiting %d second(s) before capture...\n", delaySeconds)
		for i := delaySeconds; i > 0; i-- {
			fmt.Printf("%d... ", i)
			time.Sleep(1 * time.Second)
		}
		fmt.Println("📸 Capturing!")
	}

	var rawImg image.Image
	var captureDesc string

	wLower := strings.ToLower(strings.TrimSpace(windowQuery))
	if wLower == "fullscreen" || wLower == "screen" || wLower == "all" || wLower == "" {
		fmt.Println("🖥️  Capturing full screen...")
		img, err := capture.CaptureScreen()
		if err != nil {
			return fmt.Errorf("failed to capture screen: %w", err)
		}
		rawImg = img
		captureDesc = "Fullscreen"
	} else {
		fmt.Printf("🔍 Looking for window matching '%s'...\n", windowQuery)
		win, err := capture.FindWindow(windowQuery)
		if err != nil {
			return err
		}
		fmt.Printf("🎯 Found window: '%s' (PID: %d, Size: %dx%d)\n", win.Title, win.ProcessID, win.Bounds.Dx(), win.Bounds.Dy())

		img, err := capture.CaptureWindow(win, true)
		if err != nil {
			return fmt.Errorf("failed to capture window '%s': %w", win.Title, err)
		}
		rawImg = img
		captureDesc = win.Title
	}

	hasAnnotations := len(boxQueries) > 0 || len(arrowQueries) > 0 || len(highlightList) > 0 || len(blurList) > 0 || len(badgeList) > 0 || ocrDump

	var finalImg image.Image = rawImg

	if hasAnnotations {
		fmt.Println("🔍 Running OCR engine to detect text positions...")
		ocrRes, err := ocr.RecognizeImage(rawImg)
		if err != nil {
			fmt.Printf("⚠️ OCR Warning: %v (saving raw screenshot without annotations)\n", err)
		} else {
			if ocrDump {
				fmt.Println("\n--- OCR Text Detection Dump ---")
				for idx, line := range ocrRes.Lines {
					fmt.Printf("[%d] '%s' (Bounds: %v)\n", idx+1, line.Text, line.Bounds.Rect())
				}
				fmt.Println("-------------------------------")
			}

			annotator := draw.NewAnnotator(rawImg)
			annotatedCount := 0
			bounds := rawImg.Bounds()

			baseBoxOpt := draw.BoxOptions{
				Color:       colorName,
				StrokeWidth: strokeWidth,
				Padding:     boxPadding,
				Radius:      boxRadius,
				Width:       boxWidth,
				Height:      boxHeight,
			}

			// 1. Process Boxes
			for _, target := range boxQueries {
				cleanTarget, boxOpt := draw.ParseBoxInlineOptions(target, baseBoxOpt)
				cleanTarget = applyTargetOverrides(cleanTarget)
				matches, err := ocr.FindSpecificText(ocrRes, bounds, cleanTarget, caseSensitive, markAll)
				if err != nil || len(matches) == 0 {
					fmt.Printf("⚠️ Box: Text '%s' not found in screenshot\n", cleanTarget)
					continue
				}
				for _, m := range matches {
					fmt.Printf("📦 Box added around '%s' at %v (Size: %v)\n", m.MatchedText, m.Bounds, boxOpt)
					annotator.DrawBoxWithOptions(m.Bounds, boxOpt)
					annotatedCount++
				}
			}

			// 2. Process Arrows
			for _, target := range arrowQueries {
				target = applyTargetOverrides(target)
				matches, err := ocr.FindSpecificText(ocrRes, bounds, target, caseSensitive, markAll)
				if err != nil || len(matches) == 0 {
					fmt.Printf("⚠️ Arrow: Text '%s' not found in screenshot\n", target)
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
				target = applyTargetOverrides(target)
				matches, err := ocr.FindSpecificText(ocrRes, bounds, target, caseSensitive, markAll)
				if err != nil || len(matches) == 0 {
					fmt.Printf("⚠️ Highlight: Text '%s' not found in screenshot\n", target)
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
				target = applyTargetOverrides(target)
				matches, err := ocr.FindSpecificText(ocrRes, bounds, target, caseSensitive, markAll)
				if err != nil || len(matches) == 0 {
					fmt.Printf("⚠️ Blur: Text '%s' not found in screenshot\n", target)
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

				target = applyTargetOverrides(target)
				matches, err := ocr.FindSpecificText(ocrRes, bounds, target, caseSensitive, markAll)
				if err != nil || len(matches) == 0 {
					fmt.Printf("⚠️ Badge: Text '%s' not found in screenshot\n", target)
					continue
				}
				for _, m := range matches {
					fmt.Printf("🏷️ Badge '%s' added to '%s' at %v\n", badgeText, m.MatchedText, m.Bounds.Min)
					annotator.DrawBadge(m.Bounds.Min, badgeText, colorName, "white")
					annotatedCount++
				}
			}

			finalImg = annotator.Result()
			fmt.Printf("🎨 Successfully rendered %d annotation(s)\n", annotatedCount)
		}
	}

	savedFile, err := storage.SaveImage(finalImg, outputPath)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	fmt.Printf("✅ Screenshot saved successfully:\n   📁 %s\n", savedFile)
	_ = captureDesc
	return nil
}

func PrintWindowsList() error {
	windows, err := capture.ListWindows()
	if err != nil {
		return fmt.Errorf("failed to enumerate windows: %w", err)
	}

	fmt.Println("\n================ OPEN WINDOWS ================")
	fmt.Printf("%-8s | %-12s | %-45s\n", "PID", "SIZE", "TITLE")
	fmt.Println("----------------------------------------------------------------------")
	for _, w := range windows {
		sizeStr := fmt.Sprintf("%dx%d", w.Bounds.Dx(), w.Bounds.Dy())
		title := w.Title
		if len(title) > 42 {
			title = title[:39] + "..."
		}
		fmt.Printf("%-8d | %-12s | %-45s\n", w.ProcessID, sizeStr, title)
	}
	fmt.Println("==============================================")
	fmt.Println("💡 Tip: Use `fastss -w \"<Title>\"` to capture a specific window.\n")
	return nil
}
