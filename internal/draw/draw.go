package draw

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
)

// Annotator wraps an image and provides high quality 2D drawing capabilities
type Annotator struct {
	dc  *gg.Context
	img image.Image
	w   int
	h   int
}

// BoxOptions configures dimensions, padding, radius, color, and stroke for a box
type BoxOptions struct {
	Color       string
	StrokeWidth float64
	Padding     float64
	Radius      float64
	Width       float64 // If > 0, overrides or sets fixed width centered on target
	Height      float64 // If > 0, overrides or sets fixed height centered on target
}

// NewAnnotator creates a new annotator initialized with the given image
func NewAnnotator(img image.Image) *Annotator {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dc := gg.NewContext(w, h)
	dc.DrawImage(img, 0, 0)

	return &Annotator{
		dc:  dc,
		img: img,
		w:   w,
		h:   h,
	}
}

// ParseColor parses named colors or hex strings (e.g., "red", "#FF0000", "rgba(255,0,0,0.5)")
func ParseColor(name string, defaultAlpha float64) color.RGBA {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		name = "red"
	}

	a := uint8(math.Round(defaultAlpha * 255))

	switch name {
	case "red":
		return color.RGBA{R: 235, G: 35, B: 35, A: a}
	case "green":
		return color.RGBA{R: 35, G: 200, B: 75, A: a}
	case "blue":
		return color.RGBA{R: 30, G: 144, B: 255, A: a}
	case "yellow":
		return color.RGBA{R: 255, G: 215, B: 0, A: a}
	case "orange":
		return color.RGBA{R: 255, G: 140, B: 0, A: a}
	case "magenta", "purple":
		return color.RGBA{R: 218, G: 45, B: 180, A: a}
	case "cyan":
		return color.RGBA{R: 0, G: 220, B: 220, A: a}
	case "white":
		return color.RGBA{R: 255, G: 255, B: 255, A: a}
	case "black":
		return color.RGBA{R: 0, G: 0, B: 0, A: a}
	}

	// Try Hex (#RRGGBB or #RGB)
	if strings.HasPrefix(name, "#") {
		name = name[1:]
	}
	if len(name) == 6 {
		r, _ := strconv.ParseUint(name[0:2], 16, 8)
		g, _ := strconv.ParseUint(name[2:4], 16, 8)
		b, _ := strconv.ParseUint(name[4:6], 16, 8)
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: a}
	} else if len(name) == 3 {
		r, _ := strconv.ParseUint(string(name[0])+string(name[0]), 16, 8)
		g, _ := strconv.ParseUint(string(name[1])+string(name[1]), 16, 8)
		b, _ := strconv.ParseUint(string(name[2])+string(name[2]), 16, 8)
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: a}
	}

	// Default fallback to red
	return color.RGBA{R: 235, G: 35, B: 35, A: a}
}

// DrawBoxWithOptions draws a custom styled bounding box with specific width, height, padding, and radius
func (a *Annotator) DrawBoxWithOptions(rect image.Rectangle, opt BoxOptions) {
	col := ParseColor(opt.Color, 1.0)
	strokeWidth := opt.StrokeWidth
	if strokeWidth <= 0 {
		strokeWidth = 4.0
	}
	padding := opt.Padding
	if padding <= 0 {
		padding = 6.0
	}

	w := float64(rect.Dx()) + (padding * 2)
	h := float64(rect.Dy()) + (padding * 2)

	centerX := float64(rect.Min.X) + float64(rect.Dx())/2
	centerY := float64(rect.Min.Y) + float64(rect.Dy())/2

	// If custom Width/Height are set, use them centered around the target center
	if opt.Width > 0 {
		w = opt.Width
	}
	if opt.Height > 0 {
		h = opt.Height
	}

	x := centerX - w/2
	y := centerY - h/2

	// Keep within image bounds
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+w > float64(a.w) {
		w = float64(a.w) - x
	}
	if y+h > float64(a.h) {
		h = float64(a.h) - y
	}

	a.dc.SetRGBA255(int(col.R), int(col.G), int(col.B), int(col.A))
	a.dc.SetLineWidth(strokeWidth)

	radius := opt.Radius
	if radius < 0 {
		radius = 0
	}

	if radius > 0 {
		a.dc.DrawRoundedRectangle(x, y, w, h, radius)
	} else {
		a.dc.DrawRectangle(x, y, w, h)
	}
	a.dc.Stroke()
}

