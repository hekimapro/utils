package socket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hekimapro/utils/log"
	"github.com/hekimapro/utils/snowflake"
)

// upgrader is configured to upgrade HTTP connections to WebSocket connections
// We allow all origins for development - in production you should restrict this
var upgrader = websocket.Upgrader{
	CheckOrigin: func(request *http.Request) bool {
		// Allow all origins for now (configure properly in production)
		return true
	},
}

// ConnectedClient represents a single WebSocket client connection
type ConnectedClient struct {
	ID           int64
	Connection   *websocket.Conn
	SendMessage  chan []byte
	EntityFilter map[string]bool
	mutex        sync.RWMutex
}

// SocketManager manages all WebSocket connections and broadcasts events
type SocketManager struct {
	// clients maps a client to a boolean (true if connected)
	// Using map as a set to track all connected clients
	clients map[*ConnectedClient]bool

	// broadcastChannel receives events that need to be sent to all clients
	broadcastChannel chan *DataChangeEvent

	// registerChannel receives new clients that want to connect
	registerChannel chan *ConnectedClient

	// unregisterChannel receives clients that are disconnecting
	unregisterChannel chan *ConnectedClient

	// mutex protects concurrent access to the clients map
	mutex sync.RWMutex
}

// NewSocketManager creates and initializes a new WebSocket manager
func NewSocketManager() *SocketManager {
	return &SocketManager{
		// Initialize all channels with buffers to prevent blocking
		clients:           make(map[*ConnectedClient]bool),
		broadcastChannel:  make(chan *DataChangeEvent, 100), // Buffer 100 events
		registerChannel:   make(chan *ConnectedClient),
		unregisterChannel: make(chan *ConnectedClient),
	}
}

// Run starts the WebSocket manager's main event loop
// This should be run as a goroutine
func (manager *SocketManager) Run() {
	log.Info("🚀 WebSocket manager started successfully")

	// Continuously process events from channels
	for {
		select {
		// Handle new client registration
		case newClient := <-manager.registerChannel:
			manager.mutex.Lock() // Lock to safely modify the clients map
			manager.clients[newClient] = true
			clientCount := len(manager.clients)
			manager.mutex.Unlock() // Unlock after modification

			log.Successf("✅ New client registered | ClientID: %d | Total connections: %d",
				newClient.ID, clientCount)

		// Handle client unregistration
		case disconnectingClient := <-manager.unregisterChannel:
			manager.mutex.Lock() // Lock to safely modify the clients map

			// Check if client exists in our map
			if _, exists := manager.clients[disconnectingClient]; exists {
				// Remove client from map
				delete(manager.clients, disconnectingClient)
				clientCount := len(manager.clients)

				// Close the client's send channel to signal goroutines to stop
				close(disconnectingClient.SendMessage)

				log.Infof("📤 Client disconnected | ClientID: %d | Remaining connections: %d",
					disconnectingClient.ID, clientCount)
			}
			manager.mutex.Unlock() // Unlock after modification

		// Handle broadcast events
		case event := <-manager.broadcastChannel:
			// When we receive an event to broadcast, send it to all connected clients
			manager.broadcastEventToClients(event)
		}
	}
}

