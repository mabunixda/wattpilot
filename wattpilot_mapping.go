package api
import (
	"strconv"
)

var propertyMap = map[string]string {
	"carConnected": "car",
	"mode": "lmo",
	"voltage1": "nrg",
	"voltage2": "nrg",
	"voltage3": "nrg",
	"voltageN": "nrg",
	"amps1": "nrg",
	"amps2": "nrg",
	"amps3": "nrg",
	"power1": "nrg",
	"power2": "nrg",
	"power3": "nrg",
	"power": "nrg",
	"allowCharging": "alw",
	"temp": "tma",
}


type PostFunction func(interface{}) (string, error)
var postProcess = map[string]PostFunction {
	"voltage1": voltage1Process,
	"voltage2": voltage2Process,
	"voltage3": voltage3Process,
	"voltageN": voltageNProcess,
	"amps1": amps1Process,
	"amps2": amps2Process,
	"amps3": amps3Process,
	"power1": power1Process,
	"power2": power2Process,
	"power3": power3Process,
	"power": powerProcess,
}

func voltage1Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 0)), nil
}
func voltage2Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 1)), nil
}
func voltage3Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 2)), nil
}
func voltageNProcess(data interface{}) ( string, error ){
	return float2String(voltageData(data, 3)), nil
}

func amps1Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 4)), nil
}
func amps2Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 5)), nil
}
func amps3Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 6)), nil
}

func power1Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 7) * 0.001), nil
}
func power2Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 8) * 0.001),nil
}
func power3Process(data interface{}) ( string, error ){
	return float2String(voltageData(data, 9) * 0.001), nil
}
func powerNProcess(data interface{}) ( string, error ){
	return float2String(voltageData(data, 10) * 0.001), nil
}
func powerProcess(data interface{}) ( string, error ){
	return float2String(voltageData(data, 11) * 0.001), nil
}
func voltageData(data interface{}, idx int) ( float64 ) {
	vars := data.([]interface{})
	v := vars[idx].(float64)
	return v
}
func float2String(value float64) (string) {
		return strconv.FormatFloat(value, 'f', 2, 64)
}
