package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		for {
			// Read message from browser
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Print the message to the console
			fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			send, er := CtrlMessage(msg) // ([]byte, error)
			if er != nil {
				fmt.Println("error :", er)
				return
			}

			// fmt.Println(string(send))
			if err = conn.WriteMessage(msgType, send); err != nil {
				return
			}

			// fmt.Println(msgType)
			// Write message back to browser
			// if err = conn.WriteMessage(msgType, msg); err != nil {
			// 	return
			// }
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./html/websockets.html")
	})

	http.ListenAndServe(":8080", nil)
}

type TestSample struct {
	Name  string
	Key   string
	Value string
}

func CtrlMessage(msg []byte) ([]byte, error) {

	var send []byte
	var packagedata map[string]interface{}
	err := json.Unmarshal(msg, &packagedata)
	if err != nil {
		fmt.Println("번역이 안되나보지")
		return nil, err
	}

	strCmd := packagedata["Command"]
	switch strCmd {
	case "Login":
		var tmp TestSample
		tmp.Name = "help"
		tmp.Key = "your"
		tmp.Value = "self"
		resultData, er := json.Marshal(&tmp)
		if er != nil {
			// 내용 정리 !!
			fmt.Println("json marshal error ")
			return nil, er
		}
		send = resultData
	}

	return send, nil
}
