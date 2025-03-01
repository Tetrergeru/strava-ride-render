package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"slices"
	"sync"
	"time"

	"golang.org/x/exp/constraints"
)

type Model struct {
	Name        string  `json:"name"`
	Start       string  `json:"start_time"`
	Distance    float64 `json:"distance_raw"`
	Elevation   float64 `json:"elevation_gain_raw"`
	ElapsedTime float64 `json:"elapsed_time_raw"`
	Time        float64 `json:"moving_time_raw"`
	Id          uint64  `json:"id"`
}
type Result struct {
	Models []Model `json:"models"`
}

func ReadResult[T any](fname string) T {
	file, err := os.ReadFile(fname)
	if err != nil {
		log.Fatalf("err = %v\n", err.Error())
	}

	var res T
	err = json.Unmarshal(file, &res)
	if err != nil {
		log.Fatalf("err = %v\n", err.Error())
	}

	return res
}

func TryReadResult[T any](fname string) (T, error) {
	file, err := os.ReadFile(fname)
	if err != nil {
		return *new(T), err
	}

	var res T
	err = json.Unmarshal(file, &res)
	if err != nil {
		return *new(T), err
	}

	return res, nil
}

func ConcatResults() {
	res := Result{
		Models: []Model{},
	}

	for i := 1; i <= 7; i++ {
		r := ReadResult[Result](fmt.Sprintf("result_%v.json", i))

		for _, m := range r.Models {
			res.Models = append(res.Models, m)
		}
	}

	bs, err := json.Marshal(res)
	if err != nil {
		log.Fatalf("err = %v\n", err.Error())
	}

	err = os.WriteFile("results/result.json", bs, 0644)
	if err != nil {
		log.Fatalf("err = %v\n", err.Error())
	}
}

func Stats() {
	r := ReadResult[Result]("results/result.json")
	sumDist := 0.0
	totalTime := 0.0
	totalSpeed := 0.0
	total := 0

	for _, m := range r.Models {
		sumDist += m.Distance
		totalTime += m.Time
		totalSpeed += (m.Distance / 1000.0) / (m.Time / 60.0 / 60.0)
		total++
	}

	fmt.Printf("sumDist = %v\n", int(sumDist/1000.0))
	fmt.Printf("totalTime = %v\n", int(totalTime/60.0/60.0))
	fmt.Printf("averageSpeed = %v\n", (sumDist/1000.0)/(totalTime/60.0/60.0))
	fmt.Printf("averageSpeed = %v\n", totalSpeed/float64(total))
	fmt.Printf("total = %v\n", total)
}

