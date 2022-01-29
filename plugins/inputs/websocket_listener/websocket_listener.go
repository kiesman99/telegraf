package websocket_listener

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type WebsocketConsumer struct {
	ServiceAddress string `toml:"service_address"`
	parser         parsers.Parser
	accumulator    telegraf.Accumulator
	Log            telegraf.Logger
	Conn           *websocket.Conn
	CloseSignal    chan struct{}
}

func (w *WebsocketConsumer) SetParser(parser parsers.Parser) {
	w.parser = parser
}

func (w *WebsocketConsumer) Description() string {
	return "Plugin to connect to an establised WebSocket server and listens for incoming metrics."
}

func (w *WebsocketConsumer) SampleConfig() string {
	return `
	## Address of the server to connect to
	service_address = "ws://localhost:3210/telegraf"

	## Data format to consume.
	## Each data format has its own unique set of configuration options, read
	## more about them here:
	## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
	data_format = "influx"
	`
}

func (w *WebsocketConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (w *WebsocketConsumer) Init() error {
	return nil
}

func (w *WebsocketConsumer) Start(acc telegraf.Accumulator) error {
	w.accumulator = acc
	w.Conn = w.connectToServer()
	return nil
}

func (w *WebsocketConsumer) Stop() {
	w.Log.Info("Closing connection")
	w.CloseSignal <- struct{}{}
	close(w.CloseSignal)
}

func (w *WebsocketConsumer) connectToServer() *websocket.Conn {
	w.Log.Infof("Connecting to %s\n", w.ServiceAddress)

	c, _, err := websocket.DefaultDialer.Dial(w.ServiceAddress, nil)
	if err != nil {
		w.Log.Errorf("error dialing %v\n", err)
	}

	go func() {
		w.Log.Infof("Start Listening to %s\n", w.ServiceAddress)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				w.Log.Errorf("error reading message: %v", err)
				continue
			}

			metrics, err := w.parser.Parse(message)
			if err != nil {
				w.Log.Errorf("error parsing metrics: %v", err)
				continue
			}

			for _, m := range metrics {
				w.accumulator.AddMetric(m)
			}

			w.Log.Debugf("recv: %s", message)
		}
	}()

	go func() {
		<-w.CloseSignal
		w.Conn.Close()
	}()

	return c
}

func NewWebsocketConsumer() *WebsocketConsumer {
	parser, err := parsers.NewInfluxParser()
	if err != nil {
		fmt.Printf("Error creating Parser: %v\n", err)
	}

	return &WebsocketConsumer{
		parser:      parser,
		CloseSignal: make(chan struct{}),
	}
}

func init() {
	inputs.Add("websocket_listener", func() telegraf.Input { return NewWebsocketConsumer() })
}
