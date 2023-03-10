package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	api "github.com/mabunixda/wattpilot"
)

type InputFunc func(*api.Wattpilot, []string)

var inputs = map[string]InputFunc{
	"connect":    inConnect,
	"status":     inStatus,
	"get":        inGetValue,
	"set":        inSetValue,
	"properties": inProperties,
	"dataDump":   dumpData,
}

func inStatus(w *api.Wattpilot, data []string) {
	w.StatusInfo()

	fmt.Println("")
}

func inGetValue(w *api.Wattpilot, data []string) {
	v, err := w.GetProperty(data[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(v)
}
func inSetValue(w *api.Wattpilot, data []string) {
	err := w.SetProperty(data[0], data[1])
	if err == nil {
		return
	}
	fmt.Println("error:", err)
}

func inProperties(w *api.Wattpilot, data []string) {
	keys := w.Alias()
	for idx := 0; idx < len(keys); idx += 1 {
		alias := keys[idx]
		raw := w.LookupAlias(alias)
		value, _ := w.GetProperty(alias)
		fmt.Printf("- %s: %s\n  %v\n", alias, raw, value)
	}
}

func dumpData(w *api.Wattpilot, data []string) {
	csvFile, e := os.Create("./wattpilot-data.csv")
	if e != nil {
		fmt.Println(e)
	}
	keys := w.Properties()
	writer := csv.NewWriter(csvFile)
	writer.Write(keys)
	dataSet := []string{}
	for idx := 0; idx < len(keys); idx += 1 {
		alias := keys[idx]
		value, _ := w.GetProperty(alias)
		dataSet = append(dataSet, fmt.Sprint(value))
	}
	writer.Write(dataSet)
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
	log.Println("export written to `wattpilot-data.csv`")
}

func inConnect(w *api.Wattpilot, data []string) {
	connected, err := w.Connect()
	if err != nil {
		log.Println("Could not connect", err)
	}
	if !connected || !w.IsInitialized() {
		return
	}
	log.Printf("Connected to WattPilot %s, Serial %s", w.GetName(), w.GetSerial())
}

var interrupt chan os.Signal

func main() {
	host := os.Getenv("WATTPILOT_HOST")
	pwd := os.Getenv("WATTPILOT_PASSWORD")
	if host == "" || pwd == "" {
		return
	}
	w := api.New(host, pwd)
	inConnect(w, nil)

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
				fmt.Println("Could not find command", cmd[0])
				continue
			}
			inputs[cmd[0]](w, data)
			fmt.Println("")
		}
	}
}
