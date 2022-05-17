package main

import "woungbe.addstruct.websocket/amwebsocket"

func main() {

	wss := amwebsocket.GetAMWSvr()
	wss.Init()
	/* Init 후 세팅해줘야 적용됨 */
	wss.SetDebugMode(true)
	wss.SetHandlePath("status")
	wss.SetSSL("N", "", "")
	wss.SetPort("9183")

	wss.StartWebSocketServer()

}