// DrawBox draws a bounding box around a rectangle
func (a *Annotator) DrawBox(rect image.Rectangle, colName string, strokeWidth float64, padding float64, radius float64) {
	a.DrawBoxWithOptions(rect, BoxOptions{
		Color:       colName,
		StrokeWidth: strokeWidth,
		Padding:     padding,
		Radius:      radius,
	})
}

// DrawArrow draws an arrow pointing towards the target point from a given direction
func (a *Annotator) DrawArrow(target image.Point, fromDirection string, colName string, strokeWidth float64, length float64) {
	col := ParseColor(colName, 1.0)
	if strokeWidth <= 0 {
		strokeWidth = 4.0
	}
	if length <= 0 {
		length = 60.0
	}

	tx := float64(target.X)
	ty := float64(target.Y)

	var sx, sy float64

	switch strings.ToLower(fromDirection) {
	case "top", "above":
		sx, sy = tx, ty-length
	case "bottom", "below":
		sx, sy = tx, ty+length
	case "left":
		sx, sy = tx-length, ty
	case "right":
		sx, sy = tx+length, ty
	case "top-right", "tr":
		sx, sy = tx+length*0.707, ty-length*0.707
	case "bottom-left", "bl":
		sx, sy = tx-length*0.707, ty+length*0.707
	case "bottom-right", "br":
		sx, sy = tx+length*0.707, ty+length*0.707
	case "top-left", "tl", "auto", "":
		fallthrough
	default:
		sx, sy = tx-length*0.707, ty-length*0.707
	}

	// Calculate angle of line
	angle := math.Atan2(ty-sy, tx-sx)
	headLen := strokeWidth * 4.5
	if headLen < 16 {
		headLen = 16
	}

	// End of arrow shaft slightly before the tip so the tip is sharp
	shaftEndX := tx - math.Cos(angle)*(headLen*0.6)
	shaftEndY := ty - math.Sin(angle)*(headLen*0.6)

	// Draw arrow line / shaft
	a.dc.SetRGBA255(int(col.R), int(col.G), int(col.B), int(col.A))
	a.dc.SetLineWidth(strokeWidth)
	a.dc.SetLineCapRound()
	a.dc.DrawLine(sx, sy, shaftEndX, shaftEndY)
	a.dc.Stroke()

	// Draw arrow head (triangle polygon)
	leftAngle := angle + math.Pi*0.82
	rightAngle := angle - math.Pi*0.82

	hx1 := tx + math.Cos(leftAngle)*headLen
	hy1 := ty + math.Sin(leftAngle)*headLen
	hx2 := tx + math.Cos(rightAngle)*headLen
	hy2 := ty + math.Sin(rightAngle)*headLen

	a.dc.NewSubPath()
	a.dc.MoveTo(tx, ty)
	a.dc.LineTo(hx1, hy1)
	a.dc.LineTo(hx2, hy2)
	a.dc.ClosePath()
	a.dc.Fill()
}

// DrawHighlight draws a semi-transparent colored highlight bar over the given rectangle
func (a *Annotator) DrawHighlight(rect image.Rectangle, colName string, padding float64) {
	col := ParseColor(colName, 0.4)
	if padding <= 0 {
		padding = 2.0
	}

	x := float64(rect.Min.X) - padding
	y := float64(rect.Min.Y) - padding
	w := float64(rect.Dx()) + (padding * 2)
	h := float64(rect.Dy()) + (padding * 2)

	a.dc.SetRGBA255(int(col.R), int(col.G), int(col.B), int(col.A))
	a.dc.DrawRoundedRectangle(x, y, w, h, 4)
	a.dc.Fill()
}

