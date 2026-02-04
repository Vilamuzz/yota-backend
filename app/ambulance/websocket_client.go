package ambulance

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Client represents a WebSocket connection
type Client struct {
	ID          string
	Conn        *websocket.Conn
	Hub         *Hub
	Send        chan []byte
	ClientType  ClientType
	AmbulanceID string // For mobile clients (drivers)
	mu          sync.Mutex
}

type ClientType string

const (
	ClientTypeMobile ClientType = "mobile" // Ambulance driver app
	ClientTypeWeb    ClientType = "web"    // Admin/monitoring dashboard
)

// Hub manages all WebSocket connections
type Hub struct {
	// Registered clients
	MobileClients map[string]*Client // ambulanceID -> client
	WebClients    map[string]*Client // clientID -> client

	// Channels for communication
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *BroadcastMessage

	// Location cache for quick access
	LocationCache map[string]*LocationUpdate
	cacheMu       sync.RWMutex

	mu sync.RWMutex
}

type BroadcastMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Target  string      `json:"target,omitempty"` // "all", "web", "mobile", or specific ID
}

func NewHub() *Hub {
	return &Hub{
		MobileClients: make(map[string]*Client),
		WebClients:    make(map[string]*Client),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Broadcast:     make(chan *BroadcastMessage, 256),
		LocationCache: make(map[string]*LocationUpdate),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch client.ClientType {
	case ClientTypeMobile:
		h.MobileClients[client.AmbulanceID] = client
		logrus.Infof("Mobile client registered: ambulance %s", client.AmbulanceID)

		// Notify web clients about new ambulance online
		h.notifyWebClients(&BroadcastMessage{
			Type: "ambulance_online",
			Payload: map[string]interface{}{
				"ambulance_id": client.AmbulanceID,
				"timestamp":    time.Now().Unix(),
			},
		})

	case ClientTypeWeb:
		h.WebClients[client.ID] = client
		logrus.Infof("Web client registered: %s", client.ID)

		// Send current ambulance locations to new web client
		h.sendCurrentLocations(client)
	}
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch client.ClientType {
	case ClientTypeMobile:
		if _, ok := h.MobileClients[client.AmbulanceID]; ok {
			delete(h.MobileClients, client.AmbulanceID)
			close(client.Send)
			logrus.Infof("Mobile client disconnected: ambulance %s", client.AmbulanceID)

			// Notify web clients about ambulance offline
			h.notifyWebClients(&BroadcastMessage{
				Type: "ambulance_offline",
				Payload: map[string]interface{}{
					"ambulance_id": client.AmbulanceID,
					"timestamp":    time.Now().Unix(),
				},
			})
		}

	case ClientTypeWeb:
		if _, ok := h.WebClients[client.ID]; ok {
			delete(h.WebClients, client.ID)
			close(client.Send)
			logrus.Infof("Web client disconnected: %s", client.ID)
		}
	}
}

func (h *Hub) broadcastMessage(message *BroadcastMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("Failed to marshal broadcast message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	switch message.Target {
	case "web", "":
		for _, client := range h.WebClients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.WebClients, client.ID)
			}
		}
	case "mobile":
		for _, client := range h.MobileClients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.MobileClients, client.AmbulanceID)
			}
		}
	case "all":
		for _, client := range h.WebClients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.WebClients, client.ID)
			}
		}
		for _, client := range h.MobileClients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.MobileClients, client.AmbulanceID)
			}
		}
	}
}

func (h *Hub) notifyWebClients(message *BroadcastMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	for _, client := range h.WebClients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.WebClients, client.ID)
		}
	}
}

func (h *Hub) sendCurrentLocations(client *Client) {
	h.cacheMu.RLock()
	defer h.cacheMu.RUnlock()

	locations := make([]*LocationUpdate, 0, len(h.LocationCache))
	for _, loc := range h.LocationCache {
		locations = append(locations, loc)
	}

	message := &BroadcastMessage{
		Type:    "initial_locations",
		Payload: locations,
	}

	data, _ := json.Marshal(message)
	client.Send <- data
}

func (h *Hub) UpdateLocation(update *LocationUpdate) {
	h.cacheMu.Lock()
	h.LocationCache[update.AmbulanceID] = update
	h.cacheMu.Unlock()

	// Broadcast to all web clients
	h.Broadcast <- &BroadcastMessage{
		Type:    "location_update",
		Payload: update,
		Target:  "web",
	}
}

func (h *Hub) GetOnlineAmbulances() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.MobileClients))
	for id := range h.MobileClients {
		ids = append(ids, id)
	}
	return ids
}

func (h *Hub) GetAmbulanceLocation(ambulanceID string) *LocationUpdate {
	h.cacheMu.RLock()
	defer h.cacheMu.RUnlock()
	return h.LocationCache[ambulanceID]
}
