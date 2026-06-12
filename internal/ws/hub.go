package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Aceita apenas conexões da própria origem ou do frontend configurado
	// (APP_URL). Requisições sem Origin (clientes não-browser) passam — a
	// autenticação por cookie/JWT continua obrigatória no handler.
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		if strings.EqualFold(u.Host, r.Host) {
			return true
		}
		if app := os.Getenv("APP_URL"); app != "" {
			if au, err := url.Parse(app); err == nil && strings.EqualFold(u.Host, au.Host) {
				return true
			}
		}
		return false
	},
}

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type client struct {
	albumID int64
	conn    *websocket.Conn
	send    chan []byte
	hub     *Hub
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*client]bool
}

var global = &Hub{rooms: make(map[int64]map[*client]bool)}

func GetHub() *Hub { return global }

func (h *Hub) register(c *client) {
	h.mu.Lock()
	if h.rooms[c.albumID] == nil {
		h.rooms[c.albumID] = make(map[*client]bool)
	}
	h.rooms[c.albumID][c] = true
	h.mu.Unlock()
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	if room, ok := h.rooms[c.albumID]; ok {
		if _, there := room[c]; there {
			delete(room, c)
			close(c.send)
		}
		if len(room) == 0 {
			delete(h.rooms, c.albumID)
		}
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(albumID int64, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	clients := h.rooms[albumID]
	h.mu.RUnlock()
	for c := range clients {
		select {
		case c.send <- data:
		default:
		}
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, albumID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[WS] upgrade error:", err)
		return
	}
	c := &client{
		albumID: albumID,
		conn:    conn,
		send:    make(chan []byte, 64),
		hub:     h,
	}
	h.register(c)
	go c.writePump()
	go c.readPump()
}

func (c *client) readPump() {
	defer func() {
		c.hub.unregister(c)
		c.conn.Close()
	}()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}
