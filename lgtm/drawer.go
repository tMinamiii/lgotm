package lgtm

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/pkg/errors"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed embed/NotoSansJP-Bold.otf
var NotoSansJP []byte

//go:embed embed/NotoSerifJP-Bold.otf
var NotoSerifJP []byte

//go:embed embed/gopher.png
var GopherPng []byte

type TextDrawer struct {
	MainText  *Text
	SubText   *Text
	TextColor string
	IsSerif   bool
	IsGopher  bool
}

func NewTextDrawer(main, sub *Text, color string, isSerif, isGopher bool) *TextDrawer {
	return &TextDrawer{
		MainText:  main,
		SubText:   sub,
		TextColor: color,
		IsSerif:   isSerif,
		IsGopher:  isGopher,
	}
}

func (t *TextDrawer) Draw(path string) error {
	ext, err := t.extension(path)
	if err != nil {
		return err
	}

	if ext == "gif" {
		return t.drawOnGIF(path)
	}
	return t.drawOnImage(path, ext)
}

func (t *TextDrawer) extension(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	_, format, err := image.DecodeConfig(f)
	if err != nil {
		return "", err
	}

	return format, nil
}

func (t *TextDrawer) newFilename(path, ext string) string {
	filename := filepath.Base(path)
	name := strings.Split(filename, ".")[0]
	suffix := "lgtm"
	if t.IsGopher {
		suffix = "gopher"
	}
	return fmt.Sprintf("%s-%s.%s", name, suffix, ext)
}

func (t *TextDrawer) drawOnGIF(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	orgGif, err := gif.DecodeAll(file)
	if err != nil {
		return err
	}

	newImage := make([]*image.Paletted, 0, len(orgGif.Image))
	for i, v := range orgGif.Image {
		v := v
		var img image.Image
		if t.IsGopher {
			img, err = t.embedGopher(v, i%2 == 0)
			if err != nil {
				return err
			}
		} else {
			img, err = t.drawMessageText(v)
			if err != nil {
				return err
			}
		}

		palettedImage := &image.Paletted{
			Pix:     v.Pix,
			Stride:  v.Stride,
			Rect:    v.Bounds(),
			Palette: v.Palette,
		}
		draw.Draw(palettedImage, palettedImage.Rect, img, img.Bounds().Min, draw.Over)
		newImage = append(newImage, palettedImage)
	}
	orgGif.Image = newImage

	out, err := os.Create(t.newFilename(path, "gif"))
	if err != nil {
		return err
	}
	defer out.Close()

	if err := gif.EncodeAll(out, orgGif); err != nil {
		return err
	}

	return nil
}

func (t *TextDrawer) drawOnImage(path, ext string) error {
	img, err := imaging.Open(path, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	if t.IsGopher {
		img, err = t.embedGopher(img, false)
		if err != nil {
			return err
		}
	} else {
		img, err = t.drawMessageText(img)
		if err != nil {
			return err
		}
	}

	if err := imaging.Save(img, t.newFilename(path, ext)); err != nil {
		return err
	}

	return nil
}

func (t *TextDrawer) embedGopher(src image.Image, shake bool) (image.Image, error) {
	buf := bytes.NewBuffer(GopherPng)
	gopher, err := png.Decode(buf)
	if err != nil {
		return nil, err
	}

	// if gopher image is larger than src image, resize gopher image to half size.
	if src.Bounds().Dx() <= gopher.Bounds().Dx() || src.Bounds().Dy() <= gopher.Bounds().Dy() {
		gopher = imaging.Resize(gopher, gopher.Bounds().Dx()/2, gopher.Bounds().Dy()/2, imaging.NearestNeighbor)
	}

	x := -((src.Bounds().Dx() - gopher.Bounds().Dx()) / 2)
	y := -(src.Bounds().Dy() - gopher.Bounds().Dy()) / 2
	if shake {
		x -= 3
	}

	center := image.Point{x, y}
	newImg := image.NewRGBA(src.Bounds())
	draw.Draw(newImg, newImg.Bounds(), src, image.Point{0, 0}, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), gopher, center, draw.Over)

	return newImg, nil
}

func (t *TextDrawer) drawMessageText(i image.Image) (image.Image, error) {
	img, err := t.drawString(i, t.MainText)
	if err != nil {
		return nil, err
	}

	img, err = t.drawString(img, t.SubText)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (t *TextDrawer) drawString(img image.Image, text *Text) (image.Image, error) {
	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()
	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(img, 0, 0)

	face, err := t.getFontFace(text.FontSize(img, text.Text))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse font %s", err.Error())
	}
	dc.SetFontFace(face)

	c := func() color.Gray16 {
		if t.TextColor == "white" {
			return color.White
		}
		return color.Black
	}()
	dc.SetColor(c)

	maxWidth := func() float64 {
		if imgWidth > 640 {
			return float64(imgWidth) - 60.0
		}
		return float64(imgWidth)
	}()

	pt := text.Point(img)
	dc.DrawStringWrapped(text.Text, pt.X, pt.Y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	return dc.Image(), nil
}

func (t *TextDrawer) getFontFace(size float64) (font.Face, error) {
	opts := &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingNone,
	}

	if t.IsSerif {
		otf, err := opentype.Parse(NotoSerifJP)
		if err != nil {
			return nil, err
		}
		return opentype.NewFace(otf, opts)
	}

	otf, err := opentype.Parse(NotoSansJP)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(otf, opts)
}
