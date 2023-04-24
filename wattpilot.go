package wattpilot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
	"math/rand"
	"nhooyr.io/websocket"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

const (
	CONTEXT_TIMEOUT   = 30 // seconds
	RECONNECT_TIMEOUT = 5  // seconds
)

//go:generate go run gen/generate.go

var randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

type eventFunc func(map[string]interface{})

type Pubsub struct {
	mu     sync.RWMutex
	subs   map[string][]chan interface{}
	closed bool
}

func NewPubsub() *Pubsub {
	ps := &Pubsub{}
	ps.subs = make(map[string][]chan interface{})
	return ps
}
func (ps *Pubsub) Subscribe(topic string) <-chan interface{} {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan interface{}, 1)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

func (ps *Pubsub) Publish(topic string, msg interface{}) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return
	}
	for _, ch := range ps.subs[topic] {
		go func(ch chan interface{}) {
			ch <- msg
		}(ch)
	}
}
func (ps *Pubsub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subs := range ps.subs {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}

type Wattpilot struct {
	_currentConnection *websocket.Conn
	_requestId         int
	_name              string
	_hostname          string
	_serial            string
	_version           string
	_manufacturer      string
	_devicetype        string
	_protocol          float64
	_secured           bool
	_readContext       context.Context
	_readCancel        context.CancelFunc
	_readMutex         sync.Mutex

	_token3         string
	_hashedpassword string
	_host           string
	_password       string
	_isInitialized  bool
	_isConnected    bool
	_status         map[string]interface{}
	eventHandler    map[string]eventFunc

	connected    chan bool
	initialized  chan bool
	sendResponse chan string
	interrupt    chan os.Signal
	done         chan interface{}

	_notifications *Pubsub

	_log *log.Logger
}

func New(host string, password string) *Wattpilot {

	w := &Wattpilot{
		_host:     host,
		_password: password,

		connected:    make(chan bool),
		initialized:  make(chan bool),
		sendResponse: make(chan string),
		done:         make(chan interface{}),
		interrupt:    make(chan os.Signal),

		_currentConnection: nil,
		_isConnected:       false,
		_isInitialized:     false,
		_requestId:         1,
		_status:            make(map[string]interface{}),
	}

	w._log = log.New()
	w._log.SetFormatter(&log.JSONFormatter{})
	w._log.SetLevel(log.TraceLevel)

	signal.Notify(w.interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	w._notifications = NewPubsub()

	w.eventHandler = map[string]eventFunc{
		"hello":          w.onEventHello,
		"authRequired":   w.onEventAuthRequired,
		"response":       w.onEventResponse,
		"authSuccess":    w.onEventAuthSuccess,
		"authError":      w.onEventAuthError,
		"fullStatus":     w.onEventFullStatus,
		"deltaStatus":    w.onEventDeltaStatus,
		"clearInverters": w.onEventClearInverters,
		"updateInverter": w.onEventUpdateInverter,
	}

	go w.processLoop(context.Background())

	return w

}
func (w *Wattpilot) SetLogLevel(level log.Level) {
	w._log.SetLevel(level)
}

func (w *Wattpilot) ParseLogLevel(level string) error {
	loglevel, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	w._log.SetLevel(loglevel)
	return nil
}

func (w *Wattpilot) GetName() string {
	return w._name
}

func (w *Wattpilot) GetSerial() string {
	return w._serial
}

func (w *Wattpilot) GetHost() string {
	return w._host
}

func (w *Wattpilot) IsInitialized() bool {
	return w._isInitialized
}

func (w *Wattpilot) Properties() []string {
	keys := []string{}
	for k := range w._status {
		keys = append(keys, k)
	}
	return keys
}
func (w *Wattpilot) Alias() []string {
	keys := []string{}
	for k := range propertyMap {
		keys = append(keys, k)
	}
	return keys
}
func (w *Wattpilot) LookupAlias(name string) string {
	return propertyMap[name]
}

func hasKey(data map[string]interface{}, key string) bool {
	_, isKnown := data[key]
	return isKnown
}

func sha256sum(data string) string {
	bs := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", bs)
}

func (w *Wattpilot) getRequestId() int {
	current := w._requestId
	w._requestId += 1
	return current
}

func (w *Wattpilot) onEventHello(message map[string]interface{}) {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Hello from Wattpilot")

	if hasKey(message, "hostname") {
		w._hostname = message["hostname"].(string)
	}
	if hasKey(message, "friendly_name") {
		w._name = message["friendly_name"].(string)
	} else {
		w._name = w._hostname
	}
	w._serial = message["serial"].(string)
	if hasKey(message, "version") {
		w._version = message["version"].(string)
	}
	w._manufacturer = message["manufacturer"].(string)
	w._devicetype = message["devicetype"].(string)
	w._protocol = message["protocol"].(float64)
	if hasKey(message, "secured") {
		w._secured = message["secured"].(bool)
	}

	pwd_data := pbkdf2.Key([]byte(w._password), []byte(w._serial), 100000, 256, sha512.New)
	w._hashedpassword = base64.StdEncoding.EncodeToString([]byte(pwd_data))[:32]

}

func randomHexString(n int) string {
	b := make([]byte, (n+2)/2) // can be simplified to n/2 if n is always even

	if _, err := randomSource.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)[1 : n+1]
}

func (w *Wattpilot) onEventAuthRequired(message map[string]interface{}) {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Auhtentication required")

	token1 := message["token1"].(string)
	token2 := message["token2"].(string)

	w._token3 = randomHexString(32)
	hash1 := sha256sum(token1 + w._hashedpassword)
	hash := sha256sum(w._token3 + token2 + hash1)
	response := map[string]interface{}{
		"type":   "auth",
		"token3": w._token3,
		"hash":   hash,
	}
	err := w.onSendRepsonse(false, response)
	w._isInitialized = (err != nil)
}

func (w *Wattpilot) onSendRepsonse(secured bool, message map[string]interface{}) error {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Sending data to wattpilot")

	if secured {
		msgId := message["requestId"].(int)
		payload, _ := json.Marshal(message)

		mac := hmac.New(sha256.New, []byte(w._hashedpassword))
		mac.Write(payload)
		message = make(map[string]interface{})
		message["type"] = "securedMsg"
		message["data"] = string(payload)
		message["requestId"] = strconv.Itoa(msgId) + "sm"
		message["hmac"] = hex.EncodeToString(mac.Sum(nil))
	}

	data, _ := json.Marshal(message)

	context, cancel := context.WithTimeout(w._readContext, time.Second*CONTEXT_TIMEOUT)
	defer cancel()
	err := w._currentConnection.Write(context, websocket.MessageText, data)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wattpilot) onEventResponse(message map[string]interface{}) {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Response on Event ", message["type"])

	mType := message["type"].(string)
	success, ok := message["success"]
	if ok && success.(bool) {
		return
	}
	if mType == "response" {
		w.sendResponse <- message["message"].(string)
		return
	}
}

func (w *Wattpilot) onEventAuthSuccess(message map[string]interface{}) {
	w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Auhtentication successful")

	w.connected <- true
}

func (w *Wattpilot) onEventAuthError(message map[string]interface{}) {
	w._log.WithFields(log.Fields{"wattpilot": w._host}).Error("Auhtentication error", message)
	w.connected <- false
}

func (w *Wattpilot) onEventFullStatus(message map[string]interface{}) {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Full status update - is partial: ", message["partial"])

	isPartial := message["partial"].(bool)

	w.onEventDeltaStatus(message)

	if isPartial {
		return
	}

	w.initialized <- true
	w._isInitialized = true
}
func (w *Wattpilot) onEventDeltaStatus(message map[string]interface{}) {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Delta status update")

	w._readMutex.Lock()
	defer w._readMutex.Unlock()

	status := message["status"].(map[string]interface{})
	for k, v := range status {
		w._status[k] = v
	}
	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Trigger notifications")
	for k, v := range status {
		go w._notifications.Publish(k, v)
	}
}

func (w *Wattpilot) GetNotifications(prop string) <-chan interface{} {
	return w._notifications.Subscribe(prop)
}

func (w *Wattpilot) onEventClearInverters(message map[string]interface{}) {
	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("clear inverters")
}
func (w *Wattpilot) onEventUpdateInverter(message map[string]interface{}) {
	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("update inverters")
}
func (w *Wattpilot) Disconnect() {
	w.disconnectImpl()
	<-w.interrupt
}

func (w *Wattpilot) disconnectImpl() {
	w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Disconnecting")

	if !w._isConnected {
		return
	}

	if err := w._currentConnection.Close(websocket.StatusNormalClosure, "Bye Bye"); err != nil {
		//		w._log.WithFields(log.Fields{"wattpilot": w._host}).Warn("Error on closing connection: ", err)
	}

	w._isInitialized = false
	w._isConnected = false
	w._currentConnection = nil

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("closed connection")

}

func (w *Wattpilot) Connect() error {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Connecting")

	if w._isConnected || w._isInitialized {
		w._log.WithFields(log.Fields{"wattpilot": w._host}).Debug("Already Connected")
		return nil
	}

	w._readContext, w._readCancel = context.WithCancel(context.Background())

	var err error
	dialContext, cancel := context.WithTimeout(w._readContext, time.Second*CONTEXT_TIMEOUT)
	defer cancel()
	w._currentConnection, _, err = websocket.Dial(dialContext, fmt.Sprintf("ws://%s/ws", w._host), nil)
	if err != nil {
		return err
	}

	go w.receiveHandler(w._readContext)

	w._isConnected = <-w.connected
	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Connection is ", w._isConnected)
	if !w._isConnected {
		return errors.New("could not connect")
	}

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Connected - waiting for initializiation...")

	<-w.initialized

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Connected - and initializiated")

	return nil
}

func (w *Wattpilot) reconnect() {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Reconnecting")

	if w._isConnected {
		return
	}

	for {
		w._log.WithFields(log.Fields{"wattpilot": w._host}).Debug("Reconnect is running...")
		time.Sleep(time.Second * time.Duration(RECONNECT_TIMEOUT))
		if err := w.Connect(); err != nil {
			w._log.WithFields(log.Fields{"wattpilot": w._host}).Debug("Reconnect failure: ", err)
			continue
		}
		w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Successfully reconnected")
		return
	}
}

func (w *Wattpilot) processLoop(ctx context.Context) {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Starting process loop...")
	timeout := time.After(time.Second * (CONTEXT_TIMEOUT / 2))
	for {
		select {
		case <-timeout:
			if !w._isInitialized {
				continue
			}

			w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Hello there")
			pingCtx, cancel := context.WithDeadline(ctx, time.Now().Add(RECONNECT_TIMEOUT*time.Second))
			defer cancel()
			if err := w._currentConnection.Ping(pingCtx); err != nil {
				w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Hello failed: ", err)
				break
			}
			select {
			case <-time.After((1 + RECONNECT_TIMEOUT) * time.Second):
				w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Hello: overslept")
			case <-pingCtx.Done():
				w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Hello: ", pingCtx.Err())
			}
			break
		case <-w._readContext.Done():
			w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Read context is done")
			w.disconnectImpl()
			w.reconnect()
			break

		case <-ctx.Done():
		case <-w.interrupt:
			w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("Stopping process loop...")
			w.disconnectImpl()
			return
		}
	}
}

func (w *Wattpilot) receiveHandler(ctx context.Context) {

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Starting receive handler...")

	for {
		_, msg, err := w._currentConnection.Read(ctx)
		if err != nil {
			w._log.WithFields(log.Fields{"wattpilot": w._host}).Info("Stopping receive handler...")
			w._readCancel()
			return
		}
		data := make(map[string]interface{})
		err = json.Unmarshal(msg, &data)
		if err != nil {
			continue
		}
		msgType, isTypeAvailable := data["type"]
		if !isTypeAvailable {
			continue
		}
		w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("receiving ", msgType)

		funcCall, isKnown := w.eventHandler[msgType.(string)]
		if !isKnown {
			continue
		}
		funcCall(data)
		w._log.WithFields(log.Fields{"wattpilot": w._host}).Trace("done ", msgType)
	}

}

func (w *Wattpilot) GetProperty(name string) (interface{}, error) {

	w._readMutex.Lock()
	defer w._readMutex.Unlock()

	w._log.WithFields(log.Fields{"wattpilot": w._host}).Debug("Get Property ", name)

	if !w._isInitialized {
		return nil, errors.New("connection is not valid")
	}
	origName := name
	if v, isKnown := propertyMap[name]; isKnown {
		name = v
	}
	m, post := postProcess[origName]
	if post {
		name = m.key
	}

	if !hasKey(w._status, name) {
		return nil, errors.New("could not find value of " + name)
	}
	value := w._status[name]
	if post {
		value, _ = m.f(value)
	}
	return value, nil
}

func (w *Wattpilot) SetProperty(name string, value interface{}) error {

	w._readMutex.Lock()
	defer w._readMutex.Unlock()
	w._log.WithFields(log.Fields{"wattpilot": w._host}).Debug("setting property ", name, " to ", value)

	if !w._isInitialized {
		return errors.New("connection is not valid")
	}

	if !hasKey(w._status, name) {
		return errors.New("could not find reference for update on " + name)
	}

	err := w.sendUpdate(name, value)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wattpilot) transformValue(value interface{}) interface{} {

	switch value := value.(type) {
	case int:
		return value
	case int64:
		return value
	case float64:
		return value
	}
	in_value := fmt.Sprintf("%v", value)
	if out_value, err := strconv.Atoi(in_value); err == nil {
		return out_value
	}
	if out_value, err := strconv.ParseBool(in_value); err == nil {
		return out_value
	}
	if out_value, err := strconv.ParseFloat(in_value, 64); err == nil {
		return out_value
	}

	return in_value
}

func (w *Wattpilot) sendUpdate(name string, value interface{}) error {

	message := make(map[string]interface{})
	message["type"] = "setValue"
	message["requestId"] = w.getRequestId()
	message["key"] = name
	message["value"] = w.transformValue(value)
	return w.onSendRepsonse(w._secured, message)

}

func (w *Wattpilot) StatusInfo() {

	fmt.Println("Wattpilot: " + w._name)
	fmt.Println("Serial: ", w._serial)

	v, _ := w.GetProperty("car")
	fmt.Printf("Car Connected: %v\n", v)
	v, _ = w.GetProperty("alw")
	fmt.Printf("Charge Status %v\n", v)
	v, _ = w.GetProperty("imo")
	fmt.Printf("Mode: %v\n", v)
	v, _ = w.GetProperty("amp")
	fmt.Printf("Power: %v\n\nCharge: ", v)

	v1, v2, v3, _ := w.GetVoltages()
	fmt.Printf("%v V, %v V, %v V", v1, v2, v3)
	fmt.Printf("\n\t")

	i1, i2, i3, _ := w.GetCurrents()
	fmt.Printf("%v A, %v A, %v A", i1, i2, i3)
	fmt.Printf("\n\t")

	for _, i := range []string{"power1", "power2", "power3"} {
		v, _ := w.GetProperty(i)
		fmt.Printf("%v W, ", v)
	}
	fmt.Println("")
}

func (w *Wattpilot) GetPower() (float64, error) {

	v, err := w.GetProperty("power")
	if err != nil {
		return -1, err
	}
	return strconv.ParseFloat(v.(string), 64)
}

func (w *Wattpilot) GetCurrents() (float64, float64, float64, error) {

	var currents []float64
	for _, i := range []string{"amps1", "amps2", "amps3"} {
		v, err := w.GetProperty(i)
		if err != nil {
			return -1, -1, -1, err
		}
		fi, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return -1, -1, -1, err
		}

		currents = append(currents, fi)
	}
	return currents[0], currents[1], currents[2], nil
}

func (w *Wattpilot) GetVoltages() (float64, float64, float64, error) {

	var voltages []float64
	for _, i := range []string{"voltage1", "voltage2", "voltage2"} {
		v, err := w.GetProperty(i)
		if err != nil {
			return -1, -1, -1, err
		}
		fi, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return -1, -1, -1, err
		}

		voltages = append(voltages, fi)
	}
	return voltages[0], voltages[1], voltages[2], nil
}

func (w *Wattpilot) SetCurrent(current float64) error {

	return w.SetProperty("amp", current)
}

func (w *Wattpilot) GetRFID() (string, error) {

	resp, err := w.GetProperty("cak")
	if err != nil {
		return "", err
	}
	return resp.(string), nil
}
