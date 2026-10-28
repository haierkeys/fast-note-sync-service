package app

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/lxzan/gws"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type websocketSecurityApp struct{ AppContainer }

func (*websocketSecurityApp) Logger() *zap.Logger    { return zap.NewNop() }
func (*websocketSecurityApp) IsProductionMode() bool { return true }
func (*websocketSecurityApp) IsReturnSuccess() bool  { return true }
func (*websocketSecurityApp) SubmitTaskAsync(ctx context.Context, task func(context.Context) error) error {
	return task(ctx)
}

type websocketSecurityClient struct {
	gws.BuiltinEventHandler
	messages chan []byte
}

func (c *websocketSecurityClient) OnMessage(_ *gws.Conn, message *gws.Message) {
	defer message.Close()
	c.messages <- append([]byte(nil), message.Data.Bytes()...)
}

// Exercise the real WebSocket dispatch path rather than a detached permission helper.
func TestWebsocketAnonymousBinaryAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var panics atomic.Int32
	w := NewWebsocketServer(WSConfig{GWSOption: gws.ServerOption{Recovery: func(gws.Logger) {
		if recover() != nil {
			panics.Add(1)
		}
	}}}, &websocketSecurityApp{})
	var binaryCalls atomic.Int32
	w.UseBinary("00", func(c *WebsocketClient, _ []byte) {
		binaryCalls.Add(1)
		c.ToResponse(code.Success)
	})
	w.UseWithoutAuth("Probe", func(c *WebsocketClient, _ *WebSocketMessage) { c.ToResponse(code.Success) })
	// The fixture authenticates a synthetic user so empty legacy scopes remain covered.
	w.UseWithoutAuth("FixtureAuthenticate", func(c *WebsocketClient, _ *WebSocketMessage) {
		c.User = &UserEntity{UID: 1}
		c.User.ID = "1"
		c.ToResponse(code.Success)
	})
	router := gin.New()
	router.GET("/sync", w.Run())
	server := httptest.NewServer(router)
	defer server.Close()

	connect := func(t *testing.T) (*gws.Conn, func() int) {
		t.Helper()
		events := &websocketSecurityClient{messages: make(chan []byte, 8)}
		conn, _, err := gws.NewClient(events, &gws.ClientOption{Addr: "ws" + strings.TrimPrefix(server.URL, "http") + "/sync"})
		require.NoError(t, err)
		go conn.ReadLoop()
		t.Cleanup(func() { _ = conn.NetConn().Close(); w.WaitAllClosed(time.Second) })
		return conn, func() int {
			t.Helper()
			select {
			case payload := <-events.messages:
				var response struct {
					Code int `json:"code"`
				}
				require.NoError(t, json.Unmarshal(payload, &response))
				return response.Code
			case <-time.After(time.Second):
				t.Fatal("no response: message processing may have panicked or disconnected")
				return 0
			}
		}
	}

	t.Run("anonymous upload does not reach handler", func(t *testing.T) {
		conn, receiveCode := connect(t)
		require.NoError(t, conn.WriteMessage(gws.OpcodeBinary, append([]byte("00"), make([]byte, 41)...)))
		require.Equal(t, 307, receiveCode())
		require.Zero(t, binaryCalls.Load())
	})
	for _, test := range []struct {
		name    string
		opcode  gws.Opcode
		payload string
	}{
		{"empty binary", gws.OpcodeBinary, ""},
		{"short binary", gws.OpcodeBinary, "x"},
		{"unnegotiated protobuf", gws.OpcodeBinary, "pb-invalid"},
		{"unknown binary", gws.OpcodeBinary, "zz-unknown"},
		{"malformed text", gws.OpcodeText, "no-separator"},
	} {
		t.Run(test.name, func(t *testing.T) {
			conn, receiveCode := connect(t)
			require.NoError(t, conn.WriteMessage(test.opcode, []byte(test.payload)))
			require.NoError(t, conn.WriteMessage(gws.OpcodeText, []byte("Probe|{}")))
			require.Equal(t, 1, receiveCode())
			require.Zero(t, panics.Load(), "malformed message caused a recovered panic")
		})
	}
	t.Run("authenticated legacy empty scope still works", func(t *testing.T) {
		conn, receiveCode := connect(t)
		require.NoError(t, conn.WriteMessage(gws.OpcodeText, []byte("FixtureAuthenticate|{}")))
		require.Equal(t, 1, receiveCode())
		require.NoError(t, conn.WriteMessage(gws.OpcodeBinary, []byte("00fixture")))
		require.Equal(t, 1, receiveCode())
		require.EqualValues(t, 1, binaryCalls.Load())
	})
}
