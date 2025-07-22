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
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/coder/websocket"
	"golang.org/x/crypto/pbkdf2"
)

const (
	ContextTimeout    = 60 // seconds
	ConnectionTimeout = 30 // seconds
	ReconnectTimeout  = 5  // seconds

	EventTypeHello          = "hello"
	EventTypeAuthRequired   = "authRequired"
	EventTypeResponse       = "response"
	EventTypeAuthSuccess    = "authSuccess"
	EventTypeAuthError      = "authError"
	EventTypeFullStatus     = "fullStatus"
	EventTypeDeltaStatus    = "deltaStatus"
	EventTypeClearInverters = "clearInverters"
	EventTypeUpdateInverter = "updateInverter"

	RequestTypeAuth              = "auth"
	RequestTypeRequestFullStatus = "requestFullStatus"
	RequestTypeSetValue          = "setValue"
	RequestTypeSecuredMsg        = "securedMsg"
)

//go:generate go run gen/generate.go

type eventFunc func(map[string]interface{})

type wattPilotData struct {
	name         string
	hostname     string
	serial       string
	version      string
	manufacturer string
	deviceType   string
	protocol     float64
}

type Wattpilot struct {
	wattPilotData

	requestId   int64
	connected   chan bool
	initialized chan bool
	secured     bool

	readMutex  sync.Mutex
	writeMutex sync.Mutex

	host           string
	password       string
	hashedpassword string
	isInitialized  bool
	isConnected    bool
	isConnecting   atomic.Bool // New: To prevent concurrent reconnection attempts
	data           map[string]interface{}
	eventHandler   map[string]eventFunc

	sendResponse  chan string
	interrupt     chan os.Signal
	reconnectChan chan struct{} // New: To trigger immediate reconnection
	managerOnce   sync.Once     // New: To ensure connectionManager starts only once

	notify *Pubsub
	logger *logrus.Logger
	conn   *websocket.Conn
}

func New(host string, password string) *Wattpilot {

	w := &Wattpilot{
		host:           host,
		password:       password,
		hashedpassword: "",

		connected:     make(chan bool, 1),
		initialized:   make(chan bool, 1),
		sendResponse:  make(chan string),
		interrupt:     make(chan os.Signal, 1),
		reconnectChan: make(chan struct{}, 1),

		conn:          nil,
		isConnected:   false,
		isInitialized: false,
		requestId:     0,
		data:          make(map[string]interface{}),
		logger:        logrus.New(),
		notify:        NewPubsub(),
	}

	w.logger.SetFormatter(&logrus.JSONFormatter{})
	w.logger.SetLevel(logrus.ErrorLevel)
	if level := os.Getenv("WATTPILOT_LOG"); level != "" {
		if err := w.ParseLogLevel(level); err != nil {
			w.logger.Warn("Could not parse log level setting ", err)
		}
	}

	signal.Notify(w.interrupt, os.Interrupt, syscall.SIGTERM)

	w.eventHandler = map[string]eventFunc{
		EventTypeHello:          w.onEventHello,
		EventTypeAuthRequired:   w.onEventAuthRequired,
		EventTypeResponse:       w.onEventResponse,
		EventTypeAuthSuccess:    w.onEventAuthSuccess,
		EventTypeAuthError:      w.onEventAuthError,
		EventTypeFullStatus:     w.onEventFullStatus,
		EventTypeDeltaStatus:    w.onEventDeltaStatus,
		EventTypeClearInverters: w.onEventClearInverters,
		EventTypeUpdateInverter: w.onEventUpdateInverter,
	}

	return w

}
func (w *Wattpilot) SetLogger(delegate func(string, string)) {
	w.logger.AddHook(&CallHook{Call: delegate, LogLevels: logrus.AllLevels})
}

func (w *Wattpilot) SetLogLevel(level logrus.Level) {
	w.logger.SetLevel(level)
}

func (w *Wattpilot) ParseLogLevel(level string) error {
	loglevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	w.logger.SetLevel(loglevel)
	return nil
}

func (w *Wattpilot) GetName() string {
	return w.name
}

func (w *Wattpilot) GetSerial() string {
	return w.serial
}

func (w *Wattpilot) GetHost() string {
	return w.host
}

func (w *Wattpilot) IsInitialized() bool {
	return w.isInitialized
}

