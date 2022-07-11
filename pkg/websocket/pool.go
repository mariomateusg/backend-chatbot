package websocket

import (
	"backend-chatbot/pkg/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
	}
}

func (pool *Pool) Start() {

	for {
		select {
		case client := <-pool.Register:

			resp, err := http.Get(os.Getenv("BOT_API_URL_GREETINGS"))
			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			bodyS := string(body)

			bodyS = "Chatbot: " + bodyS

			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				fmt.Println(client)
				//client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
				client.Conn.WriteJSON(Message{Type: 1, Body: bodyS})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			}
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")

			queryStruct := &models.Query{Query: message.Body}

			jsonStr, err := json.Marshal(queryStruct)

			if err != nil {
				fmt.Println(err)
			}

			resp, err := http.Post(os.Getenv("BOT_API_URL_REPLY"), "application/json", bytes.NewBuffer(jsonStr))

			var res models.Reply

			json.NewDecoder(resp.Body).Decode(&res)

			for client, _ := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}

				client.Conn.WriteJSON(Message{Type: 1, Body: "Chatbot: " + res.Result})

			}
		}
	}
}
