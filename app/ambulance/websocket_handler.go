package ambulance

import (
	"encoding/json"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

// WebSocket message types
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ReadPump handles incoming messages from WebSocket
func (c *Client) ReadPump(service Service) {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket error: %v", err)
			}
			break
		}

		if os.Getenv("DEBUG") == "true" {
			logrus.WithFields(logrus.Fields{
				"component": "websocket",
				"direction": "incoming",
				"client_id": c.ID,
			}).Info(string(message))
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			logrus.Errorf("Failed to unmarshal message: %v", err)
			continue
		}

		c.handleMessage(wsMsg, service)
	}
}

// WritePump sends messages to WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg WSMessage, service Service) {
	switch msg.Type {
	case "location_update":
		if c.ClientType != ClientTypeMobile {
			return
		}

		var update LocationUpdate
		if err := json.Unmarshal(msg.Payload, &update); err != nil {
			logrus.Errorf("Failed to unmarshal location update: %v", err)
			return
		}

		update.AmbulanceID = c.AmbulanceID
		if update.Timestamp == 0 {
			update.Timestamp = time.Now().Unix()
		}

		// Update hub cache and broadcast
		c.Hub.UpdateLocation(&update)

		// Save to database (async)
		go service.SaveLocationHistory(update)

	case "status_update":
		if c.ClientType != ClientTypeMobile {
			return
		}

		var statusUpdate struct {
			Status AmbulanceStatus `json:"status"`
		}
		if err := json.Unmarshal(msg.Payload, &statusUpdate); err != nil {
			return
		}

		go service.UpdateAmbulanceStatus(c.AmbulanceID, statusUpdate.Status)

		// Broadcast status change to web clients
		c.Hub.Broadcast <- &BroadcastMessage{
			Type: "status_update",
			Payload: map[string]interface{}{
				"ambulance_id": c.AmbulanceID,
				"status":       statusUpdate.Status,
				"timestamp":    time.Now().Unix(),
			},
			Target: "web",
		}

	case "subscribe":
		// Web clients can subscribe to specific ambulances
		if c.ClientType != ClientTypeWeb {
			return
		}
		// Handle subscription logic

	case "ping":
		// Respond with pong
		c.Send <- []byte(`{"type":"pong"}`)
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msgType string, payload interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := &BroadcastMessage{
		Type:    msgType,
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
		return nil
	default:
		return nil
	}
}
