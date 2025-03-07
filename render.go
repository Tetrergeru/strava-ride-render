package main

import (
	"image"
	"image/color"
	"image/png"
	"iter"
	"math"
	"os"
	"sort"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type Renderer struct {
	width, height int
	image         *image.RGBA
}

func NewRenderer(w, h int) *Renderer {
	if w%2 != 0 {
		w += 1
	}
	if h%2 != 0 {
		h += 1
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for i := range w {
		for j := range h {
			img.Set(i, j, RGBA(255, 255, 255, 255))
		}
	}

	return &Renderer{
		width:  w,
		height: h,
		image:  img,
	}
}

func NewRendererFromImage(img image.Image) *Renderer {
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))

	for i := range width {
		for j := range height {
			canvas.Set(i, j, img.At(i, j))
		}
	}

	return &Renderer{
		width:  width,
		height: height,
		image:  canvas,
	}
}

func (r *Renderer) RenderPoints(points iter.Seq[[2]int], totalPoints int) {
	lightestColor := RGBA(0, 0, 0, 0)

	for p := range points {
		var c color.RGBA
		if totalPoints > POINTS_TO_DRAW {
			c = lightestColor
		} else {
			c = LerpRGBA(float64(totalPoints)/float64(POINTS_TO_DRAW), RGBA(0, 0, 0, 255), lightestColor)
		}

		// r.RenderPoint(p, c)
		r.RenderSquare(p, c, 3)
		totalPoints--
	}
}

func (r *Renderer) RenderMatrix(matrix [][]int) {
	points := make([][3]int, 0)

	maxRides := 0
	for i, el := range matrix {
		for j := range el {
			maxRides = Max(maxRides, matrix[i][j])

			if matrix[i][j] != 0 {
				points = append(
					points,
					[3]int{i + 50, j + 50, matrix[i][j]},
				)
			}
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i][2] < points[j][2]
	})

	for _, p := range points {
		cl := RidesToColor(float64(p[2]), float64(maxRides))

		r.RenderSquare([2]int{p[0], p[1]}, cl, 2)
		// r.RenderPoint([2]int{p[0], p[1]}, cl)
	}
}

func (r *Renderer) SaveImage(fname string) error {
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, r.image); err != nil {
		return err
	}

	return nil
}

func (r *Renderer) RenderPoint(p [2]int, cl color.RGBA) {
	r.image.Set(p[0], p[1], cl)
}

func (r *Renderer) RenderSquare(p [2]int, cl color.RGBA, radius int) {
	for x := Max(p[0]-radius, 0); x <= Min(p[0]+radius, r.width-1); x++ {
		for y := Max(p[1]-radius, 0); y <= Min(p[1]+radius, r.height-1); y++ {
			dx := float64(x - p[0])
			dy := float64(y - p[1])

			dx *= dx
			dy *= dy

			dist := Clamp((math.Sqrt(dx+dy))/float64(radius), 0, 1)

			newCl := RGBA(cl.R, cl.G, cl.B, 255)
			oldCl := r.Get(x, y)
			r.image.Set(x, y, LerpRGBA((1.0-dist)*(float64(cl.A)/255.0), oldCl, newCl))
		}
	}
}

func (r *Renderer) Get(x, y int) color.RGBA {
	R, G, B, A := r.image.At(x, y).RGBA()
	return RGBA(byte(R), byte(G), byte(B), byte(A))
}

func (r *Renderer) RenderText(x, y int, label string, f *opentype.Font) error {
	col := color.RGBA{0, 0, 0, 255}
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    50,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return err
	}

	d := &font.Drawer{
		Dst:  r.image,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  point,
	}

	d.DrawString(label)

	return nil
}

func RidesToColor(rides, maxRides float64) color.RGBA {
	return RGBA(0, 0, 0, 255)

	if rides < 1.0 {
		return RGBA(255, 255, 255, 255)
	}
	const BORDER1 float64 = 50.0
	if rides <= BORDER1 {
		return LerpRGBA(rides/BORDER1, RGBA(233, 236, 245, 255), RGBA(21, 21, 88, 255))
	}
	const BORDER2 float64 = 150.0
	if rides <= BORDER2 {
		return LerpRGBA((rides-BORDER1)/(BORDER2-BORDER1), RGBA(21, 21, 88, 255), RGBA(239, 159, 12, 255))
	}

	// return RGBA(235, 70, 25, 255)

	v := Min((rides-BORDER2)/50.0, 1.0)
	// v = math.Pow(v, 0.5)
	c := LerpRGBA(v, RGBA(239, 159, 12, 255), RGBA(235, 70, 25, 255))

	return c
}
