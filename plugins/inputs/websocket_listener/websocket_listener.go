package websocket_listener

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

var upgrader = websocket.Upgrader{}

type WebsocketConsumer struct {
	ServiceAddress string `toml:"service_address"`
	parser         parsers.Parser
	accumulator    telegraf.Accumulator
	Log            telegraf.Logger
	Conn           *websocket.Conn
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

// func (w *WebsocketConsumer) dataGatherer(writer http.ResponseWriter, req *http.Request) {
// 	c, err := upgrader.Upgrade(writer, req, nil)
// 	if err != nil {
// 		fmt.Errorf("upgrade: %v", err)
// 		return
// 	}
// 	defer c.Close()
// 	for {
// 		_, message, err := c.ReadMessage()
// 		if err != nil {
// 			fmt.Errorf("error reading message: %v", err)
// 			break
// 		}

// 		metrics, err := w.parser.Parse(message)
// 		if err != nil {
// 			fmt.Errorf("error parsing metrics: %v", err)
// 			break
// 		}

// 		for _, m := range metrics {
// 			w.accumulator.AddMetric(m)
// 		}

// 		w.Log.Infof("recv: %s", message)
// 	}
// }

func (w *WebsocketConsumer) Start(acc telegraf.Accumulator) error {
	w.accumulator = acc
	w.Conn = w.connectToServer()
	return nil
}

func (w *WebsocketConsumer) Stop() {
	// TODO: Close connection here

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// w.Log.Info("Shutting down server")
	// if err := w.server.Shutdown(ctx); err != nil {
	// 	log.Fatalf("Error shutting down websocket_listener server, %v", err)
	// }
}

func (w *WebsocketConsumer) connectToServer() *websocket.Conn {
	// TODO: Extract port from w.ServerAddress
	// w.Log.Info("Starting Websocket Server")
	// mux := http.NewServeMux()
	// mux.HandleFunc("/watch", w.dataGatherer)

	// server := http.Server{Addr: ":3210", Handler: mux}

	// go func() {
	// 	log.Fatal(server.ListenAndServe())
	// }()

	w.Log.Info("ConnectToServer Invoked")

	c, _, err := websocket.DefaultDialer.Dial("ws://host.docker.internal:3210/telegraf", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	go func() {
		w.Log.Info("Start Listening")
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				w.Log.Errorf("error reading message: %v", err)
				continue
			}

			fmt.Printf("Received: %s\n", string(message))

			metrics, err := w.parser.Parse(message)
			if err != nil {
				w.Log.Errorf("error parsing metrics: %v", err)
				continue
			}

			for _, m := range metrics {
				w.accumulator.AddMetric(m)
			}

			w.Log.Infof("recv: %s", message)
		}
	}()

	return c
}

func NewWebsocketConsumer() *WebsocketConsumer {
	parser, err := parsers.NewInfluxParser()
	if err != nil {
		fmt.Println("Error creating Parser: %v", err)
	}

	return &WebsocketConsumer{
		parser: parser,
	}
}

func init() {
	inputs.Add("websocket_listener", func() telegraf.Input { return NewWebsocketConsumer() })
}