// broadcastEventToClients sends an event to all connected clients
// This is called internally when an event is received on the broadcast channel
func (manager *SocketManager) broadcastEventToClients(event *DataChangeEvent) {
	// Convert the event to JSON once for efficiency
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Errorf("❌ Failed to marshal event | Type: %s | Entity: %s | Error: %v",
			event.Type, event.Entity, err)
		return
	}

	// Lock for reading the clients map (multiple readers allowed)
	manager.mutex.RLock()
	clientCount := len(manager.clients)
	// Ensure we unlock when done (using defer ensures it runs even if panic occurs)
	defer manager.mutex.RUnlock()

	if clientCount == 0 {
		log.Debugf("📭 No active clients to broadcast %s event for %s",
			event.Type, event.Entity)
		return
	}

	log.Debugf("📢 Broadcasting %s event for %s to %d clients",
		event.Type, event.Entity, clientCount)

	sentCount := 0
	filteredCount := 0

	// Iterate through all connected clients
	for connectedClient := range manager.clients {
		// Check if client has entity filters
		connectedClient.mutex.RLock()
		hasFilter := len(connectedClient.EntityFilter) > 0
		shouldReceive := true

		if hasFilter {
			// Client has filters, check if they want this entity type
			_, shouldReceive = connectedClient.EntityFilter[event.Entity]
		}
		connectedClient.mutex.RUnlock()

		// Skip clients that don't want this entity type
		if hasFilter && !shouldReceive {
			filteredCount++
			continue
		}

		// Try to send the event to the client
		select {
		case connectedClient.SendMessage <- eventJSON:
			sentCount++
		default:
			// Client's send buffer is full - they might be slow or disconnected
			// We should unregister them to prevent memory leaks
			log.Warningf("⚠️ Client %d buffer full - initiating unregistration", connectedClient.ID)

			// Run unregistration in a goroutine to avoid blocking
			go func(client *ConnectedClient) {
				manager.unregisterChannel <- client
				client.Connection.Close()
			}(connectedClient)
		}
	}

	if filteredCount > 0 {
		log.Debugf("🔇 Filtered out %d clients not subscribed to %s",
			filteredCount, event.Entity)
	}

	log.Successf("✅ Event broadcast complete | Type: %s | Entity: %s | Sent: %d | Total clients: %d",
		event.Type, event.Entity, sentCount, clientCount)
}

// HandleWebSocket upgrades HTTP connections to WebSocket and manages the client lifecycle
// This is the HTTP handler function for WebSocket connections
func (manager *SocketManager) HandleWebSocket(response http.ResponseWriter, request *http.Request) {
	// Log connection attempt
	clientIP := request.RemoteAddr
	log.Infof("📡 WebSocket connection attempt from %s", clientIP)

	// Upgrade the HTTP connection to a WebSocket connection
	websocketConnection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Errorf("❌ WebSocket upgrade failed for %s | Error: %v", clientIP, err)
		return
	}

	// Create a new client for this connection
	newClient := &ConnectedClient{
		ID:           snowflake.NextID(),
		Connection:   websocketConnection,
		SendMessage:  make(chan []byte, 256), // Buffer 256 messages
		EntityFilter: make(map[string]bool),
	}

	// Register the new client with the manager
	manager.registerChannel <- newClient

	log.Successf("✅ WebSocket connection established | ClientID: %d | IP: %s",
		newClient.ID, clientIP)

	// Start two goroutines for this client:
	// 1. readPump: handles incoming messages from the client
	// 2. writePump: handles outgoing messages to the client
	go manager.readPump(newClient)
	go manager.writePump(newClient)
}

// readPump continuously reads messages from the client
// This runs in its own goroutine per client
func (manager *SocketManager) readPump(client *ConnectedClient) {
	// Ensure cleanup when this function exits
	defer func() {
		// Unregister the client
		manager.unregisterChannel <- client
		// Close the connection
		client.Connection.Close()
		log.Debugf("🔚 Read pump terminated for client %d", client.ID)
	}()

	log.Debugf("👂 Read pump started for client %d", client.ID)

	// Continuously read messages from the client
	for {
		var message map[string]any

		// Read a JSON message from the client
		err := client.Connection.ReadJSON(&message)
		if err != nil {
			// Check if the error is a normal closure
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warningf("⚠️ Unexpected connection close for client %d | Error: %v",
					client.ID, err)
			} else {
				log.Debugf("ℹ️ Client %d connection closed normally", client.ID)
			}
			break // Exit the loop on any read error
		}

		// Process the message from client
		manager.handleClientMessage(client, message)
	}
}