func (w *Wattpilot) Properties() []string {
	keys := []string{}

	w.readMutex.Lock()
	defer w.readMutex.Unlock()

	for k := range w.data {
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

func (w *Wattpilot) getRequestId() int64 {
	return atomic.AddInt64(&w.requestId, 1)
}

func (w *Wattpilot) GetNotifications(prop string) <-chan interface{} {
	return w.notify.Subscribe(prop)
}

func (w *Wattpilot) GetProperty(name string) (interface{}, error) {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Debug("Get Property ", name)

	if !w.isInitialized {
		return nil, errors.New("connection is not valid")
	}

	origName := name
	if v, isKnown := propertyMap[name]; isKnown {
		name = v
	}
	m, post := PostProcess[origName]
	if post {
		name = m.key
	}

	w.readMutex.Lock()
	defer w.readMutex.Unlock()

	if !hasKey(w.data, name) {
		return nil, errors.New("could not find value of " + name)
	}
	value := w.data[name]
	if post {
		value, _ = m.f(value)
	}
	return value, nil
}

func (w *Wattpilot) SetProperty(name string, value interface{}) error {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Debug("setting property ", name, " to ", value)

	if !w.isInitialized {
		return errors.New("connection is not valid")
	}

	w.readMutex.Lock()
	defer w.readMutex.Unlock()

	if !hasKey(w.data, name) {
		return errors.New("could not find reference for update on " + name)
	}

	return w.sendUpdate(name, value)

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
	if in_value == "nil" || in_value == "null" {
		return nil
	}
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

func (w *Wattpilot) StatusInfo() {

	fmt.Println("Wattpilot: " + w.name)
	fmt.Println("Serial: ", w.serial)

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

func (w *Wattpilot) getFloatValues(properties []string) ([]float64, error) {
	var values []float64
	for _, p := range properties {
		v, err := w.GetProperty(p)
		if err != nil {
			return nil, err
		}
		f, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return nil, err
		}
		values = append(values, f)
	}
	return values, nil
}

func (w *Wattpilot) GetCurrents() (float64, float64, float64, error) {
	values, err := w.getFloatValues([]string{"amps1", "amps2", "amps3"})
	if err != nil {
		return -1, -1, -1, err
	}
	return values[0], values[1], values[2], nil
}

func (w *Wattpilot) GetVoltages() (float64, float64, float64, error) {
	values, err := w.getFloatValues([]string{"voltage1", "voltage2", "voltage3"})
	if err != nil {
		return -1, -1, -1, err
	}
	return values[0], values[1], values[2], nil
}

func (w *Wattpilot) SetCurrent(current float64) error {

	return w.SetProperty("amp", current)
}

func (w *Wattpilot) GetRFID() (string, error) {

	resp, err := w.GetProperty("trx")
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", nil
	}
	rfid := resp.(float64)
	return fmt.Sprint(rfid), nil

}

func (w *Wattpilot) GetCarIdentifier() (string, error) {

	resp, err := w.GetProperty("cak")
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", nil
	}
	return resp.(string), nil

}

func (w *Wattpilot) RequestStatusUpdate() error {
	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Debug("requesting status update...")

	message := make(map[string]interface{})
	message["type"] = RequestTypeRequestFullStatus
	message["requestId"] = w.getRequestId()
	if err := w.onSendResponse(w.secured, message); err != nil {
		return err
	}

	return nil
}

func (w *Wattpilot) Connect() error {
	if w.IsInitialized() {
		return nil
	}

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Connecting...")

	w.managerOnce.Do(func() {
		go w.connectionManager()
	})

	if w.connectAndWait(ConnectionTimeout * time.Second) {
		return nil
	}

	return errors.New("failed to connect within timeout")
}

func (w *Wattpilot) connectAndWait(timeout time.Duration) bool {
	// Non-blocking trigger
	select {
	case w.reconnectChan <- struct{}{}:
	default:
	}

	// Wait for initialization
	select {
	case <-w.initialized:
		return true
	case <-time.After(timeout):
		return w.IsInitialized()
	}
}

func (w *Wattpilot) connectImpl() error {
	if w.isConnected {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ConnectionTimeout*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/ws", w.host), nil)
	if err != nil {
		return err
	}
	w.conn = conn

	go w.receiveHandler()

	// Wait for handshake
	select {
	case w.isConnected = <-w.connected:
		if !w.isConnected {
			w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Authentication timeout")
			w.disconnectImpl()
			return errors.New("authentication failed")
		}
	case <-ctx.Done():
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Connection handshake timeout")
		w.disconnectImpl()
		return errors.New("connection handshake timeout")
	}

	// Wait for initialization
	select {
	case <-w.initialized:
		// isInitialized is set by onEventFullStatus
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host, "initialized": w.isInitialized, "auhtenticated": w.isConnected}).Info("Initialization done")

	case <-ctx.Done():
		if !w.isInitialized {
			w.logger.WithFields(logrus.Fields{"wattpilot": w.host, "initialized": w.isInitialized, "auhtenticated": w.isConnected}).Info("Initialization timeout")
			w.disconnectImpl()
			return errors.New("initialization timeout")
		}
	}

	return nil
}

