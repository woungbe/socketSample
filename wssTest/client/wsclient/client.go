package wsclient

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type AMStatusWSConnector struct {
	con           *websocket.Conn
	send          chan []byte
	bBreakWrite   chan bool
	bConnected    bool //연결 상태
	bLogin        bool //서버 로그인상태
	bReconnecting bool //자동 재접속중인가
}

var AMStatusWSConnectorPTR *AMStatusWSConnector
var ConfigDataPTR *ConfigData // 전역 변수 컨트롤 하기

func GetAMStatusWsConnector() *AMStatusWSConnector {
	if AMStatusWSConnectorPTR == nil {
		AMStatusWSConnectorPTR = new(AMStatusWSConnector)
		ConfigDataPTR = new(ConfigData)
		ConfigDataPTR.Default()
	}
	return AMStatusWSConnectorPTR
}

// 서버 접속
func BinRestServerConnect() bool {
	return GetAMStatusWsConnector().clientConnect()
}

//IsConnected 서버와 연결상태 체크
func IsConnected() bool {
	bc := GetAMStatusWsConnector()
	if bc.con != nil && bc.bConnected == true {
		return true
	}
	return false
}

//SendPacket 메시지 전송
func SendPacket(msg []byte) {
	GetAMStatusWsConnector().SendMessageByte(msg)
}

//소켓연결 해제
func CloseConnection() {
	GetAMStatusWsConnector().con.Close()
}

//ProcClient R/W 프로세스 시작
func ProcClient() {
	go GetAMStatusWsConnector().ReadMessage()
	go GetAMStatusWsConnector().WriteMessage()
}

func AutoConnect() {
	go GetAMStatusWsConnector().reconnectingServer()
}

func (ty *AMStatusWSConnector) Init() error {
	ty.send = make(chan []byte, maxMessageSize)
	ty.bBreakWrite = make(chan bool)
	ty.bConnected = false
	ty.bLogin = false
	ty.bReconnecting = false
	return nil
}

func (ty *AMStatusWSConnector) clientConnect() bool {
	defer func() {
		if ConfigDataPTR.IsDebugmode == false {
			if err := recover(); err != nil {
				log.Fatalln("Crit Panic", "Error", err)
			}
		}
	}()

	// 연결 정보를 설정하는게 있어야겠네
	/*
		bconf := brestApiConfig.GetConfigData()
		strUrl := ""
		strUrl = fmt.Sprintf("ws://%s:%d/binancesvr", bconf.BinanceSvrAddress, bconf.BinanceSvrPort)

		r, _ := http.NewRequest("GET", strUrl, nil)
		r.Header.Add("Content-Type", "application/json")
		log.Println("Connecting", "server Connecting...", strUrl)

		c, _, err := websocket.DefaultDialer.Dial(strUrl, nil)
		ty.con = c
		if err != nil {
			log.Println("Error", "msg", err.Error())
			return false
		}
	*/

	strUrl := ConfigDataPTR.GetPath()
	c, _, err := websocket.DefaultDialer.Dial(strUrl, nil)
	ty.con = c
	if err != nil {
		log.Println("Error", "msg", err.Error())
		return false
	}

	ty.bLogin = false
	ty.bReconnecting = false
	ty.bConnected = true
	return true
}

func (ty *AMStatusWSConnector) onConnected() {
	log.Println("server Connected ")
	// ty.sendLogin() // 연결하자마자 바로 진입
}

func (ty *AMStatusWSConnector) onUnconnected() {
	log.Println("server UnConnected ")
	if ty.bConnected == true {
		ty.bBreakWrite <- true
		ty.bConnected = false
		if ty.bReconnecting == false {
			go ty.reconnectingServer()
		}
	}
}

func (ty *AMStatusWSConnector) reconnectingServer() {
	ty.bReconnecting = true
	for {
		time.Sleep(time.Duration(time.Second * 4))
		log.Println("Try Reconnect ")
		if ty.bConnected == true {
			return
		}

		if ty.clientConnect() == true {
			ProcClient()
			return
		}
	}
}

func (ty *AMStatusWSConnector) ReadMessage() {
	defer func() {
		ty.con.Close()
		ty.onUnconnected()

		if ConfigDataPTR.IsDebugmode == false {
			if err := recover(); err != nil {
				log.Fatalln("Crit Panic", "Error", err)
			}
		}
	}()

	ty.onConnected()
	ty.con.SetReadLimit(maxMessageSize)
	for {
		if ty.con != nil {
			_, message, err := ty.con.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					if ConfigDataPTR.IsDebugmode == true {
						log.Println("client close", "Msg =", err)
					}
				}
				break
			}

			if len(message) > maxMessageSize {
				log.Println("Error", "Packet size over ", len(message))
				ty.con.Close()
			} else {
				ty.messagePasering(message)
			}
		}
	}
}

func (ty *AMStatusWSConnector) WriteMessage() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		if ty != nil {
			if ty.con != nil {
				ty.con.Close()
				ty.onUnconnected()
			}
		}
		if ConfigDataPTR.IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "WriteMessage ", err)
			}
		}
	}()
	for {
		select {
		case <-ty.bBreakWrite:
			return
		case message, ok := <-ty.send:
			ty.con.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				ty.con.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			ty.con.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			ty.con.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ty.con.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (ty *AMStatusWSConnector) SendMessageByte(msg []byte) {

	defer func() {
		if ConfigDataPTR.IsDebugmode == false {
			if err := recover(); err != nil {
				log.Fatalln("Crit Panic", "Error", err)
				if ty != nil {
					if ty.con != nil {
						ty.con.Close()
					}
				}
			}
		}
	}()

	if ty == nil {
		return
	}
	if ty.con == nil {
		return
	}
	ty.send <- msg
}

func (ty *AMStatusWSConnector) messagePasering(msg []byte) {
	defer func() {
		if ConfigDataPTR.IsDebugmode == false {
			if err := recover(); err != nil {
				log.Fatalln("Crit Panic", "Error", err)
				if ty != nil {
					if ty.con != nil {
						ty.con.Close()
					}
				}
			}
		}
	}()

	var packetdata map[string]interface{}
	json.Unmarshal(msg, &packetdata)
	//fmt.Println(string(msg))
	strCmd := packetdata["Command"]

	for k, v := range packetdata {
		fmt.Println("packetdata : ", k, v)
	}

	switch strCmd {
	case "login": //현재 호가정보 전송
		// ty.onPacketLogin(msg)
	}

	/*
		var packetdata map[string]interface{}
		json.Unmarshal(msg, &packetdata)

		//fmt.Println(string(msg))
		strCmd := packetdata["Command"]
		switch strCmd {
		case "login": //현재 호가정보 전송
			ty.onPacketLogin(msg)
		}
	*/

}

// 상태값 가져오는거
/*
	// CPU, Memory, HDD,  Network
	// 명령어 줘서 살리는거 - 테스트 : 어떤 명령을 어떻게 할것
	// 소켓 보내는 것!!



*/
