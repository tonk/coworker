package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 4096
)

type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	userID      uint
	username    string
	displayName string
	avatarURL   string
	projectID   uint
	globalRole  string

	// injected handler for incoming chat messages
	onMessage func(client *Client, raw []byte)
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uint, username, displayName, avatarURL string, projectID uint, globalRole string, onMessage func(*Client, []byte)) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		userID:      userID,
		username:    username,
		displayName: displayName,
		avatarURL:   avatarURL,
		projectID:   projectID,
		globalRole:  globalRole,
		onMessage:   onMessage,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws error: %v", err)
			}
			break
		}
		if c.onMessage != nil {
			c.onMessage(c, raw)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) UserID() uint      { return c.userID }
func (c *Client) ProjectID() uint   { return c.projectID }
func (c *Client) GlobalRole() string { return c.globalRole }

func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) SendError(code, message, id string) {
	msg := Message{
		Type:    TypeError,
		Payload: ErrorPayload{Code: code, Message: message, ID: id},
	}
	data, _ := json.Marshal(msg)
	select {
	case c.send <- data:
	default:
	}
}