type Map struct {
	GradeSmooth []float64    `json:"grade_smooth"`
	LatLng      [][2]float64 `json:"latlng"`
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

func LookAtMap() {
	m := ReadResult[Map]("map.json")

	total := 0.0
	for i := 1; i < len(m.LatLng); i++ {
		c := m.LatLng[i]
		p := m.LatLng[i-1]
		delta := SphereDist(c, p)
		// break

		if delta != delta {
			fmt.Printf("NaN %v %v\n", c, p)
			continue
		}

		total += delta
	}

	total2 := 0.0
	for i := 1; i < len(m.LatLng); i++ {
		c := m.LatLng[i]
		p := m.LatLng[i-1]
		delta := SphereDist2(c, p)
		// break

		if delta != delta {
			fmt.Printf("NaN %v %v\n", c, p)
			continue
		}

		total2 += delta
	}

	// fmt.Printf("len = %v\n", len(m.LatLng))
	fmt.Printf("total = %v\n", total/1000.0)
	fmt.Printf("total2 = %v\n", total2/1000.0)
}

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

const SCALE float64 = 10000
const FRAMES int = 500
const POINTS_TO_DRAW int = 5000

func RenderPoints(dx, dy int, points [][2]float64, max, min [2]float64, fname string) {
	if dy%2 == 0 {
		dy += 1
	}
	if dx%2 == 0 {
		dx += 1
	}
	img := image.NewRGBA(image.Rect(0, 0, dy+1, dx+1))

	for j := range dy + 1 {
		for i := range dx + 1 {
			img.Set(j, dx+1-i, RGBA(255, 255, 255, 255))
		}
	}

	for _, p := range points {
		x := Max(int((p[0]-min[0])*SCALE), 0)
		y := Max(int((p[1]-min[1])*SCALE), 0)

		img.Set(y, dx+1-x, RGBA(255, 255, 255, 255))
	}
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

func MakeMatrix(dx, dy int) [][]int {
	matrix := make([][]int, dy+1)
	for j := range matrix {
		matrix[j] = make([]int, dx+1)
	}
	return matrix
}

func MapsToImage() {
	max := [2]float64{math.Inf(-1), math.Inf(-1)}
	min := [2]float64{math.Inf(1), math.Inf(1)}

	r := ReadResult[Result]("results/result.json")

	slices.SortFunc(r.Models, func(a, b Model) int {
		layout := "2006-01-02T15:04:05+0000"
		aStart, err := time.Parse(layout, a.Start)
		if err != nil {
			log.Fatal(err)
		}

		bStart, err := time.Parse(layout, b.Start)
		if err != nil {
			log.Fatal(err)
		}

		if aStart.Before(bStart) {
			return -1
		}
		if bStart.Before(aStart) {
			return 1
		}
		return 0
	})

	fmt.Printf("time1 = %v, time2 = %v\n\n", r.Models[0].Start, r.Models[1].Start)

	countRides := 0
	for _, m := range r.Models {
		fname := fmt.Sprintf("./maps/%d.json", m.Id)
		m, err := TryReadResult[Map](fname)
		if err != nil {
			continue
		}

		countRides++

		for _, p := range m.LatLng {
			max[0] = Max(max[0], p[0])
			max[1] = Max(max[1], p[1])
			min[0] = Min(min[0], p[0])
			min[1] = Min(min[1], p[1])
		}
	}

	fmt.Printf("Have %d/%d rides\n", countRides, len(r.Models))

	dx := int((max[0] - min[0]) * SCALE)
	dy := int((max[1] - min[1]) * SCALE)

	fmt.Printf("max = %v, min = %v\n", max, min)
	fmt.Printf("dx = %v, dy = %v\n", dx, dy)

	allPoints := make([][2]float64, 0)
	for _, m := range r.Models {
		fname := fmt.Sprintf("./maps/%d.json", m.Id)
		m, err := TryReadResult[Map](fname)
		if err != nil {
			continue
		}

		for _, p := range m.LatLng {
			allPoints = append(allPoints, p)
		}
	}

	pPerFrame := len(allPoints) / FRAMES
	fmt.Printf("points per frame = %v\n", pPerFrame)

	wg := sync.WaitGroup{}
	wg.Add(10)

	for thread := range 10 {
		go func() {
			for i := thread; i < FRAMES; i += 10 {
				matrix := MakeMatrix(dx, dy)
				//Max(i*pPerFrame-POINTS_TO_DRAW, 0)
				for j, p := range allPoints[0 : i*pPerFrame] {
					x := Max(int((p[0]-min[0])*SCALE), 0)
					y := Max(int((p[1]-min[1])*SCALE), 0)
					matrix[y][x] = int(float64(j) / float64(pPerFrame*100) * 255)
				}
				RenderMatrix(dx, dy, matrix, fmt.Sprintf("frames/%06d.png", i), RidesToColorBW)
				fmt.Printf("Frame %d done\n", i)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// count := 0
	// matrix := MakeMatrix(dx, dy)
	// for i, p := range allPoints {

	// 	x := Max(int((p[0]-min[0])*SCALE), 0)
	// 	y := Max(int((p[1]-min[1])*SCALE), 0)
	// 	matrix[y][x] += 1
	// 	if i%pPerFrame == 0 {
	// 		RenderMatrix(dx, dy, matrix, fmt.Sprintf("frames/%06d.png", count))
	// 		fmt.Printf("Frame %d done\n", count)
	// 		count++
	// 	}
	// }
}

func main() {
	// ConcatResults()
	// Stats()
	// LookAtMap()
	MapsToImage()

	// r := ReadResult[Result]("results/result.json")
	// builder := strings.Builder{}
	// for _, m := range r.Models {
	// 	fname := fmt.Sprintf("./maps/%d.json", m.Id)
	// 	_, err := TryReadResult[Result](fname)
	// 	if err != nil {
	// 		builder.WriteString(fmt.Sprintf("%d\n", m.Id))
	// 		// os.Remove(fname)
	// 	}
	// }

	// err := os.WriteFile("maps.txt", []byte(builder.String()), 0644)
	// if err != nil {
	// 	log.Fatalf("err = %v\n", err)
	// }
}
