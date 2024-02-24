package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demonstration purposes
	},
}

// ConnPool manages a pool of WebSocket connections.
type ConnPool struct {
	mu    sync.Mutex // Guards access to the conns slice.
	conns []*websocket.Conn
}

// Add adds a new WebSocket connection to the pool.
func (p *ConnPool) Add(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conns = append(p.conns, conn)
}

// Remove removes a WebSocket connection from the pool.
func (p *ConnPool) Remove(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, c := range p.conns {
		if c == conn {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			return
		}
	}
}

// Get returns a WebSocket connection from the pool.
func (p *ConnPool) Get() *websocket.Conn {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.conns) == 0 {
		return nil
	}
	// Simple round-robin for demonstration. Improve this for production use.
	conn := p.conns[0]
	p.conns = append(p.conns[1:], conn)
	return conn
}

var pool = ConnPool{}

func handleWebSocket(c *gin.Context) {
	w := c.Writer
	r := c.Request

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to websocket:", err)
		return
	}
	defer conn.Close()

	pool.Add(conn)
	defer pool.Remove(conn)

	for {
		// Read message from WebSocket
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		// Echo the message back to WebSocket (for demonstration)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	r := gin.Default()

	type HTTPRequest struct {
		Method  string              `json:"method"`
		Path    string              `json:"path"`
		Headers map[string][]string `json:"headers"`
		Body    string              `json:"body"`
	}

	r.POST("/send_application/", func(c *gin.Context) {
		conn := pool.Get()
		if conn == nil {
			c.String(http.StatusServiceUnavailable, "No WebSocket connections available")
			return
		}

		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Println("error reading request body:", err)
			c.String(http.StatusInternalServerError, "Failed to read request body")
			return
		}

		httpRequest := HTTPRequest{
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			Headers: c.Request.Header,
			Body:    string(bodyBytes),
		}

		requestJson, err := json.Marshal(httpRequest)
		if err != nil {
			log.Println("error marshalling request to JSON:", err)
			c.String(http.StatusInternalServerError, "Failed to serialize request")
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, requestJson)
		if err != nil {
			log.Println("error forwarding message:", err)
			c.String(http.StatusInternalServerError, "Failed to forward message")
			return
		}

		// Wait for a response from the FastAPI client
		_, responseJson, err := conn.ReadMessage()
		if err != nil {
			log.Println("error reading response message:", err)
			c.String(http.StatusInternalServerError, "Failed to read response message")
			return
		}

		c.Data(http.StatusOK, "application/json", responseJson)
	})
	// WebSocket handler
	r.GET("/ws", handleWebSocket)

	// Start the server on port 8080
	r.Run(":8080")
}
