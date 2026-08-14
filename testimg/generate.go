package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/fogleman/gg"
)

func main() {
	const w = 800
	const h = 500

	dc := gg.NewContext(w, h)
	// Background
	dc.SetRGB(0.95, 0.96, 0.98)
	dc.Clear()

	// Window bar
	dc.SetRGB(0.2, 0.25, 0.35)
	dc.DrawRectangle(0, 0, w, 40)
	dc.Fill()

	// Window title
	dc.SetRGB(1, 1, 1)
	dc.DrawString("Settings Dashboard - Application Settings", 20, 25)

	// Card 1: User Profile
	dc.SetRGB(1, 1, 1)
	dc.DrawRoundedRectangle(40, 70, 720, 160, 8)
	dc.Fill()
	dc.SetRGB(0.85, 0.85, 0.88)
	dc.SetLineWidth(1)
	dc.DrawRoundedRectangle(40, 70, 720, 160, 8)
	dc.Stroke()

	dc.SetRGB(0.1, 0.1, 0.1)
	dc.DrawString("User Profile Section", 60, 105)
	dc.SetRGB(0.4, 0.4, 0.4)
	dc.DrawString("Manage your personal account settings and secret credentials.", 60, 135)
	dc.DrawString("Secret API Token: my-secret-token-9999", 60, 175)

	// Card 2: Actions
	dc.SetRGB(1, 1, 1)
	dc.DrawRoundedRectangle(40, 260, 720, 180, 8)
	dc.Fill()
	dc.SetRGB(0.85, 0.85, 0.88)
	dc.DrawRoundedRectangle(40, 260, 720, 180, 8)
	dc.Stroke()

	dc.SetRGB(0.1, 0.1, 0.1)
	dc.DrawString("Actions & Operations", 60, 295)

	// Button: Save Changes
	dc.SetRGBA255(35, 120, 240, 255)
	dc.DrawRoundedRectangle(60, 330, 160, 45, 6)
	dc.Fill()
	dc.SetColor(color.White)
	dc.DrawString("Save Changes", 90, 358)

	// Button: Cancel
	dc.SetRGBA255(220, 225, 230, 255)
	dc.DrawRoundedRectangle(240, 330, 120, 45, 6)
	dc.Fill()
	dc.SetRGB(0.2, 0.2, 0.2)
	dc.DrawString("Cancel", 275, 358)

	// Text: Quick Help
	dc.SetRGB(0.3, 0.6, 0.3)
	dc.DrawString("Status: Online and Ready", 60, 415)

	if err := os.MkdirAll("test_assets", 0755); err != nil {
		panic(err)
	}

	outPath := "test_assets/sample_window.png"
	if err := dc.SavePNG(outPath); err != nil {
		panic(err)
	}

	fmt.Printf("Generated test window image: %s\n", outPath)
}
