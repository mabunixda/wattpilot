package main

import (
	"os"
	"fmt"
)

func main() {

	w := newWattpilot(os.Getenv("WATTPILOT_HOST"), os.Getenv("WATTPILOT_PASSWORD"))
	w.Connect()

	wac, _ := w.GetProperty("cci")

	fmt.Println(wac)
}