func (w *Wattpilot) reconnect() {
	if !w.isConnecting.CompareAndSwap(false, true) {
		return // Already reconnecting
	}
	defer w.isConnecting.Store(false)

	if w.isConnected {
		if err := w.RequestStatusUpdate(); err == nil {
			return // Healthy
		}
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Warn("Health check failed, reconnecting.")
		w.disconnectImpl()
	}

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Attempting to connect...")
	if err := w.connectImpl(); err != nil {
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Error("Failed to connect: ", err)
	} else {
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Successfully reconnected.")
	}
}

func (w *Wattpilot) Disconnect() {
	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Disconnecting client...")
	w.interrupt <- syscall.SIGINT
}

func (w *Wattpilot) disconnectImpl() {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host, "caller": getCallerFunctionName()}).Info("Closing connection...")

	if w.conn == nil {
		return // Already disconnected
	}

	if err := (*w.conn).CloseNow(); err != nil {
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Error on closing connection: ", err)
	}

	w.isInitialized = false
	w.isConnected = false
	w.conn = nil
	w.data = make(map[string]interface{})

	// Drain channels to prevent blocking future operations
	select {
	case <-w.connected:
	default:
	}
	select {
	case <-w.initialized:
	default:
	}
}

func (w *Wattpilot) connectionManager() {
	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Starting connection manager.")

	healthCheckTicker := time.NewTicker(time.Second * ContextTimeout) // Use for health checks
	defer healthCheckTicker.Stop()

	for {
		select {
		case <-healthCheckTicker.C:
			if !w.isConnected || !w.isInitialized {
				// Connection is not healthy, attempt to reconnect
				select {
				case w.reconnectChan <- struct{}{}:
				default:
				}
			}
		case <-w.reconnectChan:
			time.Sleep(ReconnectTimeout * time.Second) // Wait a bit before reconnecting
			go w.reconnect()
		case <-w.interrupt:
			w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Stopping connection manager.")
			w.disconnectImpl()
			return
		}
	}
}

func (w *Wattpilot) receiveHandler() {
	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Starting receive handler...")

	for {
		if w.conn == nil {
			w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Connection is nil, stopping receive handler.")
			return
		}
		_, msg, err := w.conn.Read(context.Background())
		if err != nil {
			w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Warn("Connection read error: ", err)
			w.disconnectImpl()
			select {
			case w.reconnectChan <- struct{}{}:
			default:
			}
			w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Stopping receive handler.")
			return
		}
		data := make(map[string]interface{})
		if err := json.Unmarshal(msg, &data); err != nil {
			w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Errorf("failed to unmarshal message: %v", err)
			continue
		}
		msgType, isTypeAvailable := data["type"]
		if !isTypeAvailable {
			continue
		}
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("receiving ", msgType)

		funcCall, isKnown := w.eventHandler[msgType.(string)]
		if !isKnown {
			continue
		}
		funcCall(data)
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("done ", msgType)
	}
}

func (w *Wattpilot) onEventHello(message map[string]interface{}) {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Hello from Wattpilot")

	if hasKey(message, "hostname") {
		w.hostname = message["hostname"].(string)
	}
	if hasKey(message, "friendlyname") {
		w.name = message["friendlyname"].(string)
	} else {
		w.name = w.hostname
	}
	w.serial = message["serial"].(string)
	if hasKey(message, "version") {
		w.version = message["version"].(string)
	}
	w.manufacturer = message["manufacturer"].(string)
	w.deviceType = message["devicetype"].(string)
	w.protocol = message["protocol"].(float64)
	if hasKey(message, "secured") {
		w.secured = message["secured"].(bool)
	}
}

