package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"

	api "github.com/mabunixda/wattpilot"
	"gopkg.in/yaml.v2"
)

const (
	fullURLFile = "https://raw.githubusercontent.com/joscha82/wattpilot/refs/heads/main/src/wattpilot/ressources/wattpilot.yaml"
	output      = "wattpilot_mapping_gen.go"
)

func downloadWattpilotYaml() ([]byte, error) {

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fullURLFile)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func main() {
	s, err := downloadWattpilotYaml()
	if err != nil {
		log.Fatalf("failed to download wattpilot yaml: %v", err)
	}
	a := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(s), &a); err != nil {
		log.Fatalf("failed to unmarshal yaml: %v", err)
	}
	propertyMap := make(map[string]string)
	for _, v := range a["properties"].([]interface{}) {
		key := ""
		alias := ""
		data := v.(map[interface{}]interface{})
		for x, y := range data {

			switch x.(string) {
			case "key":
				key = y.(string)
			case "alias":
				alias = y.(string)
			}
		}
		if key != "" && alias != "" {
			propertyMap[alias] = key
		}
	}

	f, err := os.Create(output)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := w.WriteString("package wattpilot\nvar propertyMap = map[string]string {\n"); err != nil {
		log.Fatalf("failed to write to output file: %v", err)
	}
	keys := api.Keys(propertyMap)
	sort.Strings(keys)

	for idx := 0; idx < len(keys); idx += 1 {
		i := keys[idx]
		s := propertyMap[i]
		if _, err := w.WriteString(fmt.Sprintf("\"%s\": \"%s\",\n", i, s)); err != nil {
			log.Fatalf("failed to write to output file: %v", err)
		}
	}
	if _, err := w.WriteString("}\n"); err != nil {
		log.Fatalf("failed to write to output file: %v", err)
	}
	if err := w.Flush(); err != nil {
		log.Fatalf("failed to flush writer: %v", err)
	}
}
