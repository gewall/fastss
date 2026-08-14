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
	"regexp"
	"strconv"
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

// TargetQuery holds parsed specific selection query
type TargetQuery struct {
	RawText    string
	TargetText string
	AnchorText string // For contextual: "near", "below", "above", "right-of", "left-of"
	Relation   string // "near", "below", "above", "right-of", "left-of"
	Index      int    // 0 = unspecified/all, 1 = first, 2 = second, -1 = last
	Area       string // "top", "bottom", "left", "right", "top-left", "top-right", "bottom-left", "bottom-right", "center"
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

var (
	reIndexBracket = regexp.MustCompile(`^(.*)\[(\d+|first|last|all|\-\d+)\]$`)
	reIndexColon   = regexp.MustCompile(`^(.*):(\d+|first|last|all)$`)
	reIndexHash    = regexp.MustCompile(`^(.*)#(\d+)$`)
	reRelation     = regexp.MustCompile(`(?i)^(.*?)\s+(near|below|under|above|right of|right-of|left of|left-of|after|before)\s+(.*)$`)
	reAreaPrefix   = regexp.MustCompile(`(?i)^(top-left|top-right|bottom-left|bottom-right|top|bottom|left|right|center|header|footer):(.*)$`)
)

// ParseTargetQuery parses complex targeting syntax into a structured TargetQuery
// Examples:
// - "Submit[1]" -> Target: "Submit", Index: 1
// - "Submit:last" -> Target: "Submit", Index: -1
// - "top-right:Save" -> Area: "top-right", Target: "Save"
// - "Edit near Profile" -> Target: "Edit", Anchor: "Profile", Relation: "near"
// - "Delete below 'Account Settings'" -> Target: "Delete", Anchor: "Account Settings", Relation: "below"
func ParseTargetQuery(raw string) TargetQuery {
	tq := TargetQuery{
		RawText:    raw,
		TargetText: strings.TrimSpace(raw),
	}

	// 1. Check Area Prefix (e.g. "top-right:Save")
	if m := reAreaPrefix.FindStringSubmatch(tq.TargetText); len(m) > 2 {
		tq.Area = strings.ToLower(strings.TrimSpace(m[1]))
		tq.TargetText = strings.TrimSpace(m[2])
	}

	// 2. Check Contextual Relation (e.g. "Edit near Profile" or "Save below 'Settings'")
	if m := reRelation.FindStringSubmatch(tq.TargetText); len(m) > 3 {
		tq.TargetText = strings.Trim(strings.TrimSpace(m[1]), `"'`)
		rel := strings.ToLower(strings.TrimSpace(m[2]))
		rel = strings.ReplaceAll(rel, " ", "-")
		if rel == "under" {
			rel = "below"
		}
		if rel == "before" {
			rel = "left-of"
		}
		if rel == "after" {
			rel = "right-of"
		}
		tq.Relation = rel
		tq.AnchorText = strings.Trim(strings.TrimSpace(m[3]), `"'`)
		return tq
	}

	// 3. Check Index Selectors: [1], [last], :1, #1
	if m := reIndexBracket.FindStringSubmatch(tq.TargetText); len(m) > 2 {
		tq.TargetText = strings.TrimSpace(m[1])
		idxStr := strings.ToLower(strings.TrimSpace(m[2]))
		tq.Index = parseIndexValue(idxStr)
		return tq
	}

	if m := reIndexColon.FindStringSubmatch(tq.TargetText); len(m) > 2 {
		tq.TargetText = strings.TrimSpace(m[1])
		idxStr := strings.ToLower(strings.TrimSpace(m[2]))
		tq.Index = parseIndexValue(idxStr)
		return tq
	}

	if m := reIndexHash.FindStringSubmatch(tq.TargetText); len(m) > 2 {
		tq.TargetText = strings.TrimSpace(m[1])
		idxStr := strings.TrimSpace(m[2])
		tq.Index = parseIndexValue(idxStr)
		return tq
	}

	return tq
}

func parseIndexValue(val string) int {
	if val == "first" {
		return 1
	}
	if val == "last" || val == "-1" {
		return -1
	}
	if val == "all" {
		return -2
	}
	if num, err := strconv.Atoi(val); err == nil {
		return num
	}
	return 0
}

// FindSpecificText performs targeted search with index, proximity, and spatial filtering
func FindSpecificText(res *OCRResult, imgBounds image.Rectangle, rawQuery string, caseSensitive bool, defaultAll bool) ([]MatchResult, error) {
	tq := ParseTargetQuery(rawQuery)
	if tq.TargetText == "" {
		return nil, fmt.Errorf("target query cannot be empty")
	}

	// 1. Find all raw matches for target
	matches, err := FindRawTextMatches(res, tq.TargetText, caseSensitive)
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("text '%s' not found", tq.TargetText)
	}

	// 2. Filter by Area if specified (top, bottom, left, right, top-right, etc.)
	if tq.Area != "" {
		matches = filterByArea(matches, imgBounds, tq.Area)
		if len(matches) == 0 {
			return nil, fmt.Errorf("text '%s' not found in area '%s'", tq.TargetText, tq.Area)
		}
	}

	// 3. Filter by Proximity / Contextual Relation (e.g. "Edit near Profile")
	if tq.AnchorText != "" && tq.Relation != "" {
		anchorMatches, err := FindRawTextMatches(res, tq.AnchorText, caseSensitive)
		if err != nil || len(anchorMatches) == 0 {
			return nil, fmt.Errorf("context anchor text '%s' not found for target '%s'", tq.AnchorText, tq.TargetText)
		}

		bestMatch, err := pickBySpatialRelation(matches, anchorMatches, tq.Relation)
		if err != nil {
			return nil, err
		}
		return []MatchResult{*bestMatch}, nil
	}

	// 4. Filter by Index ([1], [2], [last], [all], etc.)
	if tq.Index == -2 {
		return matches, nil // Explicitly requested ALL matches
	}
	if tq.Index != 0 {
		if tq.Index > 0 {
			if tq.Index > len(matches) {
				return nil, fmt.Errorf("match index %d out of range (only %d matches found for '%s')", tq.Index, len(matches), tq.TargetText)
			}
			return []MatchResult{matches[tq.Index-1]}, nil
		}
		if tq.Index == -1 {
			return []MatchResult{matches[len(matches)-1]}, nil
		}
	}

	// 5. Default handling: If not defaultAll and multiple matches exist, return first
	if !defaultAll && len(matches) > 1 {
		return []MatchResult{matches[0]}, nil
	}

	return matches, nil
}

