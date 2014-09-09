package dots

//resize the image to get the dots
//draw the dots

import (
	"image"
	"image/color"
	"math"

	"code.google.com/p/draw2d/draw2d"
	"github.com/nfnt/resize"
)

type dot struct {
	X     int
	Y     int
	Color color.Color
}

func getDotColors(i image.Image, dotsPerRow int) []*dot {
	b := i.Bounds()
	w := uint(dotsPerRow)
	h := uint(math.Ceil(float64(b.Dy()) * float64(dotsPerRow) / float64(b.Dx())))

	m := resize.Resize(w, h, i, resize.Bicubic)
	b = m.Bounds()
	dots := make([]*dot, 0, int(w*h))
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			dots = append(dots, &dot{X: x, Y: y, Color: m.At(x, y)})
		}
	}

	return dots
}

func drawDots(i image.Image, dotsPerRow int) (image.Image, error) {
	dots := getDotColors(i, dotsPerRow)
	b := i.Bounds()
	diameter := float64(b.Dx()) / float64(dotsPerRow)
	radius := diameter * 0.5
	m := image.NewRGBA(b)
	gc := draw2d.NewGraphicContext(m)

	for _, d := range dots {
		gc.SetFillColor(d.Color)
		gc.ArcTo(float64(d.X)*diameter+radius, float64(d.Y)*diameter+radius, radius, radius, 0, 2*math.Pi)
		gc.Fill()
	}

	return m, nil
}
