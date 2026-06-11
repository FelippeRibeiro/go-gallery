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
