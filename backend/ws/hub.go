package ws

import (
	"encoding/json"
	"sync"
)

// Hub maintains a set of clients for a project and broadcasts messages.
type Hub struct {
	projectID  uint
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

// GlobalHubs maps projectID -> Hub. Access must be guarded by hubsMu.
var (
	globalHubs = make(map[uint]*Hub)
	hubsMu     sync.Mutex
)

// GetOrCreateHub returns the hub for a project, creating it if needed.
func GetOrCreateHub(projectID uint) *Hub {
	hubsMu.Lock()
	defer hubsMu.Unlock()

	if h, ok := globalHubs[projectID]; ok {
		return h
	}
	h := &Hub{
		projectID:  projectID,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
		broadcast:  make(chan []byte, 256),
	}
	globalHubs[projectID] = h
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.broadcastPresence(client, true)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.broadcastPresence(client, false)
			h.cleanupIfEmpty()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// slow client — drop and disconnect
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) cleanupIfEmpty() {
	h.mu.RLock()
	empty := len(h.clients) == 0
	h.mu.RUnlock()
	if empty {
		hubsMu.Lock()
		delete(globalHubs, h.projectID)
		hubsMu.Unlock()
	}
}

func (h *Hub) broadcastPresence(client *Client, joined bool) {
	msgType := TypePresenceLeft
	var payload interface{}

	if joined {
		msgType = TypePresenceJoined
		payload = PresenceUser{
			ID:          client.userID,
			Username:    client.username,
			DisplayName: client.displayName,
			AvatarURL:   client.avatarURL,
		}
		// Also send presence list to the new client
		h.sendPresenceList(client)
	} else {
		payload = map[string]uint{"user_id": client.userID}
	}

	msg := Message{Type: msgType, Payload: payload}
	data, _ := json.Marshal(msg)
	h.broadcast <- data
}

func (h *Hub) sendPresenceList(target *Client) {
	h.mu.RLock()
	users := make([]PresenceUser, 0, len(h.clients))
	for c := range h.clients {
		users = append(users, PresenceUser{
			ID:          c.userID,
			Username:    c.username,
			DisplayName: c.displayName,
			AvatarURL:   c.avatarURL,
		})
	}
	h.mu.RUnlock()

	msg := Message{Type: TypePresenceList, Payload: map[string]interface{}{"users": users}}
	data, _ := json.Marshal(msg)
	select {
	case target.send <- data:
	default:
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Broadcast sends a message to all clients in this hub.
func (h *Hub) Broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.broadcast <- data
}

// BroadcastToProject sends a message to all clients of a project.
// In memory mode, delivery is direct. In Redis mode, the message is
// published to the shared channel and delivered via the subscriber.
func BroadcastToProject(projectID uint, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	if globalPubSub.IsLocal() {
		localBroadcastRaw(projectID, data)
		return
	}

	env, _ := json.Marshal(broadcastEnvelope{ProjectID: projectID, Data: json.RawMessage(data)})
	if err := globalPubSub.Publish(broadcastChannel, env); err != nil {
		// Fallback to local delivery if Redis is unavailable
		localBroadcastRaw(projectID, data)
	}
}

// localBroadcastRaw delivers a pre-serialised message to the local hub, if present.
func localBroadcastRaw(projectID uint, data []byte) {
	hubsMu.Lock()
	h, ok := globalHubs[projectID]
	hubsMu.Unlock()
	if ok {
		h.broadcast <- data
	}
}

// BroadcastToUser sends a message to all WebSocket connections belonging to a specific user.
func BroadcastToUser(userID uint, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	hubsMu.Lock()
	hubs := make([]*Hub, 0, len(globalHubs))
	for _, h := range globalHubs {
		hubs = append(hubs, h)
	}
	hubsMu.Unlock()
	for _, h := range hubs {
		h.mu.RLock()
		for c := range h.clients {
			if c.userID == userID {
				select {
				case c.send <- data:
				default:
				}
			}
		}
		h.mu.RUnlock()
	}
}

// IsUserOnline reports whether the user has at least one active WebSocket connection.
func IsUserOnline(userID uint) bool {
	hubsMu.Lock()
	hubs := make([]*Hub, 0, len(globalHubs))
	for _, h := range globalHubs {
		hubs = append(hubs, h)
	}
	hubsMu.Unlock()
	for _, h := range hubs {
		h.mu.RLock()
		for c := range h.clients {
			if c.userID == userID {
				h.mu.RUnlock()
				return true
			}
		}
		h.mu.RUnlock()
	}
	return false
}

// GetAllOnlineUsers returns a deduplicated list of all currently connected users across all hubs.
func GetAllOnlineUsers() []PresenceUser {
	hubsMu.Lock()
	hubs := make([]*Hub, 0, len(globalHubs))
	for _, h := range globalHubs {
		hubs = append(hubs, h)
	}
	hubsMu.Unlock()

	seen := make(map[uint]bool)
	var users []PresenceUser
	for _, h := range hubs {
		h.mu.RLock()
		for c := range h.clients {
			if !seen[c.userID] {
				seen[c.userID] = true
				users = append(users, PresenceUser{
					ID:          c.userID,
					Username:    c.username,
					DisplayName: c.displayName,
					AvatarURL:   c.avatarURL,
				})
			}
		}
		h.mu.RUnlock()
	}
	return users
}
