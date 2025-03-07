package main

import (
	"fmt"
	"log"
	"math"
	"slices"
	"sync"
	"time"
)

type Map struct {
	GradeSmooth []float64    `json:"grade_smooth"`
	LatLng      [][2]float64 `json:"latlng"`
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

	font, err := LoadFont("Montserrat-Regular.otf")
	if err != nil {
		log.Fatalf("Failed to load font: %v\n", err)
	}

	r := ReadSortedResults()
	maps := ReadMaps(&r)

	for _, m := range maps {
		for _, p := range m.LatLng {
			max[0] = Max(max[0], p[0])
			max[1] = Max(max[1], p[1])
			min[0] = Min(min[0], p[0])
			min[1] = Min(min[1], p[1])
		}
	}

	fmt.Printf("Have %d/%d rides\n", len(maps), len(r.Models))

	width, height := project(max, min, math.MaxInt, math.MaxInt, true)
	// height := int((max[0] - min[0]) * SCALE)
	// width := int((max[1] - min[1]) * SCALE * 0.5)

	fmt.Printf("max = %v, min = %v\n", max, min)
	fmt.Printf("height = %v, width = %v\n", height, width)

	allPoints := make([][2]float64, 0)
	rideIndices := make([]int, 0)
	for idx, m := range r.Models {
		mp, ok := maps[m.Id]
		if !ok {
			continue
		}

		allPoints = append(allPoints, mp.LatLng...)
		for _ = range mp.LatLng {
			rideIndices = append(rideIndices, idx)
		}
	}

	pPerFrame := len(allPoints) / FRAMES
	fmt.Printf("points per frame = %v\n", pPerFrame)

	background, err := OpenImageFile("maps_layers.png")
	if err != nil {
		log.Fatalf("Failed to open backgound image: %v", err)
	}
	fmt.Printf("background.height = %v, background.width = %v\n", background.Bounds().Max.Y, background.Bounds().Max.X)

	renderMode := "path"
	if renderMode == "path" {
		wg := sync.WaitGroup{}
		wg.Add(10)
		for thread := range 10 {
			go func() {
				sumDist := 0.0
				countedUntil := 1
				for i := thread; i < FRAMES; i += 10 {
					// renderer := NewRenderer(width, height)
					renderer := NewRendererFromImage(background)

					totalPoints := i * pPerFrame

					for j := countedUntil; j < totalPoints; j++ {
						dist := SphereDist2(allPoints[j-1], allPoints[j])
						if dist < 10.0 {
							sumDist += dist
						}
					}
					countedUntil = totalPoints + 1

					renderer.RenderPoints(func(yield func([2]int) bool) {
						for _, p := range allPoints[Max(0, totalPoints-POINTS_TO_DRAW):totalPoints] { // Max(0, totalPoints-POINTS_TO_DRAW)
							// y := Max(int((p[0]-min[0])*SCALE), 0)
							// x := Max(int((p[1]-min[1])*SCALE*0.5), 0)

							x, y := project(p, min, width, height, false)

							if !yield([2]int{50 + x, 50 + height - y}) {
								return
							}
						}
					}, totalPoints-Max(0, totalPoints-POINTS_TO_DRAW)) //-Max(0, totalPoints-POINTS_TO_DRAW)

					ride := r.Models[rideIndices[i*pPerFrame]]
					renderer.RenderText(width-300, 100, fmt.Sprintf("Ride: %s", ride.Name), font)
					renderer.RenderText(width-300, 200, fmt.Sprintf("Dist: %skm", FormatMetersDist(sumDist, 0)), font)

					err := renderer.SaveImage(fmt.Sprintf("frames/%06d.png", i))

					if err != nil {
						log.Fatalf("Failed to save image: %s", err.Error())
					}

					fmt.Printf("Frame %d done\n", i)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	} else if renderMode == "matrix" {
		matrix := MakeMatrix(height, width)

		maxDist := -1.0
		maxDist2 := -1.0
		sumDist := 0.0
		sumDist2 := 0.0
		for i := range allPoints[1:] {
			p := allPoints[i+1]
			prevP := allPoints[i]

			dist := SphereDist(p, prevP)
			maxDist = Max(maxDist, dist)
			sumDist += dist

			dist2 := SphereDist2(p, prevP)
			maxDist2 = Max(maxDist2, dist2)
			// if dist2 < 10.0 {
			sumDist2 += dist2
			// }

			// if math.Abs(dist2-dist) > 0.1 {
			// 	fmt.Printf("|%v-%v| = %v > 0.1\n", dist2, dist, math.Abs(dist2-dist))
			// }

			x, y := project(p, min, width, height, false)

			matrix[x][height-y] += 1
		}

		fmt.Printf("maxDist = %v\n", FormatMetersDist(maxDist, 3))
		fmt.Printf("maxDist2 = %v\n", FormatMetersDist(maxDist2, 3))
		fmt.Printf("sumDist = %v\n", FormatMetersDist(sumDist, 3))
		fmt.Printf("sumDist2 = %v\n", FormatMetersDist(sumDist2, 3))

		// renderer := NewRenderer(width, height)
		renderer := NewRendererFromImage(background)
		renderer.RenderMatrix(matrix)
		renderer.SaveImage("matrix_render.png")
	}
}

func project(p, min [2]float64, w, h int, debug bool) (int, int) {
	n_x := Clamp(int((p[1]-min[1])*SCALE), 0, w)

	y_rad := math.Pi * p[0] / 180.0
	y_proj := math.Atanh(math.Sin(y_rad))

	y_min_rad := math.Pi * min[0] / 180.0
	y_min_proj := math.Atanh(math.Sin(y_min_rad))

	n_y := Clamp(int((y_proj-y_min_proj)*SCALE*180/math.Pi), 0, h)

	if debug {
		fmt.Printf("y = %v\n", p[0])
		fmt.Printf("y_rad = %v\n", y_rad)
		fmt.Printf("y_proj = %v\n", y_proj)

		fmt.Printf("min_y = %v\n", min[0])
		fmt.Printf("y_min_rad = %v\n", y_min_rad)
		fmt.Printf("y_min_proj = %v\n", y_min_proj)

		fmt.Printf("n_y = %v\n", n_y)
	}

	return n_x, n_y
}

func ReadSortedResults() Result {
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

	return r
}

func ReadMaps(r *Result) map[uint64]Map {
	maps := make(map[uint64]Map, len(r.Models))

	for _, m := range r.Models {
		fname := fmt.Sprintf("./maps/%d.json", m.Id)

		mp, err := TryReadResult[Map](fname)
		if err != nil {
			continue
		}

		maps[m.Id] = mp
	}

	return maps
}
