package ocr

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BoundingBox represents rectangle coordinates
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"w"`
	Height float64 `json:"h"`
}

// WordResult represents an individual recognized word with its bounding box
type WordResult struct {
	Text   string      `json:"text"`
	Bounds BoundingBox `json:"bounds"`
}

// LineResult represents a recognized line composed of multiple words
type LineResult struct {
	Text   string       `json:"text"`
	Bounds BoundingBox  `json:"bounds"`
	Words  []WordResult `json:"words"`
}

// OCRResult holds full recognition data from an image
type OCRResult struct {
	Text  string       `json:"text"`
	Lines []LineResult `json:"lines"`
}

// Rect converts BoundingBox to standard image.Rectangle
func (b BoundingBox) Rect() image.Rectangle {
	return image.Rect(
		int(b.X),
		int(b.Y),
		int(b.X+b.Width),
		int(b.Y+b.Height),
	)
}

// Center returns the center point of the bounding box
func (b BoundingBox) Center() image.Point {
	return image.Pt(
		int(b.X+b.Width/2),
		int(b.Y+b.Height/2),
	)
}

// MatchResult is the location where a search query was matched
type MatchResult struct {
	MatchedText string
	Bounds      image.Rectangle
	Center      image.Point
	Score       float64
}

// RecognizeImage runs Windows Media OCR on an in-memory image
func RecognizeImage(img image.Image) (*OCRResult, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("fastss_ocr_%d.png", os.Getpid()))
	defer os.Remove(tmpFile)

	f, err := os.Create(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary image for OCR: %w", err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	f.Close()

	return RecognizeFile(tmpFile)
}

// RecognizeFile runs Windows Media OCR on an existing image file
func RecognizeFile(imagePath string) (*OCRResult, error) {
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	cleanPath := strings.ReplaceAll(absPath, `'`, `''`)

	script := fmt.Sprintf(`
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$asTaskGeneric = [System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object { 
    $_.Name -eq 'AsTask' -and $_.GetParameters().Count -eq 1 -and $_.GetParameters()[0].ParameterType.Name -eq 'IAsyncOperation`+"`"+`1' 
}[0]

function Await($WinRtTask, $ResultType) {
    $asTask = $asTaskGeneric.MakeGenericMethod($ResultType)
    $netTask = $asTask.Invoke($null, @($WinRtTask))
    $netTask.Wait(-1) | Out-Null
    $netTask.Result
}

[Windows.Storage.StorageFile, Windows.Storage, ContentType = WindowsRuntime] | Out-Null
[Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType = WindowsRuntime] | Out-Null
[Windows.Graphics.Imaging.BitmapDecoder, Windows.Graphics.Imaging, ContentType = WindowsRuntime] | Out-Null

$fullPath = '%s'
$storageFile = Await ([Windows.Storage.StorageFile]::GetFileFromPathAsync($fullPath)) ([Windows.Storage.StorageFile])
$stream = Await ($storageFile.OpenAsync([Windows.Storage.FileAccessMode]::Read)) ([Windows.Storage.Streams.IRandomAccessStream])
$decoder = Await ([Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync($stream)) ([Windows.Graphics.Imaging.BitmapDecoder])
$softwareBitmap = Await ($decoder.GetSoftwareBitmapAsync()) ([Windows.Graphics.Imaging.SoftwareBitmap])

$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
if ($engine -eq $null) {
    $engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromLanguage([Windows.Globalization.Language]::new("en-US"))
}

$ocrResult = Await ($engine.RecognizeAsync($softwareBitmap)) ([Windows.Media.Ocr.OcrResult])

$lines = @()
foreach ($line in $ocrResult.Lines) {
    $words = @()
    $lineX = 999999
    $lineY = 999999
    $lineMaxX = 0
    $lineMaxY = 0

    foreach ($word in $line.Words) {
        $r = $word.BoundingRect
        if ($r.X -lt $lineX) { $lineX = $r.X }
        if ($r.Y -lt $lineY) { $lineY = $r.Y }
        if (($r.X + $r.Width) -gt $lineMaxX) { $lineMaxX = ($r.X + $r.Width) }
        if (($r.Y + $r.Height) -gt $lineMaxY) { $lineMaxY = ($r.Y + $r.Height) }

        $words += @{
            text = $word.Text
            bounds = @{
                x = [double]$r.X
                y = [double]$r.Y
                w = [double]$r.Width
                h = [double]$r.Height
            }
        }
    }

    $lines += @{
        text = $line.Text
        bounds = @{
            x = [double]$lineX
            y = [double]$lineY
            w = [double]($lineMaxX - $lineX)
            h = [double]($lineMaxY - $lineY)
        }
        words = $words
    }
}

$output = @{
    text = $ocrResult.Text
    lines = $lines
}

$output | ConvertTo-Json -Depth 5 -Compress
`, cleanPath)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("OCR engine failed: %w, output: %s", err, string(output))
	}

	outStr := strings.TrimSpace(string(output))
	jsonStart := strings.Index(outStr, "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("OCR produced invalid output: %s", outStr)
	}
	jsonStr := outStr[jsonStart:]

	var res OCRResult
	if err := json.Unmarshal([]byte(jsonStr), &res); err != nil {
		return nil, fmt.Errorf("failed to parse OCR JSON response: %w (raw: %s)", err, jsonStr)
	}

	return &res, nil
}

