package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

		res.Models = append(res.Models, r.Models...)
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