// filterByArea keeps matches located within the designated screen sector
func filterByArea(matches []MatchResult, bounds image.Rectangle, area string) []MatchResult {
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())
	if w <= 0 || h <= 0 {
		return matches
	}

	var filtered []MatchResult
	for _, m := range matches {
		cx := float64(m.Center.X)
		cy := float64(m.Center.Y)

		inArea := false
		switch strings.ToLower(area) {
		case "top", "header":
			inArea = cy < (h * 0.4)
		case "bottom", "footer":
			inArea = cy > (h * 0.6)
		case "left":
			inArea = cx < (w * 0.4)
		case "right":
			inArea = cx > (w * 0.6)
		case "top-left", "tl":
			inArea = (cx < w*0.55) && (cy < h*0.55)
		case "top-right", "tr":
			inArea = (cx > w*0.45) && (cy < h*0.55)
		case "bottom-left", "bl":
			inArea = (cx < w*0.55) && (cy > h*0.45)
		case "bottom-right", "br":
			inArea = (cx > w*0.45) && (cy > h*0.45)
		case "center", "middle":
			inArea = (cx > w*0.25 && cx < w*0.75) && (cy > h*0.25 && cy < h*0.75)
		default:
			inArea = true
		}

		if inArea {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// pickBySpatialRelation finds the target match that best satisfies the relation with an anchor match
func pickBySpatialRelation(targets []MatchResult, anchors []MatchResult, relation string) (*MatchResult, error) {
	var bestTarget *MatchResult
	bestScore := math.MaxFloat64

	for _, t := range targets {
		for _, a := range anchors {
			dx := float64(t.Center.X - a.Center.X)
			dy := float64(t.Center.Y - a.Center.Y)
			dist := math.Sqrt(dx*dx + dy*dy)

			valid := false
			switch relation {
			case "near":
				valid = true
			case "below":
				// Target must be vertically lower than anchor
				if dy > 0 && math.Abs(dx) < 400 {
					valid = true
					dist = dy*1.5 + math.Abs(dx)
				}
			case "above":
				// Target must be vertically higher than anchor
				if dy < 0 && math.Abs(dx) < 400 {
					valid = true
					dist = math.Abs(dy)*1.5 + math.Abs(dx)
				}
			case "right-of":
				// Target must be to the right of anchor
				if dx > 0 && math.Abs(dy) < 100 {
					valid = true
					dist = dx + math.Abs(dy)*2
				}
			case "left-of":
				// Target must be to the left of anchor
				if dx < 0 && math.Abs(dy) < 100 {
					valid = true
					dist = math.Abs(dx) + math.Abs(dy)*2
				}
			}

			if valid && dist < bestScore {
				bestScore = dist
				matchCopy := t
				bestTarget = &matchCopy
			}
		}
	}

	if bestTarget == nil {
		return nil, fmt.Errorf("no target matching relation '%s' with context", relation)
	}

	return bestTarget, nil
}

// FindRawTextMatches searches the OCR results for text matching the query with exact & fuzzy support
func FindRawTextMatches(res *OCRResult, query string, caseSensitive bool) ([]MatchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query string cannot be empty")
	}

	var matches []MatchResult
	queryLower := strings.ToLower(query)

	// Step 1: Check exact / substring matching on lines and word spans
	for _, line := range res.Lines {
		subMatches := findAllWordSequencesInLine(line, query, caseSensitive, 0.95)
		matches = append(matches, subMatches...)
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

	// Step 3: Fuzzy sequence matching (threshold 0.65)
	for _, line := range res.Lines {
		subMatches := findAllWordSequencesInLine(line, query, false, 0.65)
		matches = append(matches, subMatches...)
	}

	if len(matches) > 0 {
		return matches, nil
	}

	// Step 4: Fuzzy individual word check
	for _, line := range res.Lines {
		for _, w := range line.Words {
			score := similarity(w.Text, query)
			if score >= 0.65 {
				matches = append(matches, MatchResult{
					MatchedText: w.Text,
					Bounds:      w.Bounds.Rect(),
					Center:      w.Bounds.Center(),
					Score:       score,
				})
			}
		}
	}

	return matches, nil
}

// FindText is backward compatible caller for simple query search
func FindText(res *OCRResult, query string, caseSensitive bool) ([]MatchResult, error) {
	return FindSpecificText(res, image.Rect(0, 0, 1920, 1080), query, caseSensitive, true)
}

// findAllWordSequencesInLine finds all sequences of words within a line matching query
func findAllWordSequencesInLine(line LineResult, query string, caseSensitive bool, minScore float64) []MatchResult {
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

	var results []MatchResult

	spanLen := len(qWords)
	minSpan := int(math.Max(1, float64(spanLen-1)))
	maxSpan := spanLen + 1

	for sLen := minSpan; sLen <= maxSpan; sLen++ {
		for i := 0; i <= len(words)-sLen; i++ {
			var combined []string
			for j := 0; j < sLen; j++ {
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

			if score >= minScore {
				minX := words[i].Bounds.X
				minY := words[i].Bounds.Y
				maxX := words[i].Bounds.X + words[i].Bounds.Width
				maxY := words[i].Bounds.Y + words[i].Bounds.Height

				for j := 1; j < sLen; j++ {
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

				results = append(results, MatchResult{
					MatchedText: candidate,
					Bounds:      b.Rect(),
					Center:      b.Center(),
					Score:       score,
				})
			}
		}
	}

	return results
}
