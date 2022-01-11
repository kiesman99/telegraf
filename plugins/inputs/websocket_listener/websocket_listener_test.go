package websocket_listener

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestWebsocketConsumer(t *testing.T) {
	wsc := NewWebsocketConsumer()
	wsc.Log = testutil.Logger{}
	wsc.ServiceAddress = ":3210"

	acc := &testutil.Accumulator{}
	wsc.accumulator = acc

	err := wsc.Start(acc)
	require.NoError(t, err)
	defer wsc.Stop()
}

func testWebsocketListener(t *require.TestingT, wsl *WebsocketConsumer) {
	// mstr12 := []byte("test,foo=bar v=1i 123456789\ntest,foo=baz v=2i 123456790\n")
	// mstr3 := []byte("test,foo=zab v=3i 123456791\n")

}
