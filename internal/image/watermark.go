package image

import (
	"image"
	"image/color"
	stddraw "image/draw"
	"math"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/math/fixed"
)

func ApplyWatermark(img image.Image, text string) image.Image {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	stddraw.Draw(rgba, bounds, img, bounds.Min, stddraw.Src)

	face, err := loadFont(float64(bounds.Dx()) / float64(WatermarkFontSizeRatio))
	if err != nil {
		return rgba
	}

	drawer := &font.Drawer{Face: face}

	textWidth := drawer.MeasureString(text).Ceil()
	textHeight := face.Metrics().Height.Ceil()
	descent := face.Metrics().Descent.Ceil()

	tileW := textWidth + WatermarkGapX
	tileH := textHeight + WatermarkGapY

	tile := image.NewRGBA(image.Rect(0, 0, tileW, tileH))
	tileDrawer := &font.Drawer{
		Dst:  tile,
		Src:  image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: uint8(WatermarkAlpha)}),
		Face: face,
	}
	tileDrawer.Dot = fixed.P(
		(tileW-textWidth)/2,
		(tileH+textHeight)/2-descent,
	)
	tileDrawer.DrawString(text)

	angle := float64(WatermarkAngle) * math.Pi / 180
	sin, cos := math.Sincos(angle)

	for y := bounds.Min.Y - bounds.Dy(); y < bounds.Max.Y+bounds.Dy(); y += tileH {
		for x := bounds.Min.X - bounds.Dx(); x < bounds.Max.X+bounds.Dx(); x += tileW {
			aff := f64.Aff3{
				cos, -sin, float64(x),
				sin, cos, float64(y),
			}
			draw.BiLinear.Transform(rgba, aff, tile, tile.Bounds(), draw.Over, nil)
		}
	}

	return rgba
}

func loadFont(size float64) (font.Face, error) {
	ttf, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	return opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}
