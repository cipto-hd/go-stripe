package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	*websocket.Conn
}

type WsPayload struct {
	Action      string              `json:"action"`
	Message     string              `json:"message"`
	UserName    string              `json:"username"`
	MessageType string              `json:"message_type"`
	UserID      int                 `json:"user_id"`
	ActorID     int                 `json:"auth_id"`
	Conn        WebSocketConnection `json:"-"`
}

type WsJsonResponse struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
	ActorID int    `json:"auth_id"`
}

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var clients = make(map[WebSocketConnection]string)

var wsChan = make(chan WsPayload)

func (app *application) WsEndPoint(w http.ResponseWriter, r *http.Request) {
	/* unwrap ResponseWriter to support Hijacker interface for scs v2@v2.7.0 */
	// wupgrade := w
	// if u, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
	// 	wupgrade = u.Unwrap()
	// }

	app.infoLog.Printf("w's type is %T\n", w)

	ws, err := upgradeConnection.Upgrade(w, r, nil)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Printf("Client connected from %s", r.RemoteAddr)
	var response WsJsonResponse
	response.Message = "Connected to server"

	err = ws.WriteJSON(response)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	go app.ListenForWS(&conn)
}

func (app *application) ListenForWS(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			app.errorLog.Println("ERORR:", fmt.Sprintf("%v", r))
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
		// app.infoLog.Printf("%+v", e)

		switch e.Action {
		case "deleteUser":
			response.Action = "logout"
			response.Message = "Your account has been deleted"
			response.UserID = e.UserID
			response.ActorID = e.ActorID
			app.broadcastToAll(response)
		}
	}
}

func (app *application) broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		// broadcast to every connected client
		err := client.WriteJSON(response)
		if err != nil {
			app.errorLog.Printf("Websocket err on %s: %v", response.Action, err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
