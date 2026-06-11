package image

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/gif"
	_ "image/png"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

func Decode(data []byte) (image.Image, string, error) {
	return image.Decode(bytes.NewReader(data))
}

func Resize(img image.Image, maxWidth int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxWidth {
		return img
	}

	newWidth := maxWidth
	newHeight := height * maxWidth / width

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}

func ResizeSquare(img image.Image, size int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	side := w
	if h < w {
		side = h
	}
	x0 := (w - side) / 2
	y0 := (h - side) / 2
	cropped := image.NewRGBA(image.Rect(0, 0, side, side))
	draw.Draw(cropped, cropped.Bounds(), img, image.Point{X: x0 + b.Min.X, Y: y0 + b.Min.Y}, draw.Src)
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), cropped, cropped.Bounds(), draw.Over, nil)
	return dst
}

func GenerateAvatar(img image.Image) ([]byte, error) {
	square := ResizeSquare(img, 300)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, square, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GeneratePreview(img image.Image) ([]byte, error) {
	resized := Resize(img, PreviewMaxWidth)
	watermarked := ApplyWatermark(resized, PreviewWatermarkText)

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, watermarked, &jpeg.Options{Quality: PreviewJPEGQuality})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
