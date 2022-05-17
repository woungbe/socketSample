package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"woungbe.client.websocket/wsclient"
)

func main() {
	InitService()

	done := make(chan bool)
	r := bufio.NewReader(os.Stdin)
	go func() {
		for {
			line, err := r.ReadString('\n')
			if err != nil && err.Error() != "unexpected newline" {
				log.Println(err.Error())
				//	return
				line = ""
			}

			line = strings.TrimSpace(line)

			CMDPaser(line)
		}
	}()
	<-done

}

func InitService() {
	wsclient.GetAMStatusWsConnector().Init()
	wsclient.AutoConnect()
}

func CMDPaser(strCMD string) {
	defer func() {
		if err := recover(); err != nil {

		}
	}()
	if strCMD == "" {
		return
	}
	if strCMD == "exit" {
		os.Exit(1)
		return
	}

	if strCMD == "send" {
		// c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		var sndPk wsclient.WSRequest
		sndPk.Command = "Login"
		jsondata, er := json.Marshal(&sndPk)
		if er != nil {
			log.Println("111111111111111111111111")
			return
		}
		fmt.Println(string(jsondata))
		wsclient.SendPacket(jsondata)
	}

	if strCMD == "error clear" {
		fmt.Println("시스템 오류 정보 클리어됌")
	}
	if strCMD == "view config" {
		fmt.Println("====================================================================")
		fmt.Println("====================================================================")
	}

}
