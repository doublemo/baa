package avatar

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

type syncFillImage struct {
	idx int
	img image.Image
}

// RectangleImageFill 矩形图像填充
func RectangleImageFill(max, width, height int, bg color.Color, images ...string) (*gg.Context, error) {
	imageCount := len(images)
	w := float64(width * max)
	h := float64(height * max)
	pow := max * max

	if imageCount > pow {
		images = images[0:pow]
		imageCount = pow
	} else if imageCount < pow {
		nextmax := max
		for {
			nextmax--
			if imageCount > nextmax*nextmax {
				nextmax++
				break
			}

			w = float64(width * nextmax)
			h = float64(height * nextmax)
		}
		max = nextmax
	}

	if imageCount/max != 0 {
		h = math.Ceil(float64(imageCount)/float64(max)) * float64(height)
	}

	dc := gg.NewContext(int(w), int(h))
	dc.DrawRectangle(0, 0, w, h)
	dc.SetColor(bg)
	dc.Fill()
	imagesData := make([]image.Image, imageCount)
	ch := make(chan *syncFillImage)
	for i, src := range images {
		go func(id int, path string) {
			if m, err := gg.LoadImage(path); err == nil {
				ch <- &syncFillImage{idx: id, img: m}
				return
			}

			ch <- &syncFillImage{idx: id, img: nil}
		}(i, src)
	}

	for i := 0; i < imageCount; i++ {
		img := <-ch
		if img.img == nil {
			return nil, fmt.Errorf("Unable to load image. %s", images[img.idx])
		}
		imagesData[img.idx] = img.img
	}

	y := int(math.Ceil(float64(imageCount)/float64(max))) - 1
	x := 0
	idx := 0
	mod := imageCount % max
	for {
		if idx >= imageCount {
			break
		}

		im := imagesData[idx]
		idx++
		if y == 0 && mod != 0 {
			dc.DrawImage(resize.Resize(uint(width), uint(height), im, resize.Lanczos2), int(w)/2-mod*width/2+x*width, y*height)
		} else {
			dc.DrawImage(resize.Resize(uint(width), uint(height), im, resize.Lanczos2), x*width, y*height)
		}

		x++
		if x >= max {
			y--
			x = 0
		}
	}

	return dc, nil
}

// RoundedRectangleMask 圆角矩形遮罩
func RoundedRectangleMask(src string, x, y int, w, h, r float64) (*gg.Context, error) {
	m, err := gg.LoadImage(src)
	if err != nil {
		return nil, err
	}

	dc := gg.NewContext(int(w), int(h))
	dc.DrawRoundedRectangle(0, 0, w, h, r)
	dc.Clip()
	dc.DrawImage(m, x, y)
	return dc, nil
}

// CircleMask 圆形遮罩
func CircleMask(src string, x, y int, w, h, r float64) (*gg.Context, error) {
	m, err := gg.LoadImage(src)
	if err != nil {
		return nil, err
	}

	dc := gg.NewContext(int(w), int(h))
	dc.DrawCircle(w/2, h/2, r)
	dc.Clip()
	dc.DrawImage(m, x, y)
	return dc, nil
}
