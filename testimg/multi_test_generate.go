package main

import (
	"fmt"
	"image/color"

	"github.com/fogleman/gg"
)

func main() {
	dc := gg.NewContext(800, 600)
	dc.SetRGB(0.96, 0.97, 0.98)
	dc.Clear()

	// Section 1: Profile
	dc.SetRGB(1, 1, 1)
	dc.DrawRoundedRectangle(50, 50, 700, 120, 8)
	dc.Fill()
	dc.SetRGB(0.1, 0.1, 0.1)
	dc.DrawString("Profile Settings", 70, 90)
	dc.SetRGBA255(30, 144, 255, 255)
	dc.DrawRoundedRectangle(600, 70, 100, 35, 6)
	dc.Fill()
	dc.SetColor(color.White)
	dc.DrawString("Edit", 635, 93)

	// Section 2: Billing
	dc.SetRGB(1, 1, 1)
	dc.DrawRoundedRectangle(50, 200, 700, 120, 8)
	dc.Fill()
	dc.SetRGB(0.1, 0.1, 0.1)
	dc.DrawString("Billing and Invoices", 70, 240)
	dc.SetRGBA255(30, 144, 255, 255)
	dc.DrawRoundedRectangle(600, 220, 100, 35, 6)
	dc.Fill()
	dc.SetColor(color.White)
	dc.DrawString("Edit", 635, 243)

	// Section 3: Security
	dc.SetRGB(1, 1, 1)
	dc.DrawRoundedRectangle(50, 350, 700, 120, 8)
	dc.Fill()
	dc.SetRGB(0.1, 0.1, 0.1)
	dc.DrawString("Security and Password", 70, 390)
	dc.SetRGBA255(30, 144, 255, 255)
	dc.DrawRoundedRectangle(600, 370, 100, 35, 6)
	dc.Fill()
	dc.SetColor(color.White)
	dc.DrawString("Edit", 635, 393)

	outFile := "test_assets/multi_test.png"
	_ = dc.SavePNG(outFile)
	fmt.Printf("Generated %s\n", outFile)
}