// Levenshtein distance between two strings
func levenshtein(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
		dp[i][0] = i
	}
	for j := 0; j <= m; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
			}
			dp[i][j] = min3(
				dp[i-1][j]+1,      // deletion
				dp[i][j-1]+1,      // insertion
				dp[i-1][j-1]+cost, // substitution
			)
		}
	}
	return dp[n][m]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// similarity calculates normalized string similarity between 0.0 and 1.0
func similarity(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))
	if s1 == s2 {
		return 1.0
	}
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
	if maxLen == 0 {
		return 1.0
	}
	dist := float64(levenshtein(s1, s2))
	return 1.0 - (dist / maxLen)
}

// FindText searches the OCR results for text matching the query with exact & fuzzy support
func FindText(res *OCRResult, query string, caseSensitive bool) ([]MatchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query string cannot be empty")
	}

	var matches []MatchResult
	queryLower := strings.ToLower(query)

	// Step 1: Check exact / substring matching on lines and word spans
	for _, line := range res.Lines {
		subMatch := findWordSequenceInLine(line, query, caseSensitive, 0.99)
		if subMatch != nil {
			matches = append(matches, *subMatch)
		}
	}

	if len(matches) > 0 {
		return matches, nil
	}

	// Step 2: Line-level substring check
	for _, line := range res.Lines {
		lineText := line.Text
		match := false
		if caseSensitive {
			match = strings.Contains(lineText, query)
		} else {
			match = strings.Contains(strings.ToLower(lineText), queryLower)
		}

		if match {
			rect := line.Bounds.Rect()
			matches = append(matches, MatchResult{
				MatchedText: line.Text,
				Bounds:      rect,
				Center:      line.Bounds.Center(),
				Score:       1.0,
			})
		}
	}

	if len(matches) > 0 {
		return matches, nil
	}

	// Step 3: Fuzzy sequence matching (threshold 0.68)
	var bestFuzzy *MatchResult
	bestScore := 0.0

	for _, line := range res.Lines {
		subMatch := findWordSequenceInLine(line, query, false, 0.65)
		if subMatch != nil {
			if subMatch.Score > bestScore {
				bestScore = subMatch.Score
				bestFuzzy = subMatch
			}
		}
	}

	if bestFuzzy != nil {
		matches = append(matches, *bestFuzzy)
		return matches, nil
	}

	// Step 4: Fuzzy word check
	for _, line := range res.Lines {
		for _, w := range line.Words {
			score := similarity(w.Text, query)
			if score >= 0.65 && score > bestScore {
				bestScore = score
				bestFuzzy = &MatchResult{
					MatchedText: w.Text,
					Bounds:      w.Bounds.Rect(),
					Center:      w.Bounds.Center(),
					Score:       score,
				}
			}
		}
	}

	if bestFuzzy != nil {
		matches = append(matches, *bestFuzzy)
		return matches, nil
	}

	return nil, nil
}

// findWordSequenceInLine tries to find a contiguous sequence of words within a line matching query
func findWordSequenceInLine(line LineResult, query string, caseSensitive bool, minScore float64) *MatchResult {
	words := line.Words
	if len(words) == 0 {
		return nil
	}

	qWords := strings.Fields(query)
	if len(qWords) == 0 {
		return nil
	}

	qWordsStr := strings.Join(qWords, " ")
	if !caseSensitive {
		qWordsStr = strings.ToLower(qWordsStr)
	}

	var bestMatch *MatchResult
	bestScore := minScore

	// Check spans of length from len(qWords) - 1 to len(qWords) + 1
	minSpan := int(math.Max(1, float64(len(qWords)-1)))
	maxSpan := len(qWords) + 1

	for spanLen := minSpan; spanLen <= maxSpan; spanLen++ {
		for i := 0; i <= len(words)-spanLen; i++ {
			var combined []string
			for j := 0; j < spanLen; j++ {
				combined = append(combined, words[i+j].Text)
			}
			candidate := strings.Join(combined, " ")
			check := candidate
			if !caseSensitive {
				check = strings.ToLower(check)
			}

			score := similarity(check, qWordsStr)
			if strings.Contains(check, qWordsStr) {
				score = 1.0
			}

			if score >= bestScore {
				bestScore = score

				minX := words[i].Bounds.X
				minY := words[i].Bounds.Y
				maxX := words[i].Bounds.X + words[i].Bounds.Width
				maxY := words[i].Bounds.Y + words[i].Bounds.Height

				for j := 1; j < spanLen; j++ {
					w := words[i+j]
					if w.Bounds.X < minX {
						minX = w.Bounds.X
					}
					if w.Bounds.Y < minY {
						minY = w.Bounds.Y
					}
					if (w.Bounds.X + w.Bounds.Width) > maxX {
						maxX = w.Bounds.X + w.Bounds.Width
					}
					if (w.Bounds.Y + w.Bounds.Height) > maxY {
						maxY = w.Bounds.Y + w.Bounds.Height
					}
				}

				b := BoundingBox{
					X:      minX,
					Y:      minY,
					Width:  maxX - minX,
					Height: maxY - minY,
				}

				bestMatch = &MatchResult{
					MatchedText: candidate,
					Bounds:      b.Rect(),
					Center:      b.Center(),
					Score:       score,
				}
			}
		}
	}

	return bestMatch
}
