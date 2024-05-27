package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/joho/godotenv"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
)

func MergeImageNew(base image.Image, mask image.Image, paddingX int, paddingY int, newMaskWidth uint, newMaskHeight uint) (*image.RGBA, error) {
	baseSrcBounds := base.Bounds().Max

	newWidth := baseSrcBounds.X
	newHeight := baseSrcBounds.Y

	// Resize mask
	resizedMask := resize.Resize(newMaskWidth, newMaskHeight, mask, resize.Lanczos3)

	// Get resized mask bounds
	resizedMaskBounds := resizedMask.Bounds().Max
	maskWidth := resizedMaskBounds.X
	maskHeight := resizedMaskBounds.Y

	des := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight)) // 底板
	// First draw the basemap to a new image
	draw.Draw(des, des.Bounds(), base, base.Bounds().Min, draw.Over)
	// Draw scaled overlay to new image
	draw.Draw(des, image.Rect(paddingX, newHeight-paddingY-maskHeight, (paddingX+maskWidth), (newHeight-paddingY)), resizedMask, image.Point{}, draw.Over)

	return des, nil
}
func GetImageFromFile(filePath string) (img image.Image, err error) {
	f1Src, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}
	defer f1Src.Close()

	buff := make([]byte, 512) // why 512 bytes ? see http://golang.org/pkg/net/http/#DetectContentType
	_, err = f1Src.Read(buff)

	if err != nil {
		return nil, err
	}

	filetype := http.DetectContentType(buff)

	fmt.Println(filetype)

	fSrc, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fSrc.Close()

	switch filetype {
	case "image/jpeg", "image/jpg":
		img, err = jpeg.Decode(fSrc)
		if err != nil {
			fmt.Println("jpeg error")
			return nil, err
		}

	case "image/gif":
		img, err = gif.Decode(fSrc)
		if err != nil {
			return nil, err
		}

	case "image/png":
		img, err = png.Decode(fSrc)
		if err != nil {
			return nil, err
		}
	default:
		return nil, err
	}
	return img, nil
}
func GetImageFromNet(url string) (image.Image, error) {
	res, err := http.Get(url)
	if err != nil || res.StatusCode != 200 {
		return nil, err
	}
	defer res.Body.Close()
	m, _, err := image.Decode(res.Body)
	return m, err
}
func SaveImage(targetPath string, m image.Image) error {
	fSave, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer fSave.Close()

	err = jpeg.Encode(fSave, m, nil)

	if err != nil {
		return err
	}

	return nil
}

// Font related
type TextBrush struct {
	FontType  *truetype.Font
	FontSize  float64
	FontColor *image.Uniform
	TextWidth int
}

func NewTextBrush(FontFilePath string, FontSize float64, FontColor *image.Uniform, textWidth int) (*TextBrush, error) {
	fontFile, err := os.ReadFile(FontFilePath)
	if err != nil {
		return nil, err
	}
	fontType, err := truetype.Parse(fontFile)
	if err != nil {
		return nil, err
	}
	if textWidth <= 0 {
		textWidth = 20
	}
	return &TextBrush{FontType: fontType, FontSize: FontSize, FontColor: FontColor, TextWidth: textWidth}, nil
}

// Insert text into pictures
func (fb *TextBrush) DrawFontOnRGBA(rgba *image.RGBA, pt image.Point, content string) {

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(fb.FontType)
	c.SetHinting(font.HintingFull)
	c.SetFontSize(fb.FontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fb.FontColor)

	c.DrawString(content, freetype.Pt(pt.X, pt.Y))

}

func Image2RGBA(img image.Image) *image.RGBA {

	baseSrcBounds := img.Bounds().Max

	newWidth := baseSrcBounds.X
	newHeight := baseSrcBounds.Y

	des := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight)) // 底板
	//First save an image information into jpg
	draw.Draw(des, des.Bounds(), img, img.Bounds().Min, draw.Over)

	return des
}

var (
	BaseURL     string
	MaskURL     string
	TextContent []string
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	BaseURL = os.Getenv("BaseURL")
	MaskURL = os.Getenv("MaskURL")
	TextContent = strings.Split(os.Getenv("TextContent"), ",")
}
func main() {

	base, err := GetImageFromFile(BaseURL)
	if err != nil {
		log.Fatal(err)
	}
	mask, err := GetImageFromFile(MaskURL)
	if err != nil {
		log.Fatal(err)
	}

	des, err := MergeImageNew(base, mask, 150, 320, 80, 80)
	if err != nil {
		log.Fatal(err)
	}

	textBrush, err := NewTextBrush("./assets/font.ttf", 50, image.Black, 20)
	if err != nil {
		log.Fatal(err)
	}
	textBrush.FontColor = image.NewUniform(color.RGBA{
		R: 0x8E,
		G: 0xE5,
		B: 0xEE,
		A: 255,
	})
	for i, v := range TextContent {
		textBrush.DrawFontOnRGBA(des, image.Pt(450, 100+80*i), v)
	}
	if err := SaveImage("./output.png", des); err != nil {
		log.Fatal(err)
	}
}
