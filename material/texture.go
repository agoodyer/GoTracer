package material

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"os"
	. "raytracer/common"
	"strings"
)

type Texture interface {
	Value(u float64, v float64, p *Point3) Color
}

type Solid_color struct {
	albedo Color
}

func NewSolid_color(albedo *Color) Solid_color {
	return Solid_color{albedo: *albedo}
}

func NewSolid_colorRGB(r float64, g float64, b float64) Solid_color {
	return Solid_color{albedo: NewColor(r, g, b)}
}

func (s *Solid_color) Value(u float64, v float64, p *Point3) Color {
	return s.albedo
}

type Checker_texture struct {
	inv_scale float64
	even      Texture
	odd       Texture
}

func NewChecker_texture(scale float64, c1 *Color, c2 *Color) Checker_texture {

	evenTexture := NewSolid_color(c1)
	oddTexture := NewSolid_color(c2)

	return Checker_texture{inv_scale: 1.0 / scale, even: &evenTexture, odd: &oddTexture}
}

func (t *Checker_texture) Value(u float64, v float64, p *Point3) Color {

	xInteger := int(math.Floor(t.inv_scale * p.X()))
	yInteger := int(math.Floor(t.inv_scale * p.Y()))
	zInteger := int(math.Floor(t.inv_scale * p.Z()))

	isEven := (xInteger+zInteger+yInteger)%2 == 0

	if isEven {
		return t.even.Value(u, v, p)
	} else {
		return t.odd.Value(u, v, p)
	}

}

type Image_texture struct {
	img *image.Image
}

func getImageFromFile(path string) (image.Image, error) {

	// dir, err := os.Open(".")

	// defer dir.Close()

	// filenames, _ := dir.Readdirnames(-1)

	// fmt.Print(filenames)

	f, err := os.Open(path)
	if err != nil {
		fmt.Print(err)
		return nil, err

	}
	defer f.Close()

	if strings.HasSuffix(path, ".jpg") {

		image, err := jpeg.Decode(f)
		return image, err

	} else {
		image, _, err := image.Decode(f)

		fmt.Print(image.Bounds().Dx())
		fmt.Print(image.At(0, 0))
		return image, err
	}

}

func NewImage_texture(path string) Image_texture {
	img, err := getImageFromFile(path)
	// fmt.Print(img.At(0, 0))

	// fmt.Print(img, err)

	if err != nil {
		fmt.Print("WARNING: Invalid texture path specified", err)

		// blank :=  image.NewRGBA(image.Rect(0, 0, 0, 0))
		return Image_texture{img: &img}

	}

	// fmt.Print(img.At(0, 0))

	return Image_texture{img: &img}

}

// NewImage_textureFromBytes creates a texture from embedded byte data (for WASM)
func NewImage_textureFromBytes(data []byte) Image_texture {
	reader := bytes.NewReader(data)
	
	// Try decoding as any supported format
	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Print("WARNING: Failed to decode embedded image: ", err)
		return Image_texture{img: &img}
	}
	
	return Image_texture{img: &img}
}

func (tex *Image_texture) Value(u float64, v float64, p *Point3) Color {

	img := *tex.img

	// fmt.Print(img)

	// fmt.Print(img.At(0, 0))
	if img.Bounds().Dy() <= 0 {
		return NewColor(0, 1, 1)
	}

	intvl := NewInterval(0, 1)

	// fmt.Print(u, v, "\n")
	u = intvl.Clamp(u)
	v = 1.0 - intvl.Clamp(v)

	i := int(float64(img.Bounds().Dx()) * u)
	j := int(float64(img.Bounds().Dy()) * v)

	// fmt.Print(i, "\n")
	pixel := img.At(i, j)
	// pixel := img.At(200, 300)

	// fmt.Print(img.At(i, j))

	r, g, b, _ := pixel.RGBA()

	color_scale := 1.0 / float64(255)

	return NewColor(color_scale*(float64(r)/255), color_scale*(float64(g)/255), color_scale*(float64(b)/255))

}
