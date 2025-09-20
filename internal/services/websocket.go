package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"ex4-oauth2/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketService manages real-time WebSocket connections and notifications
type WebSocketService struct {
	clients    map[string]*ClientConnection
	channels   map[models.NotificationChannel]map[string]*ClientConnection
	broadcast  chan *models.Notification
	register   chan *ClientConnection
	unregister chan *ClientConnection
	mutex      sync.RWMutex

	notificationRepo models.NotificationRepository
	auditService     *AuditService
}

// ClientConnection represents a WebSocket client connection
type ClientConnection struct {
	ID          string
	UserID      *uint
	Role        string
	IP          string
	UserAgent   string
	Conn        *websocket.Conn
	Send        chan *models.WebSocketMessage
	Channels    map[models.NotificationChannel]bool
	ConnectedAt time.Time
	LastPingAt  time.Time
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(notificationRepo models.NotificationRepository, auditService *AuditService) *WebSocketService {
	return &WebSocketService{
		clients:          make(map[string]*ClientConnection),
		channels:         make(map[models.NotificationChannel]map[string]*ClientConnection),
		broadcast:        make(chan *models.Notification, 256),
		register:         make(chan *ClientConnection),
		unregister:       make(chan *ClientConnection),
		notificationRepo: notificationRepo,
		auditService:     auditService,
	}
}

// Start begins the WebSocket service hub
func (ws *WebSocketService) Start() {
	go ws.run()

	// Start periodic cleanup
	go ws.startCleanupRoutine()
}

// run manages the WebSocket hub
func (ws *WebSocketService) run() {
	for {
		select {
		case client := <-ws.register:
			ws.registerClient(client)

		case client := <-ws.unregister:
			ws.unregisterClient(client)

		case notification := <-ws.broadcast:
			ws.broadcastNotification(notification)
		}
	}
}

// registerClient registers a new WebSocket client
func (ws *WebSocketService) registerClient(client *ClientConnection) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	ws.clients[client.ID] = client

	// Auto-subscribe based on user role
	ws.autoSubscribeChannels(client)

	// Send welcome message
	welcome := &models.WebSocketMessage{
		Type: "welcome",
		Data: map[string]interface{}{
			"client_id": client.ID,
			"channels":  ws.getClientChannels(client),
		},
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}

	select {
	case client.Send <- welcome:
	default:
		close(client.Send)
		delete(ws.clients, client.ID)
	}

	// Log connection
	if ws.auditService != nil {
		ctx := &AuditContext{
			UserID:    client.UserID,
			IPAddress: client.IP,
			UserAgent: client.UserAgent,
		}
		ws.auditService.LogAction(ctx, "websocket_connect", "connection", "connected", map[string]interface{}{
			"client_id": client.ID,
			"role":      client.Role,
		}, nil)
	}

	log.Printf("WebSocket client connected: %s (User: %v, Role: %s)",
		client.ID, client.UserID, client.Role)
}

// unregisterClient removes a WebSocket client
func (ws *WebSocketService) unregisterClient(client *ClientConnection) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if _, ok := ws.clients[client.ID]; ok {
		// Remove from channels
		for channel := range client.Channels {
			if clients, exists := ws.channels[channel]; exists {
				delete(clients, client.ID)
				if len(clients) == 0 {
					delete(ws.channels, channel)
				}
			}
		}

		// Remove client
		delete(ws.clients, client.ID)
		close(client.Send)

		// Log disconnection
		if ws.auditService != nil {
			ctx := &AuditContext{
				UserID:    client.UserID,
				IPAddress: client.IP,
				UserAgent: client.UserAgent,
			}
			ws.auditService.LogAction(ctx, "websocket_disconnect", "connection", "disconnected", map[string]interface{}{
				"client_id": client.ID,
				"duration":  time.Since(client.ConnectedAt).String(),
			}, nil)
		}

		log.Printf("WebSocket client disconnected: %s", client.ID)
	}
}

