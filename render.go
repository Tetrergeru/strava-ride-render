package main

import (
	"image"
	"image/color"
	"image/png"
	"iter"
	"log"
	"math"
	"os"
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

func (r *Renderer) RenderPoints(points iter.Seq[[2]int], totalPoints int) {
	lightestColor := RGB(230, 230, 230)

	for p := range points {
		var c color.RGBA
		if totalPoints > POINTS_TO_DRAW {
			c = lightestColor
		} else {
			c = LerpRGBA(float64(totalPoints)/float64(POINTS_TO_DRAW), lightestColor, RGB(0, 0, 0))
		}

		r.RenderPoint(p, c)
		totalPoints--
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

// ========= LEGACY =========

func RidesToColorBW(rides, maxRides, dist float64) color.RGBA {
	c := byte((255 - rides/maxRides*255.0) * dist)
	return RGBA(c, c, c, 255)
}

func RidesToColor(rides, maxRides, dist float64) color.RGBA {
	if rides < 1.0 {
		return RGBA(255, 255, 255, 255)
	}
	const BORDER1 float64 = 5.0
	if rides <= BORDER1 {
		return LerpRGBA(rides/BORDER1, RGBA(177, 185, 220, 255), RGBA(21, 21, 88, 255))
	}
	const BORDER2 float64 = 150.0
	if rides <= BORDER2 {
		return LerpRGBA((rides-BORDER1)/(BORDER2-BORDER1), RGBA(133, 51, 122, 255), RGBA(239, 159, 12, 255))
	}
	return RGBA(235, 70, 25, 255)
	// 	v := Min(rides/BORDER, 1.0)
	// 	v = math.Pow(v, 0.1)
	// 	c := byte((1 - v) * 255)
	// 	return color.RGBA{
	// 		R: c,
	// 		G: 255,
	// 		B: 255,
	// 		A: 255,
	// 	}
}

func RenderMatrix(dx, dy int, matrix [][]int, fname string, cls func(rides, maxRides, dist float64) color.RGBA) {
	if dy%2 == 0 {
		dy += 1
	}
	if dx%2 == 0 {
		dx += 1
	}
	img := image.NewRGBA(image.Rect(0, 0, dy+1, dx+1))

	maxCount := 0
	for j := range matrix {
		for i := range matrix[j] {
			maxCount = Max(maxCount, matrix[j][i])
		}
	}

	const radius = 2

	for j := range matrix {
		for i := range matrix[j] {
			if matrix[j][i] == 0 {
				img.Set(j, dx+1-i, RGBA(255, 255, 255, 255))
				continue
			}

			for y := Max(j-radius, 0); y < Min(j+radius, dy); y++ {
				for x := Max(i-radius, 0); x < Min(i+radius, dx); x++ {
					dist := Lerp((math.Abs(float64(x-i))+math.Abs(float64(y-j)))/float64(radius), 0.5, 1.0)

					c := cls(float64(matrix[j][i]), float64(maxCount), dist)

					r, _, _, _ := img.At(y, dx+1-x).RGBA()

					img.Set(y, dx+1-x, c)

					if r == 255 {
						img.Set(y, dx+1-x, c)
					} else if r > uint32(c.R) {
						img.Set(y, dx+1-x, c)
					}
				}
			}

		}
	}

	file, err := os.Create(fname)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Fatalf("Error encoding image: %v", err)
	}
}
