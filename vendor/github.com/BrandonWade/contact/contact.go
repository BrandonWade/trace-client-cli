package contact

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

// Connection - represents a websocket connection
type Connection struct {
	conn     *websocket.Conn
	buffSize int
}

// NewConnection - return a new Connection
func NewConnection(buffSize int) *Connection {
	return &Connection{
		buffSize: buffSize,
	}
}

// Open - upgrades an http request into a websocket connection
func (c *Connection) Open(w *http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  c.buffSize,
		WriteBufferSize: c.buffSize,
	}

	conn, err := upgrader.Upgrade(*w, r, nil)
	if err != nil {
		log.Fatal("error upgrading http request into websocket", err)
	}

	c.conn = conn
}

// Dial - sends an http request to be upgraded to a websocket
func (c *Connection) Dial(host, path string, params map[string]string) {
	vals := url.Values{}
	vals.Set("file", params["file"])
	rawQuery := vals.Encode()

	url := url.URL{Scheme: "ws", Host: host, Path: path, RawQuery: rawQuery}
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal("error sending websocket http request")
	}

	c.conn = conn
}

// ReadJSON - reads JSON from a websocket connection and stores it in the provided struct
func (c *Connection) ReadJSON(m interface{}) {
	c.conn.ReadJSON(&m)
	return
}

// Read - reads from a websocket connection
func (c *Connection) Read() (int, []byte, error) {
	messageType, data, err := c.conn.ReadMessage()
	return messageType, data, err
}

// WriteJSON - writes a struct to the websocket connection
func (c *Connection) WriteJSON(m interface{}) {
	c.conn.WriteJSON(m)
}

func (c *Connection) Write(s string) {
	c.conn.WriteMessage(websocket.TextMessage, []byte(s))
}

// WriteBinary - writes a binary string to the websocket connection
func (c *Connection) WriteBinary(data []byte) {
	c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Close - closes an open websocket connection
func (c *Connection) Close() {
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.conn.Close()
}