func (w *Wattpilot) onEventAuthRequired(message map[string]interface{}) {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Auhtentication required")

	pwd_data := pbkdf2.Key([]byte(w.password), []byte(w.serial), 100000, 256, sha512.New)
	w.hashedpassword = base64.StdEncoding.EncodeToString([]byte(pwd_data))[:32]

	token1 := message["token1"].(string)
	token2 := message["token2"].(string)

	authToken := randomHexString(32)
	hash1 := sha256sum(token1 + w.hashedpassword)
	hash := sha256sum(authToken + token2 + hash1)
	response := map[string]interface{}{
		"type":   RequestTypeAuth,
		"token3": authToken,
		"hash":   hash,
	}
	if err := w.onSendResponse(false, response); err != nil {
		w.isInitialized = false
		return
	}
	w.isInitialized = true
}

func (w *Wattpilot) onSendResponse(secured bool, message map[string]interface{}) error {

	w.writeMutex.Lock()
	defer w.writeMutex.Unlock()

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Sending data to wattpilot: ", message["requestId"], " secured: ", secured)

	if secured {
		msgId := message["requestId"].(int64)
		payload, err := json.Marshal(message)
		if err != nil {
			return err
		}

		mac := hmac.New(sha256.New, []byte(w.hashedpassword))
		mac.Write(payload)
		message = make(map[string]interface{})
		message["type"] = RequestTypeSecuredMsg
		message["data"] = string(payload)
		message["requestId"] = fmt.Sprintf("%d", msgId) + "sm"
		message["hmac"] = hex.EncodeToString(mac.Sum(nil))
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = w.conn.Write(context.Background(), websocket.MessageText, data)
	if err != nil {
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Sending data to wattpilot: ", message["data"], " Error: ", err)
		return err
	}
	return nil
}

func (w *Wattpilot) onEventResponse(message map[string]interface{}) {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Response on Event ", message["type"])

	mType := message["type"].(string)
	success, ok := message["success"]
	if ok && success.(bool) {
		return
	}
	if !success.(bool) {
		w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Error("Failure happened: ", message["message"])
		return
	}
	if mType == EventTypeResponse {
		w.sendResponse <- message["message"].(string)
		return
	}
}

func (w *Wattpilot) onEventAuthSuccess(message map[string]interface{}) {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Info("Auhtentication successful")
	w.connected <- true

}

func (w *Wattpilot) onEventAuthError(message map[string]interface{}) {
	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Error("Auhtentication error", message)
	w.connected <- false
}

func (w *Wattpilot) onEventFullStatus(message map[string]interface{}) {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Full status update - is partial: ", message["partial"])

	isPartial := message["partial"].(bool)

	w.updateStatus(message)

	if isPartial {
		return
	}

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Initialization done")

	if w.IsInitialized() {
		return
	}

	w.initialized <- true
	w.isInitialized = true

}

func (w *Wattpilot) onEventDeltaStatus(message map[string]interface{}) {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Delta status update")
	w.updateStatus(message)

}

func (w *Wattpilot) updateStatus(message map[string]interface{}) {

	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Enter Data-status updates")
	statusUpdates := message["status"].(map[string]interface{})
	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("Data-status gets updates #", len(statusUpdates))

	w.readMutex.Lock()
	defer w.readMutex.Unlock()

	for k, v := range statusUpdates {
		w.data[k] = v
		go w.notify.Publish(k, v)
	}
}

func (w *Wattpilot) onEventClearInverters(message map[string]interface{}) {
	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("clear inverters")
}
func (w *Wattpilot) onEventUpdateInverter(message map[string]interface{}) {
	w.logger.WithFields(logrus.Fields{"wattpilot": w.host}).Trace("update inverters")
}

func (w *Wattpilot) sendUpdate(name string, value interface{}) error {

	message := make(map[string]interface{})
	message["type"] = RequestTypeSetValue
	message["requestId"] = w.getRequestId()
	message["key"] = name
	message["value"] = w.transformValue(value)
	return w.onSendResponse(w.secured, message)

}
