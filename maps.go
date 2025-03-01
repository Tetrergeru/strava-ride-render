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

	height := int((max[0] - min[0]) * SCALE)
	width := int((max[1] - min[1]) * SCALE)

	fmt.Printf("max = %v, min = %v\n", max, min)
	fmt.Printf("dx = %v, dy = %v\n", height, width)

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

	wg := sync.WaitGroup{}
	wg.Add(10)

	for thread := range 10 {
		go func() {
			for i := thread; i < FRAMES; i += 10 {
				renderer := NewRenderer(width, height)
				totalPoints := i * pPerFrame

				renderer.RenderPoints(func(yield func([2]int) bool) {
					for _, p := range allPoints[0:totalPoints] { // Max(0, totalPoints-POINTS_TO_DRAW)
						y := Max(int((p[0]-min[0])*SCALE), 0)
						x := Max(int((p[1]-min[1])*SCALE), 0)

						if !yield([2]int{x, height - y}) {
							return
						}
					}
				}, totalPoints) //-Max(0, totalPoints-POINTS_TO_DRAW)

				ride := r.Models[rideIndices[i*pPerFrame]]
				renderer.RenderText(width-30, 10, ride.Name)

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