// autoSubscribeChannels automatically subscribes client to appropriate channels
func (ws *WebSocketService) autoSubscribeChannels(client *ClientConnection) {
	// Initialize channels if not exists
	if ws.channels[models.ChannelSystem] == nil {
		ws.channels[models.ChannelSystem] = make(map[string]*ClientConnection)
	}
	if ws.channels[models.ChannelUser] == nil {
		ws.channels[models.ChannelUser] = make(map[string]*ClientConnection)
	}
	if ws.channels[models.ChannelSecurity] == nil {
		ws.channels[models.ChannelSecurity] = make(map[string]*ClientConnection)
	}
	if ws.channels[models.ChannelAdmin] == nil {
		ws.channels[models.ChannelAdmin] = make(map[string]*ClientConnection)
	}
	if ws.channels[models.ChannelCompliance] == nil {
		ws.channels[models.ChannelCompliance] = make(map[string]*ClientConnection)
	}
	if ws.channels[models.ChannelAudit] == nil {
		ws.channels[models.ChannelAudit] = make(map[string]*ClientConnection)
	}

	// All users get system notifications
	ws.channels[models.ChannelSystem][client.ID] = client
	client.Channels[models.ChannelSystem] = true

	// Subscribe based on role
	switch client.Role {
	case "admin":
		// Admins get all channels
		for channel := range ws.channels {
			ws.channels[channel][client.ID] = client
			client.Channels[channel] = true
		}
	case "moderator":
		// Moderators get user and security channels
		ws.channels[models.ChannelUser][client.ID] = client
		ws.channels[models.ChannelSecurity][client.ID] = client
		client.Channels[models.ChannelUser] = true
		client.Channels[models.ChannelSecurity] = true
	default:
		// Regular users get user channel
		ws.channels[models.ChannelUser][client.ID] = client
		client.Channels[models.ChannelUser] = true
	}
}

// getClientChannels returns list of channels client is subscribed to
func (ws *WebSocketService) getClientChannels(client *ClientConnection) []string {
	channels := make([]string, 0, len(client.Channels))
	for channel := range client.Channels {
		channels = append(channels, string(channel))
	}
	return channels
}

// broadcastNotification sends notification to appropriate clients
func (ws *WebSocketService) broadcastNotification(notification *models.Notification) {
	message := &models.WebSocketMessage{
		Type:      "notification",
		Data:      notification,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}

	// Determine target channel
	channel := ws.getNotificationChannel(notification)

	ws.mutex.RLock()
	clients := ws.channels[channel]
	ws.mutex.RUnlock()

	if clients == nil {
		return
	}

	// Send to appropriate clients
	for _, client := range clients {
		// Check if notification is for specific user
		if notification.UserID != nil && client.UserID != nil {
			if *notification.UserID != *client.UserID {
				continue
			}
		}

		// Check role restrictions
		if notification.RecipientRole != "" && client.Role != notification.RecipientRole {
			continue
		}

		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, disconnect
			ws.unregister <- client
		}
	}
}

// getNotificationChannel determines appropriate channel for notification
func (ws *WebSocketService) getNotificationChannel(notification *models.Notification) models.NotificationChannel {
	switch notification.Type {
	case models.NotificationSecurityAlert,
		models.NotificationLoginAttempt,
		models.NotificationSuspiciousLogin,
		models.NotificationAccountLocked:
		return models.ChannelSecurity

	case models.NotificationUserRegistered,
		models.NotificationUserStatusChange,
		models.NotificationSystemAlert:
		return models.ChannelAdmin

	case models.NotificationComplianceAlert:
		return models.ChannelCompliance

	case models.NotificationProfileUpdate,
		models.NotificationPasswordChange,
		models.NotificationEmailVerified,
		models.NotificationTokenExpiry:
		return models.ChannelUser

	default:
		return models.ChannelSystem
	}
}

// SendNotification sends a notification to connected clients
func (ws *WebSocketService) SendNotification(notification *models.Notification) {
	// Set ID and timestamp if not set
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	// Send to broadcast channel
	select {
	case ws.broadcast <- notification:
	default:
		log.Printf("Broadcast channel full, dropping notification: %s", notification.ID)
	}

	// Persist notification if it's for a specific user or important
	if notification.UserID != nil || notification.Priority == models.PriorityCritical {
		ws.persistNotification(notification)
	}
}

// persistNotification saves notification to database
func (ws *WebSocketService) persistNotification(notification *models.Notification) {
	if ws.notificationRepo == nil {
		return
	}

	dataJSON := ""
	if notification.Data != nil {
		if jsonData, err := json.Marshal(notification.Data); err == nil {
			dataJSON = string(jsonData)
		}
	}

	persistent := &models.PersistentNotification{
		Type:          notification.Type,
		Priority:      notification.Priority,
		Title:         notification.Title,
		Message:       notification.Message,
		DataJSON:      dataJSON,
		UserID:        notification.UserID,
		RecipientRole: notification.RecipientRole,
		SourceIP:      notification.SourceIP,
		UserAgent:     notification.UserAgent,
		SessionID:     notification.SessionID,
		CreatedAt:     notification.CreatedAt,
		ExpiresAt:     notification.ExpiresAt,
	}

	if err := ws.notificationRepo.CreateNotification(persistent); err != nil {
		log.Printf("Failed to persist notification: %v", err)
	}
}

