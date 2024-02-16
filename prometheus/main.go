package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mabunixda/wattpilot"
)

const wattpilotPrefix = "wattpilot_%s"

type wattpilotCollector struct {
	charger *wattpilot.Wattpilot
	metrics map[string]*prometheus.Desc
}

func remove[T comparable](l []T, item T) []T {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func newWattpilotCollector(charger *wattpilot.Wattpilot) *wattpilotCollector {
	keys := remove(charger.Properties(), "wsm")
	sort.Strings(keys)
	aliases := charger.Alias()

	constLabels := make(map[string]string)
	constLabels["host"] = charger.GetHost()
	constLabels["serial"] = charger.GetSerial()

	metrics := make(map[string]*prometheus.Desc)
	for _, key := range aliases {
		alias := charger.LookupAlias(key)
		if alias == "" {
			continue
		}
		metrics[alias] = prometheus.NewDesc(
			fmt.Sprintf(wattpilotPrefix, key),
			fmt.Sprintf("Wattpilot property %s ( %s )", alias, key),
			nil, constLabels,
		)

	}
	for _, key := range keys {
		if metrics[key] != nil {
			continue
		}
		metrics[key] = prometheus.NewDesc(
			fmt.Sprintf(wattpilotPrefix, key),
			fmt.Sprintf("Wattpilot property %s", key),
			nil, constLabels,
		)
	}

	for key := range wattpilot.PostProcess {
		if metrics[key] != nil {
			continue
		}
		metrics[key] = prometheus.NewDesc(
			fmt.Sprintf(wattpilotPrefix, key),
			fmt.Sprintf("Wattpilot property %s", key),
			nil, constLabels,
		)
	}
	return &wattpilotCollector{metrics: metrics, charger: charger}
}

func (collector *wattpilotCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range collector.metrics {
		ch <- m
	}
}

// Collect implements required collect function for all promehteus collectors
func (collector *wattpilotCollector) Collect(ch chan<- prometheus.Metric) {

	for key, desc := range collector.metrics {
		data, err := collector.charger.GetProperty(key)
		if err != nil {
			continue
		}
		var value float64
		switch data := data.(type) {
		case int:
		case int64:
			value = float64(data)
			break
		case float64:
			value = data
			break
		case bool:
			value = 1.0
			if !data {
				value = 0.0
			}
			break
		default:
			in_value := fmt.Sprintf("%v", data)
			if out_value, err := strconv.ParseFloat(in_value, 64); err == nil {
				value = out_value
				break
			}
			continue
		}
		m := prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value)
		metric := prometheus.NewMetricWithTimestamp(time.Now(), m)
		ch <- metric
	}
}

func main() {

	host := os.Getenv("WATTPILOT_HOST")
	pwd := os.Getenv("WATTPILOT_PASSWORD")
	level := os.Getenv("WATTPILOT_LOG")
	if host == "" || pwd == "" {
		return
	}
	if level == "" {
		level = "WARN"
	}

	charger := wattpilot.New(host, pwd)
	if err := charger.ParseLogLevel(level); err != nil {
		log.Fatalf("Could not update loglevel to %s: %w", level, err)
	}

	charger.Connect()

	foo := newWattpilotCollector(charger)
	prometheus.MustRegister(foo)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9101", nil))
}
