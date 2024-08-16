package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Accepting all requests
	},
}

type Server struct {
	requests      []Request
	clients       map[*websocket.Conn]bool
	handleMessage func(message []byte) // New message handler
}

func StartServer(handleMessage func(message []byte)) *Server {
	server := Server{
		requests:      make([]Request, 100),
		clients:       make(map[*websocket.Conn]bool),
		handleMessage: handleMessage,
	}

	r := mux.NewRouter()
	r.HandleFunc("/post", server.handlePost).Methods("POST")
	r.HandleFunc("/get", server.echo).Methods("GET")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(index))
	})

	go http.ListenAndServe(":8081", r)

	return &server
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	request, err := NewRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.AddRequest(*request)

	if b, err := json.Marshal(request); err == nil {
		s.WriteMessage(b)
	}

	w.Write([]byte("OK"))
}

func (s *Server) AddRequest(request Request) {
	s.requests = append(s.requests, request)
}

func (server *Server) echo(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error while upgrading the connection: %v", err)
		http.Error(w, "Error while upgrading the connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	server.clients[connection] = true // Save the connection using it as a key

	server.WriteMessage([]byte("New client connected"))
	for {
		mt, message, err := connection.ReadMessage()

		if err != nil || mt == websocket.CloseMessage {
			break // Exit the loop if the client tries to close the connection or the connection is interrupted
		}

		go server.handleMessage(message)
	}

	delete(server.clients, connection) // Removing the connection

	connection.Close()
}

func (server *Server) WriteMessage(message []byte) {
	for conn := range server.clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}

type Request struct {
	Path   string      `json:"path"`
	Header http.Header `json:"header"`
	Body   []byte      `json:"body"`
}

func NewRequest(r *http.Request) (request *Request, err error) {
	request = new(Request)
	request.Header = r.Header
	request.Path = r.URL.Path
	request.Body, err = io.ReadAll(r.Body)
	if err != nil {
		return
	}

	return
}

func main() {
	server := StartServer(func(message []byte) {
		println(string(message))
	})
	_ = server
	select {}
}

var index = `
<script>
  const socket = new WebSocket('ws://localhost:8081/get');
  socket.addEventListener('open', function (event) {
	socket.send('Hello Server!');
  });
  
  socket.addEventListener('message', function (event) {
	console.log('Message from server ', event.data);
});
</script>
`
