package avatar

import (
	"image/color"
	"testing"
)

func TestRectangleImageFill(t *testing.T) {
	dc, err := RectangleImageFill(3, 300, 300, color.White, "11.png", "10.png", "3.png", "4.png", "5.jpg", "6.jpg", "8.jpg", "1.png")
	if err != nil {
		t.Fatal(err)
	}

	if err := dc.SavePNG("out.png"); err != nil {
		t.Fatal(err)
	}
}

func TestRoundedRectangleMask(t *testing.T) {
	dc, err := RoundedRectangleMask("5.jpg", -139, -120, 300, 200, 100)
	if err != nil {
		t.Fatal(err)
	}

	if err := dc.SavePNG("out1.png"); err != nil {
		t.Fatal(err)
	}
}