// GetStats returns WebSocket service statistics
func (ws *WebSocketService) GetStats() map[string]interface{} {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	channelStats := make(map[string]int)
	for channel, clients := range ws.channels {
		channelStats[string(channel)] = len(clients)
	}

	userMap := make(map[uint]bool)
	for _, client := range ws.clients {
		if client.UserID != nil {
			userMap[*client.UserID] = true
		}
	}

	return map[string]interface{}{
		"total_connections":     len(ws.clients),
		"unique_users":          len(userMap),
		"channel_subscriptions": channelStats,
		"uptime":                time.Since(time.Now()).String(),
	}
}

// startCleanupRoutine starts periodic cleanup of stale connections
func (ws *WebSocketService) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ws.cleanupStaleConnections()
		if ws.notificationRepo != nil {
			ws.notificationRepo.DeleteExpiredNotifications()
		}
	}
}

// cleanupStaleConnections removes inactive connections
func (ws *WebSocketService) cleanupStaleConnections() {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	staleThreshold := time.Now().Add(-10 * time.Minute)

	for id, client := range ws.clients {
		if client.LastPingAt.Before(staleThreshold) {
			log.Printf("Removing stale connection: %s", id)
			ws.unregister <- client
		}
	}
}

// HandleConnection handles new WebSocket connection
func (ws *WebSocketService) HandleConnection(c *gin.Context, userID *uint, role string) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %v", err)
	}

	client := &ClientConnection{
		ID:          uuid.New().String(),
		UserID:      userID,
		Role:        role,
		IP:          c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
		Conn:        conn,
		Send:        make(chan *models.WebSocketMessage, 256),
		Channels:    make(map[models.NotificationChannel]bool),
		ConnectedAt: time.Now(),
		LastPingAt:  time.Now(),
	}

	// Register client
	ws.register <- client

	// Start goroutines for reading and writing
	go ws.writePump(client)
	go ws.readPump(client)

	return nil
}

// writePump handles writing messages to WebSocket
func (ws *WebSocketService) writePump(client *ClientConnection) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump handles reading messages from WebSocket
func (ws *WebSocketService) readPump(client *ClientConnection) {
	defer func() {
		ws.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.LastPingAt = time.Now()
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message models.WebSocketMessage
		err := client.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle client messages (subscribe/unsubscribe, etc.)
		ws.handleClientMessage(client, &message)
	}
}

// handleClientMessage processes messages from clients
func (ws *WebSocketService) handleClientMessage(client *ClientConnection, message *models.WebSocketMessage) {
	switch message.Type {
	case "ping":
		client.LastPingAt = time.Now()
		pong := &models.WebSocketMessage{
			Type:      "pong",
			Timestamp: time.Now(),
			MessageID: uuid.New().String(),
		}
		select {
		case client.Send <- pong:
		default:
		}

	case "subscribe":
		if channel, ok := message.Data.(string); ok {
			ws.subscribeToChannel(client, models.NotificationChannel(channel))
		}

	case "unsubscribe":
		if channel, ok := message.Data.(string); ok {
			ws.unsubscribeFromChannel(client, models.NotificationChannel(channel))
		}
	}
}

// subscribeToChannel subscribes client to a notification channel
func (ws *WebSocketService) subscribeToChannel(client *ClientConnection, channel models.NotificationChannel) {
	// Check if user has permission for this channel
	if !ws.canAccessChannel(client, channel) {
		return
	}

	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if ws.channels[channel] == nil {
		ws.channels[channel] = make(map[string]*ClientConnection)
	}

	ws.channels[channel][client.ID] = client
	client.Channels[channel] = true
}

// unsubscribeFromChannel unsubscribes client from a notification channel
func (ws *WebSocketService) unsubscribeFromChannel(client *ClientConnection, channel models.NotificationChannel) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if clients, exists := ws.channels[channel]; exists {
		delete(clients, client.ID)
		if len(clients) == 0 {
			delete(ws.channels, channel)
		}
	}

	delete(client.Channels, channel)
}

// canAccessChannel checks if client can access a specific channel
func (ws *WebSocketService) canAccessChannel(client *ClientConnection, channel models.NotificationChannel) bool {
	switch channel {
	case models.ChannelAdmin, models.ChannelCompliance, models.ChannelAudit:
		return client.Role == "admin"
	case models.ChannelSecurity:
		return client.Role == "admin" || client.Role == "moderator"
	case models.ChannelUser, models.ChannelSystem:
		return true
	default:
		return false
	}
}
