package watch_websocket

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

type WatchWebsocketConsumer struct {
	Port        int `toml:"port"`
	parser      parsers.Parser
	accumulator telegraf.Accumulator
	Log         telegraf.Logger
}

func (w *WatchWebsocketConsumer) SetParser(parser parsers.Parser) {
	w.parser = parser
}

func (w *WatchWebsocketConsumer) Description() string {
	return "Simple Websocket Server that should be "
}

func (w *WatchWebsocketConsumer) SampleConfig() string {
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

func (w *WatchWebsocketConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (w *WatchWebsocketConsumer) Init() error {
	return nil
}

func (w *WatchWebsocketConsumer) watchDataGatherer(writer http.ResponseWriter, req *http.Request) {
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

func (w *WatchWebsocketConsumer) Start(acc telegraf.Accumulator) error {
	w.accumulator = acc
	err := w.startServer()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (w *WatchWebsocketConsumer) Stop() {
	// Clean up
}

func (w *WatchWebsocketConsumer) startServer() error {
	w.Log.Info("Starting Websocket Server")
	http.HandleFunc("/watch", w.watchDataGatherer)
	go func() {
		log.Fatal(http.ListenAndServe(":3210", nil))
	}()

	return nil
}

func NewWatchWebsocketConsumer() *WatchWebsocketConsumer {
	parser, _ := parsers.NewInfluxParser()

	return &WatchWebsocketConsumer{
		parser: parser,
	}
}

func init() {
	inputs.Add("watch_websocket", func() telegraf.Input { return NewWatchWebsocketConsumer() })
}
