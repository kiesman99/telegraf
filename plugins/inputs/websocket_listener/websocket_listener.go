package websocket_listener

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

var upgrader = websocket.Upgrader{}

type WebsocketConsumer struct {
	Port        int `toml:"port"`
	parser      parsers.Parser
	accumulator telegraf.Accumulator
	Log         telegraf.Logger
}

func (w *WebsocketConsumer) SetParser(parser parsers.Parser) {
	w.parser = parser
}

func (w *WebsocketConsumer) Description() string {
	return "Plugin to listen to an establised WebSocket server or create one to publish and listen to."
}

func (w *WebsocketConsumer) SampleConfig() string {
	return `
	## Port which will be used for the server
	# port = 3210

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

func (w *WebsocketConsumer) dataGatherer(writer http.ResponseWriter, req *http.Request) {
	c, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		fmt.Errorf("upgrade: %v", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Errorf("error reading message: %v", err)
			break
		}

		metrics, err := w.parser.Parse(message)
		if err != nil {
			fmt.Errorf("error parsing metrics: %v", err)
			break
		}

		for _, m := range metrics {
			w.accumulator.AddMetric(m)
		}

		w.Log.Infof("recv: %s", message)
	}
}

func (w *WebsocketConsumer) Start(acc telegraf.Accumulator) error {
	w.accumulator = acc
	err := w.startServer()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (w *WebsocketConsumer) Stop() {
	// Clean up
}

func (w *WebsocketConsumer) startServer() error {
	w.Log.Info("Starting Websocket Server")
	http.HandleFunc("/watch", w.dataGatherer)
	go func() {
		log.Fatal(http.ListenAndServe(":3210", nil))
	}()

	return nil
}

func NewWebsocketConsumer() *WebsocketConsumer {
	parser, _ := parsers.NewInfluxParser()

	return &WebsocketConsumer{
		parser: parser,
	}
}

func init() {
	inputs.Add("websocket_listener", func() telegraf.Input { return NewWebsocketConsumer() })
}
