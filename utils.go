package main

import (
	"image"
	"image/color"
	"math"
	"os"

	"golang.org/x/exp/constraints"
	"golang.org/x/image/font/opentype"
)

func Max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	} else {
		return y
	}
}
func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	} else {
		return y
	}
}

func Clamp[T constraints.Ordered](x, f, t T) T {
	return Max(Min(x, t), f)
}

func LerpRGBA(t float64, a, b color.RGBA) color.RGBA {
	R := byte(float64(a.R)*(1-t) + float64(b.R)*t)
	G := byte(float64(a.G)*(1-t) + float64(b.G)*t)
	B := byte(float64(a.B)*(1-t) + float64(b.B)*t)
	A := byte(float64(a.A)*(1-t) + float64(b.A)*t)
	return RGBA(R, G, B, A)
}

func Lerp(t, a, b float64) float64 {
	return float64(a)*(1-t) + float64(b)*t
}

func RGBA(r, g, b, a byte) color.RGBA {
	return color.RGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}

func RGB(r, g, b byte) color.RGBA {
	return RGBA(r, g, b, 255)
}

func SphereDist(a [2]float64, b [2]float64) float64 {
	a = [2]float64{a[1] / 180.0 * math.Pi, a[0] / 180.0 * math.Pi}
	b = [2]float64{b[1] / 180.0 * math.Pi, b[0] / 180.0 * math.Pi}

	ca0 := math.Cos(a[0])
	cb0 := math.Cos(b[0])
	sa0 := math.Sin(a[0])
	sb0 := math.Sin(b[0])

	ca1 := math.Cos(a[1])
	cb1 := math.Cos(b[1])
	sa1 := math.Sin(a[1])
	sb1 := math.Sin(b[1])

	c2 := ca1 * cb1
	d := c2*(sa0*sb0+ca0*cb0) + sa1*sb1
	return math.Acos(d) * 6356752 //6378137 //
}

func SphereDist2(a [2]float64, b [2]float64) float64 {
	a = [2]float64{a[0] / 180.0 * math.Pi, a[1] / 180.0 * math.Pi}
	b = [2]float64{b[0] / 180.0 * math.Pi, b[1] / 180.0 * math.Pi}

	d0 := b[0] - a[0]
	d1 := b[1] - a[1]

	h0 := (1 - math.Cos(d0)) / 2
	h1 := (1 - math.Cos(d1)) / 2

	h := h0 + math.Cos(b[0])*math.Cos(a[0])*h1
	d := 2 * math.Asin(math.Sqrt(h))
	return d * 6356752 //6378137 //
}

func OpenImageFile(fname string) (image.Image, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func LoadFont(fname string) (*opentype.Font, error) {
	fontBytes, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	font, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	return font, nil
}