// handleClientMessage processes messages sent by the client
// Clients can send messages to subscribe/unsubscribe from entity types
func (manager *SocketManager) handleClientMessage(client *ConnectedClient, message map[string]any) {
	// Check if the message has an "action" field
	action, actionExists := message["action"].(string)
	if !actionExists {
		log.Warningf("⚠️ Client %d sent malformed message (no action field): %v",
			client.ID, message)
		return
	}

	log.Debugf("📨 Client %d sent action: %s", client.ID, action)

	switch action {
	case "subscribe":
		// Client wants to subscribe to specific entity types
		entities, entitiesExist := message["entities"].([]any)
		if !entitiesExist {
			log.Warningf("⚠️ Client %d sent subscribe action without entities list", client.ID)
			return
		}

		// Lock to safely modify the client's filter
		client.mutex.Lock()
		defer client.mutex.Unlock()

		// Clear existing filters and set new ones
		oldFilterCount := len(client.EntityFilter)
		client.EntityFilter = make(map[string]bool)

		subscribedEntities := []string{}
		for _, entity := range entities {
			if entityString, ok := entity.(string); ok {
				client.EntityFilter[entityString] = true
				subscribedEntities = append(subscribedEntities, entityString)
			}
		}

		log.Successf("📋 Client %d subscription updated | Previous: %d entities | New: %v",
			client.ID, oldFilterCount, subscribedEntities)

	case "unsubscribe_all":
		// Client wants to unsubscribe from all entity types
		client.mutex.Lock()
		oldFilterCount := len(client.EntityFilter)
		client.EntityFilter = make(map[string]bool)
		client.mutex.Unlock()

		log.Infof("🔕 Client %d unsubscribed from all entities | Previous subscriptions: %d",
			client.ID, oldFilterCount)

	default:
		log.Warningf("⚠️ Client %d sent unknown action: %s", client.ID, action)
	}
}

// writePump continuously sends messages to the client
// This runs in its own goroutine per client
func (manager *SocketManager) writePump(client *ConnectedClient) {
	// Create a ticker for sending ping messages every 30 seconds
	pingTicker := time.NewTicker(30 * time.Second)

	// Ensure cleanup when this function exits
	defer func() {
		pingTicker.Stop()
		client.Connection.Close()
		log.Debugf("🔚 Write pump terminated for client %d", client.ID)
	}()

	log.Debugf("✍️ Write pump started for client %d", client.ID)

	// Continuously handle outgoing messages
	for {
		select {
		case message, channelOpen := <-client.SendMessage:
			// Check if the channel was closed
			if !channelOpen {
				// Channel closed, time to exit
				log.Debugf("📪 Send channel closed for client %d", client.ID)
				client.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Set a write deadline to prevent hanging
			client.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))

			// Write the message to the WebSocket connection
			err := client.Connection.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Errorf("❌ Failed to write message to client %d | Error: %v",
					client.ID, err)
				return
			}
			log.Debugf("📤 Message sent to client %d", client.ID)

		case <-pingTicker.C:
			// Time to send a ping to keep the connection alive
			client.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := client.Connection.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Warningf("⚠️ Failed to send ping to client %d | Error: %v",
					client.ID, err)
				return
			}
			log.Debugf("🏓 Ping sent to client %d", client.ID)
		}
	}
}

// BroadcastEvent broadcasts a data change event to all connected clients
// This is the public method that repositories will call
func (manager *SocketManager) BroadcastEvent(eventType EventType, entity string, data any) {
	// Create the event object
	event := &DataChangeEvent{
		Type:      eventType,
		Entity:    entity,
		Data:      data,
		Timestamp: time.Now(),
	}

	// Try to send the event to the broadcast channel
	select {
	case manager.broadcastChannel <- event:
		// Event sent successfully
		log.Infof("📨 Event queued for broadcast | Type: %s | Entity: %s",
			eventType, entity)
	default:
		// Broadcast channel is full - log this and drop the event
		// In production, you might want to handle this differently
		log.Warningf("⚠️ Broadcast channel full | Dropping %s event for %s | Channel capacity: 100",
			eventType, entity)
	}
}