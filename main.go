package main

const SCALE float64 = 10000 / 1
const FRAMES int = 500 * 4
const POINTS_TO_DRAW int = 5000

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