// DrawBlur pixelates a rectangular area to obscure sensitive information
func (a *Annotator) DrawBlur(rect image.Rectangle, blockSize int) {
	if blockSize <= 0 {
		blockSize = 10
	}

	currentImg := a.dc.Image()
	rgbaImg, ok := currentImg.(*image.RGBA)
	if !ok {
		return
	}

	minX := int(math.Max(0, float64(rect.Min.X)))
	minY := int(math.Max(0, float64(rect.Min.Y)))
	maxX := int(math.Min(float64(a.w), float64(rect.Max.X)))
	maxY := int(math.Min(float64(a.h), float64(rect.Max.Y)))

	for by := minY; by < maxY; by += blockSize {
		for bx := minX; bx < maxX; bx += blockSize {
			var rSum, gSum, bSum, count uint32
			for y := by; y < by+blockSize && y < maxY; y++ {
				for x := bx; x < bx+blockSize && x < maxX; x++ {
					c := rgbaImg.RGBAAt(x, y)
					rSum += uint32(c.R)
					gSum += uint32(c.G)
					bSum += uint32(c.B)
					count++
				}
			}
			if count == 0 {
				continue
			}

			avgCol := color.RGBA{
				R: uint8(rSum / count),
				G: uint8(gSum / count),
				B: uint8(bSum / count),
				A: 255,
			}

			for y := by; y < by+blockSize && y < maxY; y++ {
				for x := bx; x < bx+blockSize && x < maxX; x++ {
					rgbaImg.SetRGBA(x, y, avgCol)
				}
			}
		}
	}

	a.dc.DrawImage(rgbaImg, 0, 0)
}

// DrawBadge draws a small rounded pill/badge with text (e.g. "1", "NOTE") above or next to a point
func (a *Annotator) DrawBadge(pt image.Point, text string, bgColName string, textColName string) {
	bgCol := ParseColor(bgColName, 1.0)
	textCol := ParseColor(textColName, 1.0)

	padX := 8.0
	padY := 4.0
	fontSize := 14.0

	a.dc.SetRGBA255(int(textCol.R), int(textCol.G), int(textCol.B), int(textCol.A))
	w, h := a.dc.MeasureString(text)
	if w == 0 {
		w = fontSize * float64(len(text)) * 0.6
		h = fontSize
	}

	boxW := w + padX*2
	boxH := h + padY*2
	x := float64(pt.X) - boxW/2
	y := float64(pt.Y) - boxH - 6

	if x < 0 {
		x = 4
	}
	if y < 0 {
		y = 4
	}

	// Draw pill background
	a.dc.SetRGBA255(int(bgCol.R), int(bgCol.G), int(bgCol.B), int(bgCol.A))
	a.dc.DrawRoundedRectangle(x, y, boxW, boxH, boxH/2)
	a.dc.Fill()

	// Draw text
	a.dc.SetRGBA255(int(textCol.R), int(textCol.G), int(textCol.B), int(textCol.A))
	a.dc.DrawString(text, x+padX, y+padY+h*0.8)
}

// Result returns the finalized image
func (a *Annotator) Result() image.Image {
	return a.dc.Image()
}

// ParseBoxInlineOptions parses inline options from target string (e.g. "Submit|200x50" or "Submit|w=200,h=50,pad=10,r=8,color=blue")
func ParseBoxInlineOptions(raw string, baseOpt BoxOptions) (string, BoxOptions) {
	opt := baseOpt
	parts := strings.Split(raw, "|")
	if len(parts) < 2 {
		return raw, opt
	}

	cleanTarget := strings.TrimSpace(parts[0])
	optStr := strings.TrimSpace(parts[1])

	// Check format "200x50"
	if strings.Contains(optStr, "x") && !strings.Contains(optStr, "=") && !strings.Contains(optStr, ";") {
		dims := strings.Split(optStr, "x")
		if len(dims) == 2 {
			if w, err := strconv.ParseFloat(strings.TrimSpace(dims[0]), 64); err == nil {
				opt.Width = w
			}
			if h, err := strconv.ParseFloat(strings.TrimSpace(dims[1]), 64); err == nil {
				opt.Height = h
			}
		}
		return cleanTarget, opt
	}

	// Normalize delimiters: replace semicolons and spaces with commas
	optStr = strings.ReplaceAll(optStr, ";", ",")
	tokens := strings.FieldsFunc(optStr, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	for _, tok := range tokens {
		kv := strings.SplitN(strings.TrimSpace(tok), "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(kv[0]))
		v := strings.TrimSpace(kv[1])

		switch k {
		case "w", "width":
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				opt.Width = val
			}
		case "h", "height":
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				opt.Height = val
			}
		case "p", "pad", "padding":
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				opt.Padding = val
			}
		case "r", "radius":
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				opt.Radius = val
			}
		case "c", "color":
			opt.Color = v
		case "s", "stroke":
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				opt.StrokeWidth = val
			}
		}
	}

	return cleanTarget, opt
}
