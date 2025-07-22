package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	api "github.com/mabunixda/wattpilot"
)

type Executor struct {
	wattpilot *api.Wattpilot
	inputs    map[string]InputFunc
}

type InputFunc func(*api.Wattpilot, []string)

func (e *Executor) completer(d prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest
	for name := range e.inputs {
		suggestions = append(suggestions, prompt.Suggest{Text: name})
	}
	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

func (e *Executor) Execute(in string) {
	in = strings.TrimSpace(in)
	blocks := strings.Split(in, " ")

	for name, f := range e.inputs {
		if name == blocks[0] {
			f(e.wattpilot, blocks[1:])
			return
		}
	}

	fmt.Println("Sorry, I don't understand.")
}

func inStatus(w *api.Wattpilot, data []string) {
	w.StatusInfo()
	fmt.Println("")
}

func inGetValue(w *api.Wattpilot, data []string) {
	if len(data) == 0 {
		return
	}
	v, err := w.GetProperty(data[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(v)
}
func inSetValue(w *api.Wattpilot, data []string) {
	if len(data) <= 1 {
		return
	}
	if err := w.SetProperty(data[0], data[1]); err != nil {
		fmt.Println("error:", err)
	}
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
	// ... (implementation omitted for brevity)
}

func setLevel(w *api.Wattpilot, data []string) {
	if len(data) == 0 {
		return
	}
	if err := w.ParseLogLevel(data[0]); err != nil {
		fmt.Println("error on parsing: ", err)
	}
}

func inUpdateStatus(w *api.Wattpilot, data []string) {
	if err := w.RequestStatusUpdate(); err != nil {
		fmt.Println("error on update: ", err)
	}
}

func inReconnect(w *api.Wattpilot, data []string) {
	w.Disconnect()
	go func() {
		if err := w.Connect(); err != nil {
			log.Printf("Failed to reconnect to wattpilot: %v", err)
		}
	}()
}

func inDisconnect(w *api.Wattpilot, data []string) {
	w.Disconnect()
}

func quit(w *api.Wattpilot, data []string) {
	inDisconnect(w, data)
	os.Exit(0)
}

func main() {
	host := os.Getenv("WATTPILOT_HOST")
	pwd := os.Getenv("WATTPILOT_PASSWORD")
	level := os.Getenv("WATTPILOT_LOG")
	if host == "" || pwd == "" {
		log.Fatal("WATTPILOT_HOST and WATTPILOT_PASSWORD must be set")
	}
	if level == "" {
		level = "info"
	}

	w := api.New(host, pwd)
	if err := w.ParseLogLevel(level); err != nil {
		log.Fatalf("Could not update loglevel to %s: %v", level, err)
	}

	go func() {
		if err := w.Connect(); err != nil {
			log.Printf("Failed to connect to wattpilot: %v", err)
		}
	}()

	e := &Executor{
		wattpilot: w,
		inputs: map[string]InputFunc{
			"status":     inStatus,
			"get":        inGetValue,
			"set":        inSetValue,
			"disconnect": inDisconnect,
			"properties": inProperties,
			"dump":       dumpData,
			"log":        setLevel,
			"update":     inUpdateStatus,
			"reconnect":  inReconnect,
			"quit":       quit,
			"exit":       quit,
		},
	}

	p := prompt.New(
		e.Execute,
		e.completer,
		prompt.OptionTitle("wattpilot-shell"),
		prompt.OptionPrefix(">>> "),
	)

	p.Run()
}
