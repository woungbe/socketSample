package amwebsocket

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

//MfSvrLogWSSession 로그 클라이언트 섹션
type AMWSession struct {
	Index int64
	Conn  *websocket.Conn
	//로그인 되어있나
	BLogin    bool
	StrUserID string //사용자ID

	send        chan []byte
	bBreakWrite chan bool
	bConnected  bool
}

//NewSeesion 새로운섹션 생성
func NewAdminSeesionLog() *AMWSession {

	newObj := new(AMWSession)
	newObj.initObj()
	return newObj
}
func (ty *AMWSession) initObj() {

	ty.send = make(chan []byte, maxMessageSize)
	ty.bBreakWrite = make(chan bool)
	ty.bConnected = false
}

//Clear 클리어
func (ty *AMWSession) Clear() {
	ty.Conn = nil
	ty.BLogin = false
	ty.bConnected = false
	ty.StrUserID = ""
}

//ReadMessage 소켓에서 메시지 읽기
func (ty *AMWSession) ReadMessage() {
	defer func() {
		if ty != nil {
			if ty.Conn != nil {
				ty.Conn.Close()
				ty.onUnconnected()
			}
		}

		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}
	}()

	ty.Conn.SetReadLimit(maxMessageSize)
	ty.Conn.SetReadDeadline(time.Now().Add(pongWait))
	ty.Conn.SetPongHandler(func(string) error { ty.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	ty.onConnected()
	for {
		_, message, err := ty.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("AMWSession Read Error ", "Error Msg =", err)
			}
			break
		}

		if len(message) > maxMessageSize {
			log.Println("Error", "Packet size over ", len(message))
			ty.Conn.Close()

		} else {
			ty.messagePasering(message)
		}
	}
}

func (ty *AMWSession) WriteMessage() {
	ticker := time.NewTicker(pingPeriod)
	aliveChk := time.NewTicker(time.Second * liveTime)
	defer func() {
		//log.Info("쓰기정지")
		ticker.Stop()
		if ty != nil {
		}
		//	if futuMFconfig.IsDebugmode() == false {
		if err := recover(); err != nil {
			log.Fatalln("Crit Panic", "Error", err)
		}
		//	}
	}()
	for {
		select {
		case <-ty.bBreakWrite:
			return
		case message, ok := <-ty.send:
			ty.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				ty.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			ty.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			ty.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ty.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-aliveChk.C:
			if ty.BLogin == false {
				ty.Conn.Close()
			}
		}
	}
}

func (ty *AMWSession) onConnected() {
	ty.BLogin = false
	ty.bConnected = true
	//ty.BLogin = true
}
func (ty *AMWSession) onUnconnected() {

	if IsDebugmode == false {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}()
	}
	if ty.bConnected == false {
		return
	}
	ty.bConnected = false
	ty.bBreakWrite <- true

	ty.CloseJob()
	GetAMWSvr().ChanLeave <- ty

}

//CloseJob 유저 연결종료시 처리
func (ty *AMWSession) CloseJob() {
	if IsDebugmode == false {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}()
	}
	if ty.BLogin == true {
		GetAMWSvr().removeLoginUser(ty.StrUserID, ty)
	}
	ty.BLogin = false
}

//messagePasering 메시지 파싱
func (ty *AMWSession) messagePasering(msg []byte) {
	if IsDebugmode == false {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
				if ty != nil {
					if ty.Conn != nil {
						ty.Conn.Close()
					}
				}
			}
		}()
	}
	//fmt.Printf(string(msg))
	var packetdata map[string]interface{}
	json.Unmarshal(msg, &packetdata)

	strCmd := packetdata["Command"]
	//strLogType := packetdata["LogType"]
	switch strCmd {

	case "Login":

		var lgmsg WSPackageReq
		var strmsg ResultWebSocket
		strmsg.Command = "LoginResult"
		strmsg.Types = "Info"
		er := json.Unmarshal(msg, &lgmsg)
		if er == nil {
			// b. resultMsg := ty.
			// 로그인을 하세요 !!
			fmt.Println(lgmsg)
		}
		ty.BLogin = true
		ty.StrUserID = lgmsg.LoginID
		GetAMWSvr().addLoginUser(lgmsg.LoginID, ty)

		resultJson, _ := json.Marshal(strmsg)
		ty.SendMessageByte(resultJson)

	// 파서 정리중
	/*
		case "Login":
			var lgMsg dfinglobaldefine.WSPacketAdmin
			var srtMsg dfinglobaldefine.WSPacketLog
			srtMsg.Command = "LoginResult"
			srtMsg.LogType = "Info"

			er := json.Unmarshal(msg, &lgMsg)
			if er == nil {
				b, resultMsg := ty.onLogin(lgMsg.LoginID, lgMsg.OTPValue)
				if b {
					ty.StrUserID = lgMsg.LoginID
					GetAMWSvr().addLoginUser(lgMsg.LoginID, ty)
					ty.BLogin = true
					log.Println("Info", "Viewer 로그인 = ", lgMsg.LoginID)
					srtMsg.DATAS.Msg = "로그인 성공"
				} else {
					srtMsg.LogType = "Warn"
					srtMsg.DATAS.Msg = "로그인 실패 =" + resultMsg
				}
			}
			resultJson, _ := json.Marshal(srtMsg)
			ty.SendMessageByte(resultJson)
		case "LogMsg":
	*/
	default:
		ty.Conn.Close()
		log.Println("비정상패킷")
	}

}

//SendMessage 메시지전송
func (ty *AMWSession) SendMessage(msg string) {
	if IsDebugmode == false {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
				if ty != nil {
					if ty.Conn != nil {
						ty.Conn.Close()
					}
				}
			}
		}()
	}
	ty.SendMessageByte([]byte(msg))
}

//SendMessageByte 메시지전송
func (ty *AMWSession) SendMessageByte(msg []byte) {
	if IsDebugmode == false {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
				if ty != nil {
					if ty.Conn != nil {
						ty.Conn.Close()
					}
				}
			}
		}()
	}
	if ty == nil {
		return
	}
	if ty.Conn == nil {
		return
	}
	ty.send <- msg
}
