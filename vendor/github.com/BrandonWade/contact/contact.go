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
func (c *Connection) Dial(host, path string) {
	url := url.URL{Scheme: "ws", Host: host, Path: path}
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal("error sending websocket http request")
	}

	c.conn = conn
}

// Close - closes an open websocket connection
func (c *Connection) Close() {
	c.conn.WriteMessage(websocket.CloseNormalClosure, []byte(""))
	c.conn.Close()
}
