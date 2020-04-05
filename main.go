package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
var upgrader = websocket.Upgrader{}

type Message struct {
	Username string `json:"username"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Side string `json:"side"`
	Active bool `json:"active"`
}

func main()  {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/",fs)
	
	http.HandleFunc("/ws",handleConnection)
	
	go handleMessages()
	
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
			log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnection(w http.ResponseWriter, r *http.Request)  {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	log.Print(ws)

	if err != nil {
		log.Fatal(err)
	}
	clients[ws] = true

	defer ws.Close()
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
				log.Printf("error: %v", err)
				delete(clients, ws)
				break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
				err := client.WriteJSON(msg)
				if err != nil {
						log.Printf("error: %v", err)
						client.Close()
						delete(clients, client)
				}
		}
	}
}