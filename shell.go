package main

import (
	"os"
	"fmt"
	"bufio"
	"log"
	"strings"
	"github.com/mabunixda/wattpilot/api"
)

type InputFunc func(*api.Wattpilot,[]string)

var inputs = map[string]InputFunc {
	"connect": inConnect,
	"status": inStatus,
	"get": inGetValue,
	"set": inSetValue,
}

func inStatus(w *api.Wattpilot, data []string) {
	w.StatusInfo()

	fmt.Println("")
}

func inGetValue(w *api.Wattpilot, data []string) {
	v, err :=w.GetProperty(data[0])
	if(err != nil) {
		fmt.Println(err)
		return
	}
	fmt.Println(v)
}
func inSetValue(w *api.Wattpilot, data []string) {
	err := w.SetProperty(data[0], data[1])
	if(err == nil) {
		return
	}
	fmt.Println("error:",err)
}

func inConnect(w *api.Wattpilot, data []string) {
	w.Connect()
}
var interrupt chan os.Signal

func main() {

	w := api.NewWattpilot(os.Getenv("WATTPILOT_HOST"), os.Getenv("WATTPILOT_PASSWORD"))
	w.Connect()

	w.StatusInfo()

	for {
		select {
		// case <-time.After(time.Duration(1) * time.Millisecond * 1000):
		//     // Send an echo packet every second
		//     // err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
		//     if err != nil {
		//         log.Println("Error during writing to websocket:", err)
		//         return
		//     }

		case <-interrupt:

			return

		default:
			fmt.Print("wattpilot> ")
			reader := bufio.NewReader(os.Stdin)
			str, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			words := strings.Fields(str)
			if len(words) < 1 {
				continue
			}

			data := words[1:]
			cmd := words[:1]
			if _, fOk := inputs[cmd[0]]; !fOk {
				fmt.Println("Could not find command",cmd[0])
				continue
			}
			inputs[cmd[0]](w,data)
			fmt.Println("")
		}
	}
}