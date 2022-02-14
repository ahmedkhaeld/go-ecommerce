package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

// WebSocketConnection to communicate between the front end, need to send information off to the connected client
type WebSocketConnection struct {
	*websocket.Conn
}

// WsPayload defines the data that we are receiving from the client
type WsPayload struct {
	Action      string              `json:"action"`
	Message     string              `json:"message"`
	UserName    string              `json:"username"`
	MessageType string              `json:"message_type"`
	UserID      int                 `json:"user_id"`
	Conn        WebSocketConnection `json:"-"`
}

// WsJsonResponse what to send to the end-user
type WsJsonResponse struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

//clients to keep track of every client who is connected
var clients = make(map[WebSocketConnection]string)

//wsChan push to the channel any time receiving information
var wsChan = make(chan WsPayload)

func (app *application) WsEndPoint(w http.ResponseWriter, r *http.Request) {
	// 1. upgrade the connection, when a request comes from the front-end
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(fmt.Sprintf("Client connected from %s", r.RemoteAddr))
	var response WsJsonResponse
	response.Message = "Connected to server"

	// write json using gorilla websocket to write exactly the format that websocket requires
	err = ws.WriteJSON(response)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// get the connection after connected
	conn := WebSocketConnection{Conn: ws}

	// put the connection in the clients map
	clients[conn] = ""

	// run go routine in the background to listen to the websocket connection
	go app.ListenForWS(&conn)
}

func (app *application) ListenForWS(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			app.errorLog.Println("Error:", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func (app *application) ListenToWsChannel() {
	var response WsJsonResponse

	for {
		e := <-wsChan
		switch e.Action {
		case "deleteUser":
			response.Action = "logout"
			response.Message = "Your account has been deleted"
			response.UserID = e.UserID
			app.broadcastToAll(response)
		default:

		}
	}
}

func (app *application) broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		// broadcast to every connected client
		err := client.WriteJSON(response)
		if err != nil {
			app.errorLog.Printf("Websocket err on %s: %s", response.Action, err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
