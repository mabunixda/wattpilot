package wattpilot

import (
	"os"
	"fmt"
)

func main() {

	w := NewWattpilot(os.Getenv("WATTPILOT_HOST"), os.Getenv("WATTPILOT_PASSWORD"))
	w.Connect()

	wac, _ := w.GetProperty("cci")

	fmt.Println(wac)
}
